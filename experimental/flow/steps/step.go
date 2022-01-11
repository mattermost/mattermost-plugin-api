package steps

import (
	"github.com/mattermost/mattermost-plugin-api/experimental/freetextfetcher"

	"github.com/mattermost/mattermost-server/v6/model"
)

const (
	ContextStepKey     = "step"
	ContextButtonIDKey = "button_id"

	ContextPropertyKey    = "property"
	ContextButtonValueKey = "button_value"
	ContextOptionValueKey = "selected_option"
)

type Action struct {
	model.PostAction
	OnClick func() (int, Attachment)
	Dialog  *Dialog
}
type Attachment struct {
	SlackAttachment *model.SlackAttachment
	Actions         []Action
}

func (a *Attachment) ToSlackAttachment() *model.SlackAttachment {
	ret := *a.SlackAttachment
	for _, action := range a.Actions {
		postAction := action.PostAction
		ret.Actions = append(ret.Actions, &postAction)
	}

	return &ret
}

type Step interface {
	Attachment() Attachment
	GetPropertyName() string
	ShouldSkip(value interface{}) int
	IsEmpty() bool
	GetFreetextFetcher() freetextfetcher.FreetextFetcher
}
