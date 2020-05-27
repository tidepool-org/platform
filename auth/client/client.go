package client

import (
	"context"
	"net/http"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/request"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type Config struct {
	*platform.Config
	*ExternalConfig
}

func NewConfig() *Config {
	return &Config{
		Config:         platform.NewConfig(),
		ExternalConfig: NewExternalConfig(),
	}
}

func (c *Config) Load(configReporter config.Reporter) error {
	if err := c.Config.Load(configReporter); err != nil {
		return err
	}
	return c.ExternalConfig.Load(configReporter.WithScopes("external"))
}

func (c *Config) Validate() error {
	if err := c.Config.Validate(); err != nil {
		return err
	}
	return c.ExternalConfig.Validate()
}

type Client struct {
	client *platform.Client
	*External
}

func NewClient(cfg *Config, authorizeAs platform.AuthorizeAs, name string, lgr log.Logger) (*Client, error) {
	if cfg == nil {
		return nil, errors.New("config is missing")
	}
	if name == "" {
		return nil, errors.New("name is missing")
	}
	if lgr == nil {
		return nil, errors.New("logger is missing")
	}

	if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "config is invalid")
	}

	clnt, err := platform.NewClient(cfg.Config, authorizeAs)
	if err != nil {
		return nil, err
	}

	extrnl, err := NewExternal(cfg.ExternalConfig, authorizeAs, name, lgr)
	if err != nil {
		return nil, err
	}

	return &Client{
		client:   clnt,
		External: extrnl,
	}, nil
}

func (c *Client) ListUserProviderSessions(ctx context.Context, userID string, filter *auth.ProviderSessionFilter, pagination *page.Pagination) (auth.ProviderSessions, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	}
	if filter == nil {
		filter = auth.NewProviderSessionFilter()
	} else if err := structureValidator.New().Validate(filter); err != nil {
		return nil, errors.Wrap(err, "filter is invalid")
	}
	if pagination == nil {
		pagination = page.NewPagination()
	} else if err := structureValidator.New().Validate(pagination); err != nil {
		return nil, errors.Wrap(err, "pagination is invalid")
	}

	url := c.client.ConstructURL("v1", "users", userID, "provider_sessions")
	providerSessions := auth.ProviderSessions{}
	if err := c.client.RequestData(ctx, http.MethodGet, url, []request.RequestMutator{filter, pagination}, nil, &providerSessions); err != nil {
		return nil, err
	}

	return providerSessions, nil
}

func (c *Client) CreateUserProviderSession(ctx context.Context, userID string, create *auth.ProviderSessionCreate) (*auth.ProviderSession, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	}
	if create == nil {
		return nil, errors.New("create is missing")
	} else if err := structureValidator.New().Validate(create); err != nil {
		return nil, errors.Wrap(err, "create is invalid")
	}

	url := c.client.ConstructURL("v1", "users", userID, "provider_sessions")
	providerSession := &auth.ProviderSession{}
	if err := c.client.RequestData(ctx, http.MethodPost, url, nil, create, providerSession); err != nil {
		return nil, err
	}

	return providerSession, nil
}

func (c *Client) DeleteAllProviderSessions(ctx context.Context, userID string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if userID == "" {
		return errors.New("user id is missing")
	}

	url := c.client.ConstructURL("v1", "users", userID, "provider_sessions")
	return c.client.RequestData(ctx, http.MethodDelete, url, nil, nil, nil)
}

func (c *Client) GetProviderSession(ctx context.Context, id string) (*auth.ProviderSession, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	}

	url := c.client.ConstructURL("v1", "provider_sessions", id)
	providerSession := &auth.ProviderSession{}
	if err := c.client.RequestData(ctx, http.MethodGet, url, nil, nil, providerSession); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return providerSession, nil
}

func (c *Client) UpdateProviderSession(ctx context.Context, id string, update *auth.ProviderSessionUpdate) (*auth.ProviderSession, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	}
	if update == nil {
		return nil, errors.New("update is missing")
	} else if err := structureValidator.New().Validate(update); err != nil {
		return nil, errors.Wrap(err, "update is invalid")
	}

	url := c.client.ConstructURL("v1", "provider_sessions", id)
	providerSession := &auth.ProviderSession{}
	if err := c.client.RequestData(ctx, http.MethodPut, url, nil, update, providerSession); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return providerSession, nil
}

func (c *Client) DeleteProviderSession(ctx context.Context, id string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if id == "" {
		return errors.New("id is missing")
	}

	url := c.client.ConstructURL("v1", "provider_sessions", id)
	return c.client.RequestData(ctx, http.MethodDelete, url, nil, nil, nil)
}

func (c *Client) ListUserRestrictedTokens(ctx context.Context, userID string, filter *auth.RestrictedTokenFilter, pagination *page.Pagination) (auth.RestrictedTokens, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	}
	if filter == nil {
		filter = auth.NewRestrictedTokenFilter()
	} else if err := structureValidator.New().Validate(filter); err != nil {
		return nil, errors.Wrap(err, "filter is invalid")
	}
	if pagination == nil {
		pagination = page.NewPagination()
	} else if err := structureValidator.New().Validate(pagination); err != nil {
		return nil, errors.Wrap(err, "pagination is invalid")
	}

	url := c.client.ConstructURL("v1", "users", userID, "restricted_tokens")
	restrictedTokens := auth.RestrictedTokens{}
	if err := c.client.RequestData(ctx, http.MethodGet, url, []request.RequestMutator{filter, pagination}, nil, &restrictedTokens); err != nil {
		return nil, err
	}

	return restrictedTokens, nil
}

