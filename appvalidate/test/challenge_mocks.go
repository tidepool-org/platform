// Code generated by MockGen. DO NOT EDIT.
// Source: challenge.go
//
// Generated by this command:
//
//	mockgen -source=challenge.go -destination=test/challenge_mocks.go -package=test ChallengeGenerator
//

// Package test is a generated GoMock package.
package test

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockChallengeGenerator is a mock of ChallengeGenerator interface.
type MockChallengeGenerator struct {
	ctrl     *gomock.Controller
	recorder *MockChallengeGeneratorMockRecorder
	isgomock struct{}
}

// MockChallengeGeneratorMockRecorder is the mock recorder for MockChallengeGenerator.
type MockChallengeGeneratorMockRecorder struct {
	mock *MockChallengeGenerator
}

// NewMockChallengeGenerator creates a new mock instance.
func NewMockChallengeGenerator(ctrl *gomock.Controller) *MockChallengeGenerator {
	mock := &MockChallengeGenerator{ctrl: ctrl}
	mock.recorder = &MockChallengeGeneratorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockChallengeGenerator) EXPECT() *MockChallengeGeneratorMockRecorder {
	return m.recorder
}

// GenerateChallenge mocks base method.
func (m *MockChallengeGenerator) GenerateChallenge(size int) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenerateChallenge", size)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GenerateChallenge indicates an expected call of GenerateChallenge.
func (mr *MockChallengeGeneratorMockRecorder) GenerateChallenge(size any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateChallenge", reflect.TypeOf((*MockChallengeGenerator)(nil).GenerateChallenge), size)
}
