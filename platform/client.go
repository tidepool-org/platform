package platform

import (
	"context"
	"net/http"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/request"
)

type Client struct {
	*client.Client
	serviceSecret string
	httpClient    *http.Client
}

func NewClient(cfg *Config) (*Client, error) {
	if cfg == nil {
		return nil, errors.New("config is missing")
	}

	if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "config is invalid")
	}

	clnt, err := client.New(cfg.Config)
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{
		Timeout: cfg.Timeout,
	}

	return &Client{
		Client:        clnt,
		serviceSecret: cfg.ServiceSecret,
		httpClient:    httpClient,
	}, nil
}

func (c *Client) HTTPClient() *http.Client {
	return c.httpClient
}

func (c *Client) SendRequestAsUser(ctx context.Context, method string, url string, mutators []client.Mutator, requestBody interface{}, responseBody interface{}) error {
	if ctx == nil {
		return errors.New("context is missing")
	}

	details := request.DetailsFromContext(ctx)
	if details == nil {
		return errors.New("details is missing")
	}

	mutators = append(mutators, NewSessionTokenHeaderMutator(details.Token()), NewTraceMutator(ctx))

	return c.SendRequest(ctx, method, url, mutators, requestBody, responseBody, c.HTTPClient())
}

func (c *Client) SendRequestAsServer(ctx context.Context, method string, url string, mutators []client.Mutator, requestBody interface{}, responseBody interface{}) error {
	if ctx == nil {
		return errors.New("context is missing")
	}

	// TODO: Update once all services support service secret
	if c.serviceSecret != "" {
		mutators = append(mutators, NewServiceSecretHeaderMutator(c.serviceSecret))
	} else if serverSessionToken := auth.ServerSessionTokenFromContext(ctx); serverSessionToken != "" {
		mutators = append(mutators, NewSessionTokenHeaderMutator(serverSessionToken))
	} else {
		return errors.New("server session token is missing")
	}
	mutators = append(mutators, NewTraceMutator(ctx))

	return c.SendRequest(ctx, method, url, mutators, requestBody, responseBody, c.HTTPClient())
}
