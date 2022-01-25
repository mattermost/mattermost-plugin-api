package steps

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v6/model"
)

type emptyStep struct {
	title   string
	message string
}

func NewEmptyStep(title, message string) Step {
	return &emptyStep{
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

func (s *emptyStep) GetPropertyName() string {
	return ""
}

func (s *emptyStep) IsEmpty() bool {
	return true
}
