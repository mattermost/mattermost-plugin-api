package pluginapi

import (
	"io"
	"net/http"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

// pluginImplementation implements the interface currently required by HooksRPCServer.
//
// This will likely be simplified if we relocate the RPC layer and generated code into this
// package instead of having to work within the constraints of that code.
//
// Note that while this plugin implementation "implements" all hooks, we override the Implemented
// hook to only tell the Mattermost server about registered hooks, preserving the performance
// of a plugin that doesn't register a hook.
type pluginImplementation struct {
	c *Client
}

func (p *pluginImplementation) SetAPI(api plugin.API) {
	// Save the API pointer: this package wraps same.
	p.c.api = api
}

func (p *pluginImplementation) SetHelpers(helpers plugin.Helpers) {
	// Ignore the helpers: this package replaces same.This method only exists to satisfy the
	// interface currently required by the generated code in
	// github.com/mattermost-server/plugin. See above for future direction here.
}

func (p *pluginImplementation) OnActivate() error {
	if p.c.hooks.onActivate == nil {
		return nil
	}

	return p.c.hooks.onActivate()
}

func (p *pluginImplementation) OnDeactivate() error {
	if p.c.hooks.onDeactivate == nil {
		return nil
	}

	return p.c.hooks.onDeactivate()
}

func (p *pluginImplementation) OnConfigurationChange() error {
	if p.c.hooks.onConfigurationChange == nil {
		return nil
	}

	return p.c.hooks.onConfigurationChange()
}

func (p *pluginImplementation) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	if p.c.hooks.serveHTTP == nil {
		return
	}

	p.c.hooks.serveHTTP(c, w, r)
}

func (p *pluginImplementation) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	if p.c.hooks.executeCommand == nil {
		return nil, nil
	}

	commandResponse, err := p.c.hooks.executeCommand(c, args)
	if err != nil {
		return commandResponse, model.NewAppError("ExecuteCommand", "pluginImplementation.ExecuteCommand", nil, err.Error(), http.StatusInternalServerError)
	}

	return commandResponse, nil
}

func (p *pluginImplementation) UserHasBeenCreated(c *plugin.Context, user *model.User) {
	if p.c.hooks.userHasBeenCreated == nil {
		return
	}

	p.c.hooks.userHasBeenCreated(c, user)
}

func (p *pluginImplementation) UserWillLogIn(c *plugin.Context, user *model.User) string {
	if p.c.hooks.userWillLogIn == nil {
		return ""
	}

	return p.c.hooks.userWillLogIn(c, user)
}

func (p *pluginImplementation) UserHasLoggedIn(c *plugin.Context, user *model.User) {
	if p.c.hooks.userHasLoggedIn == nil {
		return
	}

	p.c.hooks.userHasLoggedIn(c, user)
}

func (p *pluginImplementation) MessageWillBePosted(c *plugin.Context, post *model.Post) (*model.Post, string) {
	if p.c.hooks.messageWillBePosted == nil {
		return nil, ""
	}

	return p.c.hooks.messageWillBePosted(c, post)
}

func (p *pluginImplementation) MessageWillBeUpdated(c *plugin.Context, newPost, oldPost *model.Post) (*model.Post, string) {
	if p.c.hooks.messageWillBeUpdated == nil {
		return nil, ""
	}

	return p.c.hooks.messageWillBeUpdated(c, newPost, oldPost)
}

func (p *pluginImplementation) MessageHasBeenPosted(c *plugin.Context, post *model.Post) {
	if p.c.hooks.messageHasBeenPosted == nil {
		return
	}

	p.c.hooks.messageHasBeenPosted(c, post)
}

func (p *pluginImplementation) MessageHasBeenUpdated(c *plugin.Context, newPost, oldPost *model.Post) {
	if p.c.hooks.messageHasBeenUpdated == nil {
		return
	}

	p.c.hooks.messageHasBeenUpdated(c, newPost, oldPost)
}

func (p *pluginImplementation) ChannelHasBeenCreated(c *plugin.Context, channel *model.Channel) {
	if p.c.hooks.channelHasBeenCreated == nil {
		return
	}

	p.c.hooks.channelHasBeenCreated(c, channel)
}

func (p *pluginImplementation) UserHasJoinedChannel(c *plugin.Context, channelMember *model.ChannelMember, actor *model.User) {
	if p.c.hooks.userHasJoinedChannel == nil {
		return
	}

	p.c.hooks.userHasJoinedChannel(c, channelMember, actor)
}

func (p *pluginImplementation) UserHasLeftChannel(c *plugin.Context, channelMember *model.ChannelMember, actor *model.User) {
	if p.c.hooks.userHasLeftChannel == nil {
		return
	}

	p.c.hooks.userHasLeftChannel(c, channelMember, actor)
}

func (p *pluginImplementation) UserHasJoinedTeam(c *plugin.Context, teamMember *model.TeamMember, actor *model.User) {
	if p.c.hooks.userHasJoinedTeam == nil {
		return
	}

	p.c.hooks.userHasJoinedTeam(c, teamMember, actor)
}

func (p *pluginImplementation) UserHasLeftTeam(c *plugin.Context, teamMember *model.TeamMember, actor *model.User) {
	if p.c.hooks.userHasLeftTeam == nil {
		return
	}

	p.c.hooks.userHasLeftTeam(c, teamMember, actor)
}

func (p *pluginImplementation) FileWillBeUploaded(c *plugin.Context, info *model.FileInfo, file io.Reader, output io.Writer) (*model.FileInfo, string) {
	if p.c.hooks.fileWillBeUploaded == nil {
		return nil, ""
	}

	return p.c.hooks.fileWillBeUploaded(c, info, file, output)
}
