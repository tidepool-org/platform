package test

import (
	"context"

	"github.com/tidepool-org/platform/devicetokens"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/request"
)

type ServerSessionTokenOutput struct {
	Token string
	Error error
}

type ValidateSessionTokenOutput struct {
	AuthDetails request.AuthDetails
	Error       error
}

type EnsureAuthorizedUserInput struct {
	TargetUserID         string
	AuthorizedPermission string
}

type EnsureAuthorizedUserOutput struct {
	AuthorizedUserID string
	Error            error
}

type UserPermissionsOutput struct {
	Permissions permission.Permissions
	Error       error
}

type ExternalAccessor struct {
	ServerSessionTokenInvocations      int
	ServerSessionTokenStub             func() (string, error)
	ServerSessionTokenOutputs          []ServerSessionTokenOutput
	ServerSessionTokenOutput           *ServerSessionTokenOutput
	ValidateSessionTokenInvocations    int
	ValidateSessionTokenInputs         []string
	ValidateSessionTokenStub           func(ctx context.Context, token string) (request.AuthDetails, error)
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
	EnsureAuthorizedUserStub           func(ctx context.Context, targetUserID string, authorizedPermission string) (string, error)
	EnsureAuthorizedUserOutputs        []EnsureAuthorizedUserOutput
	EnsureAuthorizedUserOutput         *EnsureAuthorizedUserOutput
	GetUserPermissionsOutputs          []UserPermissionsOutput
	GetUserPermissionsOutput           *UserPermissionsOutput
	GetUserPermissionsStub             func(ctx context.Context, requestUserID string, targetUserID string) (permission.Permissions, error)
}

func NewExternalAccessor() *ExternalAccessor {
	return &ExternalAccessor{}
}

func (e *ExternalAccessor) ServerSessionToken() (string, error) {
	e.ServerSessionTokenInvocations++
	if e.ServerSessionTokenStub != nil {
		return e.ServerSessionTokenStub()
	}
	if len(e.ServerSessionTokenOutputs) > 0 {
		output := e.ServerSessionTokenOutputs[0]
		e.ServerSessionTokenOutputs = e.ServerSessionTokenOutputs[1:]
		return output.Token, output.Error
	}
	if e.ServerSessionTokenOutput != nil {
		return e.ServerSessionTokenOutput.Token, e.ServerSessionTokenOutput.Error
	}
	panic("ServerSessionToken has no output")
}

func (e *ExternalAccessor) ValidateSessionToken(ctx context.Context, token string) (request.AuthDetails, error) {
	e.ValidateSessionTokenInvocations++
	e.ValidateSessionTokenInputs = append(e.ValidateSessionTokenInputs, token)
	if e.ValidateSessionTokenStub != nil {
		return e.ValidateSessionTokenStub(ctx, token)
	}
	if len(e.ValidateSessionTokenOutputs) > 0 {
		output := e.ValidateSessionTokenOutputs[0]
		e.ValidateSessionTokenOutputs = e.ValidateSessionTokenOutputs[1:]
		return output.AuthDetails, output.Error
	}
	if e.ValidateSessionTokenOutput != nil {
		return e.ValidateSessionTokenOutput.AuthDetails, e.ValidateSessionTokenOutput.Error
	}
	panic("ValidateSessionToken has no output")
}

func (e *ExternalAccessor) EnsureAuthorized(ctx context.Context) error {
	e.EnsureAuthorizedInvocations++
	if e.EnsureAuthorizedStub != nil {
		return e.EnsureAuthorizedStub(ctx)
	}
	if len(e.EnsureAuthorizedOutputs) > 0 {
		output := e.EnsureAuthorizedOutputs[0]
		e.EnsureAuthorizedOutputs = e.EnsureAuthorizedOutputs[1:]
		return output
	}
	if e.EnsureAuthorizedOutput != nil {
		return *e.EnsureAuthorizedOutput
	}
	panic("EnsureAuthorized has no output")
}

func (e *ExternalAccessor) EnsureAuthorizedService(ctx context.Context) error {
	e.EnsureAuthorizedServiceInvocations++
	if e.EnsureAuthorizedServiceStub != nil {
		return e.EnsureAuthorizedServiceStub(ctx)
	}
	if len(e.EnsureAuthorizedServiceOutputs) > 0 {
		output := e.EnsureAuthorizedServiceOutputs[0]
		e.EnsureAuthorizedServiceOutputs = e.EnsureAuthorizedServiceOutputs[1:]
		return output
	}
	if e.EnsureAuthorizedServiceOutput != nil {
		return *e.EnsureAuthorizedServiceOutput
	}
	panic("EnsureAuthorizedService has no output")
}

func (e *ExternalAccessor) EnsureAuthorizedUser(ctx context.Context, targetUserID string, authorizedPermission string) (string, error) {
	e.EnsureAuthorizedUserInvocations++
	e.EnsureAuthorizedUserInputs = append(e.EnsureAuthorizedUserInputs, EnsureAuthorizedUserInput{TargetUserID: targetUserID, AuthorizedPermission: authorizedPermission})
	if e.EnsureAuthorizedUserStub != nil {
		return e.EnsureAuthorizedUserStub(ctx, targetUserID, authorizedPermission)
	}
	if len(e.EnsureAuthorizedUserOutputs) > 0 {
		output := e.EnsureAuthorizedUserOutputs[0]
		e.EnsureAuthorizedUserOutputs = e.EnsureAuthorizedUserOutputs[1:]
		return output.AuthorizedUserID, output.Error
	}
	if e.EnsureAuthorizedUserOutput != nil {
		return e.EnsureAuthorizedUserOutput.AuthorizedUserID, e.EnsureAuthorizedUserOutput.Error
	}
	panic("EnsureAuthorizedUser has no output")
}

func (e *ExternalAccessor) AssertOutputsEmpty() {
	if len(e.ServerSessionTokenOutputs) > 0 {
		panic("ServerSessionTokenOutputs is not empty")
	}
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

func (e *ExternalAccessor) GetUserPermissions(ctx context.Context, requestUserID string, targetUserID string) (permission.Permissions, error) {
	if e.GetUserPermissionsStub != nil {
		return e.GetUserPermissionsStub(ctx, requestUserID, targetUserID)
	}
	if len(e.GetUserPermissionsOutputs) > 0 {
		output := e.GetUserPermissionsOutputs[0]
		e.GetUserPermissionsOutputs = e.GetUserPermissionsOutputs[1:]
		e.GetUserPermissionsOutput = &output
		return output.Permissions, output.Error
	}
	if e.GetUserPermissionsOutput != nil {
		return e.GetUserPermissionsOutput.Permissions, e.GetUserPermissionsOutput.Error
	}
	panic("GetUserPermissions no output")
}

func NewDeviceTokensClient() *DeviceTokensClient { return &DeviceTokensClient{} }

type DeviceTokensClient struct{}

func (c *DeviceTokensClient) GetDeviceTokens(ctx context.Context, userID string) ([]*devicetokens.DeviceToken, error) {
	return nil, nil
}
