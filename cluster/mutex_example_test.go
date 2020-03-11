package cluster_test

import (
	"github.com/lieut-data/mattermost-plugin-api/cluster"
	"github.com/mattermost/mattermost-server/v5/plugin"
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
