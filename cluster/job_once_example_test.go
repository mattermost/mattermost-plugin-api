package cluster

import (
	"github.com/mattermost/mattermost-server/v5/plugin"
)

func ExampleScheduleOnce() {
	// Use p.API from your plugin instead.
	pluginAPI := plugin.API(nil)

	callback := func(key string) error {
		if key == "the-key-i-am-watching-for" {
			// Work to do only once per cluster

			// If successful, return nil to signal the job has been completed
			return nil

			// If unsuccessful, return any error and the job will try again in a minute e.g.:
			// return errors.New("didn't work")
		}

		// Watch for the other keys...

		// If the plugin was sent a key it wasn't expecting, you probably want to return nil so that
		// it isn't rescheduled and sent to you again.
		return nil
	}

	// Set the callback. The most recently set callback will be used for all future job calls.
	ScheduleOnceRegisterCallback(callback)

	// Start the scheduler. This should only be done once per plugin instance.
	err := ScheduleOnceStartScheduler(pluginAPI)
	if err != nil {
		// You probably forgot to call ScheduleOnceRegisterCallback first.
	}

	jobs, err := ListScheduledJobs(pluginAPI)
	// Maybe you want to check the scheduled jobs, or close them.

	// main thread

	defer func() {
		for _, j := range jobs {
			j.Close()
		}
	}()
}
