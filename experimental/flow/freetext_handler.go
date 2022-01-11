package flow

import (
	"encoding/json"
)

type freetextInfo struct {
	Step     int
	Property string
	UserID   string
}

func (fc *flowController) ftOnFetch(message, payload string) {
	var ftInfo freetextInfo
	err := json.Unmarshal([]byte(payload), &ftInfo)
	if err != nil {
		fc.Logger.Warnf("cannot unmarshal free text info, err=%s", err)
		return
	}

	err = fc.SetProperty(ftInfo.UserID, ftInfo.Property, message)
	if err != nil {
		fc.Logger.Warnf("cannot set free text property %s, err=%s", ftInfo.Property, err)
		return
	}

	step := fc.GetFlow().Step(ftInfo.Step)
	if step == nil {
		fc.Logger.Warnf("There is no step %d.", step)
		return
	}

	skip := step.ShouldSkip(message)

	_ = fc.store.RemovePostID(ftInfo.UserID, ftInfo.Property)
	_ = fc.NextStep(ftInfo.UserID, ftInfo.Step, skip)
}

func (fc *flowController) ftOnCancel(payload string) {
	var ftInfo freetextInfo
	err := json.Unmarshal([]byte(payload), &ftInfo)
	if err != nil {
		fc.Logger.Errorf("cannot unmarshal free text info, err=%s", err)
		return
	}

	_ = fc.store.RemovePostID(ftInfo.UserID, ftInfo.Property)
	_ = fc.NextStep(ftInfo.UserID, ftInfo.Step, 0)
}
