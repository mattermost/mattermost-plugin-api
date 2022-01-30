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

type State map[string]string

type Flow interface {
	Step(Name) Step
	Go(userID string, to Name) error
	Start(userID string, appState State) error
}

const (
	contextStepKey   = "step"
	contextButtonKey = "button"
)

type UserFlow struct {
	Name Name

	api       *pluginapi.Client
	pluginURL string
	botUserID string

	steps map[Name]Step
	index []Name
}

var _ Flow = (*UserFlow)(nil)

// NewUserFlow creates a new flow using direct messages with the user.
//
// name must be a unique identifier for the flow within the plugin.
func NewUserFlow(name Name, api *pluginapi.Client, pluginURL, botUserID string) *UserFlow {
	return &UserFlow{
		Name:      name,
		api:       api,
		pluginURL: pluginURL,
		botUserID: botUserID,
	}
}

func (f UserFlow) WithSteps(orderedSteps ...Step) *UserFlow {
	if f.steps == nil {
		f.steps = map[Name]Step{}
	}
	for _, step := range orderedSteps {
		name := step.Name()
		if _, ok := f.steps[name]; ok {
			f.api.Log.Warn("ignored duplicate step name", "name", name, "flow", f.Name)
			continue
		}
		f.steps[name] = step
		f.index = append(f.index, name)
	}
	return &f
}

func (f UserFlow) InitHTTP(r *mux.Router) *UserFlow {
	flowRouter := r.PathPrefix("/").Subrouter()
	flowRouter.HandleFunc(makePath(f.Name)+"/button", f.handleButton).Methods(http.MethodPost)
	flowRouter.HandleFunc(makePath(f.Name)+"/dialog", f.handleDialog).Methods(http.MethodPost)
	return &f
}

func (f UserFlow) Step(name Name) Step {
	step, _ := f.steps[name]
	return step
}

func (f UserFlow) Start(userID string, appState State) error {
	if len(f.index) == 0 {
		return errors.New("no steps")
	}

	err := f.storeState(userID, flowState{
		AppState: appState,
	})
	if err != nil {
		return err
	}

	return f.Go(userID, f.index[0])
}

func (f UserFlow) Go(userID string, toName Name) error {
	state, err := f.getState(userID)
	if err != nil {
		return err
	}
	if toName == state.StepName {
		// Stay at the current step
		return nil
	}
	if toName == "" {
		return errors.New("<>/<> DONE")
	}

	to := f.Step(toName)
	if to == nil {
		return errors.Errorf("%s: step not found", toName)
	}

	post := f.renderAsPost(toName, to.Render(state.AppState, f.pluginURL, false, 0))
	err = f.api.Post.DM(f.botUserID, userID, post)
	if err != nil {
		return err
	}

	state.StepName = toName
	state.PostID = post.Id
	err = f.storeState(userID, state)
	if err != nil {
		return err
	}

	// If the "to" step is not actionable, proceed to the next step,
	// recursively.
	if len(f.Buttons(to, state.AppState)) == 0 {
		return f.Go(userID, f.next(toName))
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

func Goto(next Name) func(string, State) (Name, State) {
	return func(_ string, appState State) (Name, State) {
		return next, appState
	}
}

func DialogGoto(next Name) func(string, map[string]interface{}, State) (Name, State, string, map[string]string) {
	return func(userID string, submitted map[string]interface{}, appState State) (Name, State, string, map[string]string) {
		for k, v := range submitted {
			appState[k] = fmt.Sprintf("%v", v)
		}
		return next, appState, "", nil
	}
}
