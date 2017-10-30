package test

import (
	"context"

	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/user"
)

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
	*test.Mock
	GetUserPermissionsInvocations int
	GetUserPermissionsInputs      []GetUserPermissionsInput
	GetUserPermissionsOutputs     []GetUserPermissionsOutput
}

func NewClient() *Client {
	return &Client{
		Mock: test.NewMock(),
	}
}

func (c *Client) GetUserPermissions(ctx context.Context, requestUserID string, targetUserID string) (user.Permissions, error) {
	c.GetUserPermissionsInvocations++

	c.GetUserPermissionsInputs = append(c.GetUserPermissionsInputs, GetUserPermissionsInput{Context: ctx, RequestUserID: requestUserID, TargetUserID: targetUserID})

	gomega.Expect(c.GetUserPermissionsOutputs).ToNot(gomega.BeEmpty())

	output := c.GetUserPermissionsOutputs[0]
	c.GetUserPermissionsOutputs = c.GetUserPermissionsOutputs[1:]
	return output.Permissions, output.Error
}

func (c *Client) Expectations() {
	c.Mock.Expectations()
	gomega.Expect(c.GetUserPermissionsOutputs).To(gomega.BeEmpty())
}
