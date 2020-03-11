package cluster

import (
	"testing"
	"time"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustMakeLockKey(key string) string {
	key, err := makeLockKey(key)
	if err != nil {
		panic(err)
	}

	return key
}

func mustNewMutex(pluginAPI MutexPluginAPI, key string) *Mutex {
	m, err := NewMutex(pluginAPI, key)
	if err != nil {
		panic(err)
	}

	return m
}

func TestMakeLockKey(t *testing.T) {
	t.Run("fails when empty", func(t *testing.T) {
		key, err := makeLockKey("")
		assert.Error(t, err)
		assert.Empty(t, key)
	})

	t.Run("not-empty", func(t *testing.T) {
		testCases := map[string]string{
			"key":   mutexPrefix + "key",
			"other": mutexPrefix + "other",
		}

		for key, expected := range testCases {
			actual, err := makeLockKey(key)
			require.NoError(t, err)
			assert.Equal(t, expected, actual)
		}
	})
}

func TestLockValue(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		actual, err := getLockValue([]byte{})
		require.NoError(t, err)
		require.True(t, actual.IsZero())
	})

	t.Run("invalid", func(t *testing.T) {
		actual, err := getLockValue([]byte("abc"))
		require.Error(t, err)
		require.True(t, actual.IsZero())
	})

	t.Run("successful", func(t *testing.T) {
		testCases := []time.Time{
			time.Now().Add(-15 * time.Second),
			time.Now(),
			time.Now().Add(15 * time.Second),
		}

		for _, testCase := range testCases {
			t.Run(testCase.Format("Mon Jan 2 15:04:05 -0700 MST 2006"), func(t *testing.T) {
				actual, err := getLockValue(makeLockValue(testCase))
				require.NoError(t, err)
				require.Equal(t, testCase.Truncate(0), actual.Truncate(0))
			})
		}
	})
}

func lock(t *testing.T, m *Mutex) {
	t.Helper()

	done := make(chan bool)
	go func() {
		t.Helper()

		defer close(done)
		m.Lock()
	}()

	select {
	case <-time.After(2 * time.Second):
		require.Fail(t, "failed to lock mutex within 1 second")
	case <-done:
	}
}

func unlock(t *testing.T, m *Mutex, panics bool) {
	t.Helper()

	done := make(chan bool)
	go func() {
		t.Helper()

		defer close(done)
		if panics {
			assert.Panics(t, m.Unlock)
		} else {
			assert.NotPanics(t, m.Unlock)
		}
	}()

	select {
	case <-time.After(1 * time.Second):
		require.Fail(t, "failed to unlock mutex within 1 second")
	case <-done:
	}
}

