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

	// Start the scheduler.
	scheduler, err := StartJobOnceScheduler(pluginAPI, callback)
	if err != nil {
		// You probably forgot to call RegisterJobOnceCallback first.
		return
	}

	// main thread

	// Maybe you want to check the scheduled jobs, or close them:
	jobs, err := scheduler.ListScheduledJobs()
	defer func() {
		for _, j := range jobs {
			scheduler.Close(j.Key)
		}
	}()
}
