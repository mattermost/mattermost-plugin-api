package flow

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/mattermost/mattermost-server/v6/model"
)

var _ Step = (*InfoStep)(nil)

type InfoStep struct {
	name     Name
	title    string
	message  string
	OnRender func(userID string)
}

func NewInfoStep(name Name, title, message string) *InfoStep {
	return &InfoStep{
		name:    name,
		title:   title,
		message: message,
	}
}

func (s *InfoStep) Render(state State, pluginURL string) Attachment {
	sa := Attachment{
		SlackAttachment: &model.SlackAttachment{
			Title:    withState(s.title, state),
			Text:     withState(s.message, state),
			Fallback: fmt.Sprintf("%s: %s", s.title, s.message),
		},
	}

	return sa
}

func (s *InfoStep) Name() Name {
	return s.name
}

func withState(source string, state State) string {
	t, err := template.New("message").Parse(source)
	if err != nil {
		return source + " ###ERROR: " + err.Error()
	}
	buf := bytes.NewBuffer(nil)
	err = t.Execute(buf, state)
	if err != nil {
		return source + " ###ERROR: " + err.Error()
	}
	return buf.String()
}
