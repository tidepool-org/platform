package customerio

import (
	"context"
	"net/http"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/request"
)

type Event struct {
	Name string `json:"name"`
	ID   string `json:"id"`
	Data any    `json:"data"`
}

func (c *Client) SendEvent(ctx context.Context, cid string, event Event) error {
	ctx = log.NewContextWithLogger(ctx, c.logger)
	url := c.trackClient.ConstructURL("api", "v1", "customers", cid, "events")

	mutators := []request.RequestMutator{
		c.trackAPIAuthMutator(),
	}

	if err := c.trackClient.RequestDataWithHTTPClient(ctx, http.MethodPost, url, mutators, event, nil, nil, c.httpClient); err != nil {
		return err
	}

	return nil
}
