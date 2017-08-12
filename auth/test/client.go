package test

import (
	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/id"
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

type Client struct {
	ID                       string
	ServerTokenInvocations   int
	ServerTokenOutputs       []ServerTokenOutput
	ValidateTokenInvocations int
	ValidateTokenInputs      []ValidateTokenInput
	ValidateTokenOutputs     []ValidateTokenOutput
}

func NewClient() *Client {
	return &Client{
		ID: id.New(),
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

func (c *Client) ValidateToken(context auth.Context, token string) (auth.Details, error) {
	c.ValidateTokenInvocations++

	c.ValidateTokenInputs = append(c.ValidateTokenInputs, ValidateTokenInput{Context: context, Token: token})

	if len(c.ValidateTokenOutputs) == 0 {
		panic("Unexpected invocation of ValidateToken on Client")
	}

	output := c.ValidateTokenOutputs[0]
	c.ValidateTokenOutputs = c.ValidateTokenOutputs[1:]
	return output.Details, output.Error
}

func (c *Client) UnusedOutputsCount() int {
	return len(c.ServerTokenOutputs) +
		len(c.ValidateTokenOutputs)
}
