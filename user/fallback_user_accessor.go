package user

import (
	"context"
	"errors"
	"time"
)

// FallbackLegacyUserAccessor acts as an intermediary between seagulls profile
// and the new keycloak profile. This is because prior and during migration,
// some profiles may be still in seagull. As such, FallbackLegacyUserAccessor
// will first try to retrieve from seagull. If the profile is migrated or
// doesn't exist in seagull, then it will refer to keycloak /
// FallbackLegacyUserAccessor.accessor
type FallbackLegacyUserAccessor struct {
	seagullLegacyAccessor LegacyProfileAccessor
	accessor              ProfileAccessor
	roleGetter            RoleGetter
}

func NewFallbackLegacyUserAccessor(seagullAccessor LegacyProfileAccessor, accessor ProfileAccessor, roleGetter RoleGetter) *FallbackLegacyUserAccessor {
	return &FallbackLegacyUserAccessor{
		seagullLegacyAccessor: seagullAccessor,
		accessor:              accessor,
		roleGetter:            roleGetter,
	}
}

func (f *FallbackLegacyUserAccessor) FindUserProfile(ctx context.Context, userID string) (*LegacyUserProfile, error) {
	profile, _, err := f.findUserProfile(ctx, userID)
	return profile, err
}

func (f *FallbackLegacyUserAccessor) findUserProfile(ctx context.Context, userID string) (profile *LegacyUserProfile, retrievedFromSeagull bool, err error) {
	seagullProfile, err := f.seagullLegacyAccessor.FindUserProfile(ctx, userID)
	// A not found error is OK to proceed as it may still exist in keycloak. Any
	// other errors are unexpected.
	if err != nil && !errors.Is(err, ErrUserProfileNotFound) {
		return nil, true, err
	}

	// If a profile migration to keycloak is in progress or has recently failed,
	// return the current profile from seagull if it exists instead of waiting.
	if seagullProfile != nil && !IsMigrationCompleted(seagullProfile.MigrationStatus) {
		return seagullProfile, true, nil
	}

	profile, err = f.accessor.FindUserProfile(ctx, userID)
	if err != nil {
		return nil, false, err
	}
	if profile == nil {
		return nil, false, ErrUserProfileNotFound
	}
	return profile, false, nil
}

func (f *FallbackLegacyUserAccessor) UpdateUserProfile(ctx context.Context, id string, profile *LegacyUserProfile) error {
	// retry any updates in case a migration happens sometime during this call -
	// a migration should not take more than a few seconds so this is acceptable
	// IMO.
	arbritraryRetryLimit := 3
	var err error
	for i := range arbritraryRetryLimit {
		err = f.upsertUserProfile(ctx, id, profile)
		if errors.Is(err, ErrUserProfileMigrationInProgress) {
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}
		if err != nil {
			return err
		}
	}
	return err
}

func (f *FallbackLegacyUserAccessor) UpdateUserProfileV2(ctx context.Context, userID string, profile *UserProfile) error {
	prevProfile, retrievedFromSeagull, err := f.findUserProfile(ctx, userID)
	if err != nil && !errors.Is(err, ErrUserProfileNotFound) {
		return err
	}

	// This is only meant to be called for migrated profiles so it will return an error if the profile exists unmigrated in seagull
	if prevProfile != nil && retrievedFromSeagull && prevProfile.MigrationStatus == MigrationUnmigrated {
		return ErrProfileNotMigrated
	}

	return f.accessor.UpdateUserProfileV2(ctx, userID, profile)
}

func (f *FallbackLegacyUserAccessor) upsertUserProfile(ctx context.Context, userID string, profile *LegacyUserProfile) error {
	profile, retrievedFromSeagull, err := f.findUserProfile(ctx, userID)
	if err != nil && !errors.Is(err, ErrUserProfileNotFound) {
		return err
	}

	// Any unmigrated profile that exist in seagull should be returned to the
	// user immediately. If the profile is currently being migrated,
	// ErrUserProfileMigrationInProgress will be returned to the client to re-try
	// their update as it is not expected for a migration to take more than a few
	// seconds. There is no preemptive attempt to migrate the profile on access
	// to avoid possibly migrating the profile the same time as the migrator is
	// running.
	if profile != nil && retrievedFromSeagull && profile.MigrationStatus == MigrationUnmigrated {
		return f.seagullLegacyAccessor.UpdateUserProfile(ctx, userID, profile)
	}

	// If we've reached this point, the profile has either been migrated to
	// keycloak OR it was created AFTER the release of keycloak profiles or it
	// just doesn't exist so upsert the profile into the non-legacy ProfileAccessor
	return f.accessor.UpdateUserProfile(ctx, userID, profile)
}

func (f *FallbackLegacyUserAccessor) DeleteUserProfile(ctx context.Context, userID string) error {
	profile, retrievedFromSeagull, err := f.findUserProfile(ctx, userID)
	if err != nil && !errors.Is(err, ErrUserProfileNotFound) {
		return err
	}
	if profile != nil && retrievedFromSeagull && profile.MigrationStatus == MigrationUnmigrated {
		return f.seagullLegacyAccessor.DeleteUserProfile(ctx, userID)
	}
	return f.accessor.DeleteUserProfile(ctx, userID)
}
