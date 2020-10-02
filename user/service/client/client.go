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

	session := c.UserStructuredStore().NewSession()
	defer session.Close()

	return session.Get(ctx, id, nil)
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

	session := c.UserStructuredStore().NewSession()
	defer session.Close()

	result, err := session.Get(ctx, id, condition)
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

	deleted, err := session.Delete(ctx, id, condition)
	if err != nil {
		return false, err
	} else if !deleted {
		return false, nil
	}

	if err = c.MetricClient().RecordMetric(ctx, "users_delete", map[string]string{"userId": id}); err != nil {
		logger.WithError(err).Error("Unable to record metric for delete")
	}

	sessionSession := c.SessionStore().NewSessionsSession()
	defer sessionSession.Close()

	if err = sessionSession.DestroySessionsForUserByID(ctx, id); err != nil {
		logger.WithError(err).Error("Unable to destroy all sessions")
	}

	permissionSession := c.PermissionStore().NewPermissionsSession()
	defer permissionSession.Close()

	if err = permissionSession.DestroyPermissionsForUserByID(ctx, id); err != nil {
		logger.WithError(err).Error("Unable to destroy all permissions")
	}

	confirmationSession := c.ConfirmationStore().NewConfirmationSession()
	defer confirmationSession.Close()

	if err = confirmationSession.DeleteUserConfirmations(ctx, id); err != nil {
		logger.WithError(err).Error("Unable to destroy all confirmations")
	}

	if err = c.BlobClient().DeleteAll(ctx, id); err != nil {
		logger.WithError(err).Error("Unable to destroy all blobs")
	}

	if err = c.ImageClient().DeleteAll(ctx, id); err != nil {
		logger.WithError(err).Error("Unable to destroy all images")
	}

	messageUser := &messageStore.User{ID: id}

	profileSession := c.ProfileStore().NewSession()
	defer profileSession.Close()

	profile, err := profileSession.Get(ctx, id, nil)
	if err != nil || profile == nil || profile.FullName == nil {
		logger.WithError(err).Error("Unable to get profile name for deleted messages")
	} else {
		messageUser.FullName = *profile.FullName
	}

	messageSession := c.MessageStore().NewMessagesSession()
	defer messageSession.Close()

	if err = messageSession.DestroyMessagesForUserByID(ctx, id); err != nil {
		logger.WithError(err).Error("Unable to destroy all messages")
	}

	if err = messageSession.DeleteMessagesFromUser(ctx, messageUser); err != nil {
		logger.WithError(err).Error("Unable to delete messages from user")
	}

	if profile != nil {
		if _, err = profileSession.Destroy(ctx, id, nil); err != nil {
			logger.WithError(err).Error("Unable to destroy profile")
		}
	}

	return session.Destroy(ctx, id, nil)
}
