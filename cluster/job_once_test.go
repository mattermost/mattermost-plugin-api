package cluster

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScheduleOnceParallel(t *testing.T) {
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
	jobKey5 := makeKey()
	count5 := new(int32)

	manyJobs := make(map[string]*int32)
	for i := 0; i < 100; i++ {
		manyJobs[makeKey()] = new(int32)
	}

	callback := func(key string) {
		switch key {
		case jobKey1:
			atomic.AddInt32(count1, 1)
		case jobKey2:
			atomic.AddInt32(count2, 1)
		case jobKey3:
			return // do nothing, like an error occurred in the plugin
		case jobKey4:
			atomic.AddInt32(count4, 1)
		case jobKey5:
			atomic.AddInt32(count5, 1)
		default:
			count, ok := manyJobs[key]
			if ok {
				atomic.AddInt32(count, 1)
				return
			}
		}
	}

	mockPluginAPI := newMockPluginAPI(t)
	getVal := func(key string) []byte {
		data, _ := mockPluginAPI.KVGet(key)
		return data
	}

	scheduler, err := StartJobOnceScheduler(mockPluginAPI, callback)
	require.NoError(t, err)
	jobs, err := scheduler.ListScheduledJobs()
	require.NoError(t, err)
	require.Empty(t, jobs)

	t.Run("one scheduled job", func(t *testing.T) {
		t.Parallel()

		job, err := scheduler.ScheduleOnce(jobKey1, time.Now().Add(100*time.Millisecond))
		require.NoError(t, err)
		require.NotNil(t, job)
		assert.NotEmpty(t, getVal(oncePrefix+jobKey1))

		time.Sleep(200*time.Millisecond + scheduleOnceJitter)

		assert.Empty(t, getVal(oncePrefix+jobKey1))
		activeJobs.mu.RLock()
		assert.Empty(t, activeJobs.jobs[jobKey1])
		activeJobs.mu.RUnlock()

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

		job, err := scheduler.ScheduleOnce(jobKey2, time.Now().Add(100*time.Millisecond))
		require.NoError(t, err)
		require.NotNil(t, job)
		assert.NotEmpty(t, getVal(oncePrefix+jobKey2))

		job.Close()
		assert.Empty(t, getVal(oncePrefix+jobKey2))
		activeJobs.mu.RLock()
		assert.Empty(t, activeJobs.jobs[jobKey2])
		activeJobs.mu.RUnlock()

		time.Sleep(2 * (waitAfterFail + scheduleOnceJitter))

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

		job, err := scheduler.ScheduleOnce(jobKey3, time.Now().Add(100*time.Millisecond))
		require.NoError(t, err)
		require.NotNil(t, job)
		assert.NotEmpty(t, getVal(oncePrefix+jobKey3))

		time.Sleep(200*time.Millisecond + scheduleOnceJitter)
		assert.Empty(t, getVal(oncePrefix+jobKey3))
		activeJobs.mu.RLock()
		assert.Empty(t, activeJobs.jobs[jobKey3])
		activeJobs.mu.RUnlock()
	})

	t.Run("close and restart a job with the same key", func(t *testing.T) {
		t.Parallel()

		job, err := scheduler.ScheduleOnce(jobKey4, time.Now().Add(100*time.Millisecond))
		require.NoError(t, err)
		require.NotNil(t, job)
		assert.NotEmpty(t, getVal(oncePrefix+jobKey4))

		job.Close()
		assert.Empty(t, getVal(oncePrefix+jobKey4))
		activeJobs.mu.RLock()
		assert.Empty(t, activeJobs.jobs[jobKey4])
		activeJobs.mu.RUnlock()

		job, err = scheduler.ScheduleOnce(jobKey4, time.Now().Add(100*time.Millisecond))
		require.NoError(t, err)
		require.NotNil(t, job)
		assert.NotEmpty(t, getVal(oncePrefix+jobKey4))

		time.Sleep(200*time.Millisecond + scheduleOnceJitter)
		assert.Equal(t, int32(1), atomic.LoadInt32(count4))
		assert.Empty(t, getVal(oncePrefix+jobKey4))
		activeJobs.mu.RLock()
		assert.Empty(t, activeJobs.jobs[jobKey4])
		activeJobs.mu.RUnlock()
	})

	t.Run("many scheduled jobs", func(t *testing.T) {
		t.Parallel()

		for k := range manyJobs {
			job, err := scheduler.ScheduleOnce(k, time.Now().Add(100*time.Millisecond))
			require.NoError(t, err)
			require.NotNil(t, job)
			assert.NotEmpty(t, getVal(oncePrefix+k))
		}

		time.Sleep(200*time.Millisecond + scheduleOnceJitter)

		for k, v := range manyJobs {
			assert.Empty(t, getVal(oncePrefix+k))
			activeJobs.mu.RLock()
			assert.Empty(t, activeJobs.jobs[k])
			activeJobs.mu.RUnlock()
			assert.Equal(t, int32(1), *v)
		}
	})

	t.Run("close a job by key name", func(t *testing.T) {
		t.Parallel()

		job, err := scheduler.ScheduleOnce(jobKey5, time.Now().Add(100*time.Millisecond))
		require.NoError(t, err)
		require.NotNil(t, job)
		assert.NotEmpty(t, getVal(oncePrefix+jobKey5))
		activeJobs.mu.RLock()
		assert.NotEmpty(t, activeJobs.jobs[jobKey5])
		activeJobs.mu.RUnlock()

		jobs, err := scheduler.ListScheduledJobs()
		require.NoError(t, err)
		// simulate finding it in the list for whatever reason
		for _, jobs := range jobs {
			if jobs.Key == jobKey5 {
				scheduler.Close(jobKey5)
				break
			}
		}

		assert.Empty(t, getVal(oncePrefix+jobKey5))
		activeJobs.mu.RLock()
		assert.Empty(t, activeJobs.jobs[jobKey5])
		activeJobs.mu.RUnlock()

		// close it again doesn't do anything:
		scheduler.Close(jobKey5)

		time.Sleep(150*time.Millisecond + scheduleOnceJitter)
		assert.Equal(t, int32(0), *count5)
	})

	t.Run("starting the scheduler again will return an error", func(t *testing.T) {
		t.Parallel()

		newScheduler, err := StartJobOnceScheduler(mockPluginAPI, callback)
		require.Error(t, err)
		require.Nil(t, newScheduler)
	})
}

