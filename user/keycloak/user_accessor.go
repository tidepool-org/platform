package keycloak

import (
	"context"

	"github.com/tidepool-org/platform/pointer"
	user "github.com/tidepool-org/platform/user"
)

type keycloakUserAccessor struct {
	keycloakClient *keycloakClient
}

func NewKeycloakUserAccessor(config *KeycloakConfig) *keycloakUserAccessor {
	return &keycloakUserAccessor{
		keycloakClient: newKeycloakClient(config),
	}
}

func (m *keycloakUserAccessor) Get(ctx context.Context, id string) (*user.User, error) {
	if !user.IsValidUserID(id) {
		return nil, user.ErrUserNotFound
	}

	u, err := m.keycloakClient.GetUserById(ctx, id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, user.ErrUserNotFound
	}
	return u, nil
}

func (m *keycloakUserAccessor) FindLegacyUserProfile(ctx context.Context, id string) (*user.LegacyUserProfile, error) {
	u, err := m.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if u == nil || u.Profile == nil {
		return nil, user.ErrUserProfileNotFound
	}
	return u.Profile.ToLegacyProfile(pointer.ToStringArray(u.Roles)), nil
}

func (m *keycloakUserAccessor) Roles(ctx context.Context, userID string) ([]string, error) {
	return m.keycloakClient.GetRolesForUser(ctx, userID)
}

func (m *keycloakUserAccessor) FindUsersWithIds(ctx context.Context, ids []string) (users []*user.User, err error) {
	return m.keycloakClient.FindUsersWithIds(ctx, ids)
}

func (m *keycloakUserAccessor) UpdateLegacyUserProfile(ctx context.Context, userID string, p *user.LegacyUserProfile) error {
	roles, err := m.Roles(ctx, userID)
	if err != nil {
		return err
	}
	if !user.HasClinicOrClinicianRole(roles) && p.Clinic != nil {
		p.Clinic = nil
	}
	return m.keycloakClient.UpdateUserProfile(ctx, userID, p.ToUserProfile())
}

func (m *keycloakUserAccessor) UpdateUserProfile(ctx context.Context, userID string, p *user.Profile) error {
	return m.keycloakClient.UpdateUserProfile(ctx, userID, p)
}
