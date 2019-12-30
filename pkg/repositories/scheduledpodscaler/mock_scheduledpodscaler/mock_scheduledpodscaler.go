// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/int128/scheduled-scaler/pkg/repositories/scheduledpodscaler (interfaces: Interface)

// Package mock_scheduledpodscaler is a generated GoMock package.
package mock_scheduledpodscaler

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	v1 "github.com/int128/scheduled-scaler/api/v1"
	types "k8s.io/apimachinery/pkg/types"
	reflect "reflect"
)

// MockInterface is a mock of Interface interface
type MockInterface struct {
	ctrl     *gomock.Controller
	recorder *MockInterfaceMockRecorder
}

// MockInterfaceMockRecorder is the mock recorder for MockInterface
type MockInterfaceMockRecorder struct {
	mock *MockInterface
}

// NewMockInterface creates a new mock instance
func NewMockInterface(ctrl *gomock.Controller) *MockInterface {
	mock := &MockInterface{ctrl: ctrl}
	mock.recorder = &MockInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockInterface) EXPECT() *MockInterfaceMockRecorder {
	return m.recorder
}

// Get mocks base method
func (m *MockInterface) Get(arg0 context.Context, arg1 types.NamespacedName) (*v1.ScheduledPodScaler, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0, arg1)
	ret0, _ := ret[0].(*v1.ScheduledPodScaler)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get
func (mr *MockInterfaceMockRecorder) Get(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockInterface)(nil).Get), arg0, arg1)
}

// UpdateStatus mocks base method
func (m *MockInterface) UpdateStatus(arg0 context.Context, arg1 *v1.ScheduledPodScaler) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateStatus", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateStatus indicates an expected call of UpdateStatus
func (mr *MockInterfaceMockRecorder) UpdateStatus(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateStatus", reflect.TypeOf((*MockInterface)(nil).UpdateStatus), arg0, arg1)
}
