// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package poster

import (
	"github.com/mattermost/mattermost-server/v5/model"
)

type Poster interface {
	// DM posts a simple Direct Message to the specified user
	DM(mattermostUserID, format string, args ...interface{}) (string, error)

	// DMWithAttachments posts a Direct Message that contains Slack attachments.
	// Often used to include post actions.
	DMWithAttachments(mattermostUserID string, attachments ...*model.SlackAttachment) (string, error)

	// Ephemeral sends an ephemeral message to a user
	Ephemeral(mattermostUserID, channelID, format string, args ...interface{})

	// DMPUpdate updates the postID with the formatted message
	DMUpdate(postID, format string, args ...interface{}) error

	// DeletePost deletes a single post
	DeletePost(postID string) error

	// DMUpdatePost substitute one post with another
	UpdatePost(post *model.Post) error

	// UpdatePosterID updates the Mattermost User ID of the poster
	UpdatePosterID(id string)
}
