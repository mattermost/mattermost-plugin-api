package pluginapi

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"
	"github.com/stretchr/testify/require"
)

func TestCreateTeam(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		api.On("CreateTeam", &model.Team{Name: "1"}).Return(&model.Team{Name: "1", Id: "2"}, nil)

		team := &model.Team{Name: "1"}
		err := client.Team.Create(team)
		require.NoError(t, err)
		require.Equal(t, &model.Team{Name: "1", Id: "2"}, team)
	})

	t.Run("failure", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		appErr := model.NewAppError("here", "id", nil, "an error occurred", http.StatusInternalServerError)

		api.On("CreateTeam", &model.Team{Name: "1"}).Return(nil, appErr)

		team := &model.Team{Name: "1"}
		err := client.Team.Create(team)
		require.Equal(t, appErr, err)
		require.Equal(t, &model.Team{Name: "1"}, team)
	})
}

func TestGetTeam(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		api.On("GetTeam", "1").Return(&model.Team{Id: "2"}, nil)

		team, err := client.Team.Get("1")
		require.NoError(t, err)
		require.Equal(t, &model.Team{Id: "2"}, team)
	})

	t.Run("failure", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		appErr := model.NewAppError("here", "id", nil, "an error occurred", http.StatusInternalServerError)

		api.On("GetTeam", "1").Return(nil, appErr)

		team, err := client.Team.Get("1")
		require.Equal(t, appErr, err)
		require.Zero(t, team)
	})
}

func TestGetTeamByName(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		api.On("GetTeamByName", "1").Return(&model.Team{Id: "2"}, nil)

		team, err := client.Team.GetByName("1")
		require.NoError(t, err)
		require.Equal(t, &model.Team{Id: "2"}, team)
	})

	t.Run("failure", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		appErr := model.NewAppError("here", "id", nil, "an error occurred", http.StatusInternalServerError)

		api.On("GetTeamByName", "1").Return(nil, appErr)

		team, err := client.Team.GetByName("1")
		require.Equal(t, appErr, err)
		require.Zero(t, team)
	})
}

func TestUpdateTeam(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		api.On("UpdateTeam", &model.Team{Name: "1"}).Return(&model.Team{Name: "1", Id: "2"}, nil)

		team := &model.Team{Name: "1"}
		err := client.Team.Update(team)
		require.NoError(t, err)
		require.Equal(t, &model.Team{Name: "1", Id: "2"}, team)
	})

	t.Run("failure", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		appErr := model.NewAppError("here", "id", nil, "an error occurred", http.StatusInternalServerError)

		api.On("UpdateTeam", &model.Team{Name: "1"}).Return(nil, appErr)

		team := &model.Team{Name: "1"}
		err := client.Team.Update(team)
		require.Equal(t, appErr, err)
		require.Equal(t, &model.Team{Name: "1"}, team)
	})
}

func TestListTeams(t *testing.T) {
	t.Run("list all", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		api.On("GetTeams").Return([]*model.Team{{Id: "1"}, {Id: "2"}}, nil)

		teams, err := client.Team.List()
		require.NoError(t, err)
		require.Equal(t, []*model.Team{{Id: "1"}, {Id: "2"}}, teams)
	})

	t.Run("list scoped to user", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		api.On("GetTeamsForUser", "3").Return([]*model.Team{{Id: "1"}, {Id: "2"}}, nil)

		teams, err := client.Team.List(FilterTeamsByUser("3"))
		require.NoError(t, err)
		require.Equal(t, []*model.Team{{Id: "1"}, {Id: "2"}}, teams)
	})

	t.Run("failure", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		appErr := model.NewAppError("here", "id", nil, "an error occurred", http.StatusInternalServerError)

		api.On("GetTeams").Return(nil, appErr)

		teams, err := client.Team.List()
		require.Equal(t, appErr, err)
		require.Len(t, teams, 0)
	})
}

func TestSearchTeams(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		api.On("SearchTeams", "1").Return([]*model.Team{{Id: "1"}, {Id: "2"}}, nil)

		teams, err := client.Team.Search("1")
		require.NoError(t, err)
		require.Equal(t, []*model.Team{{Id: "1"}, {Id: "2"}}, teams)
	})

	t.Run("failure", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		appErr := model.NewAppError("here", "id", nil, "an error occurred", http.StatusInternalServerError)

		api.On("SearchTeams", "1").Return(nil, appErr)

		teams, err := client.Team.Search("1")
		require.Equal(t, appErr, err)
		require.Zero(t, teams)
	})
}

