package api

//go:generate mockgen -source=./client.go -destination=./mock.go -package api ClientInterface
//go:generate mockgen -source=./client.go -destination=./mock.go -package api ClientWithResponsesInterface

import "github.com/golang/mock/gomock"

func (m *MockClientInterface) Reset(ctrl *gomock.Controller) {
	m.ctrl = ctrl
	m.recorder = &MockClientInterfaceMockRecorder{mock: m}
}

func (m *MockClientWithResponsesInterface) Reset(ctrl *gomock.Controller) {
	m.ctrl = ctrl
	m.recorder = &MockClientWithResponsesInterfaceMockRecorder{mock: m}
}
