package deduplicator

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

import (
	"strconv"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/types/base/upload"
	"github.com/tidepool-org/platform/log"
)

type TruncateFactory struct {
}

type TruncateDeduplicator struct {
	logger           log.Logger
	dataStoreSession store.Session
	dataset          *upload.Upload
}

const TruncateDeduplicatorName = "truncate"

var ExpectedDeviceManufacturers = []string{"Animas"}

func NewTruncateFactory() (*TruncateFactory, error) {
	return &TruncateFactory{}, nil
}

func (t *TruncateFactory) CanDeduplicateDataset(dataset *upload.Upload) (bool, error) {
	if dataset == nil {
		return false, app.Error("deduplicator", "dataset is missing")
	}

	if dataset.Deduplicator != nil {
		return dataset.Deduplicator.Name == TruncateDeduplicatorName, nil
	}

	if dataset.UploadID == "" || dataset.UserID == "" || dataset.GroupID == "" {
		return false, nil
	}
	if dataset.DeviceID == nil || *dataset.DeviceID == "" {
		return false, nil
	}
	if dataset.DeviceManufacturers == nil || !app.StringsContainsAnyStrings(*dataset.DeviceManufacturers, ExpectedDeviceManufacturers) {
		return false, nil
	}

	return true, nil
}

func (t *TruncateFactory) NewDeduplicator(logger log.Logger, dataStoreSession store.Session, dataset *upload.Upload) (Deduplicator, error) {
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
	if dataset.DeviceID == nil {
		return nil, app.Error("deduplicator", "dataset device id is missing")
	}
	if *dataset.DeviceID == "" {
		return nil, app.Error("deduplicator", "dataset device id is empty")
	}
	if dataset.DeviceManufacturers == nil {
		return nil, app.Error("deduplicator", "dataset device manufacturers is missing")
	}
	if !app.StringsContainsAnyStrings(*dataset.DeviceManufacturers, ExpectedDeviceManufacturers) {
		return nil, app.Error("deduplicator", "dataset device manufacturers does not contain expected device manufacturer")
	}

	return &TruncateDeduplicator{
		logger:           logger,
		dataStoreSession: dataStoreSession,
		dataset:          dataset,
	}, nil
}

func (t *TruncateDeduplicator) InitializeDataset() error {
	t.dataset.Deduplicator = &upload.Deduplicator{Name: TruncateDeduplicatorName}

	if err := t.dataStoreSession.UpdateDataset(t.dataset); err != nil {
		return app.ExtError(err, "deduplicator", "unable to initialize dataset")
	}

	return nil
}

func (t *TruncateDeduplicator) AddDataToDataset(datasetData []data.Datum) error {
	if datasetData == nil {
		return app.Error("deduplicator", "dataset data is missing")
	}

	if err := t.dataStoreSession.CreateDatasetData(t.dataset, datasetData); err != nil {
		return app.ExtError(err, "deduplicator", "unable to add data to dataset")
	}

	return nil
}

func (t *TruncateDeduplicator) FinalizeDataset() error {
	// TODO: Technically, ActivateDatasetData could succeed, but DeleteOtherDatasetData fail. This would
	// result in duplicate (and possible incorrect) data. Is there a way to resolve this? Would be nice to have transactions.

	if err := t.dataStoreSession.ActivateDatasetData(t.dataset); err != nil {
		return app.ExtErrorf(err, "deduplicator", "unable to activate data in dataset with id %s", strconv.Quote(t.dataset.UploadID))
	}
	if err := t.dataStoreSession.DeleteOtherDatasetData(t.dataset); err != nil {
		return app.ExtErrorf(err, "deduplicator", "unable to remove all other data except dataset with id %s", strconv.Quote(t.dataset.UploadID))
	}

	return nil
}
