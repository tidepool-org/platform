package client

import (
	"context"

	"github.com/tidepool-org/platform/request"

	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/oauth"
)

type Client struct {
	client   *client.Client
	provider oauth.Provider
}

func New(cfg *client.Config, prvdr oauth.Provider) (*Client, error) {
	if cfg == nil {
		return nil, errors.New("config is missing")
	}
	if prvdr == nil {
		return nil, errors.New("provider is missing")
	}

	clnt, err := client.New(cfg)
	if err != nil {
		return nil, err
	}

	return &Client{
		client:   clnt,
		provider: prvdr,
	}, nil
}

func (c *Client) ConstructURL(paths ...string) string {
	return c.client.ConstructURL(paths...)
}

func (c *Client) AppendURLQuery(urlString string, query map[string]string) string {
	return c.client.AppendURLQuery(urlString, query)
}

func (c *Client) SendOAuthRequest(ctx context.Context, method string, url string, mutators []client.Mutator, requestBody interface{}, responseBody interface{}, tokenSource oauth.TokenSource) error {
	if tokenSource == nil {
		return errors.New("token source is missing")
	}

	httpClient, err := tokenSource.HTTPClient(ctx, c.provider)
	if err != nil {
		return err
	}

	if err = c.client.SendRequest(ctx, method, url, mutators, requestBody, responseBody, httpClient); err != nil {
		if oauth.IsAuthorizationError(err) {
			return request.ErrorUnauthenticated()
		}
		return err
	}

	return nil
}
