package service

import (
	"context"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/auth/client"
	authStore "github.com/tidepool-org/platform/auth/store"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/provider"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type Client struct {
	*client.External
	authStore       authStore.Store
	providerFactory provider.Factory
}

func NewClient(cfg *client.ExternalConfig, authorizeAs platform.AuthorizeAs, name string, logger log.Logger, authStore authStore.Store, providerFactory provider.Factory) (*Client, error) {
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

	external, err := client.NewExternal(cfg, authorizeAs, name, logger)
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
	repository := c.authStore.NewProviderSessionRepository()
	return repository.ListUserProviderSessions(ctx, userID, filter, pagination)
}

func (c *Client) CreateUserProviderSession(ctx context.Context, userID string, create *auth.ProviderSessionCreate) (*auth.ProviderSession, error) {
	prvdr, err := c.providerFactory.Get(create.Type, create.Name)
	if err != nil {
		return nil, err
	}

	if err = prvdr.BeforeCreate(ctx, userID, create); err != nil {
		log.LoggerFromContext(ctx).WithError(err).Error("Unable to prepare provider session for creation")
		return nil, err
	}

	repository := c.authStore.NewProviderSessionRepository()

	providerSession, err := repository.CreateUserProviderSession(ctx, userID, create)
	if err != nil {
		return nil, err
	}

	ctx = log.ContextWithFields(ctx, log.Fields{
		"providerSessionId":         providerSession.ID,
		"providerSessionType":       providerSession.Type,
		"providerSessionName":       providerSession.Name,
		"providerSessionExternalId": providerSession.ExternalID,
		"userId":                    providerSession.UserID,
	})

	if err = prvdr.OnCreate(ctx, providerSession.UserID, providerSession); err != nil {
		log.LoggerFromContext(ctx).WithError(err).Error("Unable to finalize creation of provider session")
		c.deleteProviderSession(ctx, repository, providerSession)
		return nil, err
	}

	return providerSession, nil
}

func (c *Client) DeleteUserProviderSessions(ctx context.Context, userID string) error {
	ctx, logger := log.ContextAndLoggerWithField(ctx, "userId", userID)

	repository := c.authStore.NewProviderSessionRepository()

	// TODO: Add pagination if/when we ever get over one page of provider sessions
	if providerSessions, err := repository.ListUserProviderSessions(ctx, userID, nil, nil); err != nil {
		logger.WithError(err).Warn("Unable to list user provider sessions")
	} else {
		for _, providerSession := range providerSessions {
			ctx, logger := log.ContextAndLoggerWithField(ctx, "providerSessionId", providerSession.ID)

			if err := c.deleteProviderSession(ctx, repository, providerSession); err != nil {
				logger.WithError(err).Warn("Unable to delete provider session")
			}
		}
	}

	return nil
}

func (c *Client) DeleteAllProviderSessions(ctx context.Context, filter auth.ProviderSessionFilter) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(&filter); err != nil {
		return errors.Wrap(err, "filter is invalid")
	} else if filter.Type == nil {
		return errors.New("filter type is missing")
	} else if filter.Name == nil {
		return errors.New("filter name is missing")
	} else if filter.ExternalID == nil {
		return errors.New("filter external id is missing")
	}

	ctx, logger := log.ContextAndLoggerWithField(ctx, "filter", filter)

	repository := c.authStore.NewProviderSessionRepository()
	if err := page.Paginate(func(pagination page.Pagination) (bool, error) {
		providerSessions, err := repository.ListAllProviderSessions(ctx, filter, pagination)
		if err != nil {
			logger.WithError(err).Warn("Unable to list all provider sessions")
			return false, errors.Wrap(err, "unable to list all provider sessions")
		}
		for _, providerSession := range providerSessions {
			ctx, logger := log.ContextAndLoggerWithField(ctx, "providerSessionId", providerSession.ID)

			if err := c.deleteProviderSession(ctx, repository, providerSession); err != nil {
				logger.WithError(err).Warn("Unable to delete all provider session")
			}
		}
		return len(providerSessions) < pagination.Size, nil
	}); err != nil {
		return err
	}

	// We are intentionally not calling the repository to make sure we don't run into
	// a race condition and delete documents from the repository without invoking the 'OnDelete'
	// callback of each provider for documents added after the loop above has completed.
	return nil
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
	ctx, logger := log.ContextAndLoggerWithFields(ctx, log.Fields{
		"providerSessionId":         providerSession.ID,
		"providerSessionType":       providerSession.Type,
		"providerSessionName":       providerSession.Name,
		"providerSessionExternalId": providerSession.ExternalID,
		"userId":                    providerSession.UserID,
	})

	prvdr, err := c.providerFactory.Get(providerSession.Type, providerSession.Name)
	if err != nil {
		logger.WithError(err).Warn("Unable to get provider")
	} else if prvdr != nil {
		if err = prvdr.OnDelete(ctx, providerSession.UserID, providerSession); err != nil {
			logger.WithError(err).Warn("Unable to finalize deletion of provider session")
			return err
		}
	}

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
