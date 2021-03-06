// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/shirasudon/go-chat/chat (interfaces: QueryService)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	action "github.com/shirasudon/go-chat/chat/action"
	queried "github.com/shirasudon/go-chat/chat/queried"
	reflect "reflect"
)

// MockQueryService is a mock of QueryService interface
type MockQueryService struct {
	ctrl     *gomock.Controller
	recorder *MockQueryServiceMockRecorder
}

// MockQueryServiceMockRecorder is the mock recorder for MockQueryService
type MockQueryServiceMockRecorder struct {
	mock *MockQueryService
}

// NewMockQueryService creates a new mock instance
func NewMockQueryService(ctrl *gomock.Controller) *MockQueryService {
	mock := &MockQueryService{ctrl: ctrl}
	mock.recorder = &MockQueryServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockQueryService) EXPECT() *MockQueryServiceMockRecorder {
	return m.recorder
}

// FindRoomInfo mocks base method
func (m *MockQueryService) FindRoomInfo(arg0 context.Context, arg1, arg2 uint64) (*queried.RoomInfo, error) {
	ret := m.ctrl.Call(m, "FindRoomInfo", arg0, arg1, arg2)
	ret0, _ := ret[0].(*queried.RoomInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindRoomInfo indicates an expected call of FindRoomInfo
func (mr *MockQueryServiceMockRecorder) FindRoomInfo(arg0, arg1, arg2 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindRoomInfo", reflect.TypeOf((*MockQueryService)(nil).FindRoomInfo), arg0, arg1, arg2)
}

// FindRoomMessages mocks base method
func (m *MockQueryService) FindRoomMessages(arg0 context.Context, arg1 uint64, arg2 action.QueryRoomMessages) (*queried.RoomMessages, error) {
	ret := m.ctrl.Call(m, "FindRoomMessages", arg0, arg1, arg2)
	ret0, _ := ret[0].(*queried.RoomMessages)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindRoomMessages indicates an expected call of FindRoomMessages
func (mr *MockQueryServiceMockRecorder) FindRoomMessages(arg0, arg1, arg2 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindRoomMessages", reflect.TypeOf((*MockQueryService)(nil).FindRoomMessages), arg0, arg1, arg2)
}

// FindUnreadRoomMessages mocks base method
func (m *MockQueryService) FindUnreadRoomMessages(arg0 context.Context, arg1 uint64, arg2 action.QueryUnreadRoomMessages) (*queried.UnreadRoomMessages, error) {
	ret := m.ctrl.Call(m, "FindUnreadRoomMessages", arg0, arg1, arg2)
	ret0, _ := ret[0].(*queried.UnreadRoomMessages)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindUnreadRoomMessages indicates an expected call of FindUnreadRoomMessages
func (mr *MockQueryServiceMockRecorder) FindUnreadRoomMessages(arg0, arg1, arg2 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindUnreadRoomMessages", reflect.TypeOf((*MockQueryService)(nil).FindUnreadRoomMessages), arg0, arg1, arg2)
}

// FindUserByNameAndPassword mocks base method
func (m *MockQueryService) FindUserByNameAndPassword(arg0 context.Context, arg1, arg2 string) (*queried.AuthUser, error) {
	ret := m.ctrl.Call(m, "FindUserByNameAndPassword", arg0, arg1, arg2)
	ret0, _ := ret[0].(*queried.AuthUser)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindUserByNameAndPassword indicates an expected call of FindUserByNameAndPassword
func (mr *MockQueryServiceMockRecorder) FindUserByNameAndPassword(arg0, arg1, arg2 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindUserByNameAndPassword", reflect.TypeOf((*MockQueryService)(nil).FindUserByNameAndPassword), arg0, arg1, arg2)
}

// FindUserRelation mocks base method
func (m *MockQueryService) FindUserRelation(arg0 context.Context, arg1 uint64) (*queried.UserRelation, error) {
	ret := m.ctrl.Call(m, "FindUserRelation", arg0, arg1)
	ret0, _ := ret[0].(*queried.UserRelation)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindUserRelation indicates an expected call of FindUserRelation
func (mr *MockQueryServiceMockRecorder) FindUserRelation(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindUserRelation", reflect.TypeOf((*MockQueryService)(nil).FindUserRelation), arg0, arg1)
}
