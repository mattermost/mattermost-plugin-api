package flow

import (
	"encoding/json"
	"fmt"

	"github.com/mattermost/mattermost-plugin-api/utils/bot/logger"
	"github.com/mattermost/mattermost-plugin-api/utils/bot/poster"
	"github.com/mattermost/mattermost-plugin-api/utils/flow/steps"
)

type FlowController interface {
	RegisterFlow(Flow, FlowStore)
	Start(userID string) error
	NextStep(userID string, from int, value interface{}) error
	GetCurrentStep(userID string) (steps.Step, int, error)
	GetHandlerURL() string
	Cancel(userID string) error
}

type flowController struct {
	poster.Poster
	logger.Logger
	flow      Flow
	store     FlowStore
	pluginURL string
}

func NewFlowController(p poster.Poster, l logger.Logger, pluginURL string) FlowController {
	return &flowController{
		Poster:    p,
		Logger:    l,
		pluginURL: pluginURL,
	}
}

func (fc *flowController) RegisterFlow(flow Flow, store FlowStore) {
	fc.flow = flow
	fc.store = store

	for _, step := range flow.Steps() {
		ftf := step.GetFreetextFetcher()
		if ftf != nil {
			ftf.UpdateHooks(nil,
				fc.ftOnFetch,
				fc.ftOnCancel,
			)
		}
	}
}

func (fc *flowController) Start(userID string) error {
	err := fc.setFlowStep(userID, 1)
	if err != nil {
		return err
	}
	return fc.processStep(userID, fc.flow.Step(1), 1)
}

func (fc *flowController) NextStep(userID string, from int, value interface{}) error {
	step, err := fc.getFlowStep(userID)
	if err != nil {
		return err
	}

	if step != from {
		return nil
	}

	skip := fc.flow.Step(step).ShouldSkip(value)
	step += 1 + skip
	if step > fc.flow.Length() {
		fc.removeFlowStep(userID)
		fc.flow.FlowDone(userID)
		return nil
	}

	err = fc.setFlowStep(userID, step)
	if err != nil {
		return err
	}

	return fc.processStep(userID, fc.flow.Step(step), step)
}

func (fc *flowController) GetCurrentStep(userID string) (steps.Step, int, error) {
	index, err := fc.getFlowStep(userID)
	if err != nil {
		return nil, 0, err
	}

	if index == 0 {
		return nil, 0, nil
	}

	step := fc.flow.Step(index)
	if step == nil {
		return nil, 0, fmt.Errorf("step %d not found", index)
	}

	return step, index, nil
}

func (fc *flowController) GetHandlerURL() string {
	return fc.pluginURL + fc.flow.URL()
}

func (fc *flowController) Cancel(userID string) error {
	stepIndex, err := fc.getFlowStep(userID)
	if err != nil {
		return err
	}

	step := fc.flow.Step(stepIndex)
	if step == nil {
		return nil
	}

	postID, err := fc.store.GetPostID(userID, step.GetPropertyName())
	if err != nil {
		return err
	}

	err = fc.DeletePost(postID)
	if err != nil {
		return err
	}

	return nil
}

func (fc *flowController) setFlowStep(userID string, step int) error {
	return fc.store.SetCurrentStep(userID, step)
}

func (fc *flowController) getFlowStep(userID string) (int, error) {
	return fc.store.GetCurrentStep(userID)
}

func (fc *flowController) removeFlowStep(userID string) error {
	return fc.store.DeleteCurrentStep(userID)
}

func (fc *flowController) processStep(userID string, step steps.Step, i int) error {
	if step == nil {
		fc.Errorf("Step nil")
	}

	if fc.flow == nil {
		fc.Errorf("Flow nil")
	}

	if fc.store == nil {
		fc.Errorf("Store nil")
	}

	postID, err := fc.DMWithAttachments(userID, step.PostSlackAttachment(fc.GetHandlerURL(), i))
	if err != nil {
		return err
	}

	if step.IsEmpty() {
		return fc.NextStep(userID, i, false)
	}

	err = fc.store.SetPostID(userID, step.GetPropertyName(), postID)
	if err != nil {
		return err
	}

	ftf := step.GetFreetextFetcher()
	if ftf == nil {
		return nil
	}

	payload, err := json.Marshal(freetextInfo{
		Step:     i,
		UserID:   userID,
		Property: step.GetPropertyName(),
	})
	if err != nil {
		return err
	}
	ftf.StartFetching(userID, string(payload))
	return nil
}
