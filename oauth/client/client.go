package client

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/oauth"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
)

type Client struct {
	baseClient        *client.Client
	httpClient        *http.Client
	tokenSourceSource oauth.TokenSourceSource
}

func New(config *client.Config, httpClient *http.Client, tokenSourceSource oauth.TokenSourceSource) (*Client, error) {
	return NewWithErrorParser(config, httpClient, tokenSourceSource, nil)
}

func NewWithErrorParser(config *client.Config, httpClient *http.Client, tokenSourceSource oauth.TokenSourceSource, errorResponseParser client.ErrorResponseParser) (*Client, error) {
	baseClient, err := client.NewWithErrorParser(config, errorResponseParser)
	if err != nil {
		return nil, err
	}

	return NewWithClient(baseClient, httpClient, tokenSourceSource)
}

func NewWithClient(baseClient *client.Client, httpClient *http.Client, tokenSourceSource oauth.TokenSourceSource) (*Client, error) {
	if baseClient == nil {
		return nil, errors.New("base client is missing")
	}
	if tokenSourceSource == nil {
		return nil, errors.New("token source source is missing")
	}

	// Ensure the HTTP client. If no timeout, then use timeout from config. The HTTP client is usually provided by the
	// provider and inherits the token timeout, but overrides the API request timeout based upon the config client
	// timeout.
	httpClient = pointer.From(pointer.Default(httpClient, *http.DefaultClient))
	if httpClient.Timeout == 0 {
		if clientTimeout := baseClient.ClientTimeout(); clientTimeout > 0 {
			httpClient.Timeout = clientTimeout
		}
	}

	return &Client{
		baseClient:        baseClient,
		httpClient:        httpClient,
		tokenSourceSource: tokenSourceSource,
	}, nil
}

func (c *Client) Client() *client.Client {
	return c.baseClient
}

func (c *Client) ConstructURL(paths ...string) string {
	return c.baseClient.ConstructURL(paths...)
}

func (c *Client) AppendURLQuery(urlString string, query map[string]string) string {
	return c.baseClient.AppendURLQuery(urlString, query)
}

func (c *Client) SendOAuthRequest(ctx context.Context, method string, url string, mutators []request.RequestMutator, requestBody any, responseBody any, inspectors []request.ResponseInspector, tokenSource oauth.TokenSource) error {
	if tokenSource == nil {
		return errors.New("token source is missing")
	}

	// Attempt with existing token
	err := c.sendOAuthRequest(ctx, method, url, mutators, requestBody, responseBody, inspectors, tokenSource)

	// If the first request results in an access token error, then mark the token as
	// expired, send request again, and it will attempt to use the refresh token to
	// generate a new access token
	if oauth.IsAccessTokenError(err) {
		if _, tokenErr := tokenSource.ExpireToken(ctx); tokenErr != nil {
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

func (c *Client) sendOAuthRequest(ctx context.Context, method string, url string, mutators []request.RequestMutator, requestBody any, responseBody any, inspectors []request.ResponseInspector, tokenSource oauth.TokenSource) error {
	if ctx == nil {
		return errors.New("context is missing")
	}

	// Inject the HTTP client to use the transport for all requests and the timeout for token requests
	ctx = context.WithValue(ctx, oauth2.HTTPClient, c.httpClient)

	httpClient, err := tokenSource.HTTPClient(ctx, c.tokenSourceSource)
	if err != nil {
		return err
	}

	// Override any client timeout from the token source for API requests
	if clientTimeout := c.Client().ClientTimeout(); clientTimeout > 0 && clientTimeout != httpClient.Timeout {
		httpClient = pointer.From(*httpClient)
		httpClient.Timeout = clientTimeout
	}

	err = c.baseClient.RequestDataWithHTTPClient(ctx, method, url, mutators, requestBody, responseBody, inspectors, httpClient)

	if _, tokenErr := tokenSource.UpdateToken(ctx); tokenErr != nil {
		log.LoggerFromContext(ctx).WithError(tokenErr).Error("unable to update token")
	}

	return err
}
