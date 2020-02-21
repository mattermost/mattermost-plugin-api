package cluster

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func lock(t *testing.T, m *Mutex) {
	t.Helper()

	done := make(chan bool)
	go func() {
		defer close(done)
		m.Lock()
	}()

	select {
	case <-time.After(1 * time.Second):
		require.Fail(t, "failed to lock mutex within 1 second")
	case <-done:
	}
}

func unlock(t *testing.T, m *Mutex, panics bool) {
	t.Helper()

	done := make(chan bool)
	go func() {
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
	t.Run("successful lock/unlock cycle", func(t *testing.T) {
		mockPluginAPI := newMockPluginAPI(t)

		m := NewMutex(mockPluginAPI, "key")
		lock(t, m)
		unlock(t, m, false)
		lock(t, m)
		unlock(t, m, false)
	})

	t.Run("unlock when not locked", func(t *testing.T) {
		mockPluginAPI := newMockPluginAPI(t)

		m := NewMutex(mockPluginAPI, "key")
		unlock(t, m, true)
	})

	t.Run("blocking lock", func(t *testing.T) {
		mockPluginAPI := newMockPluginAPI(t)

		m := NewMutex(mockPluginAPI, "key")
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
		mockPluginAPI := newMockPluginAPI(t)

		m := NewMutex(mockPluginAPI, "key")

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

	t.Run("failed unlock", func(t *testing.T) {
		mockPluginAPI := newMockPluginAPI(t)

		m := NewMutex(mockPluginAPI, "key")
		lock(t, m)

		mockPluginAPI.setFailing(true)

		unlock(t, m, false)

		// Simulate expiry
		mockPluginAPI.clear()
		mockPluginAPI.setFailing(false)

		lock(t, m)
	})

	t.Run("discrete keys", func(t *testing.T) {
		mockPluginAPI := newMockPluginAPI(t)

		m1 := NewMutex(mockPluginAPI, "key1")
		lock(t, m1)

		m2 := NewMutex(mockPluginAPI, "key2")
		lock(t, m2)

		m3 := NewMutex(mockPluginAPI, "key3")
		lock(t, m3)

		unlock(t, m1, false)
		unlock(t, m3, false)

		lock(t, m1)

		unlock(t, m2, false)
		unlock(t, m1, false)
	})
}
