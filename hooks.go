package pluginapi

import (
	"io"
	"net/http"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

type hooks struct {
	registered []int

	onActivate            func() error
	onDeactivate          func() error
	onConfigurationChange func() error
	serveHTTP             func(c *plugin.Context, w http.ResponseWriter, r *http.Request)
	executeCommand        func(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, error)
	userHasBeenCreated    func(c *plugin.Context, user *model.User)
	userWillLogIn         func(c *plugin.Context, user *model.User) string
	userHasLoggedIn       func(c *plugin.Context, user *model.User)
	messageWillBePosted   func(c *plugin.Context, post *model.Post) (*model.Post, string)
	messageWillBeUpdated  func(c *plugin.Context, newPost, oldPost *model.Post) (*model.Post, string)
	messageHasBeenPosted  func(c *plugin.Context, post *model.Post)
	messageHasBeenUpdated func(c *plugin.Context, newPost, oldPost *model.Post)
	channelHasBeenCreated func(c *plugin.Context, channel *model.Channel)
	userHasJoinedChannel  func(c *plugin.Context, channelMember *model.ChannelMember, actor *model.User)
	userHasLeftChannel    func(c *plugin.Context, channelMember *model.ChannelMember, actor *model.User)
	userHasJoinedTeam     func(c *plugin.Context, teamMember *model.TeamMember, actor *model.User)
	userHasLeftTeam       func(c *plugin.Context, teamMember *model.TeamMember, actor *model.User)
	fileWillBeUploaded    func(c *plugin.Context, info *model.FileInfo, file io.Reader, output io.Writer) (*model.FileInfo, string)
}

func (c *Client) RegisterOnActivate(callback func() error) {
	c.hooks.registered = append(c.hooks.registered, plugin.OnActivateId)
	c.hooks.onActivate = callback
}

func (c *Client) RegisterOnDeactivate(callback func() error) {
	c.hooks.registered = append(c.hooks.registered, plugin.OnDeactivateId)
	c.hooks.onDeactivate = callback
}

func (c *Client) RegisterOnConfigurationChange(callback func() error) {
	c.hooks.registered = append(c.hooks.registered, plugin.OnConfigurationChangeId)
	c.hooks.onConfigurationChange = callback
}

func (c *Client) RegisterServeHTTP(callback func(c *plugin.Context, w http.ResponseWriter, r *http.Request)) {
	c.hooks.registered = append(c.hooks.registered, plugin.ServeHTTPId)
	c.hooks.serveHTTP = callback
}

func (c *Client) RegisterExecuteCommand(callback func(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, error)) {
	c.hooks.registered = append(c.hooks.registered, plugin.ExecuteCommandId)
	c.hooks.executeCommand = callback
}

func (c *Client) RegisterUserHasBeenCreated(callback func(c *plugin.Context, user *model.User)) {
	c.hooks.registered = append(c.hooks.registered, plugin.UserHasBeenCreatedId)
	c.hooks.userHasBeenCreated = callback
}

func (c *Client) RegisterUserWillLogIn(callback func(c *plugin.Context, user *model.User) string) {
	c.hooks.registered = append(c.hooks.registered, plugin.UserWillLogInId)
	c.hooks.userWillLogIn = callback
}

func (c *Client) RegisterUserHasLoggedIn(callback func(c *plugin.Context, user *model.User)) {
	c.hooks.registered = append(c.hooks.registered, plugin.UserHasLoggedInId)
	c.hooks.userHasLoggedIn = callback
}

func (c *Client) RegisterMessageWillBePosted(callback func(c *plugin.Context, post *model.Post) (*model.Post, string)) {
	c.hooks.registered = append(c.hooks.registered, plugin.MessageWillBePostedId)
	c.hooks.messageWillBePosted = callback
}

func (c *Client) RegisterMessageWillBeUpdated(callback func(c *plugin.Context, newPost, oldPost *model.Post) (*model.Post, string)) {
	c.hooks.registered = append(c.hooks.registered, plugin.MessageWillBeUpdatedId)
	c.hooks.messageWillBeUpdated = callback
}

func (c *Client) RegisterMessageHasBeenPosted(callback func(c *plugin.Context, post *model.Post)) {
	c.hooks.registered = append(c.hooks.registered, plugin.MessageHasBeenPostedId)
	c.hooks.messageHasBeenPosted = callback
}

func (c *Client) RegisterMessageHasBeenUpdated(callback func(c *plugin.Context, newPost, oldPost *model.Post)) {
	c.hooks.registered = append(c.hooks.registered, plugin.MessageHasBeenUpdatedId)
	c.hooks.messageHasBeenUpdated = callback
}

func (c *Client) RegisterChannelHasBeenCreated(callback func(c *plugin.Context, channel *model.Channel)) {
	c.hooks.registered = append(c.hooks.registered, plugin.ChannelHasBeenCreatedId)
	c.hooks.channelHasBeenCreated = callback
}

func (c *Client) RegisterUserHasJoinedChannel(callback func(c *plugin.Context, channelMember *model.ChannelMember, actor *model.User)) {
	c.hooks.registered = append(c.hooks.registered, plugin.UserHasJoinedChannelId)
	c.hooks.userHasJoinedChannel = callback
}

func (c *Client) RegisterUserHasLeftChannel(callback func(c *plugin.Context, channelMember *model.ChannelMember, actor *model.User)) {
	c.hooks.registered = append(c.hooks.registered, plugin.UserHasLeftChannelId)
	c.hooks.userHasLeftChannel = callback
}

func (c *Client) RegisterUserHasJoinedTeam(callback func(c *plugin.Context, teamMember *model.TeamMember, actor *model.User)) {
	c.hooks.registered = append(c.hooks.registered, plugin.UserHasJoinedTeamId)
	c.hooks.userHasJoinedTeam = callback
}

func (c *Client) RegisterUserHasLeftTeam(callback func(c *plugin.Context, teamMember *model.TeamMember, actor *model.User)) {
	c.hooks.registered = append(c.hooks.registered, plugin.UserHasLeftTeamId)
	c.hooks.userHasLeftTeam = callback
}

func (c *Client) RegisterFileWillBeUploaded(callback func(c *plugin.Context, info *model.FileInfo, file io.Reader, output io.Writer) (*model.FileInfo, string)) {
	c.hooks.registered = append(c.hooks.registered, plugin.FileWillBeUploadedId)
	c.hooks.fileWillBeUploaded = callback
}
