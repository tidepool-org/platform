package service

import (
	"context"

	"github.com/tidepool-org/platform/auth"
	authClient "github.com/tidepool-org/platform/auth/client"
	authStore "github.com/tidepool-org/platform/auth/store"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/provider"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type Client struct {
	*authClient.External
	authStore       authStore.Store
	providerFactory provider.Factory
}

func NewClient(cfg *authClient.ExternalConfig, authorizeAs platform.AuthorizeAs, name string, logger log.Logger, authStore authStore.Store, providerFactory provider.Factory) (*Client, error) {
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

	external, err := authClient.NewExternal(cfg, authorizeAs, name, logger)
	if err != nil {
		return nil, err
	}

	return &Client{
		External:        external,
		authStore:       authStore,
		providerFactory: providerFactory,
	}, nil
}

func (c *Client) CreateProviderSession(ctx context.Context, create *auth.ProviderSessionCreate) (*auth.ProviderSession, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if create == nil {
		return nil, errors.New("create is missing")
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(create); err != nil {
		return nil, errors.Wrap(err, "create is invalid")
	}

	prvdr, err := c.providerFactory.Get(create.Type, create.Name)
	if err != nil {
		return nil, err
	}

	repository := c.authStore.NewProviderSessionRepository()

	providerSession, err := repository.CreateProviderSession(ctx, create)
	if err != nil {
		return nil, err
	}

	ctx = log.ContextWithField(ctx, "providerSession", log.Fields{
		"id":         providerSession.ID,
		"userId":     providerSession.UserID,
		"type":       providerSession.Type,
		"name":       providerSession.Name,
		"externalId": providerSession.ExternalID,
	})

	// From this point forward, the context should not be cancelable
	ctx = context.WithoutCancel(ctx)

	if err = prvdr.OnCreate(ctx, providerSession); err != nil {
		log.LoggerFromContext(ctx).WithError(err).Error("Unable to finalize creation of provider session")
		if err := c.deleteProviderSession(ctx, repository, providerSession); err != nil {
			log.LoggerFromContext(ctx).WithError(err).Warn("Unable to delete provider session")
		}
		return nil, err
	}

	return providerSession, nil
}

func (c *Client) DeleteProviderSessions(ctx context.Context, filter *auth.ProviderSessionFilter) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if filter == nil {
		return errors.New("filter is missing")
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(filter); err != nil {
		return errors.Wrap(err, "filter is invalid")
	}

	ctx = log.ContextWithField(ctx, "filter", filter)

	repository := c.authStore.NewProviderSessionRepository()
	_, err := page.Process(
		func(pagination page.Pagination) (auth.ProviderSessions, error) {
			return repository.ListProviderSessions(ctx, filter, &pagination)
		},
		func(providerSession *auth.ProviderSession) (*auth.ProviderSession, error) {
			ctx, logger := log.ContextAndLoggerWithField(ctx, "providerSessionId", providerSession.ID)
			if err := c.deleteProviderSession(ctx, repository, providerSession); err != nil {
				logger.WithError(err).Warn("Unable to delete provider session")
			}
			return providerSession, nil
		},
	)
	return err
}

func (c *Client) ListProviderSessions(ctx context.Context, filter *auth.ProviderSessionFilter, pagination *page.Pagination) (auth.ProviderSessions, error) {
	repository := c.authStore.NewProviderSessionRepository()
	return repository.ListProviderSessions(ctx, filter, pagination)
}

func (c *Client) GetProviderSession(ctx context.Context, id string) (*auth.ProviderSession, error) {
	repository := c.authStore.NewProviderSessionRepository()
	return repository.GetProviderSession(ctx, id)
}

func (c *Client) UpdateProviderSession(ctx context.Context, id string, update *auth.ProviderSessionUpdate) (*auth.ProviderSession, error) {
	repository := c.authStore.NewProviderSessionRepository()

	return repository.UpdateProviderSession(ctx, id, update)
}

func (c *Client) DeleteProviderSession(ctx context.Context, id string) error {
	repository := c.authStore.NewProviderSessionRepository()

	providerSession, err := repository.GetProviderSession(ctx, id)
	if err != nil {
		return err
	} else if providerSession == nil {
		return nil
	}

	return c.deleteProviderSession(ctx, repository, providerSession)
}

func (c *Client) deleteProviderSession(ctx context.Context, repository authStore.ProviderSessionRepository, providerSession *auth.ProviderSession) error {
	ctx, logger := log.ContextAndLoggerWithField(ctx, "providerSession", log.Fields{
		"id":         providerSession.ID,
		"userId":     providerSession.UserID,
		"type":       providerSession.Type,
		"name":       providerSession.Name,
		"externalId": providerSession.ExternalID,
	})

	prvdr, err := c.providerFactory.Get(providerSession.Type, providerSession.Name)
	if err != nil {
		logger.WithError(err).Warn("Unable to get provider")
	} else if prvdr != nil {
		if err = prvdr.OnDelete(ctx, providerSession); err != nil {
			logger.WithError(err).Warn("Unable to finalize deletion of provider session")
			return err
		}
	}

	// From this point forward, the context should not be cancelable
	ctx = context.WithoutCancel(ctx)

	return repository.DeleteProviderSession(ctx, providerSession.ID)
}

func (c *Client) ListUserRestrictedTokens(ctx context.Context, userID string, filter *auth.RestrictedTokenFilter, pagination *page.Pagination) (auth.RestrictedTokens, error) {
	repository := c.authStore.NewRestrictedTokenRepository()
	return repository.ListUserRestrictedTokens(ctx, userID, filter, pagination)
}

func (c *Client) CreateUserRestrictedToken(ctx context.Context, userID string, create *auth.RestrictedTokenCreate) (*auth.RestrictedToken, error) {
	repository := c.authStore.NewRestrictedTokenRepository()
	return repository.CreateUserRestrictedToken(ctx, userID, create)
}

func (c *Client) DeleteAllRestrictedTokens(ctx context.Context, userID string) error {
	repository := c.authStore.NewRestrictedTokenRepository()
	return repository.DeleteAllRestrictedTokens(ctx, userID)
}

func (c *Client) GetRestrictedToken(ctx context.Context, id string) (*auth.RestrictedToken, error) {
	repository := c.authStore.NewRestrictedTokenRepository()
	return repository.GetRestrictedToken(ctx, id)
}

func (c *Client) UpdateRestrictedToken(ctx context.Context, id string, update *auth.RestrictedTokenUpdate) (*auth.RestrictedToken, error) {
	repository := c.authStore.NewRestrictedTokenRepository()
	return repository.UpdateRestrictedToken(ctx, id, update)
}

func (c *Client) DeleteRestrictedToken(ctx context.Context, id string) error {
	repository := c.authStore.NewRestrictedTokenRepository()

	return repository.DeleteRestrictedToken(ctx, id)
}
