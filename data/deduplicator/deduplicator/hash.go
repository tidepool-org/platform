package deduplicator

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/pointer"
)

type HashOptions struct {
	Version       string
	LegacyGroupID *string
}

func NewLegacyHashOptions(legacyGroupID string) HashOptions {
	return HashOptions{
		Version:       DeviceDeactivateHashVersionLegacy,
		LegacyGroupID: &legacyGroupID,
	}
}

func NewDefaultDeviceDeactivateHashOptions() HashOptions {
	return HashOptions{
		Version: DeviceDeactivateHashVersionCurrent,
	}
}

func (d HashOptions) Validate() error {

	switch d.Version {
	case DeviceDeactivateHashVersionLegacy:
		if d.LegacyGroupID == nil || *d.LegacyGroupID == "" {
			return errors.New("missing required legacy groupId for the device deactive hash legacy version")
		}
	case DeviceDeactivateHashVersionCurrent:
		if d.LegacyGroupID != nil || *d.LegacyGroupID != "" {
			return errors.New("groupId is not required for the device deactive hash current version")
		}
	default:
		return errors.Newf("missing valid version %s", d.Version)
	}
	return nil
}

func AssignDataSetDataIdentityHashes(dataSetData data.Data, options HashOptions) error {
	if err := options.Validate(); err != nil {
		return err
	}
	for _, dataSetDatum := range dataSetData {
		var hash string

		fields, err := dataSetDatum.IdentityFields(options.Version)
		if err != nil {
			return errors.Wrap(err, "unable to gather identity fields for datum")
		}

		if options.Version == DeviceDeactivateHashVersionLegacy {
			hash, err = GenerateLegacyIdentityHash(fields)
			if err != nil {
				return errors.Wrapf(err, "unable to generate legacy identity hash for datum %T", dataSetDatum)
			}
			hash, err = GenerateLegacyIdentityHash([]string{hash, *options.LegacyGroupID})
			if err != nil {
				return errors.Wrapf(err, "unable to generate legacy identity hash with legacy groupID for datum %T", dataSetDatum)
			}
		} else {
			hash, err = GenerateIdentityHash(fields)
			if err != nil {
				return errors.Wrap(err, "unable to generate identity hash for datum")
			}
		}

		deduplicator := dataSetDatum.DeduplicatorDescriptor()
		if deduplicator == nil {
			deduplicator = data.NewDeduplicatorDescriptor()
		}
		deduplicator.Hash = pointer.FromString(hash)

		dataSetDatum.SetDeduplicatorDescriptor(deduplicator)
	}

	return nil
}

func GenerateIdentityHash(identityFields []string) (string, error) {
	if len(identityFields) == 0 {
		return "", errors.New("identity fields are missing")
	}

	for _, identityField := range identityFields {
		if identityField == "" {
			return "", errors.New("identity field is empty")
		}
	}
	identityString := strings.Join(identityFields, "|")
	identitySum := sha256.Sum256([]byte(identityString))
	identityHash := base64.StdEncoding.EncodeToString(identitySum[:])
	return identityHash, nil
}

func GenerateLegacyIdentityHash(identityFields []string) (string, error) {
	if len(identityFields) == 0 {
		return "", errors.New("identity fields are missing")
	}
	hasher := sha1.New()
	for _, val := range identityFields {
		if val == "" {
			return "", errors.New("identity field is empty")
		}
		hasher.Write([]byte(fmt.Sprintf("%v", val)))
		hasher.Write([]byte("_"))
	}

	hasher.Write([]byte("bootstrap"))
	hasher.Write([]byte("_"))
	digest := hasher.Sum(nil)
	return base32.NewEncoding("0123456789abcdefghijklmnopqrstuv").WithPadding('-').EncodeToString(digest), nil
}
