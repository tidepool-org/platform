package user

import (
	"encoding/json"
	"time"
)

// LegacyRawSeagullProfile is database model representation of the legacy seagull collection object. The value is a raw stringified JSON blob.
// TODO: delete once all profiles are migrated over
type LegacyRawSeagullProfile struct {
	UserID string `bson:"userId"`
	Value  string `bson:"value"`
	// The presence of migrationStart in the seagull collection
	// means migration of this profile is in progress.
	MigrationStart *time.Time `bson:"migrationStart,omitempty"`
	// The presence of migrationEnd means the profile is fully migrated and all reads / writes to a user profile should go through keycloak
	MigrationEnd *time.Time `bson:"migrationEnd,omitempty"`
}

// ToLegacyProfile returns an object that is suitable as a JSON response - ie, the profile is not just a stringified JSON blob.
func (up *LegacyRawSeagullProfile) ToLegacyProfile() (*LegacyUserProfile, error) {
	var value map[string]any
	if err := json.Unmarshal([]byte(up.Value), &value); err != nil {
		return nil, err
	}
	// Unfortunately since the profile is embedded within the raw string, we will need Marshal and Unmarshal to our actual LegacyUserProfile object.
	profileRaw, ok := value["profile"].(map[string]any)
	if !ok {
		return nil, ErrUserProfileNotFound
	}
	var legacyProfile LegacyUserProfile
	if err := marshalThenUnmarshal(profileRaw, &legacyProfile); err != nil {
		return nil, err
	}
	return &legacyProfile, nil
}

// marshalThenUnmarshal marshal's src into JSON, then Unmarshals
// that JSON into dst. This is needed if we only need to marshal part of a map[string]any object into dst
func marshalThenUnmarshal(src, dst any) error {
	bytes, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, dst)
}
