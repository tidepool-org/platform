package deduplicator

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/pointer"
)

func AssignDataSetDataIdentityHashes(dataSetData data.Data) error {
	for _, dataSetDatum := range dataSetData {
		fields, err := dataSetDatum.IdentityFields()
		if err != nil {
			return errors.Wrap(err, "unable to gather identity fields for datum")
		}

		hash, err := GenerateIdentityHash(fields)
		if err != nil {
			return errors.Wrap(err, "unable to generate identity hash for datum")
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

	identityString := strings.Join(identityFields, hashIdentityFieldsSeparator)
	identitySum := sha256.Sum256([]byte(identityString))
	identityHash := base64.StdEncoding.EncodeToString(identitySum[:])

	return identityHash, nil
}

const hashIdentityFieldsSeparator = "|"
