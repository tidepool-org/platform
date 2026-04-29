package test

import (
	"context"

	"github.com/tidepool-org/platform/permission"
)

type Client struct {
	*ProviderSessionAccessor
	*RestrictedTokenAccessor
	*ExternalAccessor
}

func NewClient() *Client {
	return &Client{
		ProviderSessionAccessor: NewProviderSessionAccessor(),
		RestrictedTokenAccessor: NewRestrictedTokenAccessor(),
		ExternalAccessor:        NewExternalAccessor(),
	}
}

func (c *Client) AssertOutputsEmpty() {
	c.ProviderSessionAccessor.Expectations()
	c.RestrictedTokenAccessor.Expectations()
	c.ExternalAccessor.AssertOutputsEmpty()
}

func (c *Client) GetUserPermissions(ctx context.Context, requestUserID string, targetUserID string) (permission.Permissions, error) {
	return c.ExternalAccessor.Client.GetUserPermissions(ctx, requestUserID, targetUserID)
}

func (c *Client) UpdateUserPermissions(ctx context.Context, requestUserID string, targetUserID string, permissions permission.Permissions) error {
	return c.ExternalAccessor.Client.UpdateUserPermissions(ctx, requestUserID, targetUserID, permissions)
}

func (c *Client) GroupsForUser(ctx context.Context, granteeUserID string) (permission.Permissions, error) {
	return nil, nil
}

func (c *Client) UsersInGroup(ctx context.Context, sharerID string) (permission.Permissions, error) {
	return nil, nil
}

func (c *Client) HasMembershipRelationship(ctx context.Context, granteeUserID, grantorUserID string) (has bool, err error) {
	return false, nil
}

func (c *Client) HasCustodianPermissions(ctx context.Context, granteeUserID, grantorUserID string) (has bool, err error) {
	return false, nil
}

func (c *Client) HasWritePermissions(ctx context.Context, granteeUserID, grantorUserID string) (has bool, err error) {
	return false, nil
}
