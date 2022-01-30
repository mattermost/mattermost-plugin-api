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

func (f *UserFlow) handleSubmit(userID string, fromName Name, buttonIndex int, onClickFunc func(Button, State) (Name, State, error)) (Name, *model.Post, error) {
	state, err := f.getState(userID)
	if err != nil {
		return "", nil, err
	}
	if state.StepName != fromName {
		return "", nil, errors.Errorf("click from an inactive step: %v")
	}
	from := f.Step(fromName)
	if from == nil {
		return "", nil, errors.Errorf("step %q not found", fromName)
	}

	// render the "active" state of the current step to validate the button
	// index.
	appState := state.AppState
	buttons := f.Buttons(from, appState)
	if buttonIndex > len(buttons)-1 {
		return "", nil, errors.Errorf("button number %v to high, only %v buttons", buttonIndex, len(buttons))
	}

	toName, appState, err := onClickFunc(buttons[buttonIndex], appState)
	if err != nil {
		return "", nil, err
	}

	// Empty next step name in the response indicates advancing to the next step
	// in the flow. To stay on the same step the handlers should return the step
	// name.
	if toName == "" {
		toName = f.next(fromName)
	}

	// if toName is different, render the "done" state of the "from" step.
	var post *model.Post
	if toName != fromName {
		post = f.renderAsPost(from.Name(), from.Render(appState, f.pluginURL, true, buttonIndex))
		post.Id = state.PostID
	}

	state.AppState = appState
	err = f.storeState(userID, state)
	if err != nil {
		return "", nil, err
	}

	return toName, post, nil
}

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
	i, err := strconv.Atoi(buttonStr)
	if err != nil {
		common.SlackAttachmentError(w, errors.Wrap(err, "invalid button number"))
		return
	}

	toName, updatedPost, err := f.handleSubmit(userID, fromName, i,
		func(b Button, appState State) (Name, State, error) {
			next := fromName
			if b.OnClick != nil {
				next, appState = b.OnClick(userID, appState)
			}

			if b.Dialog != nil {
				if b.OnDialogSubmit == nil {
					return "", nil, errors.Errorf("no submit function for dialog, step: %s", fromName)
				}

				dialogRequest := model.OpenDialogRequest{
					TriggerId: request.TriggerId,
					URL:       f.pluginURL + makePath(f.Name) + "/dialog",
					Dialog:    *b.Dialog,
				}
				dialogRequest.Dialog.State = fmt.Sprintf("%v,%v", fromName, i)

				err = f.api.Frontend.OpenInteractiveDialog(dialogRequest)
				if err != nil {
					return "", nil, err
				}
			}
			return next, appState, nil
		})
	if err != nil {
		common.SlackAttachmentError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(model.PostActionIntegrationResponse{
		Update: updatedPost,
	})

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
	buttonIndex, err := strconv.Atoi(data[1])
	if err != nil {
		common.DialogError(w, errors.Wrap(err, "malformed button number"))
		return
	}

	responseError := ""
	var responseErrors map[string]string
	toName, updatedPost, err := f.handleSubmit(userID, fromName, buttonIndex,
		func(b Button, appState State) (Name, State, error) {
			next := fromName
			if b.OnDialogSubmit != nil {
				next, appState, responseError, responseErrors = b.OnDialogSubmit(userID, request.Submission, appState)
			}
			return next, appState, nil
		},
	)
	if err != nil {
		common.DialogError(w, err)
		return
	}

	if responseError == "" && len(responseErrors) == 0 && updatedPost != nil {
		err = f.api.Post.UpdatePost(updatedPost)
		if err != nil {
			common.DialogError(w, errors.Wrap(err, "Failed to update post"))
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(model.SubmitDialogResponse{
		Error:  responseError,
		Errors: responseErrors,
	})
	if responseError != "" || len(responseErrors) > 0 {
		return
	}

	// advance to next step if needed.
	err = f.Go(userID, toName)
	if err != nil {
		f.api.Log.Warn("failed to advance to next step", "flow", f.Name, "step", toName)
	}
}
