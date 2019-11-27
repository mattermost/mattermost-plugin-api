// Package dlock is a distributed lock to enable advanced synchronization for Mattermost Plugins.
//
// if you're new to distributed locks and Mattermost Plugins please read sample use case scenarios
// at: https://community.mattermost.com/core/pl/bb376sjsdbym8kj7nz7zrcos7r
package dlock

import (
	"context"

	"sync"
	"time"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
)

const (
	// storePrefix used to prefix lock related keys in KV store.
	storePrefix = "dlock:"
)

const (
	// lockTTL is lock's expiry time.
	lockTTL = time.Second * 15

	// lockRefreshInterval used to determine how long to wait before refreshing
	// a lock's expiry time.
	lockRefreshInterval = time.Second

	// lockTryInterval used to wait before trying to obtain the lock again.
	lockTryInterval = time.Second
)

var (
	// ErrCouldntObtainImmediately returned when a lock couldn't be obtained immediately after
	// calling Lock().
	ErrCouldntObtainImmediately = errors.New("could not obtain immediately")
)

// Store is a data store to keep locks' state.
type Store interface {
	KVSetWithOptions(key string, newValue interface{}, options model.PluginKVSetOptions) (bool, *model.AppError)
}

// DLock is a distributed lock.
type DLock struct {
	// store used to store lock's state to do synchronization.
	store Store

	// key to lock for.
	key string

	// refreshCancel stops refreshing lock's TTL.
	refreshCancel context.CancelFunc

	// refreshWait is a waiter to make sure refreshing is finished.
	refreshWait *sync.WaitGroup
}

// New creates a new distributed lock for key on given store with options.
// think,
//   `dl := New("my-key", store)`
// as an equivalent of,
//   `var m sync.Mutex`
// and use it in the same way.
func New(key string, store Store) *DLock {
	return &DLock{
		key:   buildKey(key),
		store: store,
	}
}

// Lock obtains a new lock.
// ctx provides a context to locking. when ctx is cancelled, Lock() will stop
// blocking and retries and return with error.
// use Lock() exactly like sync.Mutex.Lock(), avoid missuses like deadlocks.
func (d *DLock) Lock(ctx context.Context) error {
	for {
		isLockObtained, err := d.lock()
		if err != nil {
			return err
		}
		if isLockObtained {
			return nil
		}
		afterC := time.After(lockTryInterval)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-afterC:
		}
	}
}

// TryLock tries to obtain the lock immediately.
// err is only filled on system failure.
func (d *DLock) TryLock() (isLockObtained bool, err error) {
	return d.lock()
}

// lock obtains a new lock and starts refreshing the lock until a call to
// Unlock() to not hit lock's TTL.
func (d *DLock) lock() (isLockObtained bool, err error) {
	kopts := model.PluginKVSetOptions{
		EncodeJSON:      true,
		Atomic:          true,
		OldValue:        nil,
		ExpireInSeconds: int64(lockTTL.Seconds()),
	}
	isLockObtained, aerr := d.store.KVSetWithOptions(d.key, true, kopts)
	if aerr != nil {
		return false, errors.Wrap(aerr, "KVSetWithOptions() error, cannot lock")
	}
	if isLockObtained {
		d.startRefreshLoop()
	}
	return isLockObtained, nil
}

// startRefreshLoop refreshes an obtained lock to not get caught by lock's TTL.
// TTL tends to hit and release the lock automatically when plugin terminates.
func (d *DLock) startRefreshLoop() {
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		t := time.NewTicker(lockRefreshInterval)
		kopts := model.PluginKVSetOptions{
			EncodeJSON:      true,
			ExpireInSeconds: int64(lockTTL.Seconds()),
		}
		for {
			select {
			case <-t.C:
				d.store.KVSetWithOptions(d.key, true, kopts)
			case <-ctx.Done():
				return
			}
		}
	}()
	d.refreshCancel = cancel
	d.refreshWait = &wg
}

// Unlock unlocks Lock().
// use Unlock() exactly like sync.Mutex.Unlock().
func (d *DLock) Unlock() error {
	d.refreshCancel()
	d.refreshWait.Wait()
	aerr := d.store.KVDelete(d.key)
	return normalizeAppErr(aerr)
}

// buildKey builds a lock key for KV store.
func buildKey(key string) string {
	return storePrefix + key
}

// normalize error normalizes Plugin API's errors.
// please see this docs to know more about what this normalization do: https://golang.org/doc/faq#nil_error
func normalizeAppErr(err *model.AppError) error {
	if err == nil {
		return nil
	}
	return err
}
