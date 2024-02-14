package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/oauth2"
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

func (m *keycloakUserAccessor) CreateUser(ctx context.Context, details *NewUserDetails) (*FullUser, error) {
	keycloakUser := &keycloakUser{
		Enabled:       details.Password != nil && *details.Password != "",
		EmailVerified: details.EmailVerified,
	}

	if keycloakUser.EmailVerified {
		// Automatically set terms accepted date when email is verified (i.e. internal usage only).
		termsAccepted := fmt.Sprintf("%v", time.Now().Unix())
		keycloakUser.Attributes = keycloakUserAttributes{
			TermsAcceptedDate: []string{termsAccepted},
		}
	}

	// Users without roles should be treated as patients to prevent keycloak from displaying
	// the role selection dialog
	if len(details.Roles) == 0 {
		details.Roles = []string{RolePatient}
	}
	if details.Username != nil {
		keycloakUser.Username = *details.Username
	}
	if len(details.Emails) > 0 {
		keycloakUser.Email = details.Emails[0]
	}
	keycloakUser.Roles = details.Roles

	keycloakUser, err := m.keycloakClient.CreateUser(ctx, keycloakUser)
	if errors.Is(err, ErrUserConflict) {
		return nil, ErrUserConflict
	}
	if err != nil {
		return nil, err
	}

	user := newUserFromKeycloakUser(keycloakUser)

	// Unclaimed custodial account should not be allowed to have a password
	if !user.IsCustodialAccount() {
		if err = m.keycloakClient.UpdateUserPassword(ctx, keycloakUser.ID, *details.Password); err != nil {
			return nil, err
		}
	}

	return user, nil
}

func (m *keycloakUserAccessor) UpdateUser(ctx context.Context, user *FullUser, details *UpdateUserDetails) (*FullUser, error) {
	emails := append([]string{}, details.Emails...)
	if details.Username != nil {
		emails = append(emails, *details.Username)
	}
	if err := m.assertEmailsUnique(ctx, user.Id, emails); err != nil {
		return nil, err
	}

	if user.IsMigrated {
		return m.updateKeycloakUser(ctx, user, details)
	}
	// expected all users to be migrated(?)
	return nil, ErrUserNotMigrated
}

func (m *keycloakUserAccessor) assertEmailsUnique(ctx context.Context, userId string, emails []string) error {
	// for _, email := range emails {
	// 	users, err := m.fallback.FindUsers(&User{
	// 		Username: email,
	// 		Emails:   emails,
	// 	})
	// 	if err != nil {
	// 		return err
	// 	}
	// 	for _, user := range users {
	// 		if user.Id != userId {
	// 			return ErrEmailConflict
	// 		}
	// 	}
	// }

	for _, email := range emails {
		user, err := m.keycloakClient.GetUserByEmail(ctx, email)
		if err != nil {
			return err
		}
		if user != nil && user.ID != userId {
			return ErrEmailConflict
		}
	}
	return nil
}

func (m *keycloakUserAccessor) updateKeycloakUser(ctx context.Context, user *FullUser, details *UpdateUserDetails) (*FullUser, error) {
	keycloakUser := userToKeycloakUser(user)
	if details.Roles != nil {
		keycloakUser.Roles = details.Roles
	}
	if details.Password != nil && len(*details.Password) > 0 {
		if err := m.keycloakClient.UpdateUserPassword(ctx, user.Id, *details.Password); err != nil {
			return nil, err
		}
		// Remove the custodial role after the password has been set
		newRoles := make([]string, 0)
		for _, role := range keycloakUser.Roles {
			if role != RoleCustodialAccount {
				newRoles = append(newRoles, role)
			}
		}
		keycloakUser.Roles = newRoles
		keycloakUser.Enabled = true
	}
	if details.Username != nil {
		keycloakUser.Username = *details.Username
	}
	if details.Emails != nil && len(details.Emails) > 0 {
		keycloakUser.Email = details.Emails[0]
	}
	if details.EmailVerified != nil {
		keycloakUser.EmailVerified = *details.EmailVerified
	}
	if details.TermsAccepted != nil && IsValidTimestamp(*details.TermsAccepted) {
		if ts, err := TimestampToUnixString(*details.TermsAccepted); err == nil {
			keycloakUser.Attributes.TermsAcceptedDate = []string{ts}
		}
	}

	err := m.keycloakClient.UpdateUser(ctx, keycloakUser)
	if err != nil {
		return nil, err
	}

	updated, err := m.keycloakClient.GetUserById(ctx, keycloakUser.ID)
	if err != nil {
		return nil, err
	}

	return newUserFromKeycloakUser(updated), nil
}

func (m *keycloakUserAccessor) FindUser(ctx context.Context, user *FullUser) (*FullUser, error) {
	var keycloakUser *keycloakUser
	var err error

	if IsValidUserID(user.Id) {
		keycloakUser, err = m.keycloakClient.GetUserById(ctx, user.Id)
	} else {
		email := ""
		if user.Emails != nil && len(user.Emails) > 0 {
			email = user.Emails[0]
		}
		keycloakUser, err = m.keycloakClient.GetUserByEmail(ctx, email)
	}

	if err != nil && err != ErrUserNotFound {
		return nil, err
	} else if err == nil && keycloakUser != nil {
		return newUserFromKeycloakUser(keycloakUser), nil
	}
	// expected all users to already be migrated(?)
	return nil, ErrUserNotMigrated
}

func (m *keycloakUserAccessor) FindUserById(ctx context.Context, id string) (*FullUser, error) {
	if !IsValidUserID(id) {
		return nil, ErrUserNotFound
	}

	keycloakUser, err := m.keycloakClient.GetUserById(ctx, id)
	if err != nil {
		return nil, err
	}
	if keycloakUser == nil {
		return nil, ErrUserNotFound
	}
	return newUserFromKeycloakUser(keycloakUser), nil
}

func (m *keycloakUserAccessor) FindUsersWithIds(ctx context.Context, ids []string) (users []*FullUser, err error) {
	keycloakUsers, err := m.keycloakClient.FindUsersWithIds(ctx, ids)
	if err != nil {
		return users, err
	}

	for _, user := range keycloakUsers {
		users = append(users, newUserFromKeycloakUser(user))
	}
	return users, nil
}

func (m *keycloakUserAccessor) RemoveUser(ctx context.Context, user *FullUser) error {
	return m.keycloakClient.DeleteUser(ctx, user.Id)
}

func (m *keycloakUserAccessor) RemoveTokensForUser(ctx context.Context, userId string) error {
	return m.keycloakClient.DeleteUserSessions(ctx, userId)
}

func (m *keycloakUserAccessor) UpdateUserProfile(ctx context.Context, userId string, p *UserProfile) error {
	return m.keycloakClient.UpdateUserProfile(ctx, userId, p)
}

func (m *keycloakUserAccessor) DeleteUserProfile(ctx context.Context, userId string) error {
	return m.keycloakClient.DeleteUserProfile(ctx, userId)
}