func TestMutex(t *testing.T) {
	t.Parallel()

	makeKey := func() string {
		return model.NewId()
	}

	t.Run("successful lock/unlock cycle", func(t *testing.T) {
		t.Parallel()

		mockPluginAPI := newMockPluginAPI(t)

		m := mustNewMutex(mockPluginAPI, makeKey())
		lock(t, m)
		unlock(t, m, false)
		lock(t, m)
		unlock(t, m, false)
	})

	t.Run("unlock when not locked", func(t *testing.T) {
		t.Parallel()

		mockPluginAPI := newMockPluginAPI(t)

		m := mustNewMutex(mockPluginAPI, makeKey())
		unlock(t, m, true)
	})

	t.Run("blocking lock", func(t *testing.T) {
		t.Parallel()

		mockPluginAPI := newMockPluginAPI(t)

		m := mustNewMutex(mockPluginAPI, makeKey())
		lock(t, m)

		done := make(chan bool)
		go func() {
			defer close(done)
			m.Lock()
		}()

		select {
		case <-time.After(1 * time.Second):
		case <-done:
			require.Fail(t, "second goroutine should not have locked")
		}

		unlock(t, m, false)

		select {
		case <-time.After(pollWaitInterval * 2):
			require.Fail(t, "second goroutine should have locked")
		case <-done:
		}
	})

	t.Run("failed lock", func(t *testing.T) {
		t.Parallel()

		mockPluginAPI := newMockPluginAPI(t)

		m := mustNewMutex(mockPluginAPI, makeKey())

		mockPluginAPI.setFailing(true)

		done := make(chan bool)
		go func() {
			defer close(done)
			m.Lock()
		}()

		select {
		case <-time.After(5 * time.Second):
		case <-done:
			require.Fail(t, "goroutine should not have locked")
		}

		mockPluginAPI.setFailing(false)

		select {
		case <-time.After(15 * time.Second):
			require.Fail(t, "goroutine should have locked")
		case <-done:
		}
	})

	t.Run("failed unlock, key deleted", func(t *testing.T) {
		t.Parallel()

		mockPluginAPI := newMockPluginAPI(t)

		key := makeKey()
		m := mustNewMutex(mockPluginAPI, key)
		lock(t, m)

		mockPluginAPI.setFailing(true)

		unlock(t, m, false)

		// Simulate expiry by deleting key
		mockPluginAPI.setFailing(false)
		mockPluginAPI.KVDelete(mustMakeLockKey(key))

		lock(t, m)
	})

	t.Run("failed unlock, key expired", func(t *testing.T) {
		t.Parallel()

		mockPluginAPI := newMockPluginAPI(t)

		key := makeKey()
		m := mustNewMutex(mockPluginAPI, key)
		lock(t, m)

		mockPluginAPI.setFailing(true)

		unlock(t, m, false)

		// Simulate expiry by writing expired value
		mockPluginAPI.setFailing(false)
		mockPluginAPI.KVSet(mustMakeLockKey(key), makeLockValue(time.Now().Add(-1*time.Second)))

		lock(t, m)
	})

	t.Run("discrete keys", func(t *testing.T) {
		t.Parallel()

		mockPluginAPI := newMockPluginAPI(t)

		m1 := mustNewMutex(mockPluginAPI, makeKey())
		lock(t, m1)

		m2 := mustNewMutex(mockPluginAPI, makeKey())
		lock(t, m2)

		m3 := mustNewMutex(mockPluginAPI, makeKey())
		lock(t, m3)

		unlock(t, m1, false)
		unlock(t, m3, false)

		lock(t, m1)

		unlock(t, m2, false)
		unlock(t, m1, false)
	})

	t.Run("expiring lock", func(t *testing.T) {
		t.Parallel()

		mockPluginAPI := newMockPluginAPI(t)

		key := makeKey()
		m := mustNewMutex(mockPluginAPI, key)

		// Simulate lock expiring in 5 seconds
		now := time.Now()
		appErr := mockPluginAPI.KVSet(mustMakeLockKey(key), makeLockValue(now.Add(5*time.Second)))
		require.Nil(t, appErr)

		done1 := make(chan bool)
		go func() {
			defer close(done1)
			m.Lock()
		}()

		done2 := make(chan bool)
		go func() {
			defer close(done2)
			m.Lock()
		}()

		select {
		case <-time.After(1 * time.Second):
		case <-done1:
			require.Fail(t, "first goroutine should not have locked yet")
		case <-done2:
			require.Fail(t, "second goroutine should not have locked yet")
		}

		select {
		case <-time.After(4*time.Second + pollWaitInterval*2):
			require.Fail(t, "some goroutine should have locked after expiry")
		case <-done1:
			m.Unlock()
			select {
			case <-done2:
			case <-time.After(pollWaitInterval * 2):
				require.Fail(t, "second goroutine should have locked")
			}

		case <-done2:
			m.Unlock()
			select {
			case <-done2:
			case <-time.After(pollWaitInterval * 2):
				require.Fail(t, "first goroutine should have locked")
			}
		}
	})

	t.Run("held lock does not expire", func(t *testing.T) {
		t.Parallel()

		mockPluginAPI := newMockPluginAPI(t)

		m := mustNewMutex(mockPluginAPI, makeKey())

		m.Lock()

		done := make(chan bool)
		go func() {
			defer close(done)
			m.Lock()
		}()

		select {
		case <-time.After(ttl + pollWaitInterval*2):
		case <-done:
			require.Fail(t, "goroutine should not have locked")
		}

		m.Unlock()

		select {
		case <-time.After(pollWaitInterval * 2):
			require.Fail(t, "goroutine should have locked after expiry")
		case <-done:
		}
	})
}
