package cluster

import (
	"github.com/mattermost/mattermost-server/v5/plugin"
)

func ExampleScheduleOnce() {
	// Use p.API from your plugin instead.
	pluginAPI := plugin.API(nil)

	callback := func(key string) {
		if key == "the-key-i-am-watching-for" {
			// Work to do only once per cluster
		}
	}

	// Set the callback. The most recently set callback will be used for all future job calls.
	RegisterJobOnceCallback(callback)

	// Start the scheduler. This should only be done once per plugin instance.
	scheduler, err := StartJobOnceScheduler(pluginAPI)
	if err != nil {
		// You probably forgot to call RegisterJobOnceCallback first.
		return
	}

	// Maybe you want to check the scheduled jobs, or close them:
	//jobs, err := scheduler.ListScheduledJobs()

	// main thread

	// Find the active jobs and close them
	activeJobs := scheduler.ListActiveJobs()
	defer func() {
		for _, j := range activeJobs {
			scheduler.Close(j.Key)
		}
	}()
}
