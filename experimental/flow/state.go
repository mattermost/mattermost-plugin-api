package flow

import (
	"errors"
)

// JSON-serializable flow state.
type flowState struct {
	// The name of the step.
	StepName Name

	// ID of the post produced by the step.
	PostID string

	// Application-level state.
	AppState State
}

func (f *UserFlow) storeState(userID string, state flowState) error {
	ok, err := f.api.KV.Set(kvKey(userID, f.Name), state)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("value not set without errors")
	}
	return nil
}

func (f *UserFlow) getState(userID string) (flowState, error) {
	state := flowState{}
	err := f.api.KV.Get(kvKey(userID, f.Name), &state)
	return state, err
}

func (f *UserFlow) removeState(userID string) error {
	return f.api.KV.Delete(kvKey(userID, f.Name))
}

func kvKey(userID string, flowName Name) string {
	return "_flow-" + userID + "-" + string(flowName)
}
