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
	f = f.forUser(userID)

	var request model.PostActionIntegrationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		common.SlackAttachmentError(w, errors.New("invalid request"))
		return
	}

	fromString, ok := request.Context[contextStepKey].(string)
	if !ok {
		common.SlackAttachmentError(w, errors.New("missing step name"))
		return
	}
	fromName := Name(fromString)
	buttonStr, ok := request.Context[contextButtonKey].(string)
	if !ok {
		common.SlackAttachmentError(w, errors.New("missing button id"))
		return
	}
	buttonIndex, err := strconv.Atoi(buttonStr)
	if err != nil {
		common.SlackAttachmentError(w, errors.Wrap(err, "invalid button number"))
		return
	}

	from, b, state, err := f.getStepButton(fromName, buttonIndex)
	if err != nil {
		common.SlackAttachmentError(w, err)
		return
	}
	appState := state.AppState

	toName := fromName
	if b.OnClick != nil {
		var updated State
		toName, updated = b.OnClick(f)
		appState = appState.MergeWith(updated)
	}
	// Empty next step name in the response indicates advancing to the next step
	// in the flow. To stay on the same step the handlers should return the step
	// name.
	if toName == "" {
		toName = f.next(fromName)
	}

	if b.Dialog != nil {
		if b.OnDialogSubmit == nil {
			common.SlackAttachmentError(w, errors.Errorf("no submit function for dialog, step: %s", fromName))
			return
		}

		d := *b.Dialog
		d.Title = formatState(d.Title, appState)
		d.IntroductionText = formatState(d.IntroductionText, appState)
		d.SubmitLabel = formatState(d.SubmitLabel, appState)
		for i := range d.Elements {
			d.Elements[i].DisplayName = formatState(d.Elements[i].DisplayName, f.appState)
			d.Elements[i].Name = formatState(d.Elements[i].Name, f.appState)
			d.Elements[i].Default = formatState(d.Elements[i].Default, f.appState)
			d.Elements[i].Placeholder = formatState(d.Elements[i].Placeholder, f.appState)
			d.Elements[i].HelpText = formatState(d.Elements[i].HelpText, f.appState)
		}

		dialogRequest := model.OpenDialogRequest{
			TriggerId: request.TriggerId,
			URL:       f.pluginURL + makePath(f.name) + "/dialog",
			Dialog:    d,
		}
		dialogRequest.Dialog.State = fmt.Sprintf("%v,%v", fromName, buttonIndex)

		err = f.api.Frontend.OpenInteractiveDialog(dialogRequest)
		if err != nil {
			common.SlackAttachmentError(w, err)
			return
		}
	}

	// if toName is different, render the "done" state of the "from" step.
	var post *model.Post
	if toName != fromName {
		post = f.renderAsPost(from.Name(), from.Render(appState, true, buttonIndex))
		post.Id = state.PostID
	}

	state.AppState = appState
	err = f.storeState(state)
	if err != nil {
		common.SlackAttachmentError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(model.PostActionIntegrationResponse{
		Update: post,
	})

	err = f.Go(toName)
	if err != nil {
		f.api.Log.Warn("failed to advance flow to next step", "flow_name", f.name, "from", fromName, "to", toName, "error", err.Error())
	}
}

func (f *UserFlow) handleDialog(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-ID")
	if userID == "" {
		common.DialogError(w, errors.New("not authorized"))
		return
	}
	f = f.forUser(userID)

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
	buttonIndex, err := strconv.Atoi(data[1])
	if err != nil {
		common.DialogError(w, errors.Wrap(err, "malformed button number"))
		return
	}

	from, b, state, err := f.getStepButton(fromName, buttonIndex)
	if err != nil {
		common.DialogError(w, err)
		return
	}

	toName := fromName
	if b.OnDialogSubmit != nil {
		next, updated, responseError, responseErrors := b.OnDialogSubmit(f, request.Submission)
		if responseError != "" || len(responseErrors) != 0 {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(model.SubmitDialogResponse{
				Error:  responseError,
				Errors: responseErrors,
			})
			return
		}
		state.AppState = state.AppState.MergeWith(updated)
		toName = next

		// if toName is different, render the "done" state of the "from" step.
		if toName != fromName {
			post := f.renderAsPost(from.Name(), from.Render(state.AppState, true, buttonIndex))
			post.Id = state.PostID
			err = f.api.Post.UpdatePost(post)
			if err != nil {
				common.DialogError(w, errors.Wrap(err, "Failed to update post"))
				return
			}
		}

		err = f.storeState(state)
		if err != nil {
			common.DialogError(w, err)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(model.SubmitDialogResponse{})

	// advance to next step if needed.
	err = f.Go(toName)
	if err != nil {
		f.api.Log.Warn("failed to advance to next step", "flow", f.name, "step", toName)
	}
}

func (f *UserFlow) getStepButton(fromName Name, buttonIndex int) (Step, Button, flowState, error) {
	state, err := f.getState()
	if err != nil {
		return nil, Button{}, flowState{}, err
	}
	if state.StepName != fromName {
		return nil, Button{}, flowState{}, errors.Errorf("click from an inactive step: %v", fromName)
	}
	from := f.steps[fromName]
	if from == nil {
		return nil, Button{}, flowState{}, errors.Errorf("step %q not found", fromName)
	}

	// render the "active" state of the current step to validate the button
	// index.
	appState := state.AppState
	buttons := f.Buttons(from, appState)
	if buttonIndex > len(buttons)-1 {
		return nil, Button{}, flowState{}, errors.Errorf("button number %v to high, only %v buttons", buttonIndex, len(buttons))
	}

	return from, buttons[buttonIndex], state, nil
}
