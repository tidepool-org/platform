package client

import (
	"context"

	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/oauth"
	"github.com/tidepool-org/platform/request"
)

type Client struct {
	client            *client.Client
	tokenSourceSource oauth.TokenSourceSource
}

func New(config *client.Config, tokenSourceSource oauth.TokenSourceSource) (*Client, error) {
	if config == nil {
		return nil, errors.New("config is missing")
	}
	if tokenSourceSource == nil {
		return nil, errors.New("token source source is missing")
	}

	clnt, err := client.New(config)
	if err != nil {
		return nil, err
	}

	return &Client{
		client:            clnt,
		tokenSourceSource: tokenSourceSource,
	}, nil
}

func (c *Client) ConstructURL(paths ...string) string {
	return c.client.ConstructURL(paths...)
}

func (c *Client) AppendURLQuery(urlString string, query map[string]string) string {
	return c.client.AppendURLQuery(urlString, query)
}

func (c *Client) SendOAuthRequest(ctx context.Context, method string, url string, mutators []request.RequestMutator, requestBody interface{}, responseBody interface{}, inspectors []request.ResponseInspector, httpClientSource oauth.HTTPClientSource) error {
	if httpClientSource == nil {
		return errors.New("http client source is missing")
	}

	httpClient, err := httpClientSource.HTTPClient(ctx, c.tokenSourceSource)
	if err != nil {
		return err
	}

	return c.client.RequestDataWithHTTPClient(ctx, method, url, mutators, requestBody, responseBody, inspectors, httpClient)
}
