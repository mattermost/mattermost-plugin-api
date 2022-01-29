package flow

import (
	"net/http"

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
	Start(userID string) error
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
	flowRouter.HandleFunc(MakePath(f.Name)+"/button", f.handleButton).Methods(http.MethodPost)
	flowRouter.HandleFunc(MakePath(f.Name)+"/dialog", f.handleDialog).Methods(http.MethodPost)
	return &f
}

func (f UserFlow) Step(name Name) Step {
	step, _ := f.steps[name]
	return step
}

func (f UserFlow) Start(userID string) error {
	if len(f.index) == 0 {
		return errors.New("no steps")
	}
	return f.Go(userID, f.index[0])
}

func (f UserFlow) Go(userID string, toName Name) error {
	state, err := getState(&f.api.KV, userID, f.Name)
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

	post := f.renderAsPost(toName, to.Render(state.AppState, f.pluginURL))
	err = f.api.Post.DM(f.botUserID, userID, post)
	if err != nil {
		return err
	}

	state.StepName = toName
	state.PostID = post.Id
	err = storeState(&f.api.KV, userID, f.Name, state)
	if err != nil {
		return err
	}

	// If the "to" step is not actionable, proceed to the next step,
	// recursively.
	if len(f.StepActions(to, state.AppState)) == 0 {
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

func (f UserFlow) renderAsPost(name Name, attachment Attachment) *model.Post {
	post := model.Post{}
	updated := []Action{}
	for i, a := range attachment.Actions {
		a.Integration = &model.PostActionIntegration{
			URL: f.pluginURL + MakePath(f.Name) + "/button",
			Context: map[string]interface{}{
				contextStepKey:   string(name),
				contextButtonKey: i,
			},
		}
		// append by value, ok to use the modified loop variable
		updated = append(updated, a)
	}
	attachment.Actions = updated

	model.ParseSlackAttachment(&post, []*model.SlackAttachment{
		attachment.asSlackAttachment(),
	})
	return &post
}
