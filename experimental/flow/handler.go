// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package flow

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-api/experimental/common"
	"github.com/mattermost/mattermost-plugin-api/experimental/flow/steps"
)

type fh struct {
	fc *flowController
}

func initHandler(r *mux.Router, fc *flowController) {
	fh := &fh{
		fc: fc,
	}

	flowRouter := r.PathPrefix("/").Subrouter()
	flowRouter.HandleFunc(fc.GetFlow().Path()+"/button", fh.handleFlowButton).Methods(http.MethodPost)
	flowRouter.HandleFunc(fc.GetFlow().Path()+"/dialog", fh.handleFlowDialog).Methods(http.MethodPost)
}

func (fh *fh) handleFlowButton(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-ID")
	if userID == "" {
		common.SlackAttachmentError(w, "Error: Not authorized")
		return
	}

	var request model.PostActionIntegrationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		common.SlackAttachmentError(w, "Error: invalid request")
		return
	}

	rawStep, ok := request.Context[steps.ContextStepKey].(string)
	if !ok {
		common.SlackAttachmentError(w, "Error: missing step number")
		return
	}

	var stepNumber int
	err := json.Unmarshal([]byte(rawStep), &stepNumber)
	if err != nil {
		common.SlackAttachmentError(w, fmt.Sprintf("Error: cannot parse step number %v - %v", err.Error(), rawStep))
		return
	}

	step := fh.fc.GetFlow().Step(stepNumber)
	if step == nil {
		common.SlackAttachmentError(w, fmt.Sprintf("Error: There is no step %d.", step))
		return
	}

	rawButtonNumber, ok := request.Context[steps.ContextButtonIDKey].(string)
	if !ok {
		common.SlackAttachmentError(w, "Error: missing button id")
		return
	}

	var buttonNumber int
	err = json.Unmarshal([]byte(rawButtonNumber), &buttonNumber)
	if err != nil {
		common.SlackAttachmentError(w, "Error: cannot parse button number")
		return
	}

	actions := step.Attachment().Actions
	if buttonNumber > len(actions)-1 {
		common.SlackAttachmentError(w, "Error: button number to high")
		return
	}

	action := actions[buttonNumber]
	skip, attachment := action.OnClick()

	response := model.PostActionIntegrationResponse{}
	post := &model.Post{}
	model.ParseSlackAttachment(post, []*model.SlackAttachment{fh.fc.toSlackAttachments(attachment, stepNumber)})
	response.Update = post

	if action.Dialog != nil {
		dialogRequest := model.OpenDialogRequest{
			TriggerId: request.TriggerId,
			URL:       fh.fc.getDialogHandlerURL(),
			Dialog:    action.Dialog.Dialog,
		}
		dialogRequest.Dialog.State = fmt.Sprintf("%v_%v", rawStep, buttonNumber)

		err = fh.fc.OpenInteractiveDialog(dialogRequest)
		if err != nil {
			fh.fc.Logger.WithError(err).Debugf("Failed to open interactive dialog")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)

	err = fh.fc.NextStep(userID, stepNumber, skip)
	if err != nil {
		fh.fc.Logger.WithError(err).Debugf("To advance to next step")
	}
}

func (fh *fh) handleFlowDialog(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-ID")
	if userID == "" {
		common.DialogError(w, errors.New("Error: Not authorized"))
		return
	}

	var request model.SubmitDialogRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		common.DialogError(w, errors.New("Error: invalid request"))
		return
	}

	states := strings.Split(request.State, "_")
	if len(states) != 2 {
		common.DialogError(w, errors.New("Error: invalid request"))
		return
	}

	stepNumber, err := strconv.Atoi(states[0])
	if err != nil {
		common.DialogError(w, errors.Wrap(err, "Error: malformed step number"))
		return
	}

	step := fh.fc.GetFlow().Step(stepNumber)
	if step == nil {
		common.DialogError(w, errors.New("Error: There is no step"))
		return
	}

	buttonNumber, err := strconv.Atoi(states[1])
	if err != nil {
		common.DialogError(w, errors.Wrap(err, "Error: malformed button number"))
		return
	}

	actions := step.Attachment().Actions
	if buttonNumber > len(actions)-1 {
		common.SlackAttachmentError(w, "Error: button number to high")
		return
	}

	skip, attachment, resposeError, resposeErrors := actions[buttonNumber].Dialog.OnDialogSubmit(request.Submission)

	response := model.SubmitDialogResponse{
		Error:  resposeError,
		Errors: resposeErrors,
	}

	if attachment != nil {
		postID, err := fh.fc.store.GetPostID(userID, step.GetPropertyName())
		if err != nil {
			common.DialogError(w, errors.Wrap(err, "Failed to get post"))
			return
		}

		post := &model.Post{
			Id: postID,
		}

		model.ParseSlackAttachment(post, []*model.SlackAttachment{fh.fc.toSlackAttachments(*attachment, stepNumber)})
		err = fh.fc.UpdatePost(post)
		if err != nil {
			common.DialogError(w, errors.Wrap(err, "Failed to update post"))
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)

	if resposeError == "" && resposeErrors == nil {
		err = fh.fc.NextStep(userID, stepNumber, skip)
		if err != nil {
			fh.fc.Logger.WithError(err).Debugf("To advance to next step")
		}
	}
}
