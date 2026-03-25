package client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/oauth"
	oauthClient "github.com/tidepool-org/platform/oauth/client"
	"github.com/tidepool-org/platform/oura"
	"github.com/tidepool-org/platform/request"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/times"
)

//go:generate mockgen -source=client.go -destination=test/client_mocks.go -package=test -typed

const (
	HeaderClientID     = "x-client-id"
	HeaderClientSecret = "x-client-secret"
)

type Provider interface {
	oura.BaseClient
}

type Client struct {
	client *oauthClient.Client
	Provider
}

func NewWithClient(client *oauthClient.Client, provider Provider) (*Client, error) {
	if client == nil {
		return nil, errors.New("client is missing")
	}
	if provider == nil {
		return nil, errors.New("provider is missing")
	}

	return &Client{
		client:   client,
		Provider: provider,
	}, nil
}

func (c *Client) ListSubscriptions(ctx context.Context) (oura.Subscriptions, error) {
	// Possible response status codes (see below for details): 200 ([]Subscription)
	url := c.client.ConstructURL("v2", "webhook", "subscription")
	subscriptions := oura.Subscriptions{}
	if err := c.sendClientRequest(ctx, http.MethodGet, url, nil, &subscriptions); err != nil {
		return nil, errors.Wrap(err, "unable to list subscriptions")
	}

	return subscriptions, nil
}

