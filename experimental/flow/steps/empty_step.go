package steps

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v6/model"
)

type emptyStep struct {
	name    string
	title   string
	message string
}

func NewEmptyStep(name, title, message string) Step {
	return &emptyStep{
		name:    name,
		title:   title,
		message: message,
	}
}

func (s *emptyStep) Attachment(pluginURL string) Attachment {
	sa := Attachment{
		SlackAttachment: &model.SlackAttachment{
			Title:    s.title,
			Text:     s.message,
			Fallback: fmt.Sprintf("%s: %s", s.title, s.message),
		},
	}

	return sa
}

func (s *emptyStep) Name() string {
	return s.name
}

func (s *emptyStep) IsEmpty() bool {
	return true
}
