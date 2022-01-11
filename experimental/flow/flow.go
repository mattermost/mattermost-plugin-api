package flow

import (
	"github.com/mattermost/mattermost-plugin-api/experimental/flow/steps"
)

type Flow interface {
	Steps() []steps.Step
	Step(i int) steps.Step
	Path() string
	Length() int
	FlowDone(userID string)
}

type flow struct {
	steps      []steps.Step
	path       string
	onFlowDone func(userID string)
}

func NewFlow(stepList []steps.Step, path string, onFlowDone func(userID string)) Flow {
	f := &flow{
		steps:      stepList,
		path:       path,
		onFlowDone: onFlowDone,
	}
	return f
}

func (f *flow) Steps() []steps.Step {
	return f.steps
}

func (f *flow) Step(i int) steps.Step {
	if i < 1 {
		return nil
	}
	if i > len(f.steps) {
		return nil
	}
	return f.steps[i-1]
}

func (f *flow) Path() string {
	return f.path
}

func (f *flow) Length() int {
	return len(f.steps)
}

func (f *flow) FlowDone(userID string) {
	if f.onFlowDone != nil {
		f.onFlowDone(userID)
	}
}
