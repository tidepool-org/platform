package client

import (
	"net/http"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
)

type Client interface {
	GetUserPermissions(context auth.Context, requestUserID string, targetUserID string) (Permissions, error)
}

type clientImpl struct {
	client *client.Client
}

func NewClient(config *client.Config) (Client, error) {
	clnt, err := client.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &clientImpl{
		client: clnt,
	}, nil
}

func (c *clientImpl) GetUserPermissions(context auth.Context, requestUserID string, targetUserID string) (Permissions, error) {
	if context == nil {
		return nil, errors.New("client", "context is missing")
	}
	if requestUserID == "" {
		return nil, errors.New("client", "request user id is missing")
	}
	if targetUserID == "" {
		return nil, errors.New("client", "target user id is missing")
	}

	context.Logger().WithFields(log.Fields{"requestUserId": requestUserID, "targetUserId": targetUserID}).Debug("Get user permissions")

	permissions := Permissions{}
	if err := c.client.SendRequestWithServerToken(context, "GET", c.client.BuildURL("access", targetUserID, requestUserID), nil, &permissions); err != nil {
		if unexpectedResponseError, ok := err.(*client.UnexpectedResponseError); ok {
			if unexpectedResponseError.StatusCode == http.StatusNotFound {
				return nil, client.NewUnauthorizedError()
			}
		}
		return nil, err
	}

	// Fix missing view and upload permissions for an owner
	if permission, ok := permissions[OwnerPermission]; ok {
		if _, ok = permissions[UploadPermission]; !ok {
			permissions[UploadPermission] = permission
		}
		if _, ok = permissions[ViewPermission]; !ok {
			permissions[ViewPermission] = permission
		}
	}

	return permissions, nil
}
