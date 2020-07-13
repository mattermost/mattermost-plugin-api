package poster

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-api/experimental/common"
)

type defaultPoster struct {
	postAPI    common.PostAPI
	channelAPI common.ChannelAPI
	id         string
}

func NewPoster(postAPI common.PostAPI, channelAPI common.ChannelAPI, id string) Poster {
	return &defaultPoster{
		postAPI:    postAPI,
		channelAPI: channelAPI,
		id:         id,
	}
}

// DM posts a simple Direct Message to the specified user
func (p *defaultPoster) DM(mattermostUserID, format string, args ...interface{}) (string, error) {
	postID, err := p.dm(mattermostUserID, &model.Post{
		Message: fmt.Sprintf(format, args...),
	})
	if err != nil {
		return "", err
	}
	return postID, nil
}

// DMWithAttachments posts a Direct Message that contains Slack attachments.
// Often used to include post actions.
func (p *defaultPoster) DMWithAttachments(mattermostUserID string, attachments ...*model.SlackAttachment) (string, error) {
	post := model.Post{}
	model.ParseSlackAttachment(&post, attachments)
	return p.dm(mattermostUserID, &post)
}

func (p *defaultPoster) dm(mattermostUserID string, post *model.Post) (string, error) {
	channel, err := p.channelAPI.GetDirect(mattermostUserID, p.id)
	if err != nil {
		return "", errors.Wrap(err, "couldn't get bot's DM channel")
	}
	post.ChannelId = channel.Id
	post.UserId = p.id
	err = p.postAPI.CreatePost(post)
	if err != nil {
		return "", err
	}
	return post.Id, nil
}

// Ephemeral sends an ephemeral message to a user
func (p *defaultPoster) Ephemeral(userID, channelID, format string, args ...interface{}) {
	post := &model.Post{
		UserId:    p.id,
		ChannelId: channelID,
		Message:   fmt.Sprintf(format, args...),
	}
	p.postAPI.SendEphemeralPost(userID, post)
}

func (p *defaultPoster) UpdatePostByID(postID, format string, args ...interface{}) error {
	post, err := p.postAPI.GetPost(postID)
	if err != nil {
		return err
	}

	post.Message = fmt.Sprintf(format, args...)
	return p.UpdatePost(post)
}

func (p *defaultPoster) DeletePost(postID string) error {
	return p.postAPI.DeletePost(postID)
}

func (p *defaultPoster) UpdatePost(post *model.Post) error {
	return p.postAPI.UpdatePost(post)
}

func (p *defaultPoster) UpdatePosterID(id string) {
	p.id = id
}
