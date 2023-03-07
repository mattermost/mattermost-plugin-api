package cluster_test

import (
	"github.com/mattermost/mattermost-plugin-api/cluster"

	"github.com/mattermost/mattermost-server/server/v7/plugin"
)

func ExampleMutex() {
	// Use p.API from your plugin instead.
	pluginAPI := plugin.API(nil)

	m, err := cluster.NewMutex(pluginAPI, "key")
	if err != nil {
		panic(err)
	}
	m.Lock()
	// critical section
	m.Unlock()
}