func TestScheduleOnceStress(t *testing.T) {
	makeKey := model.NewId

	numPagingJobs := keysPerPage*3 + 2
	testPagingJobs := make(map[string]*int32)
	for i := 0; i < numPagingJobs; i++ {
		testPagingJobs[makeKey()] = new(int32)
	}

	callback := func(key string) {
		count, ok := testPagingJobs[key]
		if ok {
			atomic.AddInt32(count, 1)
			return
		}
	}

	mockPluginAPI := newMockPluginAPI(t)
	getVal := func(key string) []byte {
		data, _ := mockPluginAPI.KVGet(key)
		return data
	}

	// reset the scheduler from previous tests:
	func() {
		activeJobs.mu.Lock()
		defer activeJobs.mu.Unlock()
		storedCallback.mu.Lock()
		defer storedCallback.mu.Unlock()
		activeJobs.jobs = make(map[string]*JobOnce)
		storedCallback.callback = nil
		mockPluginAPI.clear() // just in case?
	}()

	// add the test paging jobs before starting scheduler
	for k := range testPagingJobs {
		assert.Empty(t, getVal(oncePrefix+k))
		job, err := newJobOnce(mockPluginAPI, k, time.Now().Add(100*time.Millisecond))
		require.NoError(t, err)
		err = job.saveMetadata()
		require.NoError(t, err)
		assert.NotEmpty(t, getVal(oncePrefix+k))
	}

	keys, appErr := mockPluginAPI.KVList(0, 99999)
	require.Nil(t, appErr)
	assert.Equal(t, len(testPagingJobs), len(keys))

	// Ensure the jobs are there and haven't run yet: (double checking because of the race detector
	// problem below.
	numInDB := 0
	numCountsAtZero := 0
	for k, v := range testPagingJobs {
		if getVal(oncePrefix+k) != nil {
			numInDB++
		}
		if atomic.LoadInt32(v) == int32(0) {
			numCountsAtZero++
		}
	}

	assert.Equal(t, len(testPagingJobs), numInDB)
	assert.Equal(t, len(testPagingJobs), numCountsAtZero)

	_, err := StartJobOnceScheduler(mockPluginAPI, callback)
	require.NoError(t, err)

	t.Run("test paging keys from the db by inserting 3 pages of jobs and starting scheduler", func(t *testing.T) {
		// wait for the testPagingJobs created in the setup to finish
		time.Sleep(300 * time.Millisecond)

		numInDB := 0
		numActive := 0
		numCountsAtZero := 0
		for k, v := range testPagingJobs {
			if getVal(oncePrefix+k) != nil {
				numInDB++
			}
			activeJobs.mu.RLock()
			if activeJobs.jobs[k] != nil {
				numActive++
			}
			activeJobs.mu.RUnlock()
			if atomic.LoadInt32(v) == int32(0) {
				numCountsAtZero++
			}
		}

		assert.Equal(t, 0, numInDB)
		assert.Equal(t, 0, numActive)
		assert.Equal(t, 0, numCountsAtZero)
	})
}