func TestDeleteTeam(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		api.On("DeleteTeam", "1").Return(nil)

		err := client.Team.Delete("1")
		require.NoError(t, err)
	})

	t.Run("failure", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		appErr := model.NewAppError("here", "id", nil, "an error occurred", http.StatusInternalServerError)

		api.On("DeleteTeam", "1").Return(appErr)

		err := client.Team.Delete("1")
		require.Equal(t, appErr, err)
	})
}

func TestGetTeamIcon(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		api.On("GetTeamIcon", "1").Return([]byte{2}, nil)

		content, err := client.Team.GetIcon("1")
		require.NoError(t, err)
		contentBytes, err := ioutil.ReadAll(content)
		require.NoError(t, err)
		require.Equal(t, []byte{2}, contentBytes)
	})

	t.Run("failure", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		appErr := model.NewAppError("here", "id", nil, "an error occurred", http.StatusInternalServerError)

		api.On("GetTeamIcon", "1").Return(nil, appErr)

		content, err := client.Team.GetIcon("1")
		require.Equal(t, appErr, err)
		require.Zero(t, content)
	})
}

func TestSetIcon(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		api.On("SetTeamIcon", "1", []byte{2}).Return(nil)

		err := client.Team.SetIcon("1", bytes.NewReader([]byte{2}))
		require.NoError(t, err)
	})

	t.Run("failure", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		appErr := model.NewAppError("here", "id", nil, "an error occurred", http.StatusInternalServerError)

		api.On("SetTeamIcon", "1", []byte{2}).Return(appErr)

		err := client.Team.SetIcon("1", bytes.NewReader([]byte{2}))
		require.Equal(t, appErr, err)
	})
}

func TestDeleteIcon(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		api.On("RemoveTeamIcon", "1").Return(nil)

		err := client.Team.DeleteIcon("1")
		require.NoError(t, err)
	})

	t.Run("failure", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		appErr := model.NewAppError("here", "id", nil, "an error occurred", http.StatusInternalServerError)

		api.On("RemoveTeamIcon", "1").Return(appErr)

		err := client.Team.DeleteIcon("1")
		require.Equal(t, appErr, err)
	})
}

func TestGetTeamUsers(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		api.On("GetUsersInTeam", "1", 2, 3).Return([]*model.User{{Id: "1"}, {Id: "2"}}, nil)

		users, err := client.Team.GetUsers("1", 2, 3)
		require.NoError(t, err)
		require.Equal(t, []*model.User{{Id: "1"}, {Id: "2"}}, users)
	})

	t.Run("failure", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		appErr := model.NewAppError("here", "id", nil, "an error occurred", http.StatusInternalServerError)

		api.On("GetUsersInTeam", "1", 2, 3).Return(nil, appErr)

		users, err := client.Team.GetUsers("1", 2, 3)
		require.Equal(t, appErr, err)
		require.Len(t, users, 0)
	})
}

func TestGetTeamUnreads(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		api.On("GetTeamsUnreadForUser", "1").Return([]*model.TeamUnread{{TeamId: "1"}, {TeamId: "2"}}, nil)

		unreads, err := client.Team.GetUnreads("1")
		require.NoError(t, err)
		require.Equal(t, []*model.TeamUnread{{TeamId: "1"}, {TeamId: "2"}}, unreads)
	})

	t.Run("failure", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		appErr := model.NewAppError("here", "id", nil, "an error occurred", http.StatusInternalServerError)

		api.On("GetTeamsUnreadForUser", "1").Return(nil, appErr)

		unreads, err := client.Team.GetUnreads("1")
		require.Equal(t, appErr, err)
		require.Len(t, unreads, 0)
	})
}

func TestCreateTeamMember(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		api.On("CreateTeamMember", "1", "2").Return(&model.TeamMember{TeamId: "3"}, nil)

		member, err := client.Team.CreateMember("1", "2")
		require.NoError(t, err)
		require.Equal(t, &model.TeamMember{TeamId: "3"}, member)
	})

	t.Run("failure", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		appErr := model.NewAppError("here", "id", nil, "an error occurred", http.StatusInternalServerError)

		api.On("CreateTeamMember", "1", "2").Return(nil, appErr)

		member, err := client.Team.CreateMember("1", "2")
		require.Equal(t, appErr, err)
		require.Zero(t, member)
	})
}

