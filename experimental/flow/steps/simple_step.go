package steps

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v6/model"
)

type simpleStep struct {
	title                string
	message              string
	propertyName         string
	trueButtonMessage    string
	falseButtonMessage   string
	trueResponseMessage  string
	falseResponseMessage string
	trueSkip             int
	falseSkip            int
}

func NewSimpleStep(
	title,
	message,
	propertyName,
	trueButtonMessage,
	falseButtonMessage,
	trueResponseMessage,
	falseResponseMessage string,
	trueSkip,
	falseSkip int,
) Step {
	return &simpleStep{
		title:                title,
		message:              message,
		propertyName:         propertyName,
		trueButtonMessage:    trueButtonMessage,
		falseButtonMessage:   falseButtonMessage,
		trueResponseMessage:  trueResponseMessage,
		falseResponseMessage: falseResponseMessage,
		trueSkip:             trueSkip,
		falseSkip:            falseSkip,
	}
}

func (s *simpleStep) Attachment(pluginURL string) Attachment {
	actionTrue := Action{
		PostAction: model.PostAction{
			Type:     model.PostActionTypeButton,
			Name:     s.trueButtonMessage,
			Disabled: false,
		},
		OnClick: func(userID string) (int, Attachment) {
			return s.trueSkip, Attachment{
				SlackAttachment: &model.SlackAttachment{
					Title:    s.title,
					Text:     s.trueResponseMessage,
					Fallback: fmt.Sprintf("%s: %s", s.title, s.trueResponseMessage),
				}}
		},
	}

	actionFalse := Action{
		PostAction: model.PostAction{
			Type:     model.PostActionTypeButton,
			Name:     s.falseButtonMessage,
			Disabled: false,
		},
		OnClick: func(userID string) (int, Attachment) {
			return s.falseSkip, Attachment{
				SlackAttachment: &model.SlackAttachment{
					Title:    s.title,
					Text:     s.falseResponseMessage,
					Fallback: fmt.Sprintf("%s: %s", s.title, s.falseResponseMessage),
				},
			}
		},
	}

	a := Attachment{
		SlackAttachment: &model.SlackAttachment{
			Title:    s.title,
			Text:     s.message,
			Fallback: fmt.Sprintf("%s: %s", s.title, s.message),
		},
		Actions: []Action{actionTrue, actionFalse},
	}

	return a
}

func (s *simpleStep) GetPropertyName() string {
	return s.propertyName
}

func (s *simpleStep) IsEmpty() bool {
	return false
}
