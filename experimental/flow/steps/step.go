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
	ret.Actions = make([]*model.PostAction, len(a.Actions))

	for i := 0; i < len(a.Actions); i++ {
		postAction := a.Actions[i].PostAction
		ret.Actions[i] = &postAction
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
