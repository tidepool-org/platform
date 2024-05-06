package alerts

import (
	"context"
	"net/http"

	"github.com/kelseyhightower/envconfig"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/errors"
	platformlog "github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/request"
)

// Client for managing alerts configs.
type Client struct {
	client        PlatformClient
	logger        platformlog.Logger
	tokenProvider auth.ServerSessionTokenProvider
}

// NewClient builds a client for interacting with alerts API endpoints.
//
// If no logger is provided, a null logger is used.
func NewClient(client PlatformClient, tokenProvider auth.ServerSessionTokenProvider, logger platformlog.Logger) *Client {
	if logger == nil {
		logger = null.NewLogger()
	}
	return &Client{
		client:        client,
		logger:        logger,
		tokenProvider: tokenProvider,
	}
}

// platform.Client is one implementation
type PlatformClient interface {
	ConstructURL(paths ...string) string
	RequestData(ctx context.Context, method string, url string, mutators []request.RequestMutator,
		requestBody interface{}, responseBody interface{}, inspectors ...request.ResponseInspector) error
}

// TokenProvider retrieves session tokens for calling the alerts API.
//
// client.External is one implementation
type TokenProvider interface {
	// ServerSessionToken provides a server-to-server API authentication token.
	ServerSessionToken() (string, error)
}

// request performs common operations before passing a request off to the
// underlying platform.Client.
func (c *Client) request(ctx context.Context, method, url string, reqBody, resBody any) error {
	// Platform's client.Client expects a logger to exist in the request's
	// context. If it doesn't exist, request processing will panic.
	loggingCtx := platformlog.NewContextWithLogger(ctx, c.logger)
	// Make sure the auth token is injected into the request's headers.
	return c.requestWithAuth(loggingCtx, method, url, reqBody, resBody)
}

// requestWithAuth injects an auth token before calling platform.Client.RequestData.
//
// At time of writing, this is the only way to inject credentials into
// platform.Client. It might be nice to be able to use a mutator, but the auth
// is specifically handled by the platform.Client via the context field, and
// if left blank, platform.Client errors.
func (c *Client) requestWithAuth(ctx context.Context, method, url string, reqBody, resBody any) error {
	return c.client.RequestData(auth.NewContextWithServerSessionTokenProvider(ctx, c.tokenProvider), method, url, nil, reqBody, resBody)
}

// Upsert updates cfg if it exists or creates it if it doesn't.
func (c *Client) Upsert(ctx context.Context, cfg *Config) error {
	url := c.client.ConstructURL("v1", "users", cfg.FollowedUserID, "followers", cfg.UserID, "alerts")
	return c.request(ctx, http.MethodPost, url, cfg, nil)
}

// Delete the alerts config.
func (c *Client) Delete(ctx context.Context, cfg *Config) error {
	url := c.client.ConstructURL("v1", "users", cfg.FollowedUserID, "followers", cfg.UserID, "alerts")
	return c.request(ctx, http.MethodDelete, url, nil, nil)
}

// Get a user's alerts configuration for the followed user.
func (c *Client) Get(ctx context.Context, followedUserID, userID string) (*Config, error) {
	url := c.client.ConstructURL("v1", "users", followedUserID, "followers", userID, "alerts")
	cfg := &Config{}
	err := c.request(ctx, http.MethodGet, url, nil, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to request alerts config")
	}
	return cfg, nil
}

// List the alerts configurations that follow the given user.
//
// This method should only be called via an authenticated service session.
func (c *Client) List(ctx context.Context, followedUserID string) ([]*Config, error) {
	url := c.client.ConstructURL("v1", "users", followedUserID, "followers", "alerts")
	configs := []*Config{}
	err := c.request(ctx, http.MethodGet, url, nil, &configs)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to request alerts configs list")
	}
	return configs, nil
}

// ConfigLoader abstracts the method by which config values are loaded.
type ConfigLoader interface {
	Load(*ClientConfig) error
}

// envconfigLoader adapts envconfig to implement ConfigLoader.
type envconfigLoader struct {
	platform.ConfigLoader
}

// NewEnvconfigLoader loads values via envconfig.
//
// If loader is nil, it defaults to envconfig for platform values.
func NewEnvconfigLoader(loader platform.ConfigLoader) *envconfigLoader {
	if loader == nil {
		loader = platform.NewEnvconfigLoader(nil)
	}
	return &envconfigLoader{
		ConfigLoader: loader,
	}
}

// Load implements ConfigLoader.
func (l *envconfigLoader) Load(cfg *ClientConfig) error {
	if err := l.ConfigLoader.Load(cfg.Config); err != nil {
		return err
	}
	if err := envconfig.Process(client.EnvconfigEmptyPrefix, cfg); err != nil {
		return err
	}
	// Override client.Client.Address to point to the data service.
	cfg.Address = cfg.DataServiceAddress
	return nil
}

type ClientConfig struct {
	*platform.Config
	// DataServiceAddress is used to override client.Client.Address.
	DataServiceAddress string `envconfig:"TIDEPOOL_DATA_SERVICE_ADDRESS" required:"true"`
}
