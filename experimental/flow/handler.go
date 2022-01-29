// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package flow

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-api/experimental/common"
)

func (f *UserFlow) handleButton(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-ID")
	if userID == "" {
		common.SlackAttachmentError(w, errors.New("Not authorized"))
		return
	}

	var request model.PostActionIntegrationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		common.SlackAttachmentError(w, errors.New("invalid request"))
		return
	}

	fromName, ok := request.Context[contextStepKey].(Name)
	if !ok {
		common.SlackAttachmentError(w, errors.New("missing step name"))
		return
	}
	button, ok := request.Context[contextButtonKey].(int)
	if !ok {
		common.SlackAttachmentError(w, errors.New("missing button id"))
		return
	}

	state, err := f.getState(userID)
	if err != nil {
		common.SlackAttachmentError(w, err)
		return
	}
	if state.StepName != fromName {
		common.SlackAttachmentError(w, errors.Errorf("click from an inactive step: %v"))
		return
	}
	from := f.Step(fromName)
	if from == nil {
		common.SlackAttachmentError(w, errors.New("there is no step"))
		return
	}

	// render the "active" state of the current step to validate the button
	// index.
	actions := from.Render(state.AppState, f.pluginURL).Actions
	if button > len(actions)-1 {
		common.SlackAttachmentError(w, errors.Errorf("button number %v to high, %v buttons", button, len(actions)))
		return
	}

	action := actions[button]
	toName, updatedAttachment := action.OnClick(userID, state.AppState)
	response := model.PostActionIntegrationResponse{
		Update: f.renderAsPost(from.Name(), *updatedAttachment),
	}

	if action.OnClickDialog != nil {
		dialogRequest := model.OpenDialogRequest{
			TriggerId: request.TriggerId,
			URL:       f.pluginURL + makePath(f.Name) + "/dialog",
			Dialog:    action.OnClickDialog.Dialog,
		}
		dialogRequest.Dialog.State = fmt.Sprintf("%v,%v", fromName, button)

		err = f.api.Frontend.OpenInteractiveDialog(dialogRequest)
		if err != nil {
			common.SlackAttachmentError(w, err)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)

	err = f.Go(userID, toName)
	if err != nil {
		f.api.Log.Warn("failed to advance flow to next step", "flow_name", f.Name, "from", fromName, "to", toName, "error", err.Error())
	}
}

func (f *UserFlow) handleDialog(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-ID")
	if userID == "" {
		common.DialogError(w, errors.New("not authorized"))
		return
	}

	var request model.SubmitDialogRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		common.DialogError(w, errors.New("invalid request"))
		return
	}

	data := strings.Split(request.State, ",")
	if len(data) != 2 {
		common.DialogError(w, errors.New("invalid request"))
		return
	}
	fromName := Name(data[0])
	button, err := strconv.Atoi(data[1])
	if err != nil {
		common.DialogError(w, errors.Wrap(err, "malformed button number"))
		return
	}

	state, err := f.getState(userID)
	if err != nil {
		common.DialogError(w, err)
		return
	}
	if state.StepName != fromName {
		common.DialogError(w, errors.Errorf("click from an inactive step: %v"))
		return
	}
	from := f.Step(fromName)
	if from == nil {
		common.DialogError(w, errors.New("there is no step"))
		return
	}

	actions := from.Render(state.AppState, f.pluginURL).Actions
	if button > len(actions)-1 {
		common.DialogError(w, errors.New("button number to high"))
		return
	}
	action := actions[button]

	var (
		toName            Name
		updatedAttachment *Attachment
		responseError     string
		responseErrors    map[string]string
	)

	if request.Cancelled {
		toName, updatedAttachment = action.OnClickDialog.OnCancel(userID, state.AppState)
	} else {
		toName, updatedAttachment, responseError, responseErrors = action.OnClickDialog.OnSubmit(userID, request.Submission, state.AppState)
	}
	// Empty next step name in the response indicates advancing to the next step
	// in the flow. To stay on the same step the handlers should return the step
	// name.
	if toName == "" {
		toName = f.next(fromName)
	}

	response := model.SubmitDialogResponse{
		Error:  responseError,
		Errors: responseErrors,
	}

	if updatedAttachment != nil && state.PostID != "" {
		post := f.renderAsPost(fromName, *updatedAttachment)
		post.Id = state.PostID
		err = f.api.Post.UpdatePost(post)
		if err != nil {
			common.DialogError(w, errors.Wrap(err, "Failed to update post"))
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
	if responseError != "" || responseErrors != nil {
		return
	}

	// advance to next step if needed.
	err = f.Go(userID, toName)
	if err != nil {
		f.api.Log.Warn("failed to advance to next step", "flow", f.Name, "step", toName)
	}
}
