package keycloak

import (
	"context"

	"github.com/tidepool-org/platform/pointer"
	userlib "github.com/tidepool-org/platform/user"
)

type keycloakUserAccessor struct {
	keycloakClient *keycloakClient
}

func NewKeycloakUserAccessor(config *KeycloakConfig) *keycloakUserAccessor {
	return &keycloakUserAccessor{
		keycloakClient: newKeycloakClient(config),
	}
}

func (m *keycloakUserAccessor) FindUser(ctx context.Context, user *userlib.User) (*userlib.User, error) {
	var foundUser *userlib.User
	var err error

	if userlib.IsValidUserID(pointer.ToString(user.UserID)) {
		foundUser, err = m.keycloakClient.GetUserById(ctx, pointer.ToString(user.UserID))
	} else {
		email := ""
		if len(user.Emails) > 0 {
			email = user.Emails[0]
		}
		foundUser, err = m.keycloakClient.GetUserByEmail(ctx, email)
	}

	if err != nil && err != userlib.ErrUserNotFound {
		return nil, err
	} else if err == nil && foundUser != nil {
		return foundUser, nil
	}
	// All users should be migrated into keycloak by the time this code is released.
	return nil, userlib.ErrUserNotMigrated
}

func (m *keycloakUserAccessor) FindUserById(ctx context.Context, id string) (*userlib.User, error) {
	if !userlib.IsValidUserID(id) {
		return nil, userlib.ErrUserNotFound
	}

	user, err := m.keycloakClient.GetUserById(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, userlib.ErrUserNotFound
	}
	return user, nil
}

func (m *keycloakUserAccessor) FindLegacyUserProfile(ctx context.Context, id string) (*userlib.LegacyUserProfile, error) {
	user, err := m.FindUserById(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil || user.Profile == nil {
		return nil, userlib.ErrUserProfileNotFound
	}
	return user.Profile.ToLegacyProfile(pointer.ToStringArray(user.Roles)), nil
}

func (m *keycloakUserAccessor) Roles(ctx context.Context, userID string) ([]string, error) {
	return m.keycloakClient.GetRolesForUser(ctx, userID)
}

func (m *keycloakUserAccessor) FindUsersWithIds(ctx context.Context, ids []string) (users []*userlib.User, err error) {
	return m.keycloakClient.FindUsersWithIds(ctx, ids)
}

func (m *keycloakUserAccessor) UpdateLegacyUserProfile(ctx context.Context, userID string, p *userlib.LegacyUserProfile) error {
	roles, err := m.Roles(ctx, userID)
	if err != nil {
		return err
	}
	if !userlib.HasClinicOrClinicianRole(roles) && p.Clinic != nil {
		p.Clinic = nil
	}
	return m.keycloakClient.UpdateUserProfile(ctx, userID, p.ToUserProfile())
}

func (m *keycloakUserAccessor) UpdateUserProfile(ctx context.Context, userID string, p *userlib.Profile) error {
	return m.keycloakClient.UpdateUserProfile(ctx, userID, p)
}
