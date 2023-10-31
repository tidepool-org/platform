package test

import (
	"context"

	"github.com/tidepool-org/platform/request"
)

type ServerSessionTokenOutput struct {
	Token string
	Error error
}

type ValidateSessionTokenOutput struct {
	Details request.Details
	Error   error
}

type EnsureAuthorizedUserInput struct {
	TargetUserID string
}

type EnsureAuthorizedUserOutput struct {
	AuthorizedUserID string
	Error            error
}

type ExternalAccessor struct {
	ValidateSessionTokenInvocations    int
	ValidateSessionTokenInputs         []string
	ValidateSessionTokenStub           func(ctx context.Context, token string) (request.Details, error)
	ValidateSessionTokenOutputs        []ValidateSessionTokenOutput
	ValidateSessionTokenOutput         *ValidateSessionTokenOutput
	EnsureAuthorizedInvocations        int
	EnsureAuthorizedStub               func(ctx context.Context) error
	EnsureAuthorizedOutputs            []error
	EnsureAuthorizedOutput             *error
	EnsureAuthorizedServiceInvocations int
	EnsureAuthorizedServiceStub        func(ctx context.Context) error
	EnsureAuthorizedServiceOutputs     []error
	EnsureAuthorizedServiceOutput      *error
	EnsureAuthorizedUserInvocations    int
	EnsureAuthorizedUserInputs         []EnsureAuthorizedUserInput
	EnsureAuthorizedUserStub           func(ctx context.Context, targetUserID string) (string, error)
	EnsureAuthorizedUserOutputs        []EnsureAuthorizedUserOutput
	EnsureAuthorizedUserOutput         *EnsureAuthorizedUserOutput
}

func NewExternalAccessor() *ExternalAccessor {
	return &ExternalAccessor{}
}

func (e *ExternalAccessor) ValidateSessionToken(ctx context.Context, token string) (request.Details, error) {
	e.ValidateSessionTokenInvocations++
	e.ValidateSessionTokenInputs = append(e.ValidateSessionTokenInputs, token)
	if e.ValidateSessionTokenStub != nil {
		return e.ValidateSessionTokenStub(ctx, token)
	}
	if len(e.ValidateSessionTokenOutputs) > 0 {
		output := e.ValidateSessionTokenOutputs[0]
		e.ValidateSessionTokenOutputs = e.ValidateSessionTokenOutputs[1:]
		return output.Details, output.Error
	}
	if e.ValidateSessionTokenOutput != nil {
		return e.ValidateSessionTokenOutput.Details, e.ValidateSessionTokenOutput.Error
	}
	panic("ValidateSessionToken has no output")
}

func (e *ExternalAccessor) AssertOutputsEmpty() {
	if len(e.ValidateSessionTokenOutputs) > 0 {
		panic("ValidateSessionTokenOutputs is not empty")
	}
	if len(e.EnsureAuthorizedOutputs) > 0 {
		panic("EnsureAuthorizedOutputs is not empty")
	}
	if len(e.EnsureAuthorizedServiceOutputs) > 0 {
		panic("EnsureAuthorizedServiceOutputs is not empty")
	}
	if len(e.EnsureAuthorizedUserOutputs) > 0 {
		panic("EnsureAuthorizedUserOutputs is not empty")
	}
}
