// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/mattermost/mattermost-plugin-api/experimental/freetextfetcher (interfaces: FreetextFetcher)

// Package mock_freetext_fetcher is a generated GoMock package.
package mock_freetext_fetcher

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	logger "github.com/mattermost/mattermost-plugin-api/experimental/bot/logger"
	model "github.com/mattermost/mattermost-server/v6/model"
	plugin "github.com/mattermost/mattermost-server/v6/plugin"
)

// MockFreetextFetcher is a mock of FreetextFetcher interface.
type MockFreetextFetcher struct {
	ctrl     *gomock.Controller
	recorder *MockFreetextFetcherMockRecorder
}

// MockFreetextFetcherMockRecorder is the mock recorder for MockFreetextFetcher.
type MockFreetextFetcherMockRecorder struct {
	mock *MockFreetextFetcher
}

// NewMockFreetextFetcher creates a new mock instance.
func NewMockFreetextFetcher(ctrl *gomock.Controller) *MockFreetextFetcher {
	mock := &MockFreetextFetcher{ctrl: ctrl}
	mock.recorder = &MockFreetextFetcherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFreetextFetcher) EXPECT() *MockFreetextFetcherMockRecorder {
	return m.recorder
}

// MessageHasBeenPosted mocks base method.
func (m *MockFreetextFetcher) MessageHasBeenPosted(arg0 *plugin.Context, arg1 *model.Post, arg2 plugin.API, arg3 logger.Logger, arg4, arg5 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "MessageHasBeenPosted", arg0, arg1, arg2, arg3, arg4, arg5)
}

// MessageHasBeenPosted indicates an expected call of MessageHasBeenPosted.
func (mr *MockFreetextFetcherMockRecorder) MessageHasBeenPosted(arg0, arg1, arg2, arg3, arg4, arg5 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MessageHasBeenPosted", reflect.TypeOf((*MockFreetextFetcher)(nil).MessageHasBeenPosted), arg0, arg1, arg2, arg3, arg4, arg5)
}

// StartFetching mocks base method.
func (m *MockFreetextFetcher) StartFetching(arg0, arg1 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "StartFetching", arg0, arg1)
}

// StartFetching indicates an expected call of StartFetching.
func (mr *MockFreetextFetcherMockRecorder) StartFetching(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartFetching", reflect.TypeOf((*MockFreetextFetcher)(nil).StartFetching), arg0, arg1)
}

// URL mocks base method.
func (m *MockFreetextFetcher) URL() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "URL")
	ret0, _ := ret[0].(string)
	return ret0
}

// URL indicates an expected call of URL.
func (mr *MockFreetextFetcherMockRecorder) URL() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "URL", reflect.TypeOf((*MockFreetextFetcher)(nil).URL))
}

// UpdateHooks mocks base method.
func (m *MockFreetextFetcher) UpdateHooks(arg0 func(string) string, arg1 func(string, string), arg2 func(string)) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "UpdateHooks", arg0, arg1, arg2)
}

// UpdateHooks indicates an expected call of UpdateHooks.
func (mr *MockFreetextFetcherMockRecorder) UpdateHooks(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateHooks", reflect.TypeOf((*MockFreetextFetcher)(nil).UpdateHooks), arg0, arg1, arg2)
}
