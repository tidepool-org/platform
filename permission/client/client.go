package client

import (
	"github.com/mdblp/go-json-rest/rest"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/request"
)

type Client struct{}

func New() *Client {
	return &Client{}
}

func (c *Client) GetUserPermissions(req *rest.Request, targetUserID string) (bool, error) {
	ctx := req.Context()
	if ctx == nil {
		return false, errors.New("context is missing")
	}
	details := request.DetailsFromContext(ctx)
	if details == nil {
		return false, request.ErrorUnauthenticated()
	}
	if details.IsService() {
		return true, nil
	}
	requestUserID := details.UserID()
	if requestUserID == "" {
		return false, errors.New("request user id is missing")
	}
	if targetUserID == "" {
		return false, errors.New("target user id is missing")
	}
	if details.Role() != "patient" {
		return false, nil
	}

	if requestUserID == targetUserID {
		return true, nil
	}
	return false, nil
}

func (c *Client) GetPatientPermissions(req *rest.Request) (bool, string, error) {
	ctx := req.Context()
	if ctx == nil {
		return false, "", errors.New("context is missing")
	}
	details := request.DetailsFromContext(ctx)
	if details == nil {
		return false, "", request.ErrorUnauthenticated()
	}
	requestUserID := details.UserID()
	if requestUserID == "" {
		return false, "", errors.New("request user id is missing")
	}

	return details.Role() == "patient", requestUserID, nil
}
