package flow

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v6/model"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
)

type Name string

type Flow interface {
	ForUser(userID string) Flow

	Go(Name) error
	Start(initial State) error
	Finish() error
	State() (_ State, userID string)
}

const (
	contextStepKey   = "step"
	contextButtonKey = "button"
)

type UserFlow struct {
	name   Name
	userID string

	api       *pluginapi.Client
	pluginURL string
	botUserID string

	steps map[Name]Step
	index []Name

	appState State
	done     func(userID string, state State) error
}

var _ Flow = (*UserFlow)(nil)

// NewUserFlow creates a new flow using direct messages with the user.
//
// name must be a unique identifier for the flow within the plugin.
func NewUserFlow(name Name, api *pluginapi.Client, pluginURL, botUserID string) *UserFlow {
	return &UserFlow{
		name:      name,
		api:       api,
		pluginURL: pluginURL,
		botUserID: botUserID,
		steps:     map[Name]Step{},
	}
}

func (f UserFlow) ForUser(userID string) Flow {
	return f.forUser(userID)
}

func (f UserFlow) forUser(userID string) *UserFlow {
	f.userID = userID
	return &f
}

func (f UserFlow) WithSteps(orderedSteps ...Step) *UserFlow {
	if f.steps == nil {
		f.steps = map[Name]Step{}
	}
	for _, step := range orderedSteps {
		stepName := step.Name()
		if _, ok := f.steps[stepName]; ok {
			f.api.Log.Warn("ignored duplicate step name", "name", stepName, "flow", f.name)
			continue
		}
		f.steps[stepName] = step
		f.index = append(f.index, stepName)
	}
	return &f
}

func (f UserFlow) OnDone(done func(string, State) error) *UserFlow {
	f.done = done
	return &f
}

func (f UserFlow) InitHTTP(r *mux.Router) *UserFlow {
	flowRouter := r.PathPrefix("/").Subrouter()
	flowRouter.HandleFunc(makePath(f.name)+"/button", f.handleButton).Methods(http.MethodPost)
	flowRouter.HandleFunc(makePath(f.name)+"/dialog", f.handleDialog).Methods(http.MethodPost)
	return &f
}

func (f UserFlow) Start(appState State) error {
	if len(f.index) == 0 {
		return errors.New("no steps")
	}

	err := f.storeState(flowState{AppState: appState})
	if err != nil {
		return err
	}

	return f.Go(f.index[0])
}

func (f UserFlow) Finish() error {
	state, err := f.getState()
	if err != nil {
		return err
	}

	f.api.Log.Debug("flow: done", "flow", f.name, "state", state)
	_ = f.removeState()

	if f.done != nil {
		err = f.done(f.userID, state.AppState)
	}
	return err
}

func (f UserFlow) Go(toName Name) error {
	if toName == "" {
		return f.Finish()
	}

	state, err := f.getState()
	if err != nil {
		return err
	}
	if toName == state.StepName {
		// Stay at the current step
		return nil
	}

	f.api.Log.Debug("flow: starting step", "flow", f.name, "step", toName, "state", state)

	to := f.steps[toName]
	if to == nil {
		return errors.Errorf("%s: step not found", toName)
	}
	post := f.renderAsPost(toName, to.Render(state.AppState, false, 0))
	err = f.api.Post.DM(f.botUserID, f.userID, post)
	if err != nil {
		return err
	}
	if to.IsTerminal() {
		return f.Finish()
	}

	state.StepName = toName
	state.PostID = post.Id
	err = f.storeState(state)
	if err != nil {
		return err
	}

	// If the "to" step is not actionable, proceed to the next step,
	// recursively.
	if len(f.Buttons(to, state.AppState)) == 0 {
		return f.Go(f.next(toName))
	}

	return nil
}

func (f UserFlow) next(fromName Name) Name {
	for i, n := range f.index {
		if fromName == n {
			if i+1 < len(f.index) {
				return f.index[i+1]
			}
			return ""
		}
	}
	return ""
}

func (f UserFlow) renderAsPost(stepName Name, attachment Attachment) *model.Post {
	post := model.Post{}
	sa := *attachment.SlackAttachment
	for i, b := range attachment.Buttons {
		sa.Actions = append(sa.Actions, f.renderButton(b, stepName, i))
	}
	model.ParseSlackAttachment(&post, []*model.SlackAttachment{&sa})
	return &post
}

func makePath(name Name) string {
	return "/" + url.PathEscape(strings.Trim(string(name), "/"))
}

func Goto(toName Name) func(Flow) (Name, State) {
	return func(_ Flow) (Name, State) {
		return toName, nil
	}
}

func DialogGoto(toName Name) func(Flow, map[string]interface{}) (Name, State, string, map[string]string) {
	return func(_ Flow, submitted map[string]interface{}) (Name, State, string, map[string]string) {
		stateUpdate := State{}
		for k, v := range submitted {
			stateUpdate[k] = fmt.Sprintf("%v", v)
		}
		return toName, stateUpdate, "", nil
	}
}
