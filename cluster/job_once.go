package cluster

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
)

const (
	// oncePrefix is used to namespace key values created for a scheduleOnce job
	oncePrefix = "once_"

	// keysPerPage is the maximum number of keys to retrieve from the db per call
	keysPerPage = 1000

	// maxNumFails is the maximum number of KVStore read fails or failed attempts to run the
	// callback until the scheduler cancels a job.
	maxNumFails = 3

	// waitAfterFail is the amount of time to wait after a failure
	waitAfterFail = 1 * time.Second

	// pollNewJobsInterval is the amount of time to wait between polling the db for new scheduled jobs
	pollNewJobsInterval = 5 * time.Minute

	// scheduleOnceJitter is the range of jitter to add to intervals to avoid contention issues
	scheduleOnceJitter = 100 * time.Millisecond
)

type JobOnceMetadata struct {
	Key   string
	RunAt time.Time
}

type JobOnce struct {
	pluginAPI    JobPluginAPI
	clusterMutex *Mutex

	// key is the original key. It is prefixed with oncePrefix when used as a key in the KVStore
	key      string
	runAt    time.Time
	numFails int

	// done signals the job.run go routine to exit
	done     chan bool
	doneOnce sync.Once

	// join is a join point for the job.run() goroutine to join the calling goroutine (in this case,
	// the one calling job.Cancel)
	join     chan bool
	joinOnce sync.Once

	storedCallback *syncedCallback
	activeJobs     *syncedJobs
}

// Cancel terminates a scheduled job, preventing it from being scheduled on this plugin instance.
// It also removes the job from the db, preventing it from being run in the future.
func (j *JobOnce) Cancel() {
	s.pluginAPI.LogError("<><> jobOnce.Cancel in job: %s, locking clusteMutex", j.key)
	j.clusterMutex.Lock()
	defer j.clusterMutex.Unlock()

	s.pluginAPI.LogError("<><> jobOnce.Cancel in job: %s, calling cancelWhileHoldingMutex", j.key)
	j.cancelWhileHoldingMutex()

	// join the running goroutine
	j.joinOnce.Do(func() {
		<-j.join
	})
}

func newJobOnce(pluginAPI JobPluginAPI, key string, runAt time.Time, callback *syncedCallback, jobs *syncedJobs) (*JobOnce, error) {
	mutex, err := NewMutex(pluginAPI, key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create job mutex")
	}

	return &JobOnce{
		pluginAPI:      pluginAPI,
		clusterMutex:   mutex,
		key:            key,
		runAt:          runAt,
		done:           make(chan bool),
		join:           make(chan bool),
		storedCallback: callback,
		activeJobs:     jobs,
	}, nil
}

func (j *JobOnce) run() {
	defer close(j.join)

	wait := time.Until(j.runAt)

	for {
		select {
		case <-j.done:
			return
		case <-time.After(wait + addJitter()):
		}

		func() {
			// Acquire the cluster mutex while we're trying to do the job
			j.clusterMutex.Lock()
			defer j.clusterMutex.Unlock()

			// Check that the job has not been completed
			j.pluginAPI.LogError("<><> jobOnce.run for key: %s, timer fired!  readMetadata", j.key)
			metadata, err := readMetadata(j.pluginAPI, j.key)
			if err != nil {
				j.pluginAPI.LogError("<><> jobOnce.run for key %s, readMetadata error: %v", j.key, err)
				j.numFails++
				if j.numFails > maxNumFails {
					j.cancelWhileHoldingMutex()
					return
				}

				// wait a bit of time and try again
				wait = waitAfterFail
				return
			}

			// If key doesn't exist, the job has been completed already
			if metadata == nil {
				j.pluginAPI.LogError("<><> jobOnce.run for key %s, readMetadata was nil. Now cancel while holding mutex", j.key)
				j.cancelWhileHoldingMutex()
				return
			}

			j.pluginAPI.LogError("<><> jobOnce.run for key %s, executing job", j.key)
			j.executeJob()

			j.pluginAPI.LogError("<><> jobOnce.run for key %s, cancelWhileHoldingMutex", j.key)
			j.cancelWhileHoldingMutex()
		}()
	}
}

func (j *JobOnce) executeJob() {
	j.storedCallback.mu.Lock()
	defer j.storedCallback.mu.Unlock()

	j.storedCallback.callback(j.key)
}

// readMetadata reads the job's stored metadata. If the caller wishes to make an atomic
// read/write, the cluster mutex for job's key should be held.
func readMetadata(pluginAPI JobPluginAPI, key string) (*JobOnceMetadata, error) {
	pluginAPI.LogError("<><> readMetadata for key %s", key)
	data, appErr := pluginAPI.KVGet(oncePrefix + key)
	if appErr != nil {
		pluginAPI.LogError("<><> readMetadata error, failed to read data for key %s", key)
		return nil, errors.Wrap(normalizeAppErr(appErr), "failed to read data")
	}

	if data == nil {
		return nil, nil
	}

	var metadata JobOnceMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, errors.Wrap(err, "failed to decode data")
	}

	return &metadata, nil
}

// saveMetadata writes the job's metadata to the kvstore. saveMetadata acquires the job's cluster lock.
// saveMetadata will not overwrite an existing key.
func (j *JobOnce) saveMetadata() error {
	j.clusterMutex.Lock()
	defer j.clusterMutex.Unlock()

	metadata := JobOnceMetadata{
		Key:   j.key,
		RunAt: j.runAt,
	}
	data, err := json.Marshal(metadata)
	if err != nil {
		return errors.Wrap(err, "failed to marshal data")
	}

	ok, appErr := j.pluginAPI.KVSetWithOptions(oncePrefix+j.key, data, model.PluginKVSetOptions{
		Atomic:   true,
		OldValue: nil,
	})
	if appErr != nil {
		return normalizeAppErr(appErr)
	}
	if !ok {
		return errors.New("failed to set data")
	}

	return nil
}

// cancelWhileHoldingMutex assumes the caller holds the job's mutex.
func (j *JobOnce) cancelWhileHoldingMutex() {
	s.pluginAPI.LogError("<><> jobOnce.cancelWhileHoldingMutex in job: %s, calling pluginAPI.KVDelete key: %s", j.key, oncePrefix+j.key)

	// remove the job from the kv store, if it exists
	err := j.pluginAPI.KVDelete(oncePrefix + j.key)
	if err != nil {
		s.pluginAPI.LogError("<><> jobOnce.cancelWhileHoldingMutex in job: %s, got err: %v", j.key, err)
	}

	j.activeJobs.mu.Lock()
	defer j.activeJobs.mu.Unlock()
	s.pluginAPI.LogError("<><> jobOnce.cancelWhileHoldingMutex in job: %s, calling pluginAPI.KVDelete key: %s", j.key, oncePrefix+j.key)
	delete(j.activeJobs.jobs, j.key)

	j.doneOnce.Do(func() {
		close(j.done)
	})
}

func addJitter() time.Duration {
	return time.Duration(rand.Int63n(int64(scheduleOnceJitter)))
}

func normalizeAppErr(appErr *model.AppError) error {
	if appErr == nil {
		return nil
	}

	if appErr.StatusCode == http.StatusNotFound {
		return pluginapi.ErrNotFound
	}

	return appErr
}
