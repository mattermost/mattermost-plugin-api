package flow

import (
	"errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
)

// JSON-serializable flow state.
type flowState struct {
	// The name of the step.
	StepName Name

	// Each done step produces a post, some add the index of selected button.
	Done   bool
	PostID string
	Button int

	// Application-level state.
	AppState State
}

func storeState(kv *pluginapi.KVService, userID string, flowName Name, state flowState) error {
	ok, err := kv.Set(kvKey(userID, flowName), state)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("value not set without errors")
	}
	return nil
}

func getState(kv *pluginapi.KVService, userID string, flowName Name) (flowState, error) {
	state := flowState{}
	err := kv.Get(kvKey(userID, flowName), &state)
	return state, err
}

func removeState(kv *pluginapi.KVService, userID string, flowName Name) error {
	return kv.Delete(kvKey(userID, flowName))
}

func kvKey(userID string, flowName Name) string {
	return "_flow-" + userID + "-" + string(flowName)
}
