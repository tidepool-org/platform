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

func AssignDataSetDataIdentityHashes(dataSetData data.Data, hasher HashType) error {
	for _, dataSetDatum := range dataSetData {
		var hash string
		if hasher == PlatformHash {
			fields, err := dataSetDatum.IdentityFields()
			if err != nil {
				return errors.Wrap(err, "unable to gather identity fields for datum")
			}
			hash, err = GenerateIdentityHash(fields)
			if err != nil {
				return errors.Wrap(err, "unable to generate identity hash for datum")
			}
		} else if hasher == LegacyHash {
			fields, err := dataSetDatum.IdentityFields()
			if err != nil {
				return errors.Wrap(err, "unable to gather legacy identity fields for datum")
			}
			hash, err = GenerateLegacyIdentityHash(fields)
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
	for _, identityField := range identityFields {
		if identityField == "" {
			return "", errors.New("identity field is empty")
		}
		hasher.Write([]byte(fmt.Sprintf("%v_", identityField)))
	}
	hasher.Write([]byte("bootstrap_"))
	hash := hasher.Sum(nil)
	return base32.NewEncoding("0123456789abcdefghijklmnopqrstuv").WithPadding('-').EncodeToString(hash), nil
}
