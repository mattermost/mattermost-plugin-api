package flow

import (
	"bytes"
	"errors"
	"html/template"
)

// State is the "app"'s state
type State map[string]string

// JSON-serializable flow state.
type flowState struct {
	// The name of the step.
	StepName Name

	// ID of the post produced by the step.
	PostID string

	// Application-level state.
	AppState State
}

func (f *UserFlow) State() (_ State, userID string) {
	return f.appState.MergeWith(nil), f.userID
}

func (s State) MergeWith(update State) State {
	n := State{}
	for k, v := range s {
		n[k] = v
	}
	for k, v := range update {
		n[k] = v
	}
	return n
}

func (f *UserFlow) storeState(state flowState) error {
	if f.userID == "" {
		return errors.New("no user specified")
	}
	ok, err := f.api.KV.Set(kvKey(f.userID, f.name), state)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("value not set without errors")
	}

	f.appState = state.AppState
	return nil
}

func (f *UserFlow) getState() (flowState, error) {
	if f.userID == "" {
		return flowState{}, errors.New("no user specified")
	}
	state := flowState{}
	err := f.api.KV.Get(kvKey(f.userID, f.name), &state)
	if err != nil {
		return flowState{}, err
	}
	if state.AppState == nil {
		state.AppState = State{}
	}

	f.appState = state.AppState
	return state, err
}

func (f *UserFlow) removeState() error {
	if f.userID == "" {
		return errors.New("no user specified")
	}
	return f.api.KV.Delete(kvKey(f.userID, f.name))
}

func kvKey(userID string, flowName Name) string {
	return "_flow-" + userID + "-" + string(flowName)
}

func formatState(source string, state State) string {
	t, err := template.New("message").Parse(source)
	if err != nil {
		return source + " ###ERROR: " + err.Error()
	}
	buf := bytes.NewBuffer(nil)
	err = t.Execute(buf, state)
	if err != nil {
		return source + " ###ERROR: " + err.Error()
	}
	return buf.String()
}
