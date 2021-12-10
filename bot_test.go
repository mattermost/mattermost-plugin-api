package pluginapi

import (
	"io/ioutil"
	"testing"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"
	"github.com/mattermost/mattermost-server/v6/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateBot(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)
		client := NewClient(api, &plugintest.Driver{})

		api.On("CreateBot", &model.Bot{Username: "1"}).Return(&model.Bot{Username: "1", UserId: "2"}, nil)

		bot := &model.Bot{Username: "1"}
		err := client.Bot.Create(bot)
		require.NoError(t, err)
		require.Equal(t, &model.Bot{Username: "1", UserId: "2"}, bot)
	})

	t.Run("failure", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)
		client := NewClient(api, &plugintest.Driver{})

		appErr := newAppError()

		api.On("CreateBot", &model.Bot{Username: "1"}).Return(nil, appErr)

		bot := &model.Bot{Username: "1"}
		err := client.Bot.Create(&model.Bot{Username: "1"})
		require.Equal(t, appErr, err)
		require.Equal(t, &model.Bot{Username: "1"}, bot)
	})
}

func TestUpdateBotStatus(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)
		client := NewClient(api, &plugintest.Driver{})

		api.On("UpdateBotActive", "1", true).Return(&model.Bot{UserId: "2"}, nil)

		bot, err := client.Bot.UpdateActive("1", true)
		require.NoError(t, err)
		require.Equal(t, &model.Bot{UserId: "2"}, bot)
	})

	t.Run("failure", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)
		client := NewClient(api, &plugintest.Driver{})

		appErr := newAppError()

		api.On("UpdateBotActive", "1", true).Return(nil, appErr)

		bot, err := client.Bot.UpdateActive("1", true)
		require.Equal(t, appErr, err)
		require.Zero(t, bot)
	})
}

func TestGetBot(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)
		client := NewClient(api, &plugintest.Driver{})

		api.On("GetBot", "1", true).Return(&model.Bot{UserId: "2"}, nil)

		bot, err := client.Bot.Get("1", true)
		require.NoError(t, err)
		require.Equal(t, &model.Bot{UserId: "2"}, bot)
	})

	t.Run("failure", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)
		client := NewClient(api, &plugintest.Driver{})

		appErr := newAppError()

		api.On("GetBot", "1", true).Return(nil, appErr)

		bot, err := client.Bot.Get("1", true)
		require.Equal(t, appErr, err)
		require.Zero(t, bot)
	})
}

func TestListBot(t *testing.T) {
	tests := []struct {
		name            string
		page, count     int
		options         []BotListOption
		expectedOptions *model.BotGetOptions
		bots            []*model.Bot
		err             error
	}{
		{
			"owner filter",
			1,
			2,
			[]BotListOption{
				BotOwner("3"),
			},
			&model.BotGetOptions{
				Page:    1,
				PerPage: 2,
				OwnerId: "3",
			},
			[]*model.Bot{
				{UserId: "4"},
				{UserId: "5"},
			},
			nil,
		},
		{
			"all filter",
			1,
			2,
			[]BotListOption{
				BotOwner("3"),
				BotIncludeDeleted(),
				BotOnlyOrphans(),
			},
			&model.BotGetOptions{
				Page:           1,
				PerPage:        2,
				OwnerId:        "3",
				IncludeDeleted: true,
				OnlyOrphaned:   true,
			},
			[]*model.Bot{
				{UserId: "4"},
			},
			nil,
		},
		{
			"no filter",
			1,
			2,
			[]BotListOption{},
			&model.BotGetOptions{
				Page:    1,
				PerPage: 2,
			},
			[]*model.Bot{
				{UserId: "4"},
			},
			nil,
		},
		{
			"app error",
			1,
			2,
			[]BotListOption{
				BotOwner("3"),
			},
			&model.BotGetOptions{
				Page:    1,
				PerPage: 2,
				OwnerId: "3",
			},
			nil,
			newAppError(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			api := &plugintest.API{}
			client := NewClient(api, &plugintest.Driver{})

			api.On("GetBots", test.expectedOptions).Return(test.bots, test.err)

			bots, err := client.Bot.List(test.page, test.count, test.options...)
			if test.err != nil {
				require.Equal(t, test.err.Error(), err.Error(), test.name)
			} else {
				require.NoError(t, err, test.name)
			}
			require.Equal(t, test.bots, bots, test.name)

			api.AssertExpectations(t)
		})
	}
}

func TestDeleteBotPermanently(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)
		client := NewClient(api, &plugintest.Driver{})

		api.On("PermanentDeleteBot", "1").Return(nil)

		err := client.Bot.DeletePermanently("1")
		require.NoError(t, err)
	})

	t.Run("failure", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)
		client := NewClient(api, &plugintest.Driver{})

		appErr := newAppError()

		api.On("PermanentDeleteBot", "1").Return(appErr)

		err := client.Bot.DeletePermanently("1")
		require.Equal(t, appErr, err)
	})
}

