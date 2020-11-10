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

// To make things more predictable for clients of job once scheduler, the callback function will
// be called only once at a time.
var storedCallback struct {
	mu       sync.Mutex
	callback func(string)
}

var activeJobs = struct {
	mu   sync.RWMutex
	jobs map[string]*JobOnce
}{
	jobs: make(map[string]*JobOnce),
}

var startPollingOnce sync.Once

// StartJobOnceScheduler sets the callback function, finds all previous ScheduleOnce jobs and starts
// them running, and fires any jobs that have reached or exceeded their runAt time. Thus, even if a
// cluster goes down and is restarted, StartJobOnceScheduler will restart previously scheduled jobs.
func StartJobOnceScheduler(pluginAPI JobPluginAPI, callback func(string)) (*JobOnceScheduler, error) {
	if err := setCallback(callback); err != nil {
		return nil, err
	}

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

// ListScheduledJobs returns a list of the jobs in the db that have been scheduled. There is no
// guarantee that list is accurate by the time the caller reads the list. E.g., the jobs in the list
// may have been run, closed, or new jobs may have scheduled.
func (s *JobOnceScheduler) ListScheduledJobs() ([]JobOnceMetadata, error) {
	var ret []JobOnceMetadata
	for i := 0; ; i++ {
		keys, err := s.pluginAPI.KVList(i, keysPerPage)
		if err != nil {
			return nil, errors.Wrap(err, "error getting KVList")
		}
		for _, k := range keys {
			if strings.HasPrefix(k, oncePrefix) {
				// We do not need the cluster lock because we cannot guarantee that the list is accurate
				// by the time the caller reads it.
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

// ScheduleOnce creates a scheduled job that will run once. When the clock reaches runAt, the
// callback (set in StartJobOnceScheduler) will be called with key as the argument.
//
// If the job key already exists in the db, this will return an error. To reschedule a job, first
// close the original  then schedule the new one. For example: find the job in the list returned by
// ListActiveJobs and call Close(key).
func (s *JobOnceScheduler) ScheduleOnce(key string, runAt time.Time) (*JobOnce, error) {
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

// Close closes a job by its key. This is useful if the plugin lost the original *JobOnce, or
// is stopping a job found in ListScheduledJobs().
func (s *JobOnceScheduler) Close(key string) {
	// using an anonymous function because job.Close() below needs access to the activeJobs mutex
	job := func() *JobOnce {
		activeJobs.mu.RLock()
		defer activeJobs.mu.RUnlock()
		j, ok := activeJobs.jobs[key]
		if ok {
			return j
		}

		// Job wasn't active, so no need to call Close (which shuts down the goroutine).
		// There's a condition where another server in the cluster started the job, and the
		// current server hasn't polled for it yet. To solve that case, delete it from the db.
		mutex, err := NewMutex(s.pluginAPI, key)
		if err != nil {
			s.pluginAPI.LogError(errors.Wrap(err, "failed to create job mutex in Close for key: "+key).Error())
		}
		mutex.Lock()
		defer mutex.Unlock()

		_ = s.pluginAPI.KVDelete(oncePrefix + key)

		return nil
	}()

	if job != nil {
		job.Close()
	}
}

func (s *JobOnceScheduler) scheduleNewJobsFromDB() error {
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

		s.runAndTrack(job)
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

func setCallback(callback func(string)) error {
	storedCallback.mu.Lock()
	defer storedCallback.mu.Unlock()
	if storedCallback.callback != nil {
		return errors.New("cannot start scheduler more than once")
	}
	if callback == nil {
		return errors.New("callback cannot be nil")
	}
	storedCallback.callback = callback
	return nil
}
