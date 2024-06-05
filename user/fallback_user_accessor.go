package user

import (
	"context"
	"errors"
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

func (f *FallbackLegacyUserAccessor) FindUserProfile(ctx context.Context, id string) (*LegacyUserProfile, error) {
	seagullProfile, err := f.legacy.FindUserProfile(ctx, id)
	if err != nil && !errors.Is(err, ErrUserProfileNotFound) {
		return nil, err
	}
	if seagullProfile != nil && seagullProfile.MigrationStatus == migrationUnmigrated {
		return seagullProfile, nil
	}
	profile, err := f.accessor.FindUserProfile(ctx, id)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, ErrUserProfileNotFound
	}
	return profile.ToLegacyProfile(), nil
}

func (f *FallbackLegacyUserAccessor) UpdateUserProfile(ctx context.Context, id string, p *LegacyUserProfile) error {
	seagullProfile, err := f.legacy.FindUserProfile(ctx, id)
	if err != nil && !errors.Is(err, ErrUserProfileNotFound) {
		return err
	}
	// An unmigrated profile should be returned until the profile has been migrated
	if seagullProfile != nil && seagullProfile.MigrationStatus == migrationUnmigrated {
		return f.legacy.UpdateUserProfile(ctx, id, p)
	}
	profile := p.ToUserProfile()
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
