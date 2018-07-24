package test

import (
	"context"
)

type EnsureAuthorizedUserInput struct {
	Context              context.Context
	TargetUserID         string
	AuthorizedPermission string
}

type EnsureAuthorizedUserOutput struct {
	AuthorizedUserID string
	Error            error
}

type Client struct {
	EnsureAuthorizedInvocations        int
	EnsureAuthorizedInputs             []context.Context
	EnsureAuthorizedStub               func(ctx context.Context) error
	EnsureAuthorizedOutputs            []error
	EnsureAuthorizedOutput             *error
	EnsureAuthorizedServiceInvocations int
	EnsureAuthorizedServiceInputs      []context.Context
	EnsureAuthorizedServiceStub        func(ctx context.Context) error
	EnsureAuthorizedServiceOutputs     []error
	EnsureAuthorizedServiceOutput      *error
	EnsureAuthorizedUserInvocations    int
	EnsureAuthorizedUserInputs         []EnsureAuthorizedUserInput
	EnsureAuthorizedUserStub           func(ctx context.Context, targetUserID string, authorizedPermission string) (string, error)
	EnsureAuthorizedUserOutputs        []EnsureAuthorizedUserOutput
	EnsureAuthorizedUserOutput         *EnsureAuthorizedUserOutput
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) EnsureAuthorized(ctx context.Context) error {
	c.EnsureAuthorizedInvocations++
	c.EnsureAuthorizedInputs = append(c.EnsureAuthorizedInputs, ctx)
	if c.EnsureAuthorizedStub != nil {
		return c.EnsureAuthorizedStub(ctx)
	}
	if len(c.EnsureAuthorizedOutputs) > 0 {
		output := c.EnsureAuthorizedOutputs[0]
		c.EnsureAuthorizedOutputs = c.EnsureAuthorizedOutputs[1:]
		return output
	}
	if c.EnsureAuthorizedOutput != nil {
		return *c.EnsureAuthorizedOutput
	}
	panic("EnsureAuthorized has no output")
}

func (c *Client) EnsureAuthorizedService(ctx context.Context) error {
	c.EnsureAuthorizedServiceInvocations++
	c.EnsureAuthorizedServiceInputs = append(c.EnsureAuthorizedServiceInputs, ctx)
	if c.EnsureAuthorizedServiceStub != nil {
		return c.EnsureAuthorizedServiceStub(ctx)
	}
	if len(c.EnsureAuthorizedServiceOutputs) > 0 {
		output := c.EnsureAuthorizedServiceOutputs[0]
		c.EnsureAuthorizedServiceOutputs = c.EnsureAuthorizedServiceOutputs[1:]
		return output
	}
	if c.EnsureAuthorizedServiceOutput != nil {
		return *c.EnsureAuthorizedServiceOutput
	}
	panic("EnsureAuthorizedService has no output")
}

func (c *Client) EnsureAuthorizedUser(ctx context.Context, targetUserID string, authorizedPermission string) (string, error) {
	c.EnsureAuthorizedUserInvocations++
	c.EnsureAuthorizedUserInputs = append(c.EnsureAuthorizedUserInputs, EnsureAuthorizedUserInput{Context: ctx, TargetUserID: targetUserID, AuthorizedPermission: authorizedPermission})
	if c.EnsureAuthorizedUserStub != nil {
		return c.EnsureAuthorizedUserStub(ctx, targetUserID, authorizedPermission)
	}
	if len(c.EnsureAuthorizedUserOutputs) > 0 {
		output := c.EnsureAuthorizedUserOutputs[0]
		c.EnsureAuthorizedUserOutputs = c.EnsureAuthorizedUserOutputs[1:]
		return output.AuthorizedUserID, output.Error
	}
	if c.EnsureAuthorizedUserOutput != nil {
		return c.EnsureAuthorizedUserOutput.AuthorizedUserID, c.EnsureAuthorizedUserOutput.Error
	}
	panic("EnsureAuthorizedUser has no output")
}

func (c *Client) AssertOutputsEmpty() {
	if len(c.EnsureAuthorizedOutputs) > 0 {
		panic("EnsureAuthorizedOutputs is not empty")
	}
	if len(c.EnsureAuthorizedServiceOutputs) > 0 {
		panic("EnsureAuthorizedServiceOutputs is not empty")
	}
	if len(c.EnsureAuthorizedUserOutputs) > 0 {
		panic("EnsureAuthorizedUserOutputs is not empty")
	}
}
