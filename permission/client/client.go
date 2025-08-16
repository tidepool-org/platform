package client

import (
	"context"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/request"
)

type Client struct {
	client *platform.Client
}

func New(config *platform.Config, authorizeAs platform.AuthorizeAs) (*Client, error) {
	clnt, err := platform.NewLegacyClient(config, authorizeAs)
	if err != nil {
		return nil, err
	}

	return &Client{
		client: clnt,
	}, nil
}

func (c *Client) GetUserPermissions(ctx context.Context, requestUserID string, targetUserID string) (permission.Permissions, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if requestUserID == "" {
		return nil, errors.New("request user id is missing")
	}
	if targetUserID == "" {
		return nil, errors.New("target user id is missing")
	}

	url := c.client.ConstructURL("access", targetUserID, requestUserID)
	result := permission.Permissions{}
	if err := c.client.RequestData(ctx, "GET", url, nil, nil, &result); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, request.ErrorUnauthorized()
		}
		return nil, err
	}

	return permission.FixOwnerPermissions(result), nil
}

// GroupsForUser returns what users have shared permissions with the user with an id of granteeUserID.
// The GroupedPermissions are keyed by the id of the user who shared their permissions with granteeUserID.
func (c *Client) GroupsForUser(ctx context.Context, granteeUserID string) (permission.Permissions, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if granteeUserID == "" {
		return nil, errors.New("user id is missing")
	}

	url := c.client.ConstructURL("access", "groups", granteeUserID)
	result := permission.Permissions{}
	if err := c.client.RequestData(ctx, "GET", url, nil, nil, &result); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, request.ErrorUnauthorized()
		}
		return nil, err
	}

	return result, nil
}

func (c *Client) UsersInGroup(ctx context.Context, sharerID string) (permission.Permissions, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if sharerID == "" {
		return nil, errors.New("user id is missing")
	}

	url := c.client.ConstructURL("access", sharerID)
	result := permission.Permissions{}
	if err := c.client.RequestData(ctx, "GET", url, nil, nil, &result); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, request.ErrorUnauthorized()
		}
		return nil, err
	}

	return result, nil
}

func (c *Client) HasMembershipRelationship(ctx context.Context, granteeUserID, grantorUserID string) (has bool, err error) {
	fromTo, err := c.GetUserPermissions(ctx, granteeUserID, grantorUserID)
	if err != nil {
		return false, err
	}
	if len(fromTo) > 0 {
		return true, nil
	}
	toFrom, err := c.GetUserPermissions(ctx, grantorUserID, granteeUserID)
	if err != nil {
		return false, err
	}
	if len(toFrom) > 0 {
		return true, nil
	}
	return false, nil
}

func (c *Client) HasCustodianPermissions(ctx context.Context, granteeUserID, grantorUserID string) (has bool, err error) {
	perms, err := c.GetUserPermissions(ctx, granteeUserID, grantorUserID)
	if err != nil {
		return false, err
	}
	_, ok := perms[permission.Custodian]
	return ok, nil
}

func (c *Client) HasWritePermissions(ctx context.Context, granteeUserID, grantorUserID string) (has bool, err error) {
	if granteeUserID != "" && granteeUserID == grantorUserID {
		return true, nil
	}
	perms, err := c.GetUserPermissions(ctx, granteeUserID, grantorUserID)
	if err != nil {
		return false, err
	}
	if _, ok := perms[permission.Custodian]; ok {
		return true, nil
	}
	if _, ok := perms[permission.Write]; ok {
		return true, nil
	}
	if _, ok := perms[permission.Owner]; ok {
		return true, nil
	}
	return false, nil
}
