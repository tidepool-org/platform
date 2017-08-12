package client

import (
	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/errors"
)

type Client interface {
	DestroyDataForUserByID(context auth.Context, userID string) error
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

func (c *clientImpl) DestroyDataForUserByID(context auth.Context, userID string) error {
	if context == nil {
		return errors.New("client", "context is missing")
	}
	if userID == "" {
		return errors.New("client", "user id is missing")
	}

	context.Logger().WithField("userId", userID).Debug("Deleting data for user")

	return c.client.SendRequestWithServerToken(context, "DELETE", c.client.BuildURL("dataservices", "v1", "users", userID, "data"), nil, nil)
}
