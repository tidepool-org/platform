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
	structureValidator "github.com/tidepool-org/platform/structure/validator"
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

func NewDefaultClient(p Params) (*Client, error) {
	p.Logger.Debug("Loading user client config")

	cfg := platform.NewConfig()
	cfg.UserAgent = p.UserAgent
	if err := cfg.Load(p.ConfigReporter.WithScopes("user", "client")); err != nil {
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

	url := c.client.ConstructURL("v1", "users", id)
	result := &user.User{}
	if err := c.client.RequestData(ctx, http.MethodGet, url, nil, nil, result); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return result, nil
}

func (c *Client) Delete(ctx context.Context, id string, deleet *user.Delete, condition *request.Condition) (bool, error) {
	if ctx == nil {
		return false, errors.New("context is missing")
	}
	if id == "" {
		return false, errors.New("id is missing")
	} else if !user.IsValidID(id) {
		return false, errors.New("id is invalid")
	}
	if deleet != nil {
		if err := structureValidator.New().Validate(deleet); err != nil {
			return false, errors.Wrap(err, "delete is invalid")
		}
	}
	if condition == nil {
		condition = request.NewCondition()
	} else if err := structureValidator.New().Validate(condition); err != nil {
		return false, errors.Wrap(err, "condition is invalid")
	}

	url := c.client.ConstructURL("v1", "users", id)
	if err := c.client.RequestData(ctx, http.MethodDelete, url, []request.RequestMutator{condition}, deleet, nil); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
