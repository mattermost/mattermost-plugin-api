package flow

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-api/experimental/bot/logger"
	"github.com/mattermost/mattermost-plugin-api/experimental/bot/poster"
	"github.com/mattermost/mattermost-plugin-api/experimental/flow/steps"
)

type DialogCreater interface {
	OpenInteractiveDialog(dialog model.OpenDialogRequest) error
}

type Controller interface {
	Start(userID string) error
	NextStep(userID string, from int, skip int) error
	GetCurrentStep(userID string) (steps.Step, int, error)
	GetFlow() Flow
	Cancel(userID string) error
	SetProperty(userID, propertyName string, value interface{}) error
}

type flowController struct {
	logger.Logger
	poster.Poster
	DialogCreater
	flow          Flow
	store         Store
	propertyStore PropertyStore
	pluginURL     string
}

func NewFlowController(
	l logger.Logger,
	r *mux.Router,
	p poster.Poster,
	d DialogCreater,
	pluginURL string,
	flow Flow,
	flowStore Store,
	propertyStore PropertyStore,
) Controller {
	fc := &flowController{
		Poster:        p,
		Logger:        l,
		DialogCreater: d,
		flow:          flow,
		store:         flowStore,
		propertyStore: propertyStore,
		pluginURL:     pluginURL,
	}

	initHandler(r, fc)

	for _, step := range flow.Steps() {
		ftf := step.GetFreetextFetcher()
		if ftf != nil {
			ftf.UpdateHooks(nil,
				fc.ftOnFetch,
				fc.ftOnCancel,
			)
		}
	}

	return fc
}

func (fc *flowController) GetFlow() Flow {
	return fc.flow
}

func (fc *flowController) SetProperty(userID, propertyName string, value interface{}) error {
	return fc.propertyStore.SetProperty(userID, propertyName, value)
}

func (fc *flowController) Start(userID string) error {
	err := fc.setFlowStep(userID, 1)
	if err != nil {
		return err
	}
	return fc.processStep(userID, 1)
}

func (fc *flowController) NextStep(userID string, from, skip int) error {
	stepIndex, err := fc.getFlowStep(userID)
	if err != nil {
		return err
	}

	log.Printf("from: %#+v\n", from)
	log.Printf("skip: %#+v\n", skip)

	if stepIndex != from {
		// We are beyond the step we were supposed to come from, so we understand this step has already been processed.
		// Used to avoid rapid firing on the Slack Attachments.
		return nil
	}

	if skip == -1 {
		// Stay at the current step
		return nil
	}

	step := fc.flow.Step(stepIndex)

	err = fc.store.RemovePostID(userID, step.GetPropertyName())
	if err != nil {
		fc.Logger.WithError(err).Debugf("error removing post id")
	}

	stepIndex += 1 + skip
	log.Printf("stepIndex: %#+v\n", stepIndex)
	if stepIndex > fc.flow.Length() {
		_ = fc.removeFlowStep(userID)
		fc.flow.FlowDone(userID)
		return nil
	}

	err = fc.setFlowStep(userID, stepIndex)
	if err != nil {
		return err
	}

	return fc.processStep(userID, stepIndex)
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

func (fc *flowController) getButtonHandlerURL() string {
	return fc.pluginURL + fc.flow.Path() + "/button"
}

func (fc *flowController) getDialogHandlerURL() string {
	return fc.pluginURL + fc.flow.Path() + "/dialog"
}

func (fc *flowController) toSlackAttachments(attachment steps.Attachment, stepNumber int) *model.SlackAttachment {
	stepValue, _ := json.Marshal(stepNumber)

	var updatedActions []steps.Action
	for i, action := range attachment.Actions {
		buttonNumber, _ := json.Marshal(i)

		updatedAction := action

		updatedAction.Integration = &model.PostActionIntegration{
			URL: fc.getButtonHandlerURL(),
			Context: map[string]interface{}{
				steps.ContextStepKey:     string(stepValue),
				steps.ContextButtonIDKey: string(buttonNumber),
			},
		}

		updatedActions = append(updatedActions, updatedAction)
	}

	attachment.Actions = updatedActions

	return attachment.ToSlackAttachment()
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

func (fc *flowController) processStep(userID string, i int) error {
	step := fc.flow.Step(i)
	if step == nil {
		fc.Errorf("Step nil")
	}

	if fc.flow == nil {
		fc.Errorf("Flow nil")
	}

	if fc.store == nil {
		fc.Errorf("Store nil")
	}

	attachements := fc.toSlackAttachments(step.Attachment(), i)
	postID, err := fc.DMWithAttachments(userID, attachements)
	if err != nil {
		return err
	}

	if step.IsEmpty() {
		return fc.NextStep(userID, i, 0)
	}

	log.Println("SetPostID")
	log.Printf("userID: %#+v\n", userID)
	log.Printf("step.GetPropertyName(): %#+v\n", step.GetPropertyName())
	log.Printf("postID: %#+v\n", postID)
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
