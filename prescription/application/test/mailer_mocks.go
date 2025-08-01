// Code generated by MockGen. DO NOT EDIT.
// Source: mailer.go
//
// Generated by this command:
//
//	mockgen -source=mailer.go -destination=mailer_mocks.go -package=test MailerClient
//

// Package test is a generated GoMock package.
package test

import (
	context "context"
	reflect "reflect"

	events "github.com/tidepool-org/go-common/events"
	gomock "go.uber.org/mock/gomock"
)

// MockMailerClient is a mock of MailerClient interface.
type MockMailerClient struct {
	ctrl     *gomock.Controller
	recorder *MockMailerClientMockRecorder
	isgomock struct{}
}

// MockMailerClientMockRecorder is the mock recorder for MockMailerClient.
type MockMailerClientMockRecorder struct {
	mock *MockMailerClient
}

// NewMockMailerClient creates a new mock instance.
func NewMockMailerClient(ctrl *gomock.Controller) *MockMailerClient {
	mock := &MockMailerClient{ctrl: ctrl}
	mock.recorder = &MockMailerClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMailerClient) EXPECT() *MockMailerClientMockRecorder {
	return m.recorder
}

// SendEmailTemplate mocks base method.
func (m *MockMailerClient) SendEmailTemplate(arg0 context.Context, arg1 events.SendEmailTemplateEvent) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendEmailTemplate", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendEmailTemplate indicates an expected call of SendEmailTemplate.
func (mr *MockMailerClientMockRecorder) SendEmailTemplate(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendEmailTemplate", reflect.TypeOf((*MockMailerClient)(nil).SendEmailTemplate), arg0, arg1)
}
