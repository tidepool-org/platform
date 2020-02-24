package service

import (
	"context"
	"github.com/tidepool-org/platform/apple"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/auth/client"
	authStore "github.com/tidepool-org/platform/auth/store"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/provider"
)

type Client struct {
	*client.External
	appleDeviceChecker apple.DeviceCheck
	authStore          authStore.Store
	providerFactory    provider.Factory
}

func NewClient(cfg *client.ExternalConfig, authorizeAs platform.AuthorizeAs, name string, logger log.Logger, authStore authStore.Store, providerFactory provider.Factory, appleDeviceChecker apple.DeviceCheck) (*Client, error) {
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
	if appleDeviceChecker == nil {
		return nil, errors.New("apple device checker is missing")
	}

	if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "config is invalid")
	}

	external, err := client.NewExternal(cfg, authorizeAs, name, logger)
	if err != nil {
		return nil, err
	}

	return &Client{
		External:           external,
		appleDeviceChecker: appleDeviceChecker,
		authStore:          authStore,
		providerFactory:    providerFactory,
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

func (c *Client) DeleteAllProviderSessions(ctx context.Context, userID string) error {
	ctx, logger := log.ContextAndLoggerWithField(ctx, "userId", userID)

	ssn := c.authStore.NewProviderSessionSession()
	defer ssn.Close()

	// TODO: Add pagination if/when we ever get over one page of provider sessions
	if providerSessions, err := ssn.ListUserProviderSessions(ctx, userID, nil, nil); err != nil {
		logger.WithError(err).Warn("Unable to list user provider sessions")
	} else {
		for _, providerSession := range providerSessions {
			c.deleteProviderSession(ctx, ssn, providerSession)
		}
	}

	return ssn.DeleteAllProviderSessions(ctx, userID)
}

func (c *Client) deleteProviderSession(ctx context.Context, ssn authStore.ProviderSessionSession, providerSession *auth.ProviderSession) {
	ctx, logger := log.ContextAndLoggerWithField(ctx, "providerSession", providerSession)

	var prvdr provider.Provider
	prvdr, err := c.providerFactory.Get(providerSession.Type, providerSession.Name)
	if err != nil {
		logger.WithError(err).Warn("Unable to get provider")
	}

	if err = ssn.DeleteProviderSession(ctx, providerSession.ID); err != nil {
		logger.WithError(err).Warn("Unable to delete provider session")
	}

	if prvdr != nil {
		if err = prvdr.OnDelete(ctx, providerSession.UserID, providerSession.ID); err != nil {
			logger.WithError(err).Warn("Unable to delete provider session from provider")
		}
	}
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

func (c *Client) DeleteAllRestrictedTokens(ctx context.Context, userID string) error {
	ssn := c.authStore.NewRestrictedTokenSession()
	defer ssn.Close()

	return ssn.DeleteAllRestrictedTokens(ctx, userID)
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

func (c *Client) CreateUserDeviceAuthorization(ctx context.Context, userID string, create *auth.DeviceAuthorizationCreate) (*auth.DeviceAuthorization, error) {
	ssn := c.authStore.NewDeviceAuthorizationSession()
	defer ssn.Close()

	return ssn.CreateUserDeviceAuthorization(ctx, userID, create)
}

func (c *Client) GetUserDeviceAuthorization(ctx context.Context, userID string, id string) (*auth.DeviceAuthorization, error) {
	ssn := c.authStore.NewDeviceAuthorizationSession()
	defer ssn.Close()

	return ssn.GetUserDeviceAuthorization(ctx, userID, id)
}

func (c *Client) ListUserDeviceAuthorizations(ctx context.Context, userID string, pagination *page.Pagination) (auth.DeviceAuthorizations, error) {
	ssn := c.authStore.NewDeviceAuthorizationSession()
	defer ssn.Close()

	return ssn.ListUserDeviceAuthorizations(ctx, userID, pagination)
}

func (c *Client) UpdateDeviceAuthorization(ctx context.Context, id string, update *auth.DeviceAuthorizationUpdate) (*auth.DeviceAuthorization, error) {
	logger := log.LoggerFromContext(ctx)
	ssn := c.authStore.NewDeviceAuthorizationSession()
	defer ssn.Close()

	if err := auth.ValidateBundleId(update.BundleId); err != nil {
		logger.WithField("deviceAuthorizationId", id).WithError(err).Debug("invalid bundle id")
		update.Status = auth.DeviceAuthorizationFailed
	} else {
		valid, err := c.appleDeviceChecker.IsValidDeviceToken(update.DeviceCheckToken)
		if err != nil {
			return nil, err
		}

		if !valid {
			logger.WithField("deviceAuthorizationId", id).Debug("invalid device check token")
			update.Status = auth.DeviceAuthorizationFailed
		} else {
			update.Status = auth.DeviceAuthorizationSuccessful
		}
	}

	return ssn.UpdateDeviceAuthorization(ctx, id, update)
}

func (c *Client) GetDeviceAuthorizationByToken(ctx context.Context, token string) (*auth.DeviceAuthorization, error) {
	ssn := c.authStore.NewDeviceAuthorizationSession()
	defer ssn.Close()

	return ssn.GetDeviceAuthorizationByToken(ctx, token)
}
