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

func New(cfg *platform.Config, authorizeAs platform.AuthorizeAs) (*Client, error) {
	clnt, err := platform.NewClient(cfg, authorizeAs)
	if err != nil {
		return nil, err
	}

	return &Client{
		client: clnt,
	}, nil
}

// FUTURE: Move to auth service

func (c *Client) EnsureAuthorized(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context is missing")
	}

	if details := request.DetailsFromContext(ctx); details != nil {
		return nil
	}

	return request.ErrorUnauthorized()
}

func (c *Client) EnsureAuthorizedService(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context is missing")
	}

	if details := request.DetailsFromContext(ctx); details != nil {
		if details.IsService() {
			return nil
		}
	}

	return request.ErrorUnauthorized()
}

// FUTURE: Move to auth service

func (c *Client) EnsureAuthorizedUser(ctx context.Context, targetUserID string, authorizedPermission string) (string, error) {
	if ctx == nil {
		return "", errors.New("context is missing")
	}
	if targetUserID == "" {
		return "", errors.New("target user id is missing")
	}
	if authorizedPermission == "" {
		return "", errors.New("authorized permission is missing")
	}

	if details := request.DetailsFromContext(ctx); details != nil {
		if details.IsService() {
			return "", nil
		}

		authenticatedUserID := details.UserID()
		if authenticatedUserID == targetUserID {
			if authorizedPermission != permission.Custodian {
				return authenticatedUserID, nil
			}
		} else {
			url := c.client.ConstructURL("access", targetUserID, authenticatedUserID)
			permissions := permission.Permissions{}
			if err := c.client.RequestData(ctx, "GET", url, nil, nil, &permissions); err != nil {
				if !request.IsErrorResourceNotFound(err) {
					return "", errors.Wrap(err, "unable to get user permissions")
				}
			} else {
				permissions = permission.FixOwnerPermissions(permissions)
				if _, ok := permissions[authorizedPermission]; ok {
					return authenticatedUserID, nil
				}
			}
		}
	}

	return "", request.ErrorUnauthorized()
}
