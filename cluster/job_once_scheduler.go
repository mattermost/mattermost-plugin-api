// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package cluster

import (
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type JobOnceScheduler struct {
	pluginAPI JobPluginAPI
}

var storedCallback struct {
	mu       sync.RWMutex
	callback func(string)
}

var activeJobs = struct {
	mu   sync.RWMutex
	jobs map[string]*JobOnce
}{
	jobs: make(map[string]*JobOnce),
}

var startPollingOnce sync.Once

// RegisterJobOnceCallback registers the callback function that will be called when a
// ScheduleOnce job fires. Will ignore nil callbacks.
func RegisterJobOnceCallback(callback func(string)) {
	if callback == nil {
		return
	}

	storedCallback.mu.Lock()
	defer storedCallback.mu.Unlock()

	storedCallback.callback = callback
}

// StartJobOnceScheduler finds all previous ScheduleOnce jobs and starts them running, and
// fires any jobs that have reached or exceeded their runAt time. Therefore, even if a cluster goes
// down and is restarted, StartScheduler will restart all previously scheduled jobs. Plugins using
// ScheduleOnce should call this when ready to handle calls to the registered callback function.
func StartJobOnceScheduler(pluginAPI JobPluginAPI) (*JobOnceScheduler, error) {
	scheduler := &JobOnceScheduler{
		pluginAPI: pluginAPI,
	}
	if err := scheduler.scheduleNewJobsFromDB(); err != nil {
		return nil, errors.Wrap(err, "could not start scheduler due to error")
	}

	startPollingOnce.Do(func() {
		go scheduler.pollForNewScheduledJobs()
	})

	return scheduler, nil
}

// ListScheduledJobs returns a list of the jobs in the db that have been scheduled. The jobs may
// have been fired by the time the caller reads the list.
func (s *JobOnceScheduler) ListScheduledJobs() ([]JobOnceMetadata, error) {
	var ret []JobOnceMetadata
	for i := 0; ; i++ {
		keys, err := s.pluginAPI.KVList(i, keysPerPage)
		if err != nil {
			return nil, errors.Wrap(err, "error getting KVList")
		}
		for _, k := range keys {
			if strings.HasPrefix(k, oncePrefix) {
				// We do not need the cluster lock because we are only reading the list here.
				metadata, err := readMetadata(s.pluginAPI, k[len(oncePrefix):])
				if err != nil {
					s.pluginAPI.LogError(errors.Wrap(err, "could not retrieve data from plugin kvstore for key: "+k).Error())
					continue
				}
				if metadata == nil {
					continue
				}

				ret = append(ret, *metadata)
			}
		}

		if len(keys) < keysPerPage {
			break
		}
	}

	return ret, nil
}

// ListActiveJobs returns a list of the jobs currently running in the current plugin. The jobs may
// have been fired by the time the caller reads the list.
func (s *JobOnceScheduler) ListActiveJobs() []JobOnceMetadata {
	activeJobs.mu.RLock()
	defer activeJobs.mu.RUnlock()

	var ret []JobOnceMetadata
	for _, j := range activeJobs.jobs {
		ret = append(ret, JobOnceMetadata{
			Key:   j.key,
			RunAt: j.runAt,
		})
	}
	return ret
}

// ScheduleOnce creates a scheduled job that will run once. When the clock reaches runAt, the
// storedCallback (set using RegisterCallback) will be called with key as the argument.
//
// If the job already exists in the db, this will return an error. To reschedule the job, first
// close the original  then schedule the new one. For example: find the job in the list returned by
// ListActiveJobs and call Close(key).
func (s *JobOnceScheduler) ScheduleOnce(key string, runAt time.Time) (*JobOnce, error) {
	storedCallback.mu.RLock()
	defer storedCallback.mu.RUnlock()
	if storedCallback.callback == nil {
		return nil, errors.New("must call RegisterCallback before scheduling new jobs")
	}

	job, err := newJobOnce(s.pluginAPI, key, runAt)
	if err != nil {
		return nil, errors.Wrap(err, "could not create new job")
	}

	if err = job.saveMetadata(); err != nil {
		return nil, errors.Wrap(err, "could not save job metadata")
	}

	s.runAndTrack(job)

	return job, nil
}

// Close closes a job by its key. This is useful if the plugin lost the original JobOnce pointer.
func (s *JobOnceScheduler) Close(key string) {
	activeJobs.mu.RLock()
	j, ok := activeJobs.jobs[key]
	// cannot defer because j.Close will eventually acquire the lock
	activeJobs.mu.RUnlock()

	if !ok {
		// Job wasn't active, so no need to call Close (which shuts down the goroutine).
		// There's a condition where another server in the cluster started the job, and the
		// current server hasn't polled for it yet. To solve that case, delete it from the db.
		// Acquire the corresponding job lock when modifying the db.
		mutex, err := NewMutex(s.pluginAPI, key)
		if err != nil {
			s.pluginAPI.LogError(errors.Wrap(err, "failed to create job mutex in Close for key: "+key).Error())
		}
		mutex.Lock()
		defer mutex.Unlock()

		_ = s.pluginAPI.KVDelete(oncePrefix + key)

		return
	}

	j.Close()
}

func (s *JobOnceScheduler) scheduleNewJobsFromDB() error {
	storedCallback.mu.RLock()
	defer storedCallback.mu.RUnlock()
	if storedCallback.callback == nil {
		return errors.New("must call RegisterCallback before starting the scheduler")
	}

	scheduled, err := s.ListScheduledJobs()
	if err != nil {
		return errors.Wrap(err, "could not read scheduled jobs from db")
	}

	for _, m := range scheduled {
		job, err := newJobOnce(s.pluginAPI, m.Key, m.RunAt)
		if err != nil {
			s.pluginAPI.LogError(errors.Wrap(err, "could not create new job for key: "+m.Key).Error())
			continue
		}

		func() {
			// It's okay if the job has been removed since reading it from the db -- that would be
			// rare, and it wouldn't fire anyway since the check is done in job.run()
			s.runAndTrack(job)
		}()
	}

	return nil
}

func (s *JobOnceScheduler) runAndTrack(job *JobOnce) {
	activeJobs.mu.Lock()
	defer activeJobs.mu.Unlock()

	// has this been scheduled already on this server?
	if _, ok := activeJobs.jobs[job.key]; ok {
		return
	}

	go job.run()

	activeJobs.jobs[job.key] = job
}

// pollForNewScheduledJobs will only be started once per plugin. It doesn't need to be stopped.
func (s *JobOnceScheduler) pollForNewScheduledJobs() {
	for {
		<-time.After(pollNewJobsInterval + addJitter())

		if err := s.scheduleNewJobsFromDB(); err != nil {
			s.pluginAPI.LogError("pluginAPI scheduleOnce poller encountered an error but is still polling", "error", err)
		}
	}
}