func TestCreateTeamMembers(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		api.On("CreateTeamMembers", "1", []string{"2"}, "3").Return([]*model.TeamMember{{TeamId: "4"}, {TeamId: "5"}}, nil)

		members, err := client.Team.CreateMembers("1", []string{"2"}, "3")
		require.NoError(t, err)
		require.Equal(t, []*model.TeamMember{{TeamId: "4"}, {TeamId: "5"}}, members)
	})

	t.Run("failure", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		appErr := model.NewAppError("here", "id", nil, "an error occurred", http.StatusInternalServerError)

		api.On("CreateTeamMembers", "1", []string{"2"}, "3").Return(nil, appErr)

		members, err := client.Team.CreateMembers("1", []string{"2"}, "3")
		require.Equal(t, appErr, err)
		require.Len(t, members, 0)
	})
}

func TestGetTeamMember(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		api.On("GetTeamMember", "1", "2").Return(&model.TeamMember{TeamId: "3"}, nil)

		member, err := client.Team.GetMember("1", "2")
		require.NoError(t, err)
		require.Equal(t, &model.TeamMember{TeamId: "3"}, member)
	})

	t.Run("failure", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		appErr := model.NewAppError("here", "id", nil, "an error occurred", http.StatusInternalServerError)

		api.On("GetTeamMember", "1", "2").Return(nil, appErr)

		member, err := client.Team.GetMember("1", "2")
		require.Equal(t, appErr, err)
		require.Zero(t, member)
	})
}

func TestGetTeamMembers(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		api.On("GetTeamMembers", "1", 2, 3).Return([]*model.TeamMember{{TeamId: "4"}, {TeamId: "5"}}, nil)

		members, err := client.Team.GetMembers("1", 2, 3)
		require.NoError(t, err)
		require.Equal(t, []*model.TeamMember{{TeamId: "4"}, {TeamId: "5"}}, members)
	})

	t.Run("failure", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		appErr := model.NewAppError("here", "id", nil, "an error occurred", http.StatusInternalServerError)

		api.On("GetTeamMembers", "1", 2, 3).Return(nil, appErr)

		members, err := client.Team.GetMembers("1", 2, 3)
		require.Equal(t, appErr, err)
		require.Len(t, members, 0)
	})
}

func TestGetUserMemberships(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		api.On("GetTeamMembersForUser", "1", 2, 3).Return([]*model.TeamMember{{TeamId: "4"}, {TeamId: "5"}}, nil)

		members, err := client.Team.GetMemberships("1", 2, 3)
		require.NoError(t, err)
		require.Equal(t, []*model.TeamMember{{TeamId: "4"}, {TeamId: "5"}}, members)
	})

	t.Run("failure", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		appErr := model.NewAppError("here", "id", nil, "an error occurred", http.StatusInternalServerError)

		api.On("GetTeamMembersForUser", "1", 2, 3).Return(nil, appErr)

		members, err := client.Team.GetMemberships("1", 2, 3)
		require.Equal(t, appErr, err)
		require.Len(t, members, 0)
	})
}

func TestUpdateTeamMemberRoles(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		api.On("UpdateTeamMemberRoles", "1", "2", "3").Return(&model.TeamMember{TeamId: "3"}, nil)

		membership, err := client.Team.UpdateMemberRoles("1", "2", "3")
		require.NoError(t, err)
		require.Equal(t, &model.TeamMember{TeamId: "3"}, membership)
	})

	t.Run("failure", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		appErr := model.NewAppError("here", "id", nil, "an error occurred", http.StatusInternalServerError)

		api.On("UpdateTeamMemberRoles", "1", "2", "3").Return(nil, appErr)

		membership, err := client.Team.UpdateMemberRoles("1", "2", "3")
		require.Equal(t, appErr, err)
		require.Zero(t, membership)
	})
}

func TestDeleteTeamMember(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		api.On("DeleteTeamMember", "1", "2", "3").Return(nil)

		err := client.Team.DeleteMember("1", "2", "3")
		require.NoError(t, err)
	})

	t.Run("failure", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		appErr := model.NewAppError("here", "id", nil, "an error occurred", http.StatusInternalServerError)

		api.On("DeleteTeamMember", "1", "2", "3").Return(appErr)

		err := client.Team.DeleteMember("1", "2", "3")
		require.Equal(t, appErr, err)
	})
}