func (c *Client) CreateSubscription(ctx context.Context, create *oura.CreateSubscription) (*oura.Subscription, error) {
	if create == nil {
		return nil, errors.New("create is missing")
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(create); err != nil {
		return nil, errors.Wrap(err, "create is invalid")
	}

	// Possible response status codes (see below for details): 201 (Subscription), 422
	url := c.client.ConstructURL("v2", "webhook", "subscription")
	subscription := &oura.Subscription{}
	if err := c.sendClientRequest(ctx, http.MethodPost, url, create, subscription); err != nil {
		return nil, errors.Wrap(err, "unable to create subscription")
	}

	return subscription, nil
}

func (c *Client) UpdateSubscription(ctx context.Context, id string, update *oura.UpdateSubscription) (*oura.Subscription, error) {
	if id == "" {
		return nil, errors.New("id is missing")
	}
	if update == nil {
		return nil, errors.New("update is missing")
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(update); err != nil {
		return nil, errors.Wrap(err, "update is invalid")
	}

	// Possible response status codes (see below for details): 200 (Subscription), 403, 422
	url := c.client.ConstructURL("v2", "webhook", "subscription", id)
	subscription := &oura.Subscription{}
	if err := c.sendClientRequest(ctx, http.MethodPut, url, update, subscription); err != nil {
		if request.StatusCodeForError(err) == http.StatusForbidden {
			return nil, request.ErrorResourceNotFound() // Map unusual 403 used to indicate 404
		}
		return nil, errors.Wrap(err, "unable to update subscription")
	}

	return subscription, nil
}

func (c *Client) RenewSubscription(ctx context.Context, id string) (*oura.Subscription, error) {
	if id == "" {
		return nil, errors.New("id is missing")
	}

	// Possible response status codes (see below for details): 200 (Subscription), 403, 422
	url := c.client.ConstructURL("v2", "webhook", "subscription", "renew", id)
	subscription := &oura.Subscription{}
	if err := c.sendClientRequest(ctx, http.MethodPut, url, nil, subscription); err != nil {
		if request.StatusCodeForError(err) == http.StatusForbidden {
			return nil, request.ErrorResourceNotFound() // Map unusual 403 used to indicate 404
		}
		return nil, errors.Wrap(err, "unable to renew subscription")
	}

	return subscription, nil
}

func (c *Client) DeleteSubscription(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id is missing")
	}

	// Possible response status codes (see below for details): 204, 403, 422
	url := c.client.ConstructURL("v2", "webhook", "subscription", id)
	if err := c.sendClientRequest(ctx, http.MethodDelete, url, nil, nil); err != nil {
		if request.StatusCodeForError(err) == http.StatusForbidden {
			return request.ErrorResourceNotFound() // Map unusual 403 used to indicate 404
		}
		return errors.Wrap(err, "unable to delete subscription")
	}

	return nil
}

func (c *Client) GetPersonalInfo(ctx context.Context, tokenSource oauth.TokenSource) (*oura.PersonalInfo, error) {
	if tokenSource == nil {
		return nil, errors.New("token source is missing")
	}

	// Possible response status codes (see below for details): 200 (PersonalInfo), 400, 401, 403, 429
	url := c.client.ConstructURL("v2", "usercollection", "personal_info")
	personalInfo := &oura.PersonalInfo{}
	if err := c.sendOAuthRequest(ctx, http.MethodGet, url, nil, personalInfo, tokenSource); err != nil {
		return nil, errors.Wrap(err, "unable to get personal info")
	}

	return personalInfo, nil
}

func (c *Client) GetData(ctx context.Context, dataType string, timeRange times.TimeRange, tokenSource oauth.TokenSource) (*oura.Data, error) {
	// TODO: https://tidepool.atlassian.net/browse/BACK-4035
	return nil, nil
}

func (c *Client) GetDatum(ctx context.Context, dataType string, dataID string, tokenSource oauth.TokenSource) (*oura.Datum, error) {
	// TODO: https://tidepool.atlassian.net/browse/BACK-4034
	return nil, nil
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

func (c *Client) sendOAuthRequest(ctx context.Context, method string, url string, requestBody any, responseBody any, tokenSource oauth.TokenSource) error {
	return log.WarnIfDurationExceedsMaximum(ctx, requestDurationMaximum, url, func(ctx context.Context) error {
		return c.client.SendOAuthRequest(ctx, method, url, nil, requestBody, responseBody, []request.ResponseInspector{prometheusCodePathResponseInspector}, tokenSource)
	})
}

func (c *Client) sendClientRequest(ctx context.Context, method string, url string, requestBody any, responseBody any) error {
	mutators := []request.RequestMutator{
		request.NewHeaderMutator(HeaderClientID, c.ClientID()),
		request.NewHeaderMutator(HeaderClientSecret, c.ClientSecret()),
	}
	return c.sendBaseRequest(ctx, method, url, mutators, requestBody, responseBody, nil)
}

func (c *Client) sendBaseRequest(ctx context.Context, method string, url string, mutators []request.RequestMutator, requestBody any, responseBody any, inspectors []request.ResponseInspector) error {
	return log.WarnIfDurationExceedsMaximum(ctx, requestDurationMaximum, url, func(ctx context.Context) error {
		return c.client.Client().RequestDataWithHTTPClient(ctx, method, url, mutators, requestBody, responseBody, append(inspectors, prometheusCodePathResponseInspector), http.DefaultClient)
	})
}

const requestDurationMaximum = 30 * time.Second

var (
	prometheusCodePathPatterns = []string{
		"/v2/usercollection/daily_activity/{document_id}",
		"/v2/usercollection/daily_cardiovascular_age/{document_id}",
		"/v2/usercollection/daily_readiness/{document_id}",
		"/v2/usercollection/daily_resilience/{document_id}",
		"/v2/usercollection/daily_sleep/{document_id}",
		"/v2/usercollection/daily_spo2/{document_id}",
		"/v2/usercollection/daily_stress/{document_id}",
		"/v2/usercollection/enhanced_tag/{document_id}",
		"/v2/usercollection/rest_mode_period/{document_id}",
		"/v2/usercollection/ring_configuration/{document_id}",
		"/v2/usercollection/session/{document_id}",
		"/v2/usercollection/sleep/{document_id}",
		"/v2/usercollection/sleep_time/{document_id}",
		"/v2/usercollection/vo2_max/{document_id}",
		"/v2/usercollection/workout/{document_id}",
		"/v2/webhook/subscription/{id}",
		"/v2/webhook/subscription/renew/{id}",
		request.PatternAny,
	}
	prometheusCodePathResponseInspector = request.NewPrometheusCodePathResponseInspectorWithPatterns("tidepool_oura_api_client_requests", "Oura API client requests", prometheusCodePathPatterns...)
)

// Possible response status codes from Oura API:
// 	200: successful get/list; response body contains the requested resource(s)
// 	201: successful creation; response body contains the created resource
// 	204: successful deletion; response body empty
// 	400: invalid query parameter(s) (only on multiple document request); response body empty
// 	401: invalid or expired OAuth token; response body empty
// 	403: missing permissions or user's Oura subscription expired; response body empty
// 	404: resource id not found (only on single document request); response body empty
// 	422: request body validation error; response body is custom error (see ErrorResponseParser in oura/client/error_parser.go)
// 	429: more than 5000 requests in a 5 minute period; response body empty
