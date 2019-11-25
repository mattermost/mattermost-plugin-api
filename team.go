package pluginapi

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

// TeamService exposes functionalities to deal with teams.
type TeamService struct {
	api plugin.API
}

// Create creates a team.
//
// Minimum server version: 5.2
func (t *TeamService) Create(team *model.Team) error {
	createdTeam, appErr := t.api.CreateTeam(team)
	if appErr != nil {
		return normalizeAppErr(appErr)
	}
	*team = *createdTeam
	return nil
}

// Get gets a team.
//
// Minimum server version: 5.2
func (t *TeamService) Get(id string) (*model.Team, error) {
	team, appErr := t.api.GetTeam(id)
	return team, normalizeAppErr(appErr)
}

// GetByName gets a team by its name.
//
// Minimum server version: 5.2
func (t *TeamService) GetByName(name string) (*model.Team, error) {
	team, appErr := t.api.GetTeamByName(name)
	return team, normalizeAppErr(appErr)
}

// Update updates a team.
//
// Minimum server version: 5.2
func (t *TeamService) Update(team *model.Team) error {
	updatedTeam, appErr := t.api.UpdateTeam(team)
	if appErr != nil {
		return normalizeAppErr(appErr)
	}
	*team = *updatedTeam
	return nil
}

// TeamListOption is used to filter team listing.
type TeamListOption func(*ListTeamsOptions)

// ListTeamsOptions holds options about filter out team listing.
type ListTeamsOptions struct {
	UserID string
}

// FilterTeamsByUser option is used to filter teams by user.
func FilterTeamsByUser(userID string) TeamListOption {
	return func(o *ListTeamsOptions) {
		o.UserID = userID
	}
}

// List gets a list of teams by options.
//
// Minimum server version: 5.2
// Minimum server version when LimitTeamsToUser() option is used: 5.6
func (t *TeamService) List(options ...TeamListOption) ([]*model.Team, error) {
	opts := ListTeamsOptions{}
	for _, o := range options {
		o(&opts)
	}
	var (
		teams  []*model.Team
		appErr *model.AppError
	)
	if opts.UserID != "" {
		teams, appErr = t.api.GetTeamsForUser(opts.UserID)
	} else {
		teams, appErr = t.api.GetTeams()
	}
	return teams, normalizeAppErr(appErr)
}

// Search search for teams by term.
//
// Minimum server version: 5.8
func (t *TeamService) Search(term string) ([]*model.Team, error) {
	teams, appErr := t.api.SearchTeams(term)
	return teams, normalizeAppErr(appErr)
}

// Delete deletes a team.
//
// Minimum server version: 5.2
func (t *TeamService) Delete(id string) error {
	appErr := t.api.DeleteTeam(id)
	return normalizeAppErr(appErr)
}

// GetIcon gets the team's icon.
//
// Minimum server version: 5.6
func (t *TeamService) GetIcon(teamID string) (content io.Reader, err error) {
	contentBytes, appErr := t.api.GetTeamIcon(teamID)
	if appErr != nil {
		return nil, normalizeAppErr(appErr)
	}
	return bytes.NewReader(contentBytes), nil
}

// SetIcon sets the team's icon.

// Minimum server version: 5.6
func (t *TeamService) SetIcon(teamID string, content io.Reader) error {
	contentBytes, err := ioutil.ReadAll(content)
	if err != nil {
		return err
	}
	appErr := t.api.SetTeamIcon(teamID, contentBytes)
	return normalizeAppErr(appErr)
}

// DeleteIcon removes the team's icon.
//
// Minimum server version: 5.6
func (t *TeamService) DeleteIcon(teamID string) error {
	appErr := t.api.RemoveTeamIcon(teamID)
	return normalizeAppErr(appErr)
}

// GetUsers gets users of the team.
//
// Minimum server version: 5.6
func (t *TeamService) GetUsers(teamID string, page int, count int) ([]*model.User, error) {
	users, appErr := t.api.GetUsersInTeam(teamID, page, count)
	return users, normalizeAppErr(appErr)
}

// GetUnreads gets the unread message and mention counts for each team to which the given user belongs.
//
// Minimum server version: 5.6
func (t *TeamService) GetUnreads(userID string) ([]*model.TeamUnread, error) {
	undreads, appErr := t.api.GetTeamsUnreadForUser(userID)
	return undreads, normalizeAppErr(appErr)
}

