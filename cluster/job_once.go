package cluster

import (
	"encoding/json"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
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

type JobOnce struct {
	pluginAPI    JobPluginAPI
	clusterMutex *Mutex

	// key is the original key. It is prefixed with oncePrefix when used as a key in the KVStore
	key      string
	runAt    time.Time
	numFails int

	doneOnce sync.Once
	done     chan bool
}

type jobOnceMetadata struct {
	Key   string
	RunAt time.Time
}

var storedCallback struct {
	mu       sync.RWMutex
	callback func(string) error
}

var currentlyScheduled = struct {
	mu   sync.Mutex
	keys map[string]bool
}{
	keys: make(map[string]bool),
}

var startPollingOnce sync.Once

// NewJobOnce returns a JobOnce ready to be used in ScheduleOnce
func NewJobOnce(pluginAPI JobPluginAPI, metadata jobOnceMetadata) (*JobOnce, error) {
	mutex, err := NewMutex(pluginAPI, metadata.Key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create job mutex")
	}

	return &JobOnce{
		pluginAPI:    pluginAPI,
		clusterMutex: mutex,
		key:          metadata.Key,
		runAt:        metadata.RunAt,
		done:         make(chan bool),
	}, nil
}

// ScheduleOnceRegisterCallback registers the callback function that will be called when a
// ScheduleOnce job fires. Will ignore nil callbacks.
func ScheduleOnceRegisterCallback(callback func(string) error) {
	if callback == nil {
		return
	}

	storedCallback.mu.Lock()
	defer storedCallback.mu.Unlock()

	storedCallback.callback = callback
}

// ScheduleOnceStartScheduler finds all previous ScheduleOnce jobs and starts them running, and
// fires any jobs that have reached or exceeded their runAt time. Therefore, even if a cluster goes
// down and is restarted, StartScheduler will restart all previously scheduled jobs. Plugins using
// ScheduleOnce should call this when ready to handle calls to the registered callback function.
func ScheduleOnceStartScheduler(pluginAPI JobPluginAPI) error {
	if err := scheduleNewJobsFromDB(pluginAPI); err != nil {
		return errors.Wrap(err, "could not start scheduler due to error")
	}

	// Start polling but only on the first call
	startPollingOnce.Do(func() {
		go pollForNewScheduledJobs(pluginAPI)
	})

	return nil
}

// ListScheduledJobs returns a list of the jobs in the db that have been scheduled. The jobs may
// have been fired by the time the caller reads the list.
func ListScheduledJobs(pluginAPI JobPluginAPI) ([]*JobOnce, error) {
	var list []*JobOnce
	for i := 0; ; i++ {
		keys, err := pluginAPI.KVList(i, keysPerPage)
		if err != nil {
			return nil, errors.Wrap(err, "error getting KVList")
		}
		for _, k := range keys {
			if strings.HasPrefix(k, oncePrefix) {
				// We do not need the cluster lock because we are only reading the list here.
				metadata, err := readMetadata(pluginAPI, k[len(oncePrefix):])
				if err != nil {
					pluginAPI.LogError(errors.Wrap(err, "could not retrieve data from plugin kvstore for key: "+k).Error())
					continue
				}
				if metadata == nil {
					continue
				}

				job, err := NewJobOnce(pluginAPI, *metadata)
				if err != nil {
					pluginAPI.LogError(errors.Wrap(err, "could not create new job for key: "+k).Error())
					continue
				}
				list = append(list, job)
			}
		}

		if len(keys) < keysPerPage {
			break
		}
	}

	return list, nil
}

// ScheduleOnce creates a scheduled job that will run once. When the clock reaches runAt, the
// storedCallback (set using RegisterCallback) will be called with key as the argument.
func ScheduleOnce(pluginAPI JobPluginAPI, key string, runAt time.Time) (*JobOnce, error) {
	storedCallback.mu.RLock()
	defer storedCallback.mu.RUnlock()
	if storedCallback.callback == nil {
		return nil, errors.New("must call RegisterCallback before scheduling new jobs")
	}

	metadata := jobOnceMetadata{
		Key:   key,
		RunAt: runAt,
	}
	job, err := NewJobOnce(pluginAPI, metadata)
	if err != nil {
		return nil, errors.Wrap(err, "could not create new job")
	}

	if err = job.saveMetadata(); err != nil {
		return nil, errors.Wrap(err, "could not save job metadata")
	}

	go job.run()

	return job, nil
}

// Close terminates a scheduled job, preventing it from being scheduled on this plugin instance.
// It also removes the job from the db, preventing it from being run in the future.
func (j *JobOnce) Close() {
	// Acquire the corresponding job lock when modifying the db
	j.clusterMutex.Lock()
	defer j.clusterMutex.Unlock()

	j.closeHoldingMutex()
}

func scheduleNewJobsFromDB(pluginAPI JobPluginAPI) error {
	storedCallback.mu.RLock()
	defer storedCallback.mu.RUnlock()
	if storedCallback.callback == nil {
		return errors.New("must call RegisterCallback before starting the scheduler")
	}

	jobs, err := ListScheduledJobs(pluginAPI)
	if err != nil {
		return errors.Wrap(err, "could not read scheduled jobs from db")
	}

	// lock and hold until we're done updating the currently scheduled jobs
	currentlyScheduled.mu.Lock()
	defer currentlyScheduled.mu.Unlock()

	for _, j := range jobs {
		if currentlyScheduled.keys[j.key] {
			continue
		}

		go j.run()

		currentlyScheduled.keys[j.key] = true
	}

	return nil
}

// pollForNewScheduledJobs will only be started once per plugin. It doesn't need to be stopped.
func pollForNewScheduledJobs(pluginAPI JobPluginAPI) {
	for {
		select {
		case <-time.After(pollNewJobsInterval + addJitter()):
		}

		if err := scheduleNewJobsFromDB(pluginAPI); err != nil {
			pluginAPI.LogError("pluginAPI scheduleOnce poller encountered an error but is still polling", "error", err)
		}
	}
}

func (j *JobOnce) run() {
	wait := j.runAt.Sub(time.Now())

	for {
		select {
		case <-j.done:
			return
		case <-time.After(wait):
		}

		func() {
			// Acquire the cluster mutex while we're trying to do the job
			j.clusterMutex.Lock()
			defer j.clusterMutex.Unlock()

			// Check that the job has not been completed
			metadata, err := j.readMetadata()
			if err != nil {
				j.numFails++
				if j.numFails > maxNumFails {
					j.closeHoldingMutex()
					return
				}

				// wait a bit of time and try again
				wait = waitAfterFail + addJitter()
				return
			}

			// If key doesn't exist, the job has been completed already
			if metadata == nil {
				j.closeHoldingMutex()
				return
			}

			// Run the job
			storedCallback.mu.RLock()
			defer storedCallback.mu.RUnlock()
			err = storedCallback.callback(j.key)
			if err != nil {
				j.pluginAPI.LogError("callback returned an error for job key: " + j.key)
			}

			j.closeHoldingMutex()
		}()
	}
}

// readMetadata reads the job's stored metadata. If the caller wishes to make an atomic
// read/write, the cluster mutex for job's key should be held.
func readMetadata(pluginAPI JobPluginAPI, key string) (*jobOnceMetadata, error) {
	data, appErr := pluginAPI.KVGet(oncePrefix + key)
	if appErr != nil {
		return nil, errors.Wrap(appErr, "failed to read data")
	}

	if data == nil {
		return nil, nil
	}

	var metadata jobOnceMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, errors.Wrap(err, "failed to decode data")
	}

	return &metadata, nil
}

