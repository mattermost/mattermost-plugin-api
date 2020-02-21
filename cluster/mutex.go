package cluster

import (
	"strconv"
	"sync"
	"time"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
)

const (
	// mutexPrefix is used to namespace key values created for a mutex from other key values
	// created by a plugin.
	mutexPrefix = "mutex_"
)

const (
	// ttl is the interval after which a locked mutex will expire unless refreshed
	ttl = time.Second * 15

	// refreshInterval is the interval on which the mutex will be refreshed when locked
	refreshInterval = ttl / 2
)

// MutexPluginAPI is the plugin API interface required to manage mutexes.
type MutexPluginAPI interface {
	KVGet(key string) ([]byte, *model.AppError)
	KVSetWithOptions(key string, value []byte, options model.PluginKVSetOptions) (bool, *model.AppError)
	LogError(msg string, keyValuePairs ...interface{})
}

// Mutex is similar to sync.Mutex, except usable by multiple plugin instances across a cluster.
//
// Internally, a mutex relies on an atomic key-value set operation as exposed by the Mattermost
// plugin API. Note that it explicitly does not rely on the built-in support for key value expiry,
// since the implementation for same in the server was partially broken prior to v5.22 and thus
// unreliable for something like a mutex. Instead, we encode the desired expiry as the value of
// the mutex's key value and atomically delete when found to be expired.
//
// Mutexes with different names are unrelated. Mutexes with the same name from different plugins
// are unrelated. Pick a unique name for each mutex your plugin requires.
//
// A Mutex must not be copied after first use.
type Mutex struct {
	pluginAPI MutexPluginAPI
	key       string

	// lock guards the variables used to manage the refresh task, and is not itself related to
	// the cluster-wide lock.
	lock        sync.Mutex
	stopRefresh chan bool
	refreshDone chan bool

	// lockExpiry tracks the expiration time of the lock when last locked. It is not guarded
	// by a local mutex, since it is only read or written when the cluster lock is held.
	lockExpiry time.Time
}

// NewMutex creates a mutex with the given key name.
//
// Panics if key is empty.
func NewMutex(pluginAPI MutexPluginAPI, key string) *Mutex {
	return &Mutex{
		pluginAPI: pluginAPI,
		key:       makeLockKey(key),
	}
}

// makeLockKey returns the prefixed key used to namespace mutex keys.
func makeLockKey(key string) string {
	if len(key) == 0 {
		panic("must specify valid mutex key")
	}

	return mutexPrefix + key
}

// makeLockValue returns the encoded lock value for the given expiry timestamp.
func makeLockValue(expiresAt time.Time) []byte {
	return []byte(strconv.FormatInt(expiresAt.UnixNano(), 10))
}

// getLockValue decodes the given lock value into the expiry timestamp it potentially represents.
func getLockValue(valueBytes []byte) (time.Time, error) {
	if len(valueBytes) == 0 {
		return time.Time{}, nil
	}

	value, err := strconv.ParseInt(string(valueBytes), 10, 64)
	if err != nil {
		return time.Time{}, errors.Wrap(err, "failed to parse mutex kv value")
	}

	return time.Unix(0, value), nil
}

// lock makes a single attempt to atomically lock the mutex, returning true only if successful.
func (m *Mutex) tryLock() (bool, error) {
	now := time.Now()
	newLockExpiry := now.Add(ttl)

	ok, appErr := m.pluginAPI.KVSetWithOptions(m.key, makeLockValue(newLockExpiry), model.PluginKVSetOptions{
		Atomic:   true,
		OldValue: nil, // No existing key value.
	})
	if appErr != nil {
		return false, errors.Wrap(appErr, "failed to set mutex kv")
	}

	if ok {
		m.lockExpiry = newLockExpiry
		return true, nil
	}

	// Check to see if the lock has expired.
	valueBytes, appErr := m.pluginAPI.KVGet(m.key)
	if appErr != nil {
		return false, errors.Wrap(appErr, "failed to get mutex kv")
	}
	actualLockExpiry, err := getLockValue(valueBytes)
	if err != nil {
		return false, err
	}

	// It might have already been deleted.
	if actualLockExpiry.IsZero() {
		return false, nil
	}

	// It might still be valid.
	if actualLockExpiry.After(now) {
		return false, nil
	}

	// Atomically delete the expired lock and try again.
	ok, appErr = m.pluginAPI.KVSetWithOptions(m.key, nil, model.PluginKVSetOptions{
		Atomic:   true,
		OldValue: valueBytes,
	})
	if err != nil {
		return false, errors.Wrap(err, "failed to delete mutex kv")
	}

	return false, nil
}

// refreshLock rewrites the lock key value with a new expiry, returning true only if successful.
func (m *Mutex) refreshLock() error {
	now := time.Now()

	newLockExpiry := now.Add(ttl)

	ok, err := m.pluginAPI.KVSetWithOptions(m.key, makeLockValue(newLockExpiry), model.PluginKVSetOptions{
		Atomic:   true,
		OldValue: makeLockValue(m.lockExpiry),
	})
	if err != nil {
		return errors.Wrap(err, "failed to refresh mutex kv")
	} else if !ok {
		return errors.New("unexpectedly failed to refresh mutex kv")
	}

	m.lockExpiry = newLockExpiry

	return nil
}

// Lock locks m. If the mutex is already locked by any plugin instance, including the current one,
// the calling goroutine blocks until the mutex can be locked.
func (m *Mutex) Lock() {
	var waitInterval time.Duration

	for {
		time.Sleep(waitInterval)

		locked, err := m.tryLock()
		if err != nil {
			m.pluginAPI.LogError("failed to lock mutex", "err", err, "lock_key", m.key)
			waitInterval = nextWaitInterval(waitInterval, err)
			continue
		} else if !locked {
			waitInterval = nextWaitInterval(waitInterval, err)
			continue
		}

		stop := make(chan bool)
		done := make(chan bool)
		go func() {
			defer close(done)
			t := time.NewTicker(refreshInterval)
			for {
				select {
				case <-t.C:
					err := m.refreshLock()
					if err != nil {
						m.pluginAPI.LogError("failed to refresh mutex", "err", err, "lock_key", m.key)
						return
					}
				case <-stop:
					return
				}
			}
		}()

		m.lock.Lock()
		m.stopRefresh = stop
		m.refreshDone = done
		m.lock.Unlock()

		return
	}
}

// Unlock unlocks m. It is a run-time error if m is not locked on entry to Unlock.
//
// Just like sync.Mutex, a locked Lock is not associated with a particular goroutine or plugin
// instance. It is allowed for one goroutine or plugin instance to lock a Lock and then arrange
// for another goroutine or plugin instance to unlock it. In practice, ownership of the lock should
// remain within a single plugin instance.
func (m *Mutex) Unlock() {
	m.lock.Lock()
	if m.stopRefresh == nil {
		m.lock.Unlock()
		panic("mutex has not been acquired")
	}

	close(m.stopRefresh)
	m.stopRefresh = nil
	<-m.refreshDone
	m.lock.Unlock()

	// If an error occurs deleting, the mutex kv will still expire, allowing later retry.
	_, _ = m.pluginAPI.KVSetWithOptions(m.key, nil, model.PluginKVSetOptions{})
}
