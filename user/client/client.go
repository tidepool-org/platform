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

func New(cfg *platform.Config) (*Client, error) {
	clnt, err := platform.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &Client{
		client: clnt,
	}, nil
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
	if err := c.client.SendRequestAsServer(ctx, "GET", c.client.ConstructURL("access", targetUserID, requestUserID), nil, nil, &permissions); err != nil {
		if errors.Code(err) == request.ErrorCodeResourceNotFound {
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
