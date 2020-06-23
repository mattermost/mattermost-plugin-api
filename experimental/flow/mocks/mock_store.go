// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/mattermost/mattermost-plugin-api/experimental/flow (interfaces: Store)

// Package mock_flow is a generated GoMock package.
package mock_flow

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockStore is a mock of Store interface
type MockStore struct {
	ctrl     *gomock.Controller
	recorder *MockStoreMockRecorder
}

// MockStoreMockRecorder is the mock recorder for MockStore
type MockStoreMockRecorder struct {
	mock *MockStore
}

// NewMockStore creates a new mock instance
func NewMockStore(ctrl *gomock.Controller) *MockStore {
	mock := &MockStore{ctrl: ctrl}
	mock.recorder = &MockStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockStore) EXPECT() *MockStoreMockRecorder {
	return m.recorder
}

// DeleteCurrentStep mocks base method
func (m *MockStore) DeleteCurrentStep(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteCurrentStep", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteCurrentStep indicates an expected call of DeleteCurrentStep
func (mr *MockStoreMockRecorder) DeleteCurrentStep(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteCurrentStep", reflect.TypeOf((*MockStore)(nil).DeleteCurrentStep), arg0)
}

// GetCurrentStep mocks base method
func (m *MockStore) GetCurrentStep(arg0 string) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCurrentStep", arg0)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCurrentStep indicates an expected call of GetCurrentStep
func (mr *MockStoreMockRecorder) GetCurrentStep(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCurrentStep", reflect.TypeOf((*MockStore)(nil).GetCurrentStep), arg0)
}

// GetPostID mocks base method
func (m *MockStore) GetPostID(arg0, arg1 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPostID", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPostID indicates an expected call of GetPostID
func (mr *MockStoreMockRecorder) GetPostID(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPostID", reflect.TypeOf((*MockStore)(nil).GetPostID), arg0, arg1)
}

// RemovePostID mocks base method
func (m *MockStore) RemovePostID(arg0, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemovePostID", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemovePostID indicates an expected call of RemovePostID
func (mr *MockStoreMockRecorder) RemovePostID(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemovePostID", reflect.TypeOf((*MockStore)(nil).RemovePostID), arg0, arg1)
}

// SetCurrentStep mocks base method
func (m *MockStore) SetCurrentStep(arg0 string, arg1 int) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetCurrentStep", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetCurrentStep indicates an expected call of SetCurrentStep
func (mr *MockStoreMockRecorder) SetCurrentStep(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetCurrentStep", reflect.TypeOf((*MockStore)(nil).SetCurrentStep), arg0, arg1)
}

// SetPostID mocks base method
func (m *MockStore) SetPostID(arg0, arg1, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetPostID", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetPostID indicates an expected call of SetPostID
func (mr *MockStoreMockRecorder) SetPostID(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetPostID", reflect.TypeOf((*MockStore)(nil).SetPostID), arg0, arg1, arg2)
}
