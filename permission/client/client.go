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
	clnt, err := platform.NewClient(config, authorizeAs)
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
	return len(perms[permission.Custodian]) > 0, nil
}
