// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package cluster

import (
	"strings"
	"sync"
	"time"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
)

type syncedCallbacks struct {
	mu        sync.Mutex
	callbacks map[string]func(string)
}

type syncedJobs struct {
	mu   sync.RWMutex
	jobs map[string]*JobOnce
}

type JobOnceScheduler struct {
	pluginAPI JobPluginAPI

	startedMu sync.RWMutex
	started   bool

	activeJobs      *syncedJobs
	storedCallbacks *syncedCallbacks
}

var schedulerOnce sync.Once
var s *JobOnceScheduler

// GetJobOnceScheduler returns a scheduler which is ready to accept callbacks. Repeated calls will
// return the same scheduler.
func GetJobOnceScheduler(pluginAPI JobPluginAPI) *JobOnceScheduler {
	schedulerOnce.Do(func() {
		s = &JobOnceScheduler{
			pluginAPI: pluginAPI,
			activeJobs: &syncedJobs{
				jobs: make(map[string]*JobOnce),
			},
			storedCallbacks: &syncedCallbacks{
				callbacks: make(map[string]func(string)),
			},
		}
	})
	return s
}

// Start finds all previous ScheduleOnce jobs and starts them running, and fires any jobs that have
// reached or exceeded their runAt time. Thus, even if a cluster goes down and is restarted, Start
// will restart previously scheduled jobs.
func (s *JobOnceScheduler) Start() error {
	s.startedMu.Lock()
	defer s.startedMu.Unlock()
	if s.started {
		return errors.New("scheduler has already been started")
	}

	if err := s.verifyCallbackExists(); err != nil {
		return err
	}

	if err := s.scheduleNewJobsFromDB(); err != nil {
		return errors.Wrap(err, "could not start scheduler due to error")
	}

	go s.pollForNewScheduledJobs()

	s.started = true

	return nil
}

// AddCallback adds a callback to the list. Each will be called with the job's id when the
// job fires. Returns the id of the callback which can be used to remove the callback in
// RemoveCallback.
func (s *JobOnceScheduler) AddCallback(callback func(string)) (string, error) {
	if callback == nil {
		return "", errors.New("callback cannot be nil")
	}

	s.storedCallbacks.mu.Lock()
	defer s.storedCallbacks.mu.Unlock()

	id := model.NewId()
	s.storedCallbacks.callbacks[id] = callback
	return id, nil
}

// RemoveCallback will remove a callback by its id.
func (s *JobOnceScheduler) RemoveCallback(id string) {
	s.storedCallbacks.mu.Lock()
	defer s.storedCallbacks.mu.Unlock()
	delete(s.storedCallbacks.callbacks, id)
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

// ScheduleOnce creates a scheduled job that will run once. When the clock reaches runAt, all
// callbacks will be called with key as the argument.
//
// If the job key already exists in the db, this will return an error. To reschedule a job, first
// close the original then schedule the new one. For example: find the job in the list returned by
// ListActiveJobs and call Close(key).
func (s *JobOnceScheduler) ScheduleOnce(key string, runAt time.Time) (*JobOnce, error) {
	s.startedMu.RLock()
	defer s.startedMu.RUnlock()
	if !s.started {
		return nil, errors.New("start the scheduler before adding jobs")
	}
	if err := s.verifyCallbackExists(); err != nil {
		return nil, err
	}

	job, err := newJobOnce(s.pluginAPI, key, runAt, s.storedCallbacks, s.activeJobs)
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
		s.activeJobs.mu.RLock()
		defer s.activeJobs.mu.RUnlock()
		j, ok := s.activeJobs.jobs[key]
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
		job, err := newJobOnce(s.pluginAPI, m.Key, m.RunAt, s.storedCallbacks, s.activeJobs)
		if err != nil {
			s.pluginAPI.LogError(errors.Wrap(err, "could not create new job for key: "+m.Key).Error())
			continue
		}

		s.runAndTrack(job)
	}

	return nil
}

func (s *JobOnceScheduler) runAndTrack(job *JobOnce) {
	s.activeJobs.mu.Lock()
	defer s.activeJobs.mu.Unlock()

	// has this been scheduled already on this server?
	if _, ok := s.activeJobs.jobs[job.key]; ok {
		return
	}

	go job.run()

	s.activeJobs.jobs[job.key] = job
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

func (s *JobOnceScheduler) verifyCallbackExists() error {
	s.storedCallbacks.mu.Lock()
	defer s.storedCallbacks.mu.Unlock()

	if len(s.storedCallbacks.callbacks) == 0 {
		return errors.New("add callbacks before starting the scheduler")
	}
	return nil
}
