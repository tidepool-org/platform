package client

import (
	"context"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/user"
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

func (c *Client) EnsureAuthorizedUser(ctx context.Context, targetUserID string, permission string) (string, error) {
	if ctx == nil {
		return "", errors.New("context is missing")
	}
	if targetUserID == "" {
		return "", errors.New("target user id is missing")
	}
	if permission == "" {
		return "", errors.New("permission is missing")
	}

	if details := request.DetailsFromContext(ctx); details != nil {
		if details.IsService() {
			return "", nil
		}

		authenticatedUserID := details.UserID()
		if authenticatedUserID == targetUserID {
			if permission != user.CustodianPermission {
				return authenticatedUserID, nil
			}
		} else {
			permissions, err := c.GetUserPermissions(ctx, authenticatedUserID, targetUserID)
			if err != nil {
				if !request.IsErrorUnauthorized(err) {
					return "", errors.Wrap(err, "unable to get user permissions")
				}
			} else if _, ok := permissions[permission]; ok {
				return authenticatedUserID, nil
			}
		}
	}

	return "", request.ErrorUnauthorized()
}

func (c *Client) GetUserPermissions(ctx context.Context, requestUserID string, targetUserID string) (user.Permissions, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if requestUserID == "" {
		return nil, errors.New("request user id is missing")
	}
	if targetUserID == "" {
		return nil, errors.New("target user id is missing")
	}

	log.LoggerFromContext(ctx).WithFields(log.Fields{"requestUserId": requestUserID, "targetUserId": targetUserID}).Debug("Get user permissions")

	permissions := user.Permissions{}
	if err := c.client.RequestData(ctx, "GET", c.client.ConstructURL("access", targetUserID, requestUserID), nil, nil, &permissions); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, request.ErrorUnauthorized()
		}
		return nil, err
	}

	// Fix missing view and upload permissions for an owner
	if permission, ok := permissions[user.OwnerPermission]; ok {
		if _, ok = permissions[user.UploadPermission]; !ok {
			permissions[user.UploadPermission] = permission
		}
		if _, ok = permissions[user.ViewPermission]; !ok {
			permissions[user.ViewPermission] = permission
		}
	}

	return permissions, nil
}
