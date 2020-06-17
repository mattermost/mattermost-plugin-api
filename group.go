package pluginapi

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

// GroupService exposes methods to manipulate groups.
type GroupService struct {
	api plugin.API
}

// Get gets a group by ID.
//
// Minimum server version: 5.18
func (g *GroupService) Get(groupID string) (*model.Group, error) {
	group, appErr := g.api.GetGroup(groupID)

	return group, normalizeAppErr(appErr)
}

// GetByName gets a group by name.
// Because of the change to 5.24, opts is a varargs. If the caller provides an opts struct,
// only the first will be used.
//
// Minimum server version: 5.24
func (g *GroupService) GetByName(name string, opts ...model.GroupSearchOpts) (*model.Group, error) {
	options := model.GroupSearchOpts{}
	if len(opts) > 0 {
		options = opts[0]
	}
	group, appErr := g.api.GetGroupByName(name, options)

	return group, normalizeAppErr(appErr)
}

// ListForUser gets the groups a user is in.
//
// Minimum server version: 5.18
func (g *GroupService) ListForUser(userID string) ([]*model.Group, error) {
	groups, appErr := g.api.GetGroupsForUser(userID)

	return groups, normalizeAppErr(appErr)
}