// CreateMember creates a team membership.
//
// Minimum server version: 5.2
func (t *TeamService) CreateMember(teamID, userID string) (*model.TeamMember, error) {
	membership, appErr := t.api.CreateTeamMember(teamID, userID)
	return membership, normalizeAppErr(appErr)
}

// CreateMembers creates a team membership for all provided user ids.
//
// Minimum server version: 5.2
func (t *TeamService) CreateMembers(teamID string, userIDs []string, requestorID string) ([]*model.TeamMember, error) {
	memberships, appErr := t.api.CreateTeamMembers(teamID, userIDs, requestorID)
	return memberships, normalizeAppErr(appErr)
}

// GetMember returns a specific membership.
//
// Minimum server version: 5.2
func (t *TeamService) GetMember(teamID, userID string) (*model.TeamMember, error) {
	member, appErr := t.api.GetTeamMember(teamID, userID)
	return member, normalizeAppErr(appErr)
}

// GetMembers returns the memberships of a specific team.
//
// Minimum server version: 5.2
func (t *TeamService) GetMembers(teamID string, page, count int) ([]*model.TeamMember, error) {
	members, appErr := t.api.GetTeamMembers(teamID, page, count)
	return members, normalizeAppErr(appErr)
}

// GetMemberships returns all team memberships of a user.
//
// Minimum server version: 5.10
func (t *TeamService) GetMemberships(userId string, page int, count int) ([]*model.TeamMember, error) {
	memberships, appErr := t.api.GetTeamMembersForUser(userId, page, count)
	return memberships, normalizeAppErr(appErr)
}

// UpdateMemberRoles updates the role for a team membership.
//
// Minimum server version: 5.2
func (t *TeamService) UpdateMemberRoles(teamID, userID, newRoles string) (*model.TeamMember, error) {
	membership, appErr := t.api.UpdateTeamMemberRoles(teamID, userID, newRoles)
	return membership, normalizeAppErr(appErr)
}

// DeleteMember deletes a team membership.
//
// User Minimum server version: 5.2
func (t *TeamService) DeleteMember(teamID, userID, requestorID string) error {
	appErr := t.api.DeleteTeamMember(teamID, userID, requestorID)
	return normalizeAppErr(appErr)
}

// GetChannelByName gets a channel by its name, given a team name.
//
// Minimum server version: 5.2
func (t *TeamService) GetChannelByName(teamName, channelName string, includeDeleted bool) (*model.Channel, error) {
	channel, appErr := t.api.GetChannelByNameForTeamName(teamName, channelName, includeDeleted)
	return channel, normalizeAppErr(appErr)
}

// GetUserChannels gets a list of channels for given user in given team.
//
// Minimum server version: 5.6
func (t *TeamService) GetUserChannels(teamID, userID string, includeDeleted bool) ([]*model.Channel, error) {
	channels, appErr := t.api.GetChannelsForTeamForUser(teamID, userID, includeDeleted)
	return channels, normalizeAppErr(appErr)
}

// GetPublicChannels gets a list of all channels.
//
// Minimum server version: 5.2
func (t *TeamService) GetPublicChannels(teamID string, page, count int) ([]*model.Channel, error) {
	channels, appErr := t.api.GetPublicChannelsForTeam(teamID, page, count)
	return channels, normalizeAppErr(appErr)
}

// SearchPosts returns a list of posts in a specific team that match the given search options.
//
// Minimum server version: 5.10
func (t *TeamService) SearchPosts(teamID string, options []*model.SearchParams) ([]*model.Post, error) {
	posts, appErr := t.api.SearchPostsInTeam(teamID, options)
	return posts, normalizeAppErr(appErr)
}

// GetStats gets a team's statistics.
//
// Minimum server version: 5.8
func (t *TeamService) GetStats(teamID string) (*model.TeamStats, error) {
	stats, appErr := t.api.GetTeamStats(teamID)
	return stats, normalizeAppErr(appErr)
}

// HasUserPermission checks if a user has the given permission at team's scope.
//
// Minimum server version: 5.3
func (t *TeamService) HasUserPermission(userID, teamID string, permission *model.Permission) bool {
	return t.api.HasPermissionToTeam(userID, teamID, permission)
}
