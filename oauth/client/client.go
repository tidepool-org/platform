package client

import (
	"context"
	"net/http"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/oauth"
	"github.com/tidepool-org/platform/request"
)

type BaseClient interface {
	ConstructURL(paths ...string) string
	AppendURLQuery(urlString string, query map[string]string) string
	RequestDataWithHTTPClient(ctx context.Context, method string, url string, mutators []request.RequestMutator, requestBody interface{}, responseBody interface{}, inspectors []request.ResponseInspector, httpClient *http.Client) error
}

type Client struct {
	baseClient        BaseClient
	tokenSourceSource oauth.TokenSourceSource
}

func New(baseClient BaseClient, tokenSourceSource oauth.TokenSourceSource) (*Client, error) {
	if baseClient == nil {
		return nil, errors.New("base client is missing")
	}
	if tokenSourceSource == nil {
		return nil, errors.New("token source source is missing")
	}

	return &Client{
		baseClient:        baseClient,
		tokenSourceSource: tokenSourceSource,
	}, nil
}

func (c *Client) ConstructURL(paths ...string) string {
	return c.baseClient.ConstructURL(paths...)
}

func (c *Client) AppendURLQuery(urlString string, query map[string]string) string {
	return c.baseClient.AppendURLQuery(urlString, query)
}

func (c *Client) SendOAuthRequest(ctx context.Context, method string, url string, mutators []request.RequestMutator, requestBody interface{}, responseBody interface{}, inspectors []request.ResponseInspector, tokenSource oauth.TokenSource) error {
	if tokenSource == nil {
		return errors.New("http client source is missing")
	}

	httpClient, err := tokenSource.HTTPClient(ctx, c.tokenSourceSource)
	if err != nil {
		return err
	}

	return c.baseClient.RequestDataWithHTTPClient(ctx, method, url, mutators, requestBody, responseBody, inspectors, httpClient)
}
