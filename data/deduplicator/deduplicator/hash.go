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

type deviceDeactivateHashOptions struct {
	version       DeviceDeactivateHashVersion
	legacyGroupID *string
}

func NewLegacyDeviceDeactivateHashOptions(legacyGroupID string) deviceDeactivateHashOptions {
	return deviceDeactivateHashOptions{
		version:       DeviceDeactivateHashVersionLegacy,
		legacyGroupID: &legacyGroupID,
	}
}

func NewDefaultDeviceDeactivateHashOptions() deviceDeactivateHashOptions {
	return deviceDeactivateHashOptions{
		version: DeviceDeactivateHashVersionCurrent,
	}
}

func (d deviceDeactivateHashOptions) ValidateLegacy() error {
	if d.version == DeviceDeactivateHashVersionLegacy {
		if d.legacyGroupID == nil || *d.legacyGroupID == "" {
			return errors.New("missing required legacy groupId for the device deactive hash legacy version")
		}
	}
	return nil
}

func AssignDataSetDataIdentityHashes(dataSetData data.Data, opts deviceDeactivateHashOptions) error {
	for _, dataSetDatum := range dataSetData {
		var hash string
		if opts.version == DeviceDeactivateHashVersionLegacy {
			if err := opts.ValidateLegacy(); err != nil {
				return err
			}
			fields, err := dataSetDatum.LegacyIdentityFields()
			if err != nil {
				return errors.Wrapf(err, "unable to gather legacy identity fields for datum %T", dataSetDatum)
			}

			hash, err = GenerateLegacyIdentityHash(fields)
			if err != nil {
				return errors.Wrapf(err, "unable to generate legacy identity hash for datum %T", dataSetDatum)
			}
			hash, err = GenerateLegacyIdentityHash([]string{hash, *opts.legacyGroupID})
			if err != nil {
				return errors.Wrapf(err, "unable to generate legacy identity hash with legacy groupID for datum %T", dataSetDatum)
			}
		} else {
			fields, err := dataSetDatum.IdentityFields()
			if err != nil {
				return errors.Wrap(err, "unable to gather identity fields for datum")
			}
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
