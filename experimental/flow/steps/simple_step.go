package steps

import (
	"encoding/json"
	"fmt"

	"github.com/mattermost/mattermost-plugin-api/experimental/freetextfetcher"

	"github.com/mattermost/mattermost-server/v6/model"
)

type simpleStep struct {
	Title                string
	Message              string
	PropertyName         string
	TrueButtonMessage    string
	FalseButtonMessage   string
	TrueResponseMessage  string
	FalseResponseMessage string
	TrueSkip             int
	FalseSkip            int
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
		Title:                title,
		Message:              message,
		PropertyName:         propertyName,
		TrueButtonMessage:    trueButtonMessage,
		FalseButtonMessage:   falseButtonMessage,
		TrueResponseMessage:  trueResponseMessage,
		FalseResponseMessage: falseResponseMessage,
		TrueSkip:             trueSkip,
		FalseSkip:            falseSkip,
	}
}

func (s *simpleStep) Attachment(pluginURL string) Attachment {
	actionTrue := Action{
		PostAction: model.PostAction{
			Type:     model.PostActionTypeButton,
			Name:     s.TrueButtonMessage,
			Disabled: false,
		},
		OnClick: func(userID string) (int, Attachment) {
			return s.TrueSkip, Attachment{
				SlackAttachment: &model.SlackAttachment{
					Title:    s.Title,
					Text:     s.TrueResponseMessage,
					Fallback: fmt.Sprintf("%s: %s", s.Title, s.TrueResponseMessage),
				}}
		},
	}

	actionFalse := Action{
		PostAction: model.PostAction{
			Type:     model.PostActionTypeButton,
			Name:     s.FalseButtonMessage,
			Disabled: false,
		},
		OnClick: func(userID string) (int, Attachment) {
			return s.FalseSkip, Attachment{
				SlackAttachment: &model.SlackAttachment{
					Title:    s.Title,
					Text:     s.FalseResponseMessage,
					Fallback: fmt.Sprintf("%s: %s", s.Title, s.FalseResponseMessage),
				},
			}
		},
	}

	a := Attachment{
		SlackAttachment: &model.SlackAttachment{
			Title:    s.Title,
			Text:     s.Message,
			Fallback: fmt.Sprintf("%s: %s", s.Title, s.Message),
		},
		Actions: []Action{actionTrue, actionFalse},
	}

	return a
}

func (s *simpleStep) GetPropertyName() string {
	return s.PropertyName
}

func (s *simpleStep) ShouldSkip(rawValue interface{}) int {
	value := s.parseValue(rawValue)

	if value {
		return s.TrueSkip
	}

	return s.FalseSkip
}

func (s *simpleStep) IsEmpty() bool {
	return false
}

func (*simpleStep) parseValue(rawValue interface{}) (value bool) {
	err := json.Unmarshal([]byte(rawValue.(string)), &value)
	if err != nil {
		return false
	}

	return value
}

func (*simpleStep) GetFreetextFetcher() freetextfetcher.FreetextFetcher {
	return nil
}
