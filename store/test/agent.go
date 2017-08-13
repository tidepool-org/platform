package test

import "github.com/tidepool-org/platform/test"

type Agent struct {
	*test.Mock
	IsServerInvocations int
	IsServerOutputs     []bool
	UserIDInvocations   int
	UserIDOutputs       []string
}

func NewAgent() *Agent {
	return &Agent{
		Mock: test.NewMock(),
	}
}

func (a *Agent) IsServer() bool {
	a.IsServerInvocations++

	if len(a.IsServerOutputs) == 0 {
		panic("Unexpected invocation of IsServer on Agent")
	}

	output := a.IsServerOutputs[0]
	a.IsServerOutputs = a.IsServerOutputs[1:]
	return output
}

func (a *Agent) UserID() string {
	a.UserIDInvocations++

	if len(a.UserIDOutputs) == 0 {
		panic("Unexpected invocation of UserID on Agent")
	}

	output := a.UserIDOutputs[0]
	a.UserIDOutputs = a.UserIDOutputs[1:]
	return output
}

func (a *Agent) UnusedOutputsCount() int {
	return len(a.IsServerOutputs) +
		len(a.UserIDOutputs)
}
