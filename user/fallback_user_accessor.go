package user

import (
	"context"
	"errors"
	"time"
)

// FallbackLegacyUserAccessor acts as an intermediary between seagulls
// profile and the new keycloak profile. This is because prior and during
// migration, some profiles may be still in seagull. As such,
// FallbackLegacyUserAccessor will first try to retrieve first. TODO:
// once all profiles are migrated, we can use UserProfileAccessor
// directly and get rid of this.
type FallbackLegacyUserAccessor struct {
	legacy   LegacyUserProfileAccessor
	accessor UserProfileAccessor
}

func NewFallbackLegacyUserAccessor(legacy LegacyUserProfileAccessor, accessor UserProfileAccessor) *FallbackLegacyUserAccessor {
	return &FallbackLegacyUserAccessor{
		legacy:   legacy,
		accessor: accessor,
	}
}

func (f *FallbackLegacyUserAccessor) FindUserProfile(ctx context.Context, id string) (*UserProfile, error) {
	seagullProfile, err := f.legacy.FindUserProfile(ctx, id)
	if err != nil && !errors.Is(err, ErrUserProfileNotFound) {
		return nil, err
	}
	if seagullProfile != nil && seagullProfile.MigrationStatus == migrationUnmigrated {
		return seagullProfile.ToUserProfile(), nil
	}
	profile, err := f.accessor.FindUserProfile(ctx, id)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, ErrUserProfileNotFound
	}
	return profile, nil
}

func (f *FallbackLegacyUserAccessor) UpdateUserProfile(ctx context.Context, id string, profile *UserProfile) error {
	// retry in case a migration happens during this call - a migration should not take more than a few seconds
	// so this is acceptable IMO.
	retryLimit := 3
	var err error
	for i := 0; i < retryLimit; i++ {
		err = f.updateUserProfile(ctx, id, profile)
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

func (f *FallbackLegacyUserAccessor) updateUserProfile(ctx context.Context, id string, profile *UserProfile) error {
	seagullProfile, err := f.legacy.FindUserProfile(ctx, id)
	if err != nil && !errors.Is(err, ErrUserProfileNotFound) {
		return err
	}
	// An unmigrated profile should be returned until the profile has been migrated
	if seagullProfile != nil && seagullProfile.MigrationStatus == migrationUnmigrated {
		// During an attempt to update a seagull profile, the migration process may have started in b/t the previous call and the attempt to update.
		// In this we will retry as the migration time.
		return f.legacy.UpdateUserProfile(ctx, id, profile)
	}
	return f.accessor.UpdateUserProfile(ctx, id, profile)
}

func (f *FallbackLegacyUserAccessor) DeleteUserProfile(ctx context.Context, id string) error {
	seagullProfile, err := f.legacy.FindUserProfile(ctx, id)
	if err != nil && !errors.Is(err, ErrUserProfileNotFound) {
		return err
	}
	if seagullProfile != nil && seagullProfile.MigrationStatus == migrationUnmigrated {
		return f.legacy.DeleteUserProfile(ctx, id)
	}
	return f.accessor.DeleteUserProfile(ctx, id)

}
