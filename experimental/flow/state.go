package flow

import (
	"bytes"
	"errors"
	"html/template"
)

// State is the "app"'s state
type State map[string]string

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

// JSON-serializable flow state.
type flowState struct {
	// The name of the step.
	StepName Name

	Done bool

	// ID of the post produced by the step.
	PostID string

	// Application-level state.
	AppState State
}

func (f *Flow) storeState(state flowState) error {
	if f.UserID == "" {
		return errors.New("no user specified")
	}
	ok, err := f.api.KV.Set(kvKey(f.UserID, f.name), state)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("value not set without errors")
	}

	f.State = state.AppState
	return nil
}

func (f *Flow) getState() (flowState, error) {
	if f.UserID == "" {
		return flowState{}, errors.New("no user specified")
	}
	state := flowState{}
	err := f.api.KV.Get(kvKey(f.UserID, f.name), &state)
	if err != nil {
		return flowState{}, err
	}
	if state.AppState == nil {
		return flowState{}, errors.New("flow not found")
	}

	f.State = state.AppState
	return state, err
}

func (f *Flow) removeState() error {
	if f.UserID == "" {
		return errors.New("no user specified")
	}
	f.State = State{}
	return f.api.KV.Delete(kvKey(f.UserID, f.name))
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
