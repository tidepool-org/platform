package user

import (
	"context"
	"errors"

	"github.com/Nerzal/gocloak/v13/pkg/jwx"
)

const (
	serverRole = "backend_service"

	TimestampFormat = "2006-01-02T15:04:05-07:00"
)

//go:generate mockgen -build_flags=--mod=mod -destination=./user_mock.go -package=user . ProfileAccessor,UserAccessor

var (
	ShorelineManagedRoles = map[string]struct{}{"patient": {}, "clinic": {}, "clinician": {}, "custodial_account": {}}

	ErrUserNotFound        = errors.New("user not found")
	ErrUserProfileNotFound = errors.New("profile not found")
	ErrUserConflict        = errors.New("user already exists")
	ErrEmailConflict       = errors.New("email already exists")
	ErrUserNotMigrated     = errors.New("user has not been migrated")
	ErrProfileNotMigrated  = errors.New("profile has not been migrated")

	// ErrUserProfileMigrationInProgress means a specific user profile is
	// currently being migrated so the client should ideally wait and
	// retry their operation again since the migration for a single user
	// should be no longer than a few seconds.
	ErrUserProfileMigrationInProgress = errors.New("user migration is in progress")
)

type LegacyProfileAccessor interface {
	FindUserProfile(ctx context.Context, userID string) (*LegacyUserProfile, error)
	UpdateUserProfile(ctx context.Context, userID string, p *LegacyUserProfile) error
	DeleteUserProfile(ctx context.Context, userID string) error
}

type ProfileAccessor interface {
	LegacyProfileAccessor
	UpdateUserProfileV2(ctx context.Context, userID string, p *UserProfile) error
}

type RoleGetter interface {
	Roles(ctx context.Context, userID string) ([]string, error)
}

// UserAccessor is the interface that can retrieve users.
// It is the equivalent of shoreline's shoreline's Storage
// interface, but for now will only retrieve user
// information.
type UserAccessor interface {
	ProfileAccessor
	RoleGetter
	FindUser(ctx context.Context, user *User) (*User, error)
	FindUserById(ctx context.Context, id string) (*User, error)
	FindUsersWithIds(ctx context.Context, ids []string) ([]*User, error)
}

type TokenIntrospectionResult struct {
	Active           bool        `json:"active"`
	Subject          string      `json:"sub"`
	EmailVerified    bool        `json:"email_verified"`
	ExpiresAt        int64       `json:"eat"`
	RealmAccess      RealmAccess `json:"realm_access"`
	IdentityProvider string      `json:"identityProvider"`
}

type AccessTokenCustomClaims struct {
	jwx.Claims
	IdentityProvider string `json:"identity_provider,omitempty"`
}

type RealmAccess struct {
	Roles []string `json:"roles"`
}

func (t *TokenIntrospectionResult) IsServerToken() bool {
	if len(t.RealmAccess.Roles) > 0 {
		for _, role := range t.RealmAccess.Roles {
			if role == serverRole {
				return true
			}
		}
	}

	return false
}