func (c *Client) DeleteAllRestrictedTokens(ctx context.Context, userID string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if userID == "" {
		return errors.New("user id is missing")
	}

	url := c.client.ConstructURL("v1", "users", userID, "restricted_tokens")
	return c.client.RequestData(ctx, http.MethodDelete, url, nil, nil, nil)
}

func (c *Client) CreateUserRestrictedToken(ctx context.Context, userID string, create *auth.RestrictedTokenCreate) (*auth.RestrictedToken, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	}
	if create == nil {
		return nil, errors.New("create is missing")
	} else if err := structureValidator.New().Validate(create); err != nil {
		return nil, errors.Wrap(err, "create is invalid")
	}

	url := c.client.ConstructURL("v1", "users", userID, "restricted_tokens")
	restrictedToken := &auth.RestrictedToken{}
	if err := c.client.RequestData(ctx, http.MethodPost, url, nil, create, restrictedToken); err != nil {
		return nil, err
	}

	return restrictedToken, nil
}

func (c *Client) GetRestrictedToken(ctx context.Context, id string) (*auth.RestrictedToken, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	}

	url := c.client.ConstructURL("v1", "restricted_tokens", id)
	restrictedToken := &auth.RestrictedToken{}
	if err := c.client.RequestData(ctx, http.MethodGet, url, nil, nil, restrictedToken); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return restrictedToken, nil
}

func (c *Client) UpdateRestrictedToken(ctx context.Context, id string, update *auth.RestrictedTokenUpdate) (*auth.RestrictedToken, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	}
	if update == nil {
		return nil, errors.New("update is missing")
	} else if err := structureValidator.New().Validate(update); err != nil {
		return nil, errors.Wrap(err, "update is invalid")
	}

	url := c.client.ConstructURL("v1", "restricted_tokens", id)
	restrictedToken := &auth.RestrictedToken{}
	if err := c.client.RequestData(ctx, http.MethodPut, url, nil, update, restrictedToken); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return restrictedToken, nil
}

func (c *Client) DeleteRestrictedToken(ctx context.Context, id string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if id == "" {
		return errors.New("id is missing")
	}

	url := c.client.ConstructURL("v1", "restricted_tokens", id)
	return c.client.RequestData(ctx, http.MethodDelete, url, nil, nil, nil)
}

func (c *Client) GetUserDeviceAuthorization(ctx context.Context, userID string, id string) (*auth.DeviceAuthorization, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	}

	url := c.client.ConstructURL("v1", "users", userID, "device_authorizations", id)
	deviceAuthorization := &auth.DeviceAuthorization{}
	if err := c.client.RequestData(ctx, http.MethodGet, url, nil, nil, deviceAuthorization); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return deviceAuthorization, nil
}

func (c *Client) ListUserDeviceAuthorizations(ctx context.Context, userID string, pagination *page.Pagination) (auth.DeviceAuthorizations, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	}
	if pagination == nil {
		pagination = page.NewPagination()
	} else if err := structureValidator.New().Validate(pagination); err != nil {
		return nil, errors.Wrap(err, "pagination is invalid")
	}

	url := c.client.ConstructURL("v1", "users", userID, "device_authorizations")
	deviceAuthorizations := auth.DeviceAuthorizations{}
	if err := c.client.RequestData(ctx, http.MethodGet, url, []request.RequestMutator{pagination}, nil, &deviceAuthorizations); err != nil {
		return nil, err
	}

	return deviceAuthorizations, nil
}

func (c *Client) CreateUserDeviceAuthorization(ctx context.Context, userID string, create *auth.DeviceAuthorizationCreate) (*auth.DeviceAuthorization, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("userID is missing")
	}
	if create == nil {
		return nil, errors.New("update is missing")
	} else if err := structureValidator.New().Validate(create); err != nil {
		return nil, errors.Wrap(err, "create is invalid")
	}

	url := c.client.ConstructURL("v1", "users", userID, "device_authorizations")
	deviceAuthorization := &auth.DeviceAuthorization{}
	if err := c.client.RequestData(ctx, http.MethodPost, url, nil, create, deviceAuthorization); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return deviceAuthorization, nil
}

func (c *Client) UpdateDeviceAuthorization(ctx context.Context, id string, update *auth.DeviceAuthorizationUpdate) (*auth.DeviceAuthorization, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	}
	if update == nil {
		return nil, errors.New("update is missing")
	} else if err := structureValidator.New().Validate(update); err != nil {
		return nil, errors.Wrap(err, "update is invalid")
	}

	url := c.client.ConstructURL("v1", "device_authorization", id)
	deviceAuthorization := &auth.DeviceAuthorization{}
	if err := c.client.RequestData(ctx, http.MethodPost, url, nil, update, deviceAuthorization); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return deviceAuthorization, nil
}

func (c *Client) GetDeviceAuthorizationByToken(ctx context.Context, token string) (*auth.DeviceAuthorization, error) {
	// private api
	return nil, errors.New("not implemented")
}
