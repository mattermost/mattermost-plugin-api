package pluginapi_test

import (
	"pluginapi"
)

type Plugin struct {
	client *pluginapi.Client
}

func (p *Plugin) OnActivate() error {
	p.client = pluginapi.NewClient(p.API)
	return nil
}

func Example() {
	p := &Plugin{}
	client := pluginapi.NewClient(nil)
	client.RegisterOnActivate(p.OnActivate)
	client.Start()
}
