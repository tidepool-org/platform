package test

import (
	"context"

	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/test"
)

type ServerSessionTokenOutput struct {
	Token string
	Error error
}

type ValidateSessionTokenInput struct {
	Context context.Context
	Token   string
}

type ValidateSessionTokenOutput struct {
	Details request.Details
	Error   error
}

type ExternalAccessor struct {
	*test.Mock
	ServerSessionTokenInvocations   int
	ServerSessionTokenOutputs       []ServerSessionTokenOutput
	ValidateSessionTokenInvocations int
	ValidateSessionTokenInputs      []ValidateSessionTokenInput
	ValidateSessionTokenOutputs     []ValidateSessionTokenOutput
}

func NewExternalAccessor() *ExternalAccessor {
	return &ExternalAccessor{
		Mock: test.NewMock(),
	}
}

func (e *ExternalAccessor) ServerSessionToken() (string, error) {
	e.ServerSessionTokenInvocations++

	gomega.Expect(e.ServerSessionTokenOutputs).ToNot(gomega.BeEmpty())

	output := e.ServerSessionTokenOutputs[0]
	e.ServerSessionTokenOutputs = e.ServerSessionTokenOutputs[1:]
	return output.Token, output.Error
}

func (e *ExternalAccessor) ValidateSessionToken(ctx context.Context, token string) (request.Details, error) {
	e.ValidateSessionTokenInvocations++

	e.ValidateSessionTokenInputs = append(e.ValidateSessionTokenInputs, ValidateSessionTokenInput{Context: ctx, Token: token})

	gomega.Expect(e.ValidateSessionTokenOutputs).ToNot(gomega.BeEmpty())

	output := e.ValidateSessionTokenOutputs[0]
	e.ValidateSessionTokenOutputs = e.ValidateSessionTokenOutputs[1:]
	return output.Details, output.Error
}

func (e *ExternalAccessor) Expectations() {
	e.Mock.Expectations()
	gomega.Expect(e.ServerSessionTokenOutputs).To(gomega.BeEmpty())
	gomega.Expect(e.ValidateSessionTokenOutputs).To(gomega.BeEmpty())
}
