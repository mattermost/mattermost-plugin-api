package flow

import (
	"github.com/mattermost/mattermost-plugin-api/utils/flow/steps"
)

type Flow interface {
	Steps() []steps.Step
	Step(i int) steps.Step
	URL() string
	Length() int
	StepDone(userID string, step int, value interface{})
	FlowDone(userID string)
}

type flow struct {
	steps      []steps.Step
	url        string
	controller FlowController
	onFlowDone func(userID string)
}

func NewFlow(stepList []steps.Step, url string, fc FlowController, onFlowDone func(userID string)) Flow {
	f := &flow{
		steps:      stepList,
		url:        url,
		controller: fc,
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

func (f *flow) URL() string {
	return f.url
}

func (f *flow) Length() int {
	return len(f.steps)
}

func (f *flow) StepDone(userID string, step int, value interface{}) {
	f.controller.NextStep(userID, step, value)
}

func (f *flow) FlowDone(userID string) {
	if f.onFlowDone != nil {
		f.onFlowDone(userID)
	}
}
