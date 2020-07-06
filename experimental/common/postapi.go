package common

import "github.com/mattermost/mattermost-server/v5/model"

type PostAPI interface {
	CreatePost(post *model.Post) error
	GetPost(postID string) (*model.Post, error)
	UpdatePost(post *model.Post) error
	DeletePost(postID string) error
	SendEphemeralPost(userID string, post *model.Post)
}
