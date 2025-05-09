// Code generated by MockGen. DO NOT EDIT.
// Source: shortener/internal/user (interfaces: UserService)

// Package mocks is a generated GoMock package.
package mocks

import (
	http "net/http"
	reflect "reflect"
	user "shortener/internal/user"

	gomock "github.com/golang/mock/gomock"
)

// MockUserService is a mock of UserService interface.
type MockUserService struct {
	ctrl     *gomock.Controller
	recorder *MockUserServiceMockRecorder
}

// MockUserServiceMockRecorder is the mock recorder for MockUserService.
type MockUserServiceMockRecorder struct {
	mock *MockUserService
}

// NewMockUserService creates a new mock instance.
func NewMockUserService(ctrl *gomock.Controller) *MockUserService {
	mock := &MockUserService{ctrl: ctrl}
	mock.recorder = &MockUserServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUserService) EXPECT() *MockUserServiceMockRecorder {
	return m.recorder
}

// AddURLs mocks base method.
func (m *MockUserService) AddURLs(arg0, arg1, arg2, arg3 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddURLs", arg0, arg1, arg2, arg3)
}

// AddURLs indicates an expected call of AddURLs.
func (mr *MockUserServiceMockRecorder) AddURLs(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddURLs", reflect.TypeOf((*MockUserService)(nil).AddURLs), arg0, arg1, arg2, arg3)
}

// GetUserIDFromCookie mocks base method.
func (m *MockUserService) GetUserIDFromCookie(arg0 *http.Request) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserIDFromCookie", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserIDFromCookie indicates an expected call of GetUserIDFromCookie.
func (mr *MockUserServiceMockRecorder) GetUserIDFromCookie(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserIDFromCookie", reflect.TypeOf((*MockUserService)(nil).GetUserIDFromCookie), arg0)
}

// GetUserURLs mocks base method.
func (m *MockUserService) GetUserURLs(arg0 string) ([]user.UserURL, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserURLs", arg0)
	ret0, _ := ret[0].([]user.UserURL)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// GetUserURLs indicates an expected call of GetUserURLs.
func (mr *MockUserServiceMockRecorder) GetUserURLs(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserURLs", reflect.TypeOf((*MockUserService)(nil).GetUserURLs), arg0)
}

// InitUserURLs mocks base method.
func (m *MockUserService) InitUserURLs(arg0 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "InitUserURLs", arg0)
}

// InitUserURLs indicates an expected call of InitUserURLs.
func (mr *MockUserServiceMockRecorder) InitUserURLs(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InitUserURLs", reflect.TypeOf((*MockUserService)(nil).InitUserURLs), arg0)
}

// SetUserIDCookie mocks base method.
func (m *MockUserService) SetUserIDCookie(arg0 http.ResponseWriter, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetUserIDCookie", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetUserIDCookie indicates an expected call of SetUserIDCookie.
func (mr *MockUserServiceMockRecorder) SetUserIDCookie(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetUserIDCookie", reflect.TypeOf((*MockUserService)(nil).SetUserIDCookie), arg0, arg1)
}
