package poster

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"
	"github.com/stretchr/testify/assert"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-api/experimental/common/mock_api"
)

const (
	botID       = "test-bot-user"
	userID      = "test-user-1"
	dmChannelID = "dm-channel-id"
)

func TestInterface(t *testing.T) {
	t.Run("Plugin API satisfy the interface", func(t *testing.T) {
		api := &plugintest.API{}
		client := pluginapi.NewClient(api)
		_ = NewPoster(&client.Post, &client.Channel, botID)
	})
}

func TestDM(t *testing.T) {
	format := "test format, string: %s int: %d value: %v"
	args := []interface{}{"some string", 5, 8.423}
	expectedMessage := "test format, string: some string int: 5 value: 8.423"

	expectedPostID := "expected-post-id"

	post := &model.Post{
		UserId:    botID,
		ChannelId: dmChannelID,
		Message:   expectedMessage,
	}

	postWithID := model.Post{
		Id:        expectedPostID,
		UserId:    botID,
		ChannelId: dmChannelID,
		Message:   expectedMessage,
	}

	channel := &model.Channel{
		Id: dmChannelID,
	}

	mockError := errors.New("mock channel error")

	t.Run("DM Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		postAPI := mock_api.NewMockPostAPI(ctrl)
		channelAPI := mock_api.NewMockChannelAPI(ctrl)

		poster := NewPoster(postAPI, channelAPI, botID)

		channelAPI.
			EXPECT().
			GetDirect(userID, botID).
			Return(channel, nil).
			Times(1)

		//nolint:govet //copy lock, but only used in tests
		postAPI.
			EXPECT().
			CreatePost(post).
			SetArg(0, postWithID).
			Return(nil).
			Times(1)

		postID, err := poster.DM(userID, format, args...)
		assert.Equal(t, expectedPostID, postID)
		assert.NoError(t, err)
	})

	t.Run("Channel error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		postAPI := mock_api.NewMockPostAPI(ctrl)
		channelAPI := mock_api.NewMockChannelAPI(ctrl)

		poster := NewPoster(postAPI, channelAPI, botID)

		channelAPI.
			EXPECT().
			GetDirect(userID, botID).
			Return(nil, mockError).
			Times(1)

		_, err := poster.DM(userID, format, args...)
		assert.Error(t, err)
	})

	t.Run("Post creation error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		postAPI := mock_api.NewMockPostAPI(ctrl)
		channelAPI := mock_api.NewMockChannelAPI(ctrl)

		poster := NewPoster(postAPI, channelAPI, botID)

		channelAPI.
			EXPECT().
			GetDirect(userID, botID).
			Return(channel, nil).
			Times(1)

		postAPI.
			EXPECT().
			CreatePost(post).
			Return(mockError).
			Times(1)

		_, err := poster.DM(userID, format, args...)
		assert.Error(t, err)
	})
}

func TestDMWithAttachments(t *testing.T) {
	expectedPostID := "expected-post-id"

	attachments := []*model.SlackAttachment{
		{},
		{},
	}

	post := &model.Post{
		UserId:    botID,
		ChannelId: dmChannelID,
	}

	model.ParseSlackAttachment(post, attachments)

	postWithID := model.Post{
		Id:        expectedPostID,
		UserId:    botID,
		ChannelId: dmChannelID,
		Type:      model.POST_SLACK_ATTACHMENT,
		Props: model.StringInterface{
			"attachments": attachments,
		},
	}

	channel := &model.Channel{
		Id: dmChannelID,
	}

	mockError := errors.New("mock channel error")
	t.Run("DM Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		postAPI := mock_api.NewMockPostAPI(ctrl)
		channelAPI := mock_api.NewMockChannelAPI(ctrl)

		poster := NewPoster(postAPI, channelAPI, botID)

		channelAPI.
			EXPECT().
			GetDirect(userID, botID).
			Return(channel, nil).
			Times(1)

		//nolint:copylocks //only used in tests
		postAPI.
			EXPECT().
			CreatePost(post).
			SetArg(0, postWithID).
			Return(nil).
			Times(1)

		postID, err := poster.DMWithAttachments(userID, attachments...)
		assert.Equal(t, expectedPostID, postID)
		assert.NoError(t, err)
	})

	t.Run("Channel error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		postAPI := mock_api.NewMockPostAPI(ctrl)
		channelAPI := mock_api.NewMockChannelAPI(ctrl)

		poster := NewPoster(postAPI, channelAPI, botID)

		channelAPI.
			EXPECT().
			GetDirect(userID, botID).
			Return(nil, mockError).
			Times(1)

		_, err := poster.DMWithAttachments(userID, attachments...)
		assert.Error(t, err)
	})

	t.Run("Post creation error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		postAPI := mock_api.NewMockPostAPI(ctrl)
		channelAPI := mock_api.NewMockChannelAPI(ctrl)

		poster := NewPoster(postAPI, channelAPI, botID)

		channelAPI.
			EXPECT().
			GetDirect(userID, botID).
			Return(channel, nil).
			Times(1)

		postAPI.
			EXPECT().
			CreatePost(post).
			Return(mockError).
			Times(1)

		_, err := poster.DMWithAttachments(userID, attachments...)
		assert.Error(t, err)
	})
}
