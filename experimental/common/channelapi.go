package common

import "github.com/mattermost/mattermost-server/v5/model"

type ChannelAPI interface {
	GetDirect(userID1, userID2 string) (*model.Channel, error)
}