func TestScheduleOnceSequential(t *testing.T) {
	makeKey := model.NewId

	resetScheduler := func() {
		activeJobs.mu.Lock()
		defer activeJobs.mu.Unlock()
		storedCallback.mu.Lock()
		defer storedCallback.mu.Unlock()
		activeJobs.jobs = make(map[string]*JobOnce)
		storedCallback.callback = nil
	}

	t.Run("starting the scheduler with a nil callback will return an error", func(t *testing.T) {
		resetScheduler()

		mockPluginAPI := newMockPluginAPI(t)

		scheduler, err := StartJobOnceScheduler(mockPluginAPI, nil)
		require.Error(t, err)
		require.Nil(t, scheduler)
	})

	t.Run("failed at the db", func(t *testing.T) {
		resetScheduler()

		jobKey1 := makeKey()
		count1 := new(int32)

		callback := func(key string) {
			if key == jobKey1 {
				atomic.AddInt32(count1, 1)
			}
		}

		mockPluginAPI := newMockPluginAPI(t)
		getVal := func(key string) []byte {
			data, _ := mockPluginAPI.KVGet(key)
			return data
		}

		scheduler, err := StartJobOnceScheduler(mockPluginAPI, callback)
		require.NoError(t, err)
		jobs, err := scheduler.ListScheduledJobs()
		require.NoError(t, err)
		require.Empty(t, jobs)

		job, err := scheduler.ScheduleOnce(jobKey1, time.Now().Add(100*time.Millisecond))
		require.NoError(t, err)
		require.NotNil(t, job)
		assert.NotEmpty(t, getVal(oncePrefix+jobKey1))
		assert.NotEmpty(t, activeJobs.jobs[jobKey1])
		mockPluginAPI.setFailingWithPrefix(oncePrefix)

		// wait until the metadata has failed to read
		time.Sleep((maxNumFails + 1) * (waitAfterFail + scheduleOnceJitter))
		assert.Equal(t, int32(0), *count1)
		assert.Nil(t, getVal(oncePrefix+jobKey1))

		assert.Empty(t, activeJobs.jobs[jobKey1])
		assert.Empty(t, getVal(oncePrefix+jobKey1))
		assert.Equal(t, int32(0), *count1)
	})

	t.Run("simulate starting the plugin with 3 pending jobs in the db", func(t *testing.T) {
		resetScheduler()

		jobKeys := make(map[string]*int32)
		for i := 0; i < 3; i++ {
			jobKeys[makeKey()] = new(int32)
		}

		callback := func(key string) {
			count, ok := jobKeys[key]
			if ok {
				atomic.AddInt32(count, 1)
			}
		}

		mockPluginAPI := newMockPluginAPI(t)
		getVal := func(key string) []byte {
			data, _ := mockPluginAPI.KVGet(key)
			return data
		}

		for k := range jobKeys {
			job, err := newJobOnce(mockPluginAPI, k, time.Now().Add(100*time.Millisecond))
			require.NoError(t, err)
			err = job.saveMetadata()
			require.NoError(t, err)
			assert.NotEmpty(t, getVal(oncePrefix+k))
		}

		// The jobs are in the db, start the plugin and make sure it runs those jobs.
		scheduler, err := StartJobOnceScheduler(mockPluginAPI, callback)
		require.NoError(t, err)

		// double checking they're in the db:
		jobs, err := scheduler.ListScheduledJobs()
		require.NoError(t, err)
		require.Len(t, jobs, 3)

		time.Sleep(120*time.Millisecond + scheduleOnceJitter)

		for k, v := range jobKeys {
			assert.Empty(t, getVal(oncePrefix+k))
			assert.Empty(t, activeJobs.jobs[k])
			assert.Equal(t, int32(1), *v)
		}
		jobs, err = scheduler.ListScheduledJobs()
		require.NoError(t, err)
		require.Empty(t, jobs)
	})

	t.Run("calling close on a job from the db causes deadlock", func(t *testing.T) {
		resetScheduler()

		jobKeys := make(map[string]*int32)
		for i := 0; i < 3; i++ {
			jobKeys[makeKey()] = new(int32)
		}

		callback := func(key string) {
			count, ok := jobKeys[key]
			if ok {
				atomic.AddInt32(count, 1)
			}
		}

		mockPluginAPI := newMockPluginAPI(t)
		getVal := func(key string) []byte {
			data, _ := mockPluginAPI.KVGet(key)
			return data
		}

		for k := range jobKeys {
			job, err := newJobOnce(mockPluginAPI, k, time.Now().Add(100*time.Millisecond))
			require.NoError(t, err)
			err = job.saveMetadata()
			require.NoError(t, err)
			assert.NotEmpty(t, getVal(oncePrefix+k))
		}

		// The jobs are in the db, start the plugin and make sure it runs those jobs.
		scheduler, err := StartJobOnceScheduler(mockPluginAPI, callback)
		require.NoError(t, err)

		// double checking they're in the db:
		jobs, err := scheduler.ListScheduledJobs()
		require.NoError(t, err)
		require.Len(t, jobs, 3)

		time.Sleep(200*time.Millisecond + scheduleOnceJitter)

		for k, v := range jobKeys {
			assert.Empty(t, getVal(oncePrefix+k))
			assert.Empty(t, activeJobs.jobs[k])
			assert.Equal(t, int32(1), *v)
		}
		jobs, err = scheduler.ListScheduledJobs()
		require.NoError(t, err)
		require.Empty(t, jobs)
	})

	t.Run("starting a job and polling before it's finished results in only one job running", func(t *testing.T) {
		resetScheduler()

		jobKey := makeKey()
		count := new(int32)

		callback := func(key string) {
			if key == jobKey {
				atomic.AddInt32(count, 1)
			}
		}

		mockPluginAPI := newMockPluginAPI(t)
		getVal := func(key string) []byte {
			data, _ := mockPluginAPI.KVGet(key)
			return data
		}

		scheduler, err := StartJobOnceScheduler(mockPluginAPI, callback)
		require.NoError(t, err)
		jobs, err := scheduler.ListScheduledJobs()
		require.NoError(t, err)
		require.Empty(t, jobs)

		job, err := scheduler.ScheduleOnce(jobKey, time.Now().Add(100*time.Millisecond))
		require.NoError(t, err)
		require.NotNil(t, job)
		assert.NotEmpty(t, getVal(oncePrefix+jobKey))
		activeJobs.mu.Lock()
		assert.NotEmpty(t, activeJobs.jobs[jobKey])
		assert.Len(t, activeJobs.jobs, 1)
		activeJobs.mu.Unlock()

		// simulate what the polling function will do for a long running job:
		err = scheduler.scheduleNewJobsFromDB()
		require.NoError(t, err)
		err = scheduler.scheduleNewJobsFromDB()
		require.NoError(t, err)
		err = scheduler.scheduleNewJobsFromDB()
		require.NoError(t, err)
		assert.NotEmpty(t, getVal(oncePrefix+jobKey))
		activeJobs.mu.Lock()
		assert.NotEmpty(t, activeJobs.jobs[jobKey])
		assert.Len(t, activeJobs.jobs, 1)
		activeJobs.mu.Unlock()

		// now wait for it to complete
		time.Sleep(120*time.Millisecond + scheduleOnceJitter)
		assert.Equal(t, int32(1), atomic.LoadInt32(count))
		assert.Empty(t, getVal(oncePrefix+jobKey))
		activeJobs.mu.Lock()
		assert.Empty(t, activeJobs.jobs)
		activeJobs.mu.Unlock()
	})

	t.Run("starting the same job again while it's still active will fail", func(t *testing.T) {
		resetScheduler()

		jobKey := makeKey()
		count := new(int32)

		callback := func(key string) {
			if key == jobKey {
				atomic.AddInt32(count, 1)
			}
		}

		mockPluginAPI := newMockPluginAPI(t)
		getVal := func(key string) []byte {
			data, _ := mockPluginAPI.KVGet(key)
			return data
		}

		scheduler, err := StartJobOnceScheduler(mockPluginAPI, callback)
		require.NoError(t, err)
		jobs, err := scheduler.ListScheduledJobs()
		require.NoError(t, err)
		require.Empty(t, jobs)

		job, err := scheduler.ScheduleOnce(jobKey, time.Now().Add(100*time.Millisecond))
		require.NoError(t, err)
		require.NotNil(t, job)
		assert.NotEmpty(t, getVal(oncePrefix+jobKey))
		assert.NotEmpty(t, activeJobs.jobs[jobKey])
		assert.Len(t, activeJobs.jobs, 1)

		// a plugin tries to start the same jobKey again:
		job, err = scheduler.ScheduleOnce(jobKey, time.Now().Add(10000*time.Millisecond))
		require.Error(t, err)
		require.Nil(t, job)

		// now wait for first job to complete
		time.Sleep(120*time.Millisecond + scheduleOnceJitter)
		assert.Equal(t, int32(1), atomic.LoadInt32(count))
		assert.Empty(t, getVal(oncePrefix+jobKey))
		assert.Empty(t, activeJobs.jobs)
	})
}
