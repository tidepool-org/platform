package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/oauth"
	oauthClient "github.com/tidepool-org/platform/oauth/client"
	"github.com/tidepool-org/platform/oura"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/times"
)

type Provider interface {
	oauth.TokenSourceSource

	ClientID() string
	ClientSecret() string

	PartnerURL() *url.URL
	PartnerSecret() string
}

type Client struct {
	client   *oauthClient.Client
	provider Provider
}

func NewWithClient(client *oauthClient.Client, provider Provider) (*Client, error) {
	if client == nil {
		return nil, errors.New("oauth client is missing")
	}
	if provider == nil {
		return nil, errors.New("provider is missing")
	}

	return &Client{
		client:   client,
		provider: provider,
	}, nil
}

func (c *Client) GetPersonalInfo(ctx context.Context, tokenSource oauth.TokenSource) (*oura.PersonalInfo, error) {
	url := c.client.ConstructURL("v2", "usercollection", "personal_info")
	responseBody := &oura.PersonalInfo{}

	if err := c.sendOuraRequest(ctx, http.MethodGet, url, nil, responseBody, tokenSource); err != nil {
		return nil, errors.Wrap(err, "unable to get user personal info")
	}

	return responseBody, nil
}

func (c *Client) GetDatum(ctx context.Context, dataType string, dataID string, tokenSource oauth.TokenSource) (*oura.Datum, error) {
	return nil, nil
}

func (c *Client) GetData(ctx context.Context, dataType string, timeRange times.TimeRange, tokenSource oauth.TokenSource) (oura.Data, error) {
	return nil, nil
}

func (c *Client) ListSubscriptions(ctx context.Context) (oura.Subscriptions, error) {
	return nil, nil
}

func (c *Client) CreateSubscription(ctx context.Context, create *oura.CreateSubscription) (*oura.Subscription, error) {
	if create == nil {
		return nil, errors.New("create is missing")
	}

	return nil, nil
}

func (c *Client) RenewSubscription(ctx context.Context, id string) (*oura.Subscription, error) {
	return nil, nil
}

func (c *Client) DeleteSubscription(ctx context.Context, id string) error {
	return nil
}

func (c *Client) RevokeOAuthToken(ctx context.Context, oauthToken *auth.OAuthToken) error {
	if oauthToken == nil {
		return errors.New("oauth token is missing")
	}

	url := c.client.ConstructURL("oauth", "revoke")
	mutators := []request.RequestMutator{request.NewHeaderMutator("Authorization", fmt.Sprintf("%s %s", oauthToken.TokenType, oauthToken.RefreshToken))}

	if err := c.sendBaseRequest(ctx, http.MethodPost, url, mutators, nil, nil, nil); err != nil {
		return errors.Wrap(err, "unable to revoke oauth token")
	}

	return nil
}

func (c *Client) sendBaseRequest(ctx context.Context, method string, url string, mutators []request.RequestMutator, requestBody any, responseBody any, inspectors []request.ResponseInspector) error {
	return log.WarnIfDurationExceedsMaximum(ctx, requestDurationMaximum, url, func(ctx context.Context) error {
		return c.client.Client().RequestDataWithHTTPClient(ctx, method, url, mutators, requestBody, responseBody, append(inspectors, prometheusCodePathResponseInspector), http.DefaultClient)
	})
}

func (c *Client) sendOuraRequest(ctx context.Context, method string, url string, requestBody any, responseBody any, tokenSource oauth.TokenSource) error {
	return log.WarnIfDurationExceedsMaximum(ctx, requestDurationMaximum, url, func(ctx context.Context) error {
		return c.client.SendOAuthRequest(ctx, method, url, nil, requestBody, responseBody, []request.ResponseInspector{prometheusCodePathResponseInspector}, tokenSource)
	})
}

const requestDurationMaximum = 30 * time.Second

var (
	prometheusCodePathPatterns          = []string{"/oauth/revoke", "/v2/usercollection/personal_info"}
	prometheusCodePathResponseInspector = request.NewPrometheusCodePathResponseInspectorWithPatterns("tidepool_oura_api_client_requests", "Oura API client requests", prometheusCodePathPatterns...)
)