func TestGetTeamChannelByName(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		api.On("GetChannelByNameForTeamName", "1", "2", true).Return(&model.Channel{TeamId: "3"}, nil)

		channel, err := client.Team.GetChannelByName("1", "2", true)
		require.NoError(t, err)
		require.Equal(t, &model.Channel{TeamId: "3"}, channel)
	})

	t.Run("failure", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		appErr := model.NewAppError("here", "id", nil, "an error occurred", http.StatusInternalServerError)

		api.On("GetChannelByNameForTeamName", "1", "2", true).Return(nil, appErr)

		channel, err := client.Team.GetChannelByName("1", "2", true)
		require.Equal(t, appErr, err)
		require.Zero(t, channel)
	})
}

func TestGetTeamUserChannels(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		api.On("GetChannelsForTeamForUser", "1", "2", true).Return([]*model.Channel{{TeamId: "3"}, {TeamId: "4"}}, nil)

		channels, err := client.Team.GetUserChannels("1", "2", true)
		require.NoError(t, err)
		require.Equal(t, []*model.Channel{{TeamId: "3"}, {TeamId: "4"}}, channels)
	})

	t.Run("failure", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		appErr := model.NewAppError("here", "id", nil, "an error occurred", http.StatusInternalServerError)

		api.On("GetChannelsForTeamForUser", "1", "2", true).Return(nil, appErr)

		channels, err := client.Team.GetUserChannels("1", "2", true)
		require.Equal(t, appErr, err)
		require.Len(t, channels, 0)
	})
}

func TestGetPublicTeamChannels(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		api.On("GetPublicChannelsForTeam", "1", 2, 3).Return([]*model.Channel{{TeamId: "3"}, {TeamId: "4"}}, nil)

		channels, err := client.Team.GetPublicChannels("1", 2, 3)
		require.NoError(t, err)
		require.Equal(t, []*model.Channel{{TeamId: "3"}, {TeamId: "4"}}, channels)
	})

	t.Run("failure", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		appErr := model.NewAppError("here", "id", nil, "an error occurred", http.StatusInternalServerError)

		api.On("GetPublicChannelsForTeam", "1", 2, 3).Return(nil, appErr)

		channels, err := client.Team.GetPublicChannels("1", 2, 3)
		require.Equal(t, appErr, err)
		require.Len(t, channels, 0)
	})
}

func TestSearchTeamPosts(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		api.
			On("SearchPostsInTeam", "1", []*model.SearchParams{{Terms: "2"}, {Terms: "3"}}).
			Return([]*model.Post{{Id: "3"}, {Id: "4"}}, nil)

		posts, err := client.Team.SearchPosts("1", []*model.SearchParams{{Terms: "2"}, {Terms: "3"}})
		require.NoError(t, err)
		require.Equal(t, []*model.Post{{Id: "3"}, {Id: "4"}}, posts)
	})

	t.Run("failure", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		appErr := model.NewAppError("here", "id", nil, "an error occurred", http.StatusInternalServerError)

		api.
			On("SearchPostsInTeam", "1", []*model.SearchParams{{Terms: "2"}, {Terms: "3"}}).
			Return(nil, appErr)

		posts, err := client.Team.SearchPosts("1", []*model.SearchParams{{Terms: "2"}, {Terms: "3"}})
		require.Equal(t, appErr, err)
		require.Len(t, posts, 0)
	})
}

func TestGetTeamStats(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		api.On("GetTeamStats", "1").Return(&model.TeamStats{TeamId: "3"}, nil)

		stats, err := client.Team.GetStats("1")
		require.NoError(t, err)
		require.Equal(t, &model.TeamStats{TeamId: "3"}, stats)
	})

	t.Run("failure", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		client := NewClient(api)

		appErr := model.NewAppError("here", "id", nil, "an error occurred", http.StatusInternalServerError)

		api.On("GetTeamStats", "1").Return(nil, appErr)

		stats, err := client.Team.GetStats("1")
		require.Equal(t, appErr, err)
		require.Zero(t, stats)
	})
}

func TestHasTeamUserPermission(t *testing.T) {
	api := &plugintest.API{}
	defer api.AssertExpectations(t)

	client := NewClient(api)

	api.On("HasPermissionToTeam", "1", "2", &model.Permission{Id: "3"}).Return(true)

	ok := client.Team.HasUserPermission("1", "2", &model.Permission{Id: "3"})
	require.True(t, ok)
}