func (j *JobOnce) readMetadata() (*jobOnceMetadata, error) {
	return readMetadata(j.pluginAPI, j.key)
}

// saveMetadata writes the job's metadata to the kvstore.
func (j *JobOnce) saveMetadata() error {
	metadata := jobOnceMetadata{
		Key:   j.key,
		RunAt: j.runAt,
	}
	data, err := json.Marshal(metadata)
	if err != nil {
		return errors.Wrap(err, "failed to marshal data")
	}

	j.clusterMutex.Lock()
	defer j.clusterMutex.Unlock()

	ok, appErr := j.pluginAPI.KVSetWithOptions(oncePrefix+j.key, data, model.PluginKVSetOptions{})
	if appErr != nil || !ok {
		return errors.Wrap(appErr, "failed to set data")
	}

	return nil
}

// closeHoldingMutex assumes the job's mutex is held.
func (j *JobOnce) closeHoldingMutex() {
	// remove the job from the kv store, if it exists
	_ = j.pluginAPI.KVDelete(oncePrefix + j.key)

	// remove the job from the currentlyScheduled map so we can reschedule if needed later
	currentlyScheduled.mu.Lock()
	defer currentlyScheduled.mu.Unlock()
	delete(currentlyScheduled.keys, j.key)

	j.doneOnce.Do(func() {
		close(j.done)
	})
}

func addJitter() time.Duration {
	return time.Duration(rand.Int63n(int64(scheduleOnceJitter)))
}