func TestEnsureBot(t *testing.T) {
	testbot := &model.Bot{
		Username:    "testbot",
		DisplayName: "Test Bot",
		Description: "testbotdescription",
	}

	m := testMutex{}

	t.Run("server version incompatible", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)
		client := NewClient(api, &plugintest.Driver{})

		api.On("GetServerVersion").Return("5.9.0")

		_, err := client.Bot.ensureBot(m, nil)
		require.Error(t, err)
		assert.Equal(t,
			"failed to ensure bot: incompatible server version for plugin, minimum required version: 5.10.0, current version: 5.9.0",
			err.Error(),
		)
	})

	t.Run("bad parameters", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)
		client := NewClient(api, &plugintest.Driver{})

		api.On("GetServerVersion").Return("5.10.0")

		t.Run("no bot", func(t *testing.T) {
			botID, err := client.Bot.ensureBot(m, nil)
			require.Error(t, err)
			assert.Equal(t, "", botID)
		})

		t.Run("bad username", func(t *testing.T) {
			botID, err := client.Bot.ensureBot(m, &model.Bot{
				Username: "",
			})
			require.Error(t, err)
			assert.Equal(t, "", botID)
		})
	})

	t.Run("if bot already exists", func(t *testing.T) {
		t.Run("should find and return the existing bot ID", func(t *testing.T) {
			api := &plugintest.API{}
			defer api.AssertExpectations(t)
			client := NewClient(api, &plugintest.Driver{})

			expectedBotID := model.NewId()

			api.On("GetServerVersion").Return("5.10.0")
			api.On("KVGet", plugin.BotUserKey).Return([]byte(expectedBotID), nil)
			api.On("PatchBot", expectedBotID, &model.BotPatch{
				Username:    &testbot.Username,
				DisplayName: &testbot.DisplayName,
				Description: &testbot.Description,
			}).Return(nil, nil)

			botID, err := client.Bot.ensureBot(m, testbot)

			require.NoError(t, err)
			assert.Equal(t, expectedBotID, botID)
		})

		t.Run("should return an error if unable to get bot", func(t *testing.T) {
			api := &plugintest.API{}
			defer api.AssertExpectations(t)
			client := NewClient(api, &plugintest.Driver{})

			api.On("GetServerVersion").Return("5.10.0")
			api.On("KVGet", plugin.BotUserKey).Return(nil, &model.AppError{})

			botID, err := client.Bot.ensureBot(m, testbot)

			require.Error(t, err)
			assert.Equal(t, "", botID)
		})

		t.Run("should set the bot profile image when specified", func(t *testing.T) {
			api := &plugintest.API{}
			defer api.AssertExpectations(t)
			client := NewClient(api, &plugintest.Driver{})

			expectedBotID := model.NewId()

			profileImageFile, err := ioutil.TempFile("", "profile_image")
			require.NoError(t, err)

			profileImageBytes := []byte("profile image")
			err = ioutil.WriteFile(profileImageFile.Name(), profileImageBytes, 0600)
			require.NoError(t, err)

			api.On("KVGet", plugin.BotUserKey).Return([]byte(expectedBotID), nil)
			api.On("GetBundlePath").Return("", nil)
			api.On("SetProfileImage", expectedBotID, profileImageBytes).Return(nil)
			api.On("GetServerVersion").Return("5.10.0")
			api.On("PatchBot", expectedBotID, &model.BotPatch{
				Username:    &testbot.Username,
				DisplayName: &testbot.DisplayName,
				Description: &testbot.Description,
			}).Return(nil, nil)

			botID, err := client.Bot.ensureBot(m, testbot, ProfileImagePath(profileImageFile.Name()))
			require.NoError(t, err)
			assert.Equal(t, expectedBotID, botID)
		})

		t.Run("should find and update the bot with new bot details", func(t *testing.T) {
			api := &plugintest.API{}
			defer api.AssertExpectations(t)
			client := NewClient(api, &plugintest.Driver{})

			expectedBotID := model.NewId()
			expectedBotUsername := "updated_testbot"
			expectedBotDisplayName := "Updated Test Bot"
			expectedBotDescription := "updated testbotdescription"

			profileImageFile, err := ioutil.TempFile("", "profile_image")
			require.NoError(t, err)

			profileImageBytes := []byte("profile image")
			err = ioutil.WriteFile(profileImageFile.Name(), profileImageBytes, 0600)
			require.NoError(t, err)

			iconImageFile, err := ioutil.TempFile("", "profile_image")
			require.NoError(t, err)

			iconImageBytes := []byte("icon image")
			err = ioutil.WriteFile(iconImageFile.Name(), iconImageBytes, 0600)
			require.NoError(t, err)

			api.On("GetServerVersion").Return("5.10.0")
			api.On("KVGet", plugin.BotUserKey).Return([]byte(expectedBotID), nil)
			api.On("GetBundlePath").Return("", nil)
			api.On("SetProfileImage", expectedBotID, profileImageBytes).Return(nil)
			api.On("PatchBot", expectedBotID, &model.BotPatch{
				Username:    &expectedBotUsername,
				DisplayName: &expectedBotDisplayName,
				Description: &expectedBotDescription,
			}).Return(nil, nil)

			updatedTestBot := &model.Bot{
				Username:    expectedBotUsername,
				DisplayName: expectedBotDisplayName,
				Description: expectedBotDescription,
			}
			botID, err := client.Bot.ensureBot(m,
				updatedTestBot,
				ProfileImagePath(profileImageFile.Name()),
			)
			require.NoError(t, err)
			assert.Equal(t, expectedBotID, botID)
		})
	})

	t.Run("if bot doesn't exist", func(t *testing.T) {
		t.Run("should create the bot and return the ID", func(t *testing.T) {
			api := &plugintest.API{}
			defer api.AssertExpectations(t)
			client := NewClient(api, &plugintest.Driver{})

			expectedBotID := model.NewId()

			api.On("GetServerVersion").Return("5.10.0")
			api.On("KVGet", plugin.BotUserKey).Return(nil, nil)
			api.On("GetUserByUsername", testbot.Username).Return(nil, nil)
			api.On("CreateBot", testbot).Return(&model.Bot{
				UserId: expectedBotID,
			}, nil)
			api.On("KVSet", plugin.BotUserKey, []byte(expectedBotID)).Return(nil)

			botID, err := client.Bot.ensureBot(m, testbot)
			require.NoError(t, err)
			assert.Equal(t, expectedBotID, botID)
		})

		t.Run("should claim existing bot and return the ID", func(t *testing.T) {
			api := &plugintest.API{}
			defer api.AssertExpectations(t)
			client := NewClient(api, &plugintest.Driver{})

			expectedBotID := model.NewId()

			api.On("GetServerVersion").Return("5.10.0")
			api.On("KVGet", plugin.BotUserKey).Return(nil, nil)
			api.On("GetUserByUsername", testbot.Username).Return(&model.User{
				Id:    expectedBotID,
				IsBot: true,
			}, nil)
			api.On("KVSet", plugin.BotUserKey, []byte(expectedBotID)).Return(nil)

			botID, err := client.Bot.ensureBot(m, testbot)
			require.NoError(t, err)
			assert.Equal(t, expectedBotID, botID)
		})

		t.Run("should return the non-bot account but log a message if user exists with the same name and is not a bot", func(t *testing.T) {
			api := &plugintest.API{}
			defer api.AssertExpectations(t)
			client := NewClient(api, &plugintest.Driver{})

			expectedBotID := model.NewId()

			api.On("GetServerVersion").Return("5.10.0")
			api.On("KVGet", plugin.BotUserKey).Return(nil, nil)
			api.On("GetUserByUsername", testbot.Username).Return(&model.User{
				Id:    expectedBotID,
				IsBot: false,
			}, nil)
			api.On("LogError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

			botID, err := client.Bot.ensureBot(m, testbot)
			require.NoError(t, err)
			assert.Equal(t, expectedBotID, botID)
		})

		t.Run("should fail if create bot fails", func(t *testing.T) {
			api := &plugintest.API{}
			defer api.AssertExpectations(t)
			client := NewClient(api, &plugintest.Driver{})

			api.On("GetServerVersion").Return("5.10.0")
			api.On("KVGet", plugin.BotUserKey).Return(nil, nil)
			api.On("GetUserByUsername", testbot.Username).Return(nil, nil)
			api.On("CreateBot", testbot).Return(nil, &model.AppError{})

			botID, err := client.Bot.ensureBot(m, testbot)
			require.Error(t, err)
			assert.Equal(t, "", botID)
		})

		t.Run("should create bot and set the bot profile image when specified", func(t *testing.T) {
			api := &plugintest.API{}
			defer api.AssertExpectations(t)
			client := NewClient(api, &plugintest.Driver{})

			expectedBotID := model.NewId()

			profileImageFile, err := ioutil.TempFile("", "profile_image")
			require.NoError(t, err)

			profileImageBytes := []byte("profile image")
			err = ioutil.WriteFile(profileImageFile.Name(), profileImageBytes, 0600)
			require.NoError(t, err)

			api.On("KVGet", plugin.BotUserKey).Return(nil, nil)
			api.On("GetUserByUsername", testbot.Username).Return(nil, nil)
			api.On("CreateBot", testbot).Return(&model.Bot{
				UserId: expectedBotID,
			}, nil)
			api.On("KVSet", plugin.BotUserKey, []byte(expectedBotID)).Return(nil)
			api.On("GetBundlePath").Return("", nil)
			api.On("SetProfileImage", expectedBotID, profileImageBytes).Return(nil)
			api.On("GetServerVersion").Return("5.10.0")

			botID, err := client.Bot.ensureBot(m, testbot, ProfileImagePath(profileImageFile.Name()))
			require.NoError(t, err)
			assert.Equal(t, expectedBotID, botID)
		})
	})
}
