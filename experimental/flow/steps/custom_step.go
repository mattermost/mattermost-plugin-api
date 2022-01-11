package steps

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-api/experimental/freetextfetcher"
)

type ButtonStyle string

const (
	Default ButtonStyle = "default"
	Primary ButtonStyle = "primary"
	Success ButtonStyle = "success"
	Good    ButtonStyle = "good"
	Warning ButtonStyle = "warning"
	Danger  ButtonStyle = "danger"
)

type Button struct {
	Name     string
	Disabled bool
	Style    ButtonStyle
	OnClick  func() int

	Dialog *Dialog
}

type customStepBuilder struct {
	step customStep
}

func NewCustomStepBuilder(title, message string) *customStepBuilder {
	return &customStepBuilder{
		step: customStep{
			ID:      model.NewId(),
			Title:   title,
			Message: message,
		},
	}
}

func (b *customStepBuilder) WithButton(button Button) *customStepBuilder {
	b.step.Buttons = append(b.step.Buttons, button)

	return b
}

func (b *customStepBuilder) WithPretext(text string) *customStepBuilder {
	b.step.Pretext = text

	return b
}

func (b *customStepBuilder) Build() Step {
	return &b.step
}

type customStep struct {
	ID      string
	Title   string
	Message string

	Pretext string
	Buttons []Button

	PropertyName string
}

func (s *customStep) Attachment() Attachment {
	a := Attachment{
		SlackAttachment: &model.SlackAttachment{
			Pretext:  s.Pretext,
			Title:    s.Title,
			Text:     s.Message,
			Fallback: fmt.Sprintf("%s: %s", s.Title, s.Message),
		},
		Actions: s.getActions(),
	}

	return a
}

func (s *customStep) getActions() []Action {
	if s.Buttons == nil {
		return nil
	}

	var actions []Action
	for i, b := range s.Buttons {
		onClick := b.OnClick
		j := i

		dialog := b.Dialog

		action := Action{
			PostAction: model.PostAction{
				Type:     model.PostActionTypeButton,
				Name:     b.Name,
				Disabled: b.Disabled,
				Style:    string(b.Style),
			},
			OnClick: func() (int, Attachment) {
				skip := 0
				if onClick != nil {
					skip = onClick()
				}

				var newActions []Action
				if skip == -1 {
					// Keep full list
					newActions = s.getActions()
				} else {
					// Only list the selected one
					action := s.getActions()[j]
					action.Disabled = true

					newActions = []Action{action}
				}

				attachment := Attachment{
					SlackAttachment: &model.SlackAttachment{
						Pretext:  s.Pretext,
						Title:    s.Title,
						Text:     s.Message,
						Fallback: fmt.Sprintf("%s: %s", s.Title, s.Message),
					},
					Actions: newActions,
				}

				return skip, attachment
			},
		}

		if dialog != nil {
			action.Dialog = &Dialog{
				Dialog: dialog.Dialog,

				OnDialogSubmit: func(submission map[string]interface{}) (int, *Attachment, string, map[string]string) {
					skip, _, resposeError, resposeErrors := dialog.OnDialogSubmit(submission)

					var newActions []Action
					if skip == -1 || resposeError != "" || len(resposeErrors) != 0 {
						// Keep full list
						newActions = s.getActions()
					} else {
						// Only list the selected one
						newAction := s.getActions()[j]
						newAction.Disabled = true

						newActions = []Action{newAction}
					}

					attachment := &Attachment{
						SlackAttachment: &model.SlackAttachment{
							Pretext:  s.Pretext,
							Title:    s.Title,
							Text:     s.Message,
							Fallback: fmt.Sprintf("%s: %s", s.Title, s.Message),
						},
						Actions: newActions,
					}

					return skip, attachment, resposeError, resposeErrors
				},
			}
		}

		actions = append(actions, action)
	}

	return actions
}

func (s *customStep) GetPropertyName() string {
	return s.ID
}

func (s *customStep) ShouldSkip(rawValue interface{}) int {
	i, err := s.parseValue(rawValue)
	if err != nil {
		// TODO: properly handle this case
		return 0
	}

	if i > len(s.Buttons)-1 {
		// TODO: properly handle this case
		return 0
	}

	b := s.Buttons[i]

	if b.Dialog != nil {
		return -1 // Go back to the current step
	}

	return 100
}

func (*customStep) parseValue(rawValue interface{}) (int, error) {
	v, ok := rawValue.(string)
	if !ok {
		return 0, errors.New("value is not a string")
	}

	var i int
	err := json.Unmarshal([]byte(v), &i)
	if err != nil {
		return 0, errors.Wrap(err, "failed to unmarshal json value")
	}

	return i, nil
}

func (s *customStep) IsEmpty() bool {
	return len(s.Buttons) == 0
}

func (*customStep) GetFreetextFetcher() freetextfetcher.FreetextFetcher {
	return nil
}
