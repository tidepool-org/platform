package test

import (
	"context"

	"github.com/tidepool-org/platform/user"
)

type EnsureAuthorizedUserInput struct {
	Context      context.Context
	TargetUserID string
	Permission   string
}

type EnsureAuthorizedUserOutput struct {
	AuthorizedUserID string
	Error            error
}

type GetUserPermissionsInput struct {
	Context       context.Context
	RequestUserID string
	TargetUserID  string
}

type GetUserPermissionsOutput struct {
	Permissions user.Permissions
	Error       error
}

type Client struct {
	EnsureAuthorizedServiceInvocations int
	EnsureAuthorizedServiceInputs      []context.Context
	EnsureAuthorizedServiceStub        func(ctx context.Context) error
	EnsureAuthorizedServiceOutputs     []error
	EnsureAuthorizedServiceOutput      *error
	EnsureAuthorizedUserInvocations    int
	EnsureAuthorizedUserInputs         []EnsureAuthorizedUserInput
	EnsureAuthorizedUserStub           func(ctx context.Context, targetUserID string, permission string) (string, error)
	EnsureAuthorizedUserOutputs        []EnsureAuthorizedUserOutput
	EnsureAuthorizedUserOutput         *EnsureAuthorizedUserOutput
	GetUserPermissionsInvocations      int
	GetUserPermissionsInputs           []GetUserPermissionsInput
	GetUserPermissionsStub             func(ctx context.Context, requestUserID string, targetUserID string) (user.Permissions, error)
	GetUserPermissionsOutputs          []GetUserPermissionsOutput
	GetUserPermissionsOutput           *GetUserPermissionsOutput
}

func NewClient() *Client {
	return &Client{}
}

func (r *Client) EnsureAuthorizedService(ctx context.Context) error {
	r.EnsureAuthorizedServiceInvocations++
	r.EnsureAuthorizedServiceInputs = append(r.EnsureAuthorizedServiceInputs, ctx)
	if r.EnsureAuthorizedServiceStub != nil {
		return r.EnsureAuthorizedServiceStub(ctx)
	}
	if len(r.EnsureAuthorizedServiceOutputs) > 0 {
		output := r.EnsureAuthorizedServiceOutputs[0]
		r.EnsureAuthorizedServiceOutputs = r.EnsureAuthorizedServiceOutputs[1:]
		return output
	}
	if r.EnsureAuthorizedServiceOutput != nil {
		return *r.EnsureAuthorizedServiceOutput
	}
	panic("EnsureAuthorizedService has no output")
}

func (r *Client) EnsureAuthorizedUser(ctx context.Context, targetUserID string, permission string) (string, error) {
	r.EnsureAuthorizedUserInvocations++
	r.EnsureAuthorizedUserInputs = append(r.EnsureAuthorizedUserInputs, EnsureAuthorizedUserInput{Context: ctx, TargetUserID: targetUserID, Permission: permission})
	if r.EnsureAuthorizedUserStub != nil {
		return r.EnsureAuthorizedUserStub(ctx, targetUserID, permission)
	}
	if len(r.EnsureAuthorizedUserOutputs) > 0 {
		output := r.EnsureAuthorizedUserOutputs[0]
		r.EnsureAuthorizedUserOutputs = r.EnsureAuthorizedUserOutputs[1:]
		return output.AuthorizedUserID, output.Error
	}
	if r.EnsureAuthorizedUserOutput != nil {
		return r.EnsureAuthorizedUserOutput.AuthorizedUserID, r.EnsureAuthorizedUserOutput.Error
	}
	panic("EnsureAuthorizedUser has no output")
}

func (r *Client) GetUserPermissions(ctx context.Context, requestUserID string, targetUserID string) (user.Permissions, error) {
	r.GetUserPermissionsInvocations++
	r.GetUserPermissionsInputs = append(r.GetUserPermissionsInputs, GetUserPermissionsInput{Context: ctx, RequestUserID: requestUserID, TargetUserID: targetUserID})
	if r.GetUserPermissionsStub != nil {
		return r.GetUserPermissionsStub(ctx, requestUserID, targetUserID)
	}
	if len(r.GetUserPermissionsOutputs) > 0 {
		output := r.GetUserPermissionsOutputs[0]
		r.GetUserPermissionsOutputs = r.GetUserPermissionsOutputs[1:]
		return output.Permissions, output.Error
	}
	if r.GetUserPermissionsOutput != nil {
		return r.GetUserPermissionsOutput.Permissions, r.GetUserPermissionsOutput.Error
	}
	panic("GetUserPermissions has no output")
}

func (r *Client) AssertOutputsEmpty() {
	if len(r.EnsureAuthorizedServiceOutputs) > 0 {
		panic("EnsureAuthorizedServiceOutputs is not empty")
	}
	if len(r.EnsureAuthorizedUserOutputs) > 0 {
		panic("EnsureAuthorizedUserOutputs is not empty")
	}
	if len(r.GetUserPermissionsOutputs) > 0 {
		panic("GetUserPermissionsOutputs is not empty")
	}
}
