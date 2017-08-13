package test

import (
	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/test"
)

type ServerTokenOutput struct {
	Token string
	Error error
}

type ValidateTokenInput struct {
	Context auth.Context
	Token   string
}

type ValidateTokenOutput struct {
	Details auth.Details
	Error   error
}

type GetStatusOutput struct {
	Status *auth.Status
	Error  error
}

type Client struct {
	*test.Mock
	ServerTokenInvocations   int
	ServerTokenOutputs       []ServerTokenOutput
	ValidateTokenInvocations int
	ValidateTokenInputs      []ValidateTokenInput
	ValidateTokenOutputs     []ValidateTokenOutput
	GetStatusInvocations     int
	GetStatusInputs          []auth.Context
	GetStatusOutputs         []GetStatusOutput
}

func NewClient() *Client {
	return &Client{
		Mock: test.NewMock(),
	}
}

func (c *Client) ServerToken() (string, error) {
	c.ServerTokenInvocations++

	if len(c.ServerTokenOutputs) == 0 {
		panic("Unexpected invocation of ServerToken on Client")
	}

	output := c.ServerTokenOutputs[0]
	c.ServerTokenOutputs = c.ServerTokenOutputs[1:]
	return output.Token, output.Error
}

func (c *Client) ValidateToken(ctx auth.Context, token string) (auth.Details, error) {
	c.ValidateTokenInvocations++

	c.ValidateTokenInputs = append(c.ValidateTokenInputs, ValidateTokenInput{Context: ctx, Token: token})

	if len(c.ValidateTokenOutputs) == 0 {
		panic("Unexpected invocation of ValidateToken on Client")
	}

	output := c.ValidateTokenOutputs[0]
	c.ValidateTokenOutputs = c.ValidateTokenOutputs[1:]
	return output.Details, output.Error
}

func (c *Client) GetStatus(ctx auth.Context) (*auth.Status, error) {
	c.GetStatusInvocations++

	c.GetStatusInputs = append(c.GetStatusInputs, ctx)

	if len(c.GetStatusOutputs) == 0 {
		panic("Unexpected invocation of GetStatus on Client")
	}

	output := c.GetStatusOutputs[0]
	c.GetStatusOutputs = c.GetStatusOutputs[1:]
	return output.Status, output.Error
}

func (c *Client) UnusedOutputsCount() int {
	return len(c.ServerTokenOutputs) +
		len(c.ValidateTokenOutputs) +
		len(c.GetStatusOutputs)
}
