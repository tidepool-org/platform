package user

import (
	"context"
	"errors"
)

//go:generate mockgen -build_flags=--mod=mod -destination=./user_mock.go -package=user . ProfileAccessor,UserAccessor

var (
	ShorelineManagedRoles = map[string]struct{}{"patient": {}, "clinic": {}, "clinician": {}, "custodial_account": {}}

	ErrUserNotFound        = errors.New("user not found")
	ErrUserProfileNotFound = errors.New("profile not found")
	ErrUserNotMigrated     = errors.New("user has not been migrated")
	ErrProfileNotMigrated  = errors.New("profile has not been migrated")

	// ErrUserProfileMigrationInProgress means a specific user profile is
	// currently being migrated so the client should ideally wait and
	// retry their operation again since the migration for a single user
	// should be no longer than a few seconds.
	ErrUserProfileMigrationInProgress = errors.New("user migration is in progress")
)

type LegacyProfileAccessor interface {
	FindLegacyUserProfile(ctx context.Context, userID string) (*LegacyUserProfile, error)
	UpdateLegacyUserProfile(ctx context.Context, userID string, p *LegacyUserProfile) error
}

type ProfileAccessor interface {
	LegacyProfileAccessor
	UpdateUserProfile(ctx context.Context, userID string, p *Profile) error
}

type RoleGetter interface {
	Roles(ctx context.Context, userID string) ([]string, error)
}

// UserAccessor is the interface that can retrieve users.
// It is the equivalent of shoreline's Storage interface,
// but for now will only retrieve user information.
type UserAccessor interface {
	ProfileAccessor
	RoleGetter
	Client
}
