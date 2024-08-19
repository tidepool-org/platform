package client

import (
	"context"
	"net/http"

	"go.uber.org/fx"

	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/log"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/user"
)

type Client struct {
	client *platform.Client
}

var ClientModule = fx.Provide(NewDefaultClient)

type Params struct {
	fx.In

	ConfigReporter config.Reporter
	Logger         log.Logger
	UserAgent      string `name:"userAgent"`
}

func NewDefaultClient(p Params) (user.Client, error) {
	p.Logger.Debug("Loading user client config")

	cfg := platform.NewConfig()
	cfg.UserAgent = p.UserAgent
	reporter := p.ConfigReporter.WithScopes("user", "client")
	loader := platform.NewConfigReporterLoader(reporter)
	if err := cfg.Load(loader); err != nil {
		return nil, errors.Wrap(err, "unable to get user client config")
	}

	p.Logger.Debug("Creating user client")

	clnt, err := New(cfg, platform.AuthorizeAsService)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create user client")
	}

	return clnt, nil
}

func New(config *platform.Config, authorizeAs platform.AuthorizeAs) (*Client, error) {
	client, err := platform.NewClient(config, authorizeAs)
	if err != nil {
		return nil, err
	}

	return &Client{
		client: client,
	}, nil
}

func (c *Client) Get(ctx context.Context, id string) (*user.User, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	} else if !user.IsValidID(id) {
		return nil, errors.New("id is invalid")
	}

	url := c.client.ConstructURL("auth", "user", id)
	result := &user.User{}
	if err := c.client.RequestData(ctx, http.MethodGet, url, nil, nil, result); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return result, nil
}
