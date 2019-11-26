package pluginapi

import (
	"net/rpc"

	hashicorpPlugin "github.com/hashicorp/go-plugin"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

// hooksRPCServer embeds the Mattermost-server definition of same, but overrides the Implemented
// hook to customize the list of hooks.
type hooksRPCServer struct {
	*plugin.HooksRPCServer
	client *Client
}

// Implemented informs the Mattermost server of the set of hooks the active plugin has registered.
func (s *hooksRPCServer) Implemented(args struct{}, reply *[]string) error {
	methodIdToName := map[int]string{
		plugin.OnActivateId:            "OnActivate",
		plugin.OnDeactivateId:          "OnDeactivate",
		plugin.ServeHTTPId:             "ServeHTTP",
		plugin.OnConfigurationChangeId: "OnConfigurationChange",
		plugin.ExecuteCommandId:        "ExecuteCommand",
		plugin.MessageWillBePostedId:   "MessageWillBePosted",
		plugin.MessageWillBeUpdatedId:  "MessageWillBeUpdated",
		plugin.MessageHasBeenPostedId:  "MessageHasBeenPosted",
		plugin.MessageHasBeenUpdatedId: "MessageHasBeenUpdated",
		plugin.UserHasJoinedChannelId:  "UserHasJoinedChannel",
		plugin.UserHasLeftChannelId:    "UserHasLeftChannel",
		plugin.UserHasJoinedTeamId:     "UserHasJoinedTeam",
		plugin.UserHasLeftTeamId:       "UserHasLeftTeam",
		plugin.ChannelHasBeenCreatedId: "ChannelHasBeenCreated",
		plugin.FileWillBeUploadedId:    "FileWillBeUploaded",
		plugin.UserWillLogInId:         "UserWillLogIn",
		plugin.UserHasLoggedInId:       "UserHasLoggedIn",
		plugin.UserHasBeenCreatedId:    "UserHasBeenCreated",
	}

	methodNames := []string{}
	for _, hookId := range s.client.hooks.registered {
		if methodName, ok := methodIdToName[hookId]; ok {
			methodNames = append(methodNames, methodName)
		}
	}

	*reply = methodNames
	return nil
}

// hooksPlugin implements hashicorp/go-plugin/plugin.Plugin interface to connect plugin hooks.
type hooksPlugin struct {
	hooks  interface{}
	client *Client
}

func (p *hooksPlugin) Server(b *hashicorpPlugin.MuxBroker) (interface{}, error) {
	return &hooksRPCServer{
		plugin.NewHooksRPCServer(b, p.hooks),
		p.client,
	}, nil
}

func (p *hooksPlugin) Client(b *hashicorpPlugin.MuxBroker, client *rpc.Client) (interface{}, error) {
	// The implementation of the client is located in github.com/mattermost-server/plugin,
	// and is not duplicated here at this time. We may later choose to host the entirety of the
	// server/client API here, along with the generated code.
	return nil, nil
}

// Start is the entry point to be called from a plugin's main() function.
func (c *Client) Start() {
	hashicorpPlugin.Serve(&hashicorpPlugin.ServeConfig{
		HandshakeConfig: plugin.Handshake,
		Plugins: map[string]hashicorpPlugin.Plugin{
			"hooks": &hooksPlugin{hooks: &pluginImplementation{c}, client: c},
		},
	})
}
