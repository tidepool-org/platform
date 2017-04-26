package deduplicator

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
)

const _HashIdentityFieldsSeparator = "|"

func AssignDatasetDataIdentityHashes(datasetData []data.Datum) ([]string, error) {
	if len(datasetData) == 0 {
		return nil, nil
	}

	hashes := []string{}
	for _, datasetDatum := range datasetData {
		fields, err := datasetDatum.IdentityFields()
		if err != nil {
			return nil, app.ExtError(err, "deduplicator", "unable to gather identity fields for datum")
		}

		hash, err := GenerateIdentityHash(fields)
		if err != nil {
			return nil, app.ExtError(err, "deduplicator", "unable to generate identity hash for datum")
		}

		deduplicatorDescriptor := datasetDatum.DeduplicatorDescriptor()
		if deduplicatorDescriptor == nil {
			deduplicatorDescriptor = data.NewDeduplicatorDescriptor()
		}
		deduplicatorDescriptor.Hash = hash

		datasetDatum.SetDeduplicatorDescriptor(deduplicatorDescriptor)

		hashes = append(hashes, hash)
	}

	return hashes, nil
}

func GenerateIdentityHash(identityFields []string) (string, error) {
	if len(identityFields) == 0 {
		return "", app.Error("deduplicator", "identity fields are missing")
	}

	for _, identityField := range identityFields {
		if identityField == "" {
			return "", app.Error("deduplicator", "identity field is empty")
		}
	}

	identityString := strings.Join(identityFields, _HashIdentityFieldsSeparator)
	identitySum := sha256.Sum256([]byte(identityString))
	identityHash := base64.StdEncoding.EncodeToString(identitySum[:])

	return identityHash, nil
}
