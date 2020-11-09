package cluster

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScheduleOnceParallel(t *testing.T) {
	t.Parallel()

	makeKey := model.NewId

	// there is only one callback by design, so all tests need to add their key
	// and callback handling code here.
	jobKey1 := makeKey()
	count1 := new(int32)
	jobKey2 := makeKey()
	count2 := new(int32)
	jobKey3 := makeKey()
	jobKey4 := makeKey()
	count4 := new(int32)

	manyJobs := make(map[string]*int32)
	for i := 0; i < 100; i++ {
		manyJobs[makeKey()] = new(int32)
	}

	callback := func(key string) error {
		switch key {
		case jobKey1:
			atomic.AddInt32(count1, 1)
			return nil
		case jobKey2:
			atomic.AddInt32(count2, 1)
			return nil
		case jobKey3:
			return errors.New("failed at the plugin")
		case jobKey4:
			atomic.AddInt32(count4, 1)
			return nil
		default:
			count, ok := manyJobs[key]
			if ok {
				atomic.AddInt32(count, 1)
				return nil
			}
		}

		return errors.New("error")
	}

	ScheduleOnceRegisterCallback(callback)
	mockPluginAPI := newMockPluginAPI(t)
	err := ScheduleOnceStartScheduler(mockPluginAPI)
	require.NoError(t, err)
	jobs, err := ListScheduledJobs(mockPluginAPI)
	require.NoError(t, err)
	require.Empty(t, jobs)

	t.Run("one scheduled job", func(t *testing.T) {
		t.Parallel()

		job, err := ScheduleOnce(mockPluginAPI, jobKey1, time.Now().Add(100*time.Millisecond))
		require.NoError(t, err)
		require.NotNil(t, job)
		assert.NotNil(t, mockPluginAPI.keyValues[oncePrefix+jobKey1])

		time.Sleep(150 * time.Millisecond)

		assert.Empty(t, mockPluginAPI.keyValues[oncePrefix+jobKey1])
		assert.Empty(t, currentlyScheduled.keys[jobKey1])

		// It's okay to close jobs extra times, even if they're completed.
		job.Close()
		job.Close()
		job.Close()
		job.Close()

		// Should have been called once
		assert.Equal(t, int32(1), *count1)
	})

	t.Run("one job, stopped before firing", func(t *testing.T) {
		t.Parallel()

		job, err := ScheduleOnce(mockPluginAPI, jobKey2, time.Now().Add(100*time.Millisecond))
		require.NoError(t, err)
		require.NotNil(t, job)
		assert.NotNil(t, mockPluginAPI.keyValues[oncePrefix+jobKey2])

		job.Close()
		assert.Empty(t, mockPluginAPI.keyValues[oncePrefix+jobKey2])
		assert.Empty(t, currentlyScheduled.keys[jobKey2])

		time.Sleep(2 * waitAfterFail)

		// Should not have been called
		assert.Equal(t, int32(0), *count2)

		// It's okay to close jobs extra times, even if they're completed.
		job.Close()
		job.Close()
		job.Close()
		job.Close()
	})

	t.Run("failed at the plugin, job removed from db", func(t *testing.T) {
		t.Parallel()

		job, err := ScheduleOnce(mockPluginAPI, jobKey3, time.Now().Add(100*time.Millisecond))
		require.NoError(t, err)
		require.NotNil(t, job)
		assert.NotNil(t, mockPluginAPI.keyValues[oncePrefix+jobKey3])

		time.Sleep(200 * time.Millisecond)
		assert.Empty(t, mockPluginAPI.keyValues[oncePrefix+jobKey3])
		assert.Empty(t, currentlyScheduled.keys[jobKey3])
	})

	t.Run("close and restart a job with the same key", func(t *testing.T) {
		t.Parallel()

		job, err := ScheduleOnce(mockPluginAPI, jobKey4, time.Now().Add(100*time.Millisecond))
		require.NoError(t, err)
		require.NotNil(t, job)
		assert.NotNil(t, mockPluginAPI.keyValues[oncePrefix+jobKey4])

		job.Close()
		assert.Empty(t, mockPluginAPI.keyValues[oncePrefix+jobKey4])
		assert.Empty(t, currentlyScheduled.keys[jobKey4])

		job, err = ScheduleOnce(mockPluginAPI, jobKey4, time.Now().Add(100*time.Millisecond))
		require.NoError(t, err)
		require.NotNil(t, job)
		assert.NotNil(t, mockPluginAPI.keyValues[oncePrefix+jobKey4])

		time.Sleep(110 * time.Millisecond)
		assert.Equal(t, int32(1), *count4)
		assert.Empty(t, mockPluginAPI.keyValues[oncePrefix+jobKey4])
		assert.Empty(t, currentlyScheduled.keys[jobKey4])
	})

	t.Run("many scheduled jobs", func(t *testing.T) {
		t.Parallel()

		for k := range manyJobs {
			job, err := ScheduleOnce(mockPluginAPI, k, time.Now().Add(100*time.Millisecond))
			require.NoError(t, err)
			require.NotNil(t, job)
			assert.NotNil(t, mockPluginAPI.keyValues[oncePrefix+k])
		}

		time.Sleep(150 * time.Millisecond)

		for k, v := range manyJobs {
			assert.Empty(t, mockPluginAPI.keyValues[oncePrefix+k])
			assert.Empty(t, currentlyScheduled.keys[k])
			assert.Equal(t, int32(1), *v)
		}
	})

	//t.Run("multi-threaded, single job", func(t *testing.T) {
	//	t.Parallel()
	//
	//	mockPluginAPI := newMockPluginAPI(t)
	//
	//	count := new(int32)
	//	callback := func() {
	//		atomic.AddInt32(count, 1)
	//	}
	//
	//	var jobs []*Job
	//
	//	key := makeKey()
	//
	//	for i := 0; i < 3; i++ {
	//		job, err := Schedule(mockPluginAPI, key, MakeWaitForInterval(100*time.Millisecond), callback)
	//		require.NoError(t, err)
	//		require.NotNil(t, job)
	//
	//		jobs = append(jobs, job)
	//	}
	//
	//	time.Sleep(1 * time.Second)
	//
	//	var wg sync.WaitGroup
	//	for i := 0; i < 3; i++ {
	//		job := jobs[i]
	//		wg.Add(1)
	//		go func() {
	//			defer wg.Done()
	//			err := job.Close()
	//			require.NoError(t, err)
	//		}()
	//	}
	//	wg.Wait()
	//
	//	time.Sleep(1 * time.Second)
	//
	//	// Shouldn't have hit 20 in this time frame
	//	assert.Less(t, *count, int32(20))
	//
	//	// Should have hit at least 5 in this time frame
	//	assert.Greater(t, *count, int32(5))
	//})
	//
	//t.Run("multi-threaded, multiple jobs", func(t *testing.T) {
	//	t.Parallel()
	//
	//	mockPluginAPI := newMockPluginAPI(t)
	//
	//	countA := new(int32)
	//	callbackA := func() {
	//		atomic.AddInt32(countA, 1)
	//	}
	//
	//	countB := new(int32)
	//	callbackB := func() {
	//		atomic.AddInt32(countB, 1)
	//	}
	//
	//	keyA := makeKey()
	//	keyB := makeKey()
	//
	//	var jobs []*Job
	//	for i := 0; i < 3; i++ {
	//		var key string
	//		var callback func()
	//		if i <= 1 {
	//			key = keyA
	//			callback = callbackA
	//		} else {
	//			key = keyB
	//			callback = callbackB
	//		}
	//
	//		job, err := Schedule(mockPluginAPI, key, MakeWaitForInterval(100*time.Millisecond), callback)
	//		require.NoError(t, err)
	//		require.NotNil(t, job)
	//
	//		jobs = append(jobs, job)
	//	}
	//
	//	time.Sleep(1 * time.Second)
	//
	//	var wg sync.WaitGroup
	//	for i := 0; i < 3; i++ {
	//		job := jobs[i]
	//		wg.Add(1)
	//		go func() {
	//			defer wg.Done()
	//			err := job.Close()
	//			require.NoError(t, err)
	//		}()
	//	}
	//	wg.Wait()
	//
	//	time.Sleep(1 * time.Second)
	//
	//	// Shouldn't have hit 20 in this time frame
	//	assert.Less(t, *countA, int32(20))
	//
	//	// Should have hit at least 5 in this time frame
	//	assert.Greater(t, *countA, int32(5))
	//
	//	// Shouldn't have hit 20 in this time frame
	//	assert.Less(t, *countB, int32(20))
	//
	//	// Should have hit at least 5 in this time frame
	//	assert.Greater(t, *countB, int32(5))
	//})
}

