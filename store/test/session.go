package test

import (
	"github.com/tidepool-org/platform/log"
	nullLog "github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/store"
	"github.com/tidepool-org/platform/test"
)

type Session struct {
	*test.Mock
	IsClosedInvocations int
	IsClosedOutputs     []bool
	CloseInvocations    int
	LoggerInvocations   int
	LoggerImpl          log.Logger
	SetAgentInvocations int
	SetAgentInputs      []store.Agent
}

func NewSession() *Session {
	return &Session{
		Mock:       test.NewMock(),
		LoggerImpl: nullLog.NewLogger(),
	}
}

func (s *Session) IsClosed() bool {
	s.IsClosedInvocations++

	if len(s.IsClosedOutputs) == 0 {
		panic("Unexpected invocation of IsClosed on Session")
	}

	output := s.IsClosedOutputs[0]
	s.IsClosedOutputs = s.IsClosedOutputs[1:]
	return output
}

func (s *Session) Close() {
	s.CloseInvocations++
}

func (s *Session) Logger() log.Logger {
	s.LoggerInvocations++

	return s.LoggerImpl
}

func (s *Session) SetAgent(agent store.Agent) {
	s.SetAgentInvocations++

	s.SetAgentInputs = append(s.SetAgentInputs, agent)
}

func (s *Session) UnusedOutputsCount() int {
	return len(s.IsClosedOutputs)
}
