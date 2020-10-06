package client

import (
	"context"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/blob"
	confirmationStore "github.com/tidepool-org/platform/confirmation/store"
	dataClient "github.com/tidepool-org/platform/data/client"
	dataSource "github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/image"
	"github.com/tidepool-org/platform/log"
	messageStore "github.com/tidepool-org/platform/message/store"
	"github.com/tidepool-org/platform/metric"
	"github.com/tidepool-org/platform/permission"
	permissionStore "github.com/tidepool-org/platform/permission/store"
	profileStoreStructured "github.com/tidepool-org/platform/profile/store/structured"
	"github.com/tidepool-org/platform/request"
	sessionStore "github.com/tidepool-org/platform/session/store"
	"github.com/tidepool-org/platform/user"
	userStoreStructured "github.com/tidepool-org/platform/user/store/structured"
)

type PasswordHasher interface {
	HashPassword(userID string, password string) string
}

type Provider interface {
	AuthClient() auth.Client
	BlobClient() blob.Client
	DataClient() dataClient.Client
	DataSourceClient() dataSource.Client
	ImageClient() image.Client
	MetricClient() metric.Client
	PermissionClient() permission.Client

	ConfirmationStore() confirmationStore.Store
	MessageStore() messageStore.Store
	PermissionStore() permissionStore.Store
	ProfileStore() profileStoreStructured.Store
	SessionStore() sessionStore.Store
	UserStructuredStore() userStoreStructured.Store

	PasswordHasher() PasswordHasher
}

type Client struct {
	Provider
}

func New(provider Provider) (*Client, error) {
	if provider == nil {
		return nil, errors.New("provider is missing")
	}

	return &Client{
		Provider: provider,
	}, nil
}

func (c *Client) Get(ctx context.Context, id string) (*user.User, error) {
	ctx = log.ContextWithField(ctx, "id", id)

	if !c.canAccessUserAccount(ctx, id) {
		return nil, request.ErrorUnauthorized()
	}

	repository := c.UserStructuredStore().NewUserRepository()

	return repository.Get(ctx, id, nil)
}

func (c *Client) canAccessUserAccount(ctx context.Context, id string) bool {
	if _, err := c.AuthClient().EnsureAuthorizedUser(ctx, id, permission.Owner); err == nil {
		return true
	}
	if err := c.AuthClient().EnsureAuthorizedService(ctx); err == nil {
		return true
	}
	return false
}

func (c *Client) Delete(ctx context.Context, id string, deleet *user.Delete, condition *request.Condition) (bool, error) {
	ctx, logger := log.ContextAndLoggerWithFields(ctx, log.Fields{"id": id, "condition": condition})

	var requiresPassword bool
	if err := c.AuthClient().EnsureAuthorizedService(ctx); err == nil {
		requiresPassword = false
	} else if _, err = c.AuthClient().EnsureAuthorizedUser(ctx, id, permission.Owner); err == nil {
		requiresPassword = true
	} else if _, err = c.AuthClient().EnsureAuthorizedUser(ctx, id, permission.Custodian); err == nil {
		requiresPassword = false
	} else {
		return false, err
	}

	repository := c.UserStructuredStore().NewUserRepository()

	// Use a Background context for deletions, so that they still occur even if the DELETE API request times out.
	result, err := repository.Get(context.Background(), id, condition)
	if err != nil {
		return false, err
	} else if result == nil {
		return false, nil
	}

	if result.HasRole(user.RoleClinic) {
		return false, request.ErrorUnauthorized()
	}

	if deleet != nil && deleet.Password != nil {
		if result.PasswordHash == nil || *result.PasswordHash != c.PasswordHasher().HashPassword(*result.UserID, *deleet.Password) {
			return false, request.ErrorUnauthorized()
		}
	} else if requiresPassword {
		return false, request.ErrorUnauthorized()
	}

	deleted, err := repository.Delete(context.Background(), id, condition)
	if err != nil {
		return false, err
	} else if !deleted {
		return false, nil
	}

	if err = c.MetricClient().RecordMetric(context.Background(), "users_delete", map[string]string{"userId": id}); err != nil {
		logger.WithError(err).Error("Unable to record metric for delete")
	}

	tokenRepository := c.SessionStore().NewTokenRepository()

	if err = tokenRepository.DestroySessionsForUserByID(context.Background(), id); err != nil {
		logger.WithError(err).Error("Unable to destroy all sessions")
	}

	if err = c.AuthClient().DeleteAllRestrictedTokens(context.Background(), id); err != nil {
		logger.WithError(err).Error("Unable to destroy all restricted tokens")
	}

	if err = c.AuthClient().DeleteAllProviderSessions(context.Background(), id); err != nil {
		logger.WithError(err).Error("Unable to destroy all provider sessions")
	}

	permissionsRepository := c.PermissionStore().NewPermissionsRepository()

	if err = permissionsRepository.DestroyPermissionsForUserByID(context.Background(), id); err != nil {
		logger.WithError(err).Error("Unable to destroy all permissions")
	}

	confirmationRepository := c.ConfirmationStore().NewConfirmationRepository()

	if err = confirmationRepository.DeleteUserConfirmations(context.Background(), id); err != nil {
		logger.WithError(err).Error("Unable to destroy all confirmations")
	}

	if err = c.BlobClient().DeleteAll(context.Background(), id); err != nil {
		logger.WithError(err).Error("Unable to destroy all blobs")
	}

	if err = c.DataClient().DestroyDataForUserByID(context.Background(), id); err != nil {
		logger.WithError(err).Error("Unable to destroy all data")
	}

	if err = c.DataSourceClient().DeleteAll(context.Background(), id); err != nil {
		logger.WithError(err).Error("Unable to destroy all data sources")
	}

	if err = c.ImageClient().DeleteAll(context.Background(), id); err != nil {
		logger.WithError(err).Error("Unable to destroy all images")
	}

	messageUser := &messageStore.User{ID: id}

	profileRepository := c.ProfileStore().NewMetaRepository()

	profile, err := profileRepository.Get(context.Background(), id, nil)
	if err != nil || profile == nil || profile.FullName == nil {
		logger.WithError(err).Error("Unable to get profile name for deleted messages")
	} else {
		messageUser.FullName = *profile.FullName
	}

	messageRepository := c.MessageStore().NewMessageRepository()

	if err = messageRepository.DestroyMessagesForUserByID(context.Background(), id); err != nil {
		logger.WithError(err).Error("Unable to destroy all messages")
	}

	if err = messageRepository.DeleteMessagesFromUser(context.Background(), messageUser); err != nil {
		logger.WithError(err).Error("Unable to delete messages from user")
	}

	if profile != nil {
		if _, err = profileRepository.Destroy(context.Background(), id, nil); err != nil {
			logger.WithError(err).Error("Unable to destroy profile")
		}
	}

	return repository.Destroy(context.Background(), id, nil)
}
