package client

import (
	"context"
	"net/http"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
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

	// Attempt with existing token
	err := c.sendOAuthRequest(ctx, method, url, mutators, requestBody, responseBody, inspectors, tokenSource)

	// If the first request results in an access token error, then mark the token as
	// expired, send request again, and it will attempt to use the refresh token to
	// generate a new access token
	if oauth.IsAccessTokenError(err) {
		if tokenErr := tokenSource.ExpireToken(); tokenErr != nil {
			log.LoggerFromContext(ctx).WithError(tokenErr).Error("unable to expire token")
		}
		err = c.sendOAuthRequest(ctx, method, url, mutators, requestBody, responseBody, inspectors, tokenSource)
	}

	// If a request results in a refresh token error, then mark it as unauthenticated
	if oauth.IsRefreshTokenError(err) {
		err = errors.Wrap(request.ErrorUnauthenticated(), err.Error())
	}

	return err
}

func (c *Client) sendOAuthRequest(ctx context.Context, method string, url string, mutators []request.RequestMutator, requestBody interface{}, responseBody interface{}, inspectors []request.ResponseInspector, tokenSource oauth.TokenSource) error {
	httpClient, err := tokenSource.HTTPClient(ctx, c.tokenSourceSource)
	if err != nil {
		return err
	}

	err = c.baseClient.RequestDataWithHTTPClient(ctx, method, url, mutators, requestBody, responseBody, inspectors, httpClient)

	if tokenErr := tokenSource.UpdateToken(); tokenErr != nil {
		log.LoggerFromContext(ctx).WithError(tokenErr).Error("unable to update token")
	}

	return err
}
