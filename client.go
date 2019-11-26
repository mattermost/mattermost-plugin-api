package pluginapi

import "github.com/mattermost/mattermost-server/v5/plugin"

// Client is a streamlined wrapper over the mattermost plugin API.
type Client struct {
	api   plugin.API
	hooks hooks

	User     UserService
	Post     PostService
	Reaction ReactionService
	Emoji    EmojiService
	File     FileService
	KV       KVService
	Bot      BotService
}

// NewClient creates a new instance of Client, blocking until the hosting server initializes the
// plugin API.
func NewClient() *Client {
	return &Client{
		User:     UserService{api},
		Post:     PostService{api},
		Reaction: ReactionService{api},
		Emoji:    EmojiService{api},
		File:     FileService{api},
		KV:       KVService{api},
		Bot:      BotService{api},
	}
}
