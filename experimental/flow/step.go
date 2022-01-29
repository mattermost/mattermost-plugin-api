package flow

import (
	"github.com/mattermost/mattermost-server/v6/model"
)

type Step interface {
	Name() Name
	Render(state State, pluginURL string) Attachment
}

type Attachment struct {
	SlackAttachment *model.SlackAttachment
	Actions         []Action
}

type Action struct {
	model.PostAction
	OnClickDialog *Dialog
	OnClick       func(userID string, state State) (next Name, updated *Attachment)
}

type Dialog struct {
	Dialog   model.Dialog
	OnSubmit func(userID string, submitted map[string]interface{}, appState State) (next Name, updated *Attachment, generalError string, fieldErrors map[string]string)
	OnCancel func(userID string, appState State) (Name, *Attachment)
}

func (a *Attachment) asSlackAttachment() *model.SlackAttachment {
	ret := *a.SlackAttachment
	ret.Actions = make([]*model.PostAction, len(a.Actions))

	for i := 0; i < len(a.Actions); i++ {
		postAction := a.Actions[i].PostAction
		ret.Actions[i] = &postAction
	}

	return &ret
}

func (f UserFlow) StepActions(step Step, appState State) []Action {
	return step.Render(appState, f.pluginURL).Actions
}
