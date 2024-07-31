package user

import (
	"cmp"
	"encoding/json"
	"errors"
	"time"

	"github.com/tidepool-org/platform/pointer"
)

var (
	ErrSeagullFieldNotFound = errors.New("seagull field not found within value object string")
)

// LegacySeagullDocument is the database model representation of the legacy
// seagull collection object. The value is a raw stringified JSON blob. TODO:
// delete once all profiles are migrated over
type LegacySeagullDocument struct {
	UserID string `bson:"userId"`
	Value  string `bson:"value"`

	// The presence of these various migration markers indicate the migration
	// status of a seagull profile into keycloak. A non nil MigrationStart and
	// nil MigrationEnd indicates an inprogress migration UNLESS MigrationError
	// is non empty, in which migration should be reattempted.
	MigrationStart *time.Time `bson:"_migrationStart,omitempty"`
	// The presence of migrationEnd means the profile is fully migrated and all reads / writes to a user profile should go through keycloak
	MigrationEnd       *time.Time `bson:"_migrationEnd,omitempty"`
	MigrationError     *string    `bson:"_migrationError,omitempty"`
	MigrationErrorTime *time.Time `bson:"_migrationErrorTime,omitempty"`
}

// ToLegacyProfile returns an object that is suitable as a JSON response - ie, the profile is not just a stringified JSON blob.
func (doc *LegacySeagullDocument) ToLegacyProfile() (*LegacyUserProfile, error) {
	valueObj, err := extractSeagullValue(doc.Value)
	if err != nil {
		return nil, err
	}
	// Unfortunately since the profile is embedded within the raw string and unmarshaled to a map[string]any, we will need Marshal and Unmarshal to our actual LegacyUserProfile object.
	profileRaw, ok := valueObj["profile"].(map[string]any)
	if !ok {
		return nil, ErrUserProfileNotFound
	}
	var legacyProfile LegacyUserProfile
	if err := MarshalThenUnmarshal(profileRaw, &legacyProfile); err != nil {
		return nil, err
	}

	// Add some default names if it is an empty name for the fake child or parent of them
	isFakeChild := legacyProfile.Patient != nil && legacyProfile.Patient.IsOtherPerson
	if isFakeChild {
		// Some fake child accounts have profiles w/ an empty patient fullName or profile fullName (but not both).
		// In this case, use the non empty name for both.
		parentName := legacyProfile.FullName
		childName := pointer.ToString(legacyProfile.Patient.FullName)
		var fullName string
		if parentName == "" || childName == "" {
			fullName = cmp.Or(parentName, childName)
			legacyProfile.Patient.FullName = &fullName
			legacyProfile.FullName = fullName
		}
	}

	legacyProfile.MigrationStatus = doc.MigrationStatus()
	return &legacyProfile, nil
}

func (doc *LegacySeagullDocument) RawValue() (valueAsMap map[string]any, err error) {
	return extractSeagullValue(doc.Value)
}

// SetRawValueProfile updates the document's jsonified Value field to contain a "profile" field with the given profile
func (doc *LegacySeagullDocument) SetRawValueProfile(profile map[string]any) error {
	valueObj, err := doc.RawValue()
	// If there was an error, just make a new field "value" value.
	if err != nil {
		valueObj = map[string]any{}
	}
	valueObj["profile"] = profile
	bytes, err := json.Marshal(valueObj)
	if err != nil {
		return err
	}
	doc.Value = string(bytes)
	return nil
}

// extractSeagullValue unmarshals the jsonified string field "value" in the
// seagull collection to a map[string]any - the reason the fields aren't
// explicitly defined is because there is / was no defined schema at the
// time for seagull, so we should preserve these fields.
func extractSeagullValue(valueRaw string) (valueAsMap map[string]any, err error) {
	var value map[string]any
	if err := json.Unmarshal([]byte(valueRaw), &value); err != nil {
		return nil, err
	}
	return value, nil
}

// AddProfileToSeagullValue takes a legacy profile and adds it to an
// existing valueObj (the unmarshaled "value" of the seagull
// collection"), then returns the marshaled version of it. It returns
// this new object as a raw string to be compatible with the seagull
// collection. This is done to preserve any non profile fields that were
// stored in the "value" field
func AddProfileToSeagullValue(valueRaw string, profile *LegacyUserProfile) (updatedValueRaw string, err error) {
	valueObj, err := extractSeagullValue(valueRaw)
	// If there was an error, just make a new field "value" value.
	if err != nil {
		valueObj = map[string]any{}
	}
	valueObj["profile"] = profile
	bytes, err := json.Marshal(valueObj)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// MarshalThenUnmarshal marshal's src into JSON, then Unmarshals that
// JSON into dst. This is useful if src has some fields fields common to
// dst but are defined explicitly or in the same way.
func MarshalThenUnmarshal(src any, dst *LegacyUserProfile) error {
	bytes, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, dst)
}

func (doc *LegacySeagullDocument) MigrationStatus() migrationStatus {
	if doc.MigrationStart != nil && doc.MigrationEnd != nil {
		return migrationCompleted
	}
	if doc.MigrationStart != nil && doc.MigrationEnd == nil && doc.MigrationError == nil {
		return migrationInProgress
	}
	if doc.MigrationStart != nil && doc.MigrationError != nil {
		return migrationError
	}
	return migrationUnmigrated
}

func (doc *LegacySeagullDocument) IsMigrating() bool {
	return doc.MigrationStatus() != migrationUnmigrated
}
