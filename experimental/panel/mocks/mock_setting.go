// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/mattermost/mattermost-plugin-api/experimental/panel/settings (interfaces: Setting)

// Package mock_panel is a generated GoMock package.
package mock_panel

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	freetextfetcher "github.com/mattermost/mattermost-plugin-api/experimental/freetextfetcher"
	model "github.com/mattermost/mattermost-server/v6/model"
)

// MockSetting is a mock of Setting interface.
type MockSetting struct {
	ctrl     *gomock.Controller
	recorder *MockSettingMockRecorder
}

// MockSettingMockRecorder is the mock recorder for MockSetting.
type MockSettingMockRecorder struct {
	mock *MockSetting
}

// NewMockSetting creates a new mock instance.
func NewMockSetting(ctrl *gomock.Controller) *MockSetting {
	mock := &MockSetting{ctrl: ctrl}
	mock.recorder = &MockSettingMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSetting) EXPECT() *MockSettingMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockSetting) Get(arg0 string) (interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0)
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockSettingMockRecorder) Get(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockSetting)(nil).Get), arg0)
}

// GetDependency mocks base method.
func (m *MockSetting) GetDependency() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDependency")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetDependency indicates an expected call of GetDependency.
func (mr *MockSettingMockRecorder) GetDependency() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDependency", reflect.TypeOf((*MockSetting)(nil).GetDependency))
}

// GetDescription mocks base method.
func (m *MockSetting) GetDescription() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDescription")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetDescription indicates an expected call of GetDescription.
func (mr *MockSettingMockRecorder) GetDescription() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDescription", reflect.TypeOf((*MockSetting)(nil).GetDescription))
}

// GetFreetextFetcher mocks base method.
func (m *MockSetting) GetFreetextFetcher() freetextfetcher.FreetextFetcher {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFreetextFetcher")
	ret0, _ := ret[0].(freetextfetcher.FreetextFetcher)
	return ret0
}

// GetFreetextFetcher indicates an expected call of GetFreetextFetcher.
func (mr *MockSettingMockRecorder) GetFreetextFetcher() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFreetextFetcher", reflect.TypeOf((*MockSetting)(nil).GetFreetextFetcher))
}

// GetID mocks base method.
func (m *MockSetting) GetID() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetID")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetID indicates an expected call of GetID.
func (mr *MockSettingMockRecorder) GetID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetID", reflect.TypeOf((*MockSetting)(nil).GetID))
}

// GetSlackAttachments mocks base method.
func (m *MockSetting) GetSlackAttachments(arg0, arg1 string, arg2 bool) (*model.SlackAttachment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSlackAttachments", arg0, arg1, arg2)
	ret0, _ := ret[0].(*model.SlackAttachment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSlackAttachments indicates an expected call of GetSlackAttachments.
func (mr *MockSettingMockRecorder) GetSlackAttachments(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSlackAttachments", reflect.TypeOf((*MockSetting)(nil).GetSlackAttachments), arg0, arg1, arg2)
}

// GetTitle mocks base method.
func (m *MockSetting) GetTitle() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTitle")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetTitle indicates an expected call of GetTitle.
func (mr *MockSettingMockRecorder) GetTitle() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTitle", reflect.TypeOf((*MockSetting)(nil).GetTitle))
}

// IsDisabled mocks base method.
func (m *MockSetting) IsDisabled(arg0 interface{}) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsDisabled", arg0)
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsDisabled indicates an expected call of IsDisabled.
func (mr *MockSettingMockRecorder) IsDisabled(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsDisabled", reflect.TypeOf((*MockSetting)(nil).IsDisabled), arg0)
}

// Set mocks base method.
func (m *MockSetting) Set(arg0 string, arg1 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Set", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Set indicates an expected call of Set.
func (mr *MockSettingMockRecorder) Set(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Set", reflect.TypeOf((*MockSetting)(nil).Set), arg0, arg1)
}
