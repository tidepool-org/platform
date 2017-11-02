package service

import (
	"context"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/auth/client"
	authStore "github.com/tidepool-org/platform/auth/store"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/provider"
)

type Client struct {
	*client.External
	authStore       authStore.Store
	providerFactory provider.Factory
}

func NewClient(cfg *client.ExternalConfig, name string, logger log.Logger, authStore authStore.Store, providerFactory provider.Factory) (*Client, error) {
	if cfg == nil {
		return nil, errors.New("config is missing")
	}
	if name == "" {
		return nil, errors.New("name is missing")
	}
	if logger == nil {
		return nil, errors.New("logger is missing")
	}
	if authStore == nil {
		return nil, errors.New("auth store is missing")
	}
	if providerFactory == nil {
		return nil, errors.New("provider factory is missing")
	}

	if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "config is invalid")
	}

	external, err := client.NewExternal(cfg, name, logger)
	if err != nil {
		return nil, err
	}

	return &Client{
		External:        external,
		authStore:       authStore,
		providerFactory: providerFactory,
	}, nil
}

func (c *Client) ListUserProviderSessions(ctx context.Context, userID string, filter *auth.ProviderSessionFilter, pagination *page.Pagination) (auth.ProviderSessions, error) {
	ssn := c.authStore.NewProviderSessionSession()
	defer ssn.Close()

	return ssn.ListUserProviderSessions(ctx, userID, filter, pagination)
}

func (c *Client) CreateUserProviderSession(ctx context.Context, userID string, create *auth.ProviderSessionCreate) (*auth.ProviderSession, error) {
	prvdr, err := c.providerFactory.Get(create.Type, create.Name)
	if err != nil {
		return nil, err
	}

	ssn := c.authStore.NewProviderSessionSession()
	defer ssn.Close()

	providerSession, err := ssn.CreateUserProviderSession(ctx, userID, create)
	if err != nil {
		return nil, err
	}

	if err = prvdr.OnCreate(ctx, providerSession.UserID, providerSession.ID); err != nil {
		log.LoggerFromContext(ctx).WithError(err).WithField("providerSessionId", providerSession.ID).Error("unable to finalize creation of provider session")
		ssn.DeleteProviderSession(ctx, providerSession.ID)
		return nil, err
	}

	return providerSession, nil
}

func (c *Client) GetProviderSession(ctx context.Context, id string) (*auth.ProviderSession, error) {
	ssn := c.authStore.NewProviderSessionSession()
	defer ssn.Close()

	return ssn.GetProviderSession(ctx, id)
}

func (c *Client) UpdateProviderSession(ctx context.Context, id string, update *auth.ProviderSessionUpdate) (*auth.ProviderSession, error) {
	ssn := c.authStore.NewProviderSessionSession()
	defer ssn.Close()

	return ssn.UpdateProviderSession(ctx, id, update)
}

func (c *Client) DeleteProviderSession(ctx context.Context, id string) error {
	ssn := c.authStore.NewProviderSessionSession()
	defer ssn.Close()

	providerSession, err := ssn.GetProviderSession(ctx, id)
	if err != nil {
		return err
	} else if providerSession == nil {
		return nil
	}

	prvdr, err := c.providerFactory.Get(providerSession.Type, providerSession.Name)
	if err != nil {
		return err
	}

	if err = ssn.DeleteProviderSession(ctx, id); err != nil {
		return err
	}

	return prvdr.OnDelete(ctx, providerSession.UserID, providerSession.ID)
}

func (c *Client) ListUserRestrictedTokens(ctx context.Context, userID string, filter *auth.RestrictedTokenFilter, pagination *page.Pagination) (auth.RestrictedTokens, error) {
	ssn := c.authStore.NewRestrictedTokenSession()
	defer ssn.Close()

	return ssn.ListUserRestrictedTokens(ctx, userID, filter, pagination)
}

func (c *Client) CreateUserRestrictedToken(ctx context.Context, userID string, create *auth.RestrictedTokenCreate) (*auth.RestrictedToken, error) {
	ssn := c.authStore.NewRestrictedTokenSession()
	defer ssn.Close()

	return ssn.CreateUserRestrictedToken(ctx, userID, create)
}

func (c *Client) GetRestrictedToken(ctx context.Context, id string) (*auth.RestrictedToken, error) {
	ssn := c.authStore.NewRestrictedTokenSession()
	defer ssn.Close()

	return ssn.GetRestrictedToken(ctx, id)
}

func (c *Client) UpdateRestrictedToken(ctx context.Context, id string, update *auth.RestrictedTokenUpdate) (*auth.RestrictedToken, error) {
	ssn := c.authStore.NewRestrictedTokenSession()
	defer ssn.Close()

	return ssn.UpdateRestrictedToken(ctx, id, update)
}

func (c *Client) DeleteRestrictedToken(ctx context.Context, id string) error {
	ssn := c.authStore.NewRestrictedTokenSession()
	defer ssn.Close()

	return ssn.DeleteRestrictedToken(ctx, id)
}
