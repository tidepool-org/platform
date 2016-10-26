package deduplicator

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import (
	"crypto/sha256"
	"encoding/base64"
	"strconv"
	"strings"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/types/base/upload"
	"github.com/tidepool-org/platform/log"
)

type HashFactory struct{}

type HashDeduplicator struct {
	logger           log.Logger
	dataStoreSession store.Session
	dataset          *upload.Upload
}

const HashDeduplicatorName = "hash"
const HashIdentityFieldsSeparator = "|"

func NewHashFactory() (*HashFactory, error) {
	return &HashFactory{}, nil
}

func (h *HashFactory) CanDeduplicateDataset(dataset *upload.Upload) (bool, error) {
	if dataset == nil {
		return false, app.Error("deduplicator", "dataset is missing")
	}

	if dataset.Deduplicator != nil {
		return dataset.Deduplicator.Name == HashDeduplicatorName, nil
	}

	if dataset.UploadID == "" || dataset.UserID == "" || dataset.GroupID == "" {
		return false, nil
	}

	return true, nil
}

func (h *HashFactory) NewDeduplicator(logger log.Logger, dataStoreSession store.Session, dataset *upload.Upload) (data.Deduplicator, error) {
	if logger == nil {
		return nil, app.Error("deduplicator", "logger is missing")
	}
	if dataStoreSession == nil {
		return nil, app.Error("deduplicator", "data store session is missing")
	}
	if dataset == nil {
		return nil, app.Error("deduplicator", "dataset is missing")
	}
	if dataset.UploadID == "" {
		return nil, app.Error("deduplicator", "dataset id is missing")
	}
	if dataset.UserID == "" {
		return nil, app.Error("deduplicator", "dataset user id is missing")
	}
	if dataset.GroupID == "" {
		return nil, app.Error("deduplicator", "dataset group id is missing")
	}

	return &HashDeduplicator{
		logger:           logger,
		dataStoreSession: dataStoreSession,
		dataset:          dataset,
	}, nil
}

func (h *HashDeduplicator) InitializeDataset() error {
	h.dataset.SetDeduplicatorDescriptor(&data.DeduplicatorDescriptor{Name: HashDeduplicatorName})

	if err := h.dataStoreSession.UpdateDataset(h.dataset); err != nil {
		return app.ExtError(err, "deduplicator", "unable to initialize dataset")
	}

	return nil
}

func (h *HashDeduplicator) AddDataToDataset(datasetData []data.Datum) error {
	if datasetData == nil {
		return app.Error("deduplicator", "dataset data is missing")
	}

	if len(datasetData) == 0 {
		return nil
	}

	hashes := []string{}
	for _, datasetDatum := range datasetData {
		fields, err := datasetDatum.IdentityFields()
		if err != nil {
			return app.ExtError(err, "deduplicator", "unable to gather identity fields for datum")
		}

		hash, err := generateIdentityHash(fields)
		if err != nil {
			return app.ExtError(err, "deduplicator", "unable to generate identity hash for datum")
		}

		datasetDatum.SetDeduplicatorDescriptor(&data.DeduplicatorDescriptor{Name: HashDeduplicatorName, Hash: hash})

		hashes = append(hashes, hash)
	}

	hashes, err := h.dataStoreSession.FindDatasetDataDeduplicatorHashes(h.dataset.UserID, hashes)
	if err != nil {
		return app.ExtError(err, "deduplicator", "unable to find existing identity hashes")
	}

	uniqueDatasetData := []data.Datum{}
	for _, datasetDatum := range datasetData {
		if !app.StringsContainsString(hashes, datasetDatum.DeduplicatorDescriptor().Hash) {
			uniqueDatasetData = append(uniqueDatasetData, datasetDatum)
		}
	}

	if len(uniqueDatasetData) == 0 {
		return nil
	}

	if err = h.dataStoreSession.CreateDatasetData(h.dataset, uniqueDatasetData); err != nil {
		return app.ExtError(err, "deduplicator", "unable to add data to dataset")
	}

	return nil
}

func (h *HashDeduplicator) FinalizeDataset() error {
	if err := h.dataStoreSession.ActivateDatasetData(h.dataset); err != nil {
		return app.ExtErrorf(err, "deduplicator", "unable to activate data in dataset with id %s", strconv.Quote(h.dataset.UploadID))
	}

	return nil
}

func generateIdentityHash(identityFields []string) (string, error) {
	if len(identityFields) == 0 {
		return "", app.Error("deduplicator", "identity fields are missing")
	}

	for _, identityField := range identityFields {
		if identityField == "" {
			return "", app.Error("deduplicator", "identity field is empty")
		}
	}

	identityString := strings.Join(identityFields, HashIdentityFieldsSeparator)
	identitySum := sha256.Sum256([]byte(identityString))
	identityHash := base64.StdEncoding.EncodeToString(identitySum[:])

	return identityHash, nil
}
