package keycloak

import (
	"context"
	"time"

	"golang.org/x/oauth2"

	"github.com/tidepool-org/platform/pointer"
	userLib "github.com/tidepool-org/platform/user"
)

type keycloakUserAccessor struct {
	cfg                      *KeycloakConfig
	adminToken               *oauth2.Token
	adminTokenRefreshExpires time.Time
	keycloakClient           *keycloakClient
}

func NewKeycloakUserAccessor(config *KeycloakConfig) *keycloakUserAccessor {
	newKeycloakClient(config)
	return &keycloakUserAccessor{
		cfg:            config,
		keycloakClient: newKeycloakClient(config),
	}
}

func (m *keycloakUserAccessor) FindUser(ctx context.Context, user *userLib.User) (*userLib.User, error) {
	var keycloakUser *keycloakUser
	var err error

	if userLib.IsValidUserID(pointer.ToString(user.UserID)) {
		keycloakUser, err = m.keycloakClient.GetUserById(ctx, pointer.ToString(user.UserID))
	} else {
		email := ""
		if user.Emails != nil && len(user.Emails) > 0 {
			email = user.Emails[0]
		}
		keycloakUser, err = m.keycloakClient.GetUserByEmail(ctx, email)
	}

	if err != nil && err != userLib.ErrUserNotFound {
		return nil, err
	} else if err == nil && keycloakUser != nil {
		return newUserFromKeycloakUser(keycloakUser), nil
	}
	// expected all users to already be migrated(?)
	return nil, userLib.ErrUserNotMigrated
}

func (m *keycloakUserAccessor) FindUserById(ctx context.Context, id string) (*userLib.User, error) {
	if !userLib.IsValidUserID(id) {
		return nil, userLib.ErrUserNotFound
	}

	keycloakUser, err := m.keycloakClient.GetUserById(ctx, id)
	if err != nil {
		return nil, err
	}
	if keycloakUser == nil {
		return nil, userLib.ErrUserNotFound
	}
	return newUserFromKeycloakUser(keycloakUser), nil
}

func (m *keycloakUserAccessor) FindUsersWithIds(ctx context.Context, ids []string) (users []*userLib.User, err error) {
	keycloakUsers, err := m.keycloakClient.FindUsersWithIds(ctx, ids)
	if err != nil {
		return users, err
	}

	for _, user := range keycloakUsers {
		users = append(users, newUserFromKeycloakUser(user))
	}
	return users, nil
}

func (m *keycloakUserAccessor) UpdateUserProfile(ctx context.Context, userId string, p *userLib.UserProfile) error {
	return m.keycloakClient.UpdateUserProfile(ctx, userId, p)
}

func (m *keycloakUserAccessor) DeleteUserProfile(ctx context.Context, userId string) error {
	return m.keycloakClient.DeleteUserProfile(ctx, userId)
}
