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

var (
	ShorelineManagedRoles = map[string]struct{}{"patient": {}, "clinic": {}, "clinician": {}, "custodial_account": {}}

	ErrUserNotFound    = errors.New("user not found")
	ErrUserConflict    = errors.New("user already exists")
	ErrEmailConflict   = errors.New("email already exists")
	ErrUserNotMigrated = errors.New("user has not been migrated")
)

// UserAccessor is the interface that can retrieve users.
// It is the equivalent of shoreline's shoreline's Storage
// interface, but for now will only retrieve user
// information.
type UserAccessor interface {
	FindUser(ctx context.Context, user *User) (*User, error)
	FindUserById(ctx context.Context, id string) (*User, error)
	FindUsersWithIds(ctx context.Context, ids []string) ([]*User, error)
	UpdateUserProfile(ctx context.Context, id string, p *UserProfile) error
	DeleteUserProfile(ctx context.Context, id string) error
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
