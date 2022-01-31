package flow

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mattermost/mattermost-server/v6/model"
)

type Color string

const (
	ColorDefault Color = "default"
	ColorPrimary Color = "primary"
	ColorSuccess Color = "success"
	ColorGood    Color = "good"
	ColorWarning Color = "warning"
	ColorDanger  Color = "danger"
)

type Step interface {
	Name() Name
	IsTerminal() bool
	Render(state State, pluginURL string, done bool, selectedButton int) Attachment
}

type Attachment struct {
	SlackAttachment *model.SlackAttachment
	Buttons         []Button
}

type Button struct {
	Name     string
	Disabled bool
	Color    Color

	// OnClick is called when the button is clicked. It returns the next step's
	// name and the state updates to apply.
	//
	// If Dialog is also specified, OnClick is executed first.
	OnClick func(f Flow) (Name, State)

	// Dialog is the interactive dialog to display if the button is clicked
	// (OnClick is executed first). OnDialogSubmit must be provided.
	Dialog *model.Dialog

	// Function that is called when the dialog box is submitted. It can return a
	// general error, or field-specific errors. On success it returns the name
	// of the next step, and the state updates to apply.
	OnDialogSubmit func(f Flow, submitted map[string]interface{}) (Name, State, string, map[string]string)
}

type step struct {
	model.SlackAttachment

	name     Name
	terminal bool
	buttons  []Button
}

var _ Step = (*step)(nil)

func NewStep(name Name) *step {
	return &step{
		name: name,
	}
}

func (s step) Terminal() *step {
	s.terminal = true
	return &s
}

func (s step) WithColor(color Color) *step {
	s.Color = string(color)
	return &s
}

func (s step) WithTitle(text string) *step {
	s.Title = text
	return &s
}

func (s step) WithMessage(text string) *step {
	s.Text = text
	return &s
}

func (s step) WithPretext(text string) *step {
	s.Pretext = text
	return &s
}

func (s step) WithButton(button Button) *step {
	s.buttons = append(s.buttons, button)
	return &s
}

func (s step) WithImage(pluginURL, path string) *step {
	if path != "" {
		s.ImageURL = pluginURL + "/" + strings.TrimPrefix(path, "/")
	}
	return &s
}

func (s *step) Render(state State, pluginURL string, done bool, button int) Attachment {
	buttons := s.getButtons(state)
	// if moving to a different step, indicate the performed selection by
	// including only the selected button, disabled.
	if done {
		selected := buttons[button]
		selected.Disabled = true
		buttons = []Button{selected}
	}

	a := Attachment{
		SlackAttachment: s.getSlackAttachment(state, pluginURL),
		Buttons:         buttons,
	}
	return a
}

func (s *step) getSlackAttachment(state State, pluginURL string) *model.SlackAttachment {
	a := s.SlackAttachment
	a.Pretext = formatState(s.Pretext, state)
	a.Title = formatState(s.Title, state)
	a.Text = formatState(s.Text, state)
	a.Fallback = fmt.Sprintf("%s: %s", a.Title, a.Text)
	return &a
}

func (s *step) getButtons(state State) []Button {
	var buttons []Button
	for _, b := range s.buttons {
		button := b
		button.Name = formatState(b.Name, state)
		buttons = append(buttons, b)
	}
	return buttons
}

func (s *step) Name() Name {
	return s.name
}

func (s *step) IsTerminal() bool {
	return s.terminal
}

func (f UserFlow) renderButton(b Button, stepName Name, i int) *model.PostAction {
	return &model.PostAction{
		Name:     b.Name,
		Disabled: b.Disabled,
		Style:    string(b.Color),
		Integration: &model.PostActionIntegration{
			URL: f.pluginURL + makePath(f.name) + "/button",
			Context: map[string]interface{}{
				contextStepKey:   string(stepName),
				contextButtonKey: strconv.Itoa(i),
			},
		},
	}
}

func (f UserFlow) Buttons(step Step, appState State) []Button {
	return step.Render(appState, f.pluginURL, false, 0).Buttons
}