func TestScheduleOnceSequential(t *testing.T) {
	makeKey := model.NewId

	t.Run("failed at the db", func(t *testing.T) {
		// there is only one callback by design, so all tests need to add their key
		// and callback handling code here.
		jobKey1 := makeKey()
		count1 := new(int32)

		callback := func(key string) error {
			switch key {
			case jobKey1:
				atomic.AddInt32(count1, 1)
				return nil
			}
			return errors.New("error")
		}

		ScheduleOnceRegisterCallback(callback)
		mockPluginAPI := newMockPluginAPI(t)
		err := ScheduleOnceStartScheduler(mockPluginAPI)
		require.NoError(t, err)
		jobs, err := ListScheduledJobs(mockPluginAPI)
		require.NoError(t, err)
		require.Empty(t, jobs)

		job, err := ScheduleOnce(mockPluginAPI, jobKey1, time.Now().Add(100*time.Millisecond))
		require.NoError(t, err)
		require.NotNil(t, job)
		assert.NotNil(t, mockPluginAPI.keyValues[oncePrefix+jobKey1])
		assert.NotNil(t, currentlyScheduled.keys[jobKey1])
		mockPluginAPI.failingWithPrefix = oncePrefix

		// wait until the metadata has failed to read
		time.Sleep((maxNumFails + 1) * waitAfterFail)
		assert.Equal(t, int32(0), *count1)
		assert.Nil(t, mockPluginAPI.keyValues[oncePrefix+jobKey1])

		assert.Empty(t, currentlyScheduled.keys[jobKey1])
		assert.Empty(t, mockPluginAPI.keyValues[oncePrefix+jobKey1])
		assert.Equal(t, int32(0), *count1)
	})

	t.Run("simulate starting the plugin with 3 pending jobs in the db", func(t *testing.T) {
		jobKeys := make(map[string]*int32)
		for i := 0; i < 3; i++ {
			jobKeys[makeKey()] = new(int32)
		}

		callback := func(key string) error {
			count, ok := jobKeys[key]
			if ok {
				atomic.AddInt32(count, 1)
				return nil
			} else {
				return errors.New("error")
			}
		}

		mockPluginAPI := newMockPluginAPI(t)

		for k := range jobKeys {
			metadata := jobOnceMetadata{
				Key:   k,
				RunAt: time.Now().Add(100*time.Millisecond + addJitter()),
			}
			job, err := NewJobOnce(mockPluginAPI, metadata)
			require.NoError(t, err)
			err = job.saveMetadata()
			require.NoError(t, err)
			assert.NotEmpty(t, mockPluginAPI.keyValues[oncePrefix+k])
		}

		jobs, err := ListScheduledJobs(mockPluginAPI)
		require.NoError(t, err)
		require.Len(t, jobs, 3)

		// The jobs are in the db, start the plugin and make sure it runs those jobs.
		ScheduleOnceRegisterCallback(callback)
		err = ScheduleOnceStartScheduler(mockPluginAPI)
		require.NoError(t, err)

		time.Sleep(200 * time.Millisecond)

		for k, v := range jobKeys {
			assert.Empty(t, mockPluginAPI.keyValues[oncePrefix+k])
			assert.Empty(t, currentlyScheduled.keys[k])
			assert.Equal(t, int32(1), *v)
		}
		jobs, err = ListScheduledJobs(mockPluginAPI)
		require.NoError(t, err)
		require.Empty(t, jobs)
	})
}
