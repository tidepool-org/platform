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

type AfterFactory struct{}

type AfterDeduplicator struct {
	logger           log.Logger
	dataStoreSession store.Session
	dataset          *upload.Upload
}

const AfterDeduplicatorName = "after"

// TODO: Consider using Device Model NOT Device Manufacturer to be more accurate

var AfterExpectedDeviceManufacturers = []string{"Medtronic"}

func NewAfterFactory() (*AfterFactory, error) {
	return &AfterFactory{}, nil
}

func (a *AfterFactory) CanDeduplicateDataset(dataset *upload.Upload) (bool, error) {
	if dataset == nil {
		return false, app.Error("deduplicator", "dataset is missing")
	}

	if dataset.Deduplicator != nil {
		return dataset.Deduplicator.Name == AfterDeduplicatorName, nil
	}

	if dataset.UploadID == "" || dataset.UserID == "" || dataset.GroupID == "" {
		return false, nil
	}
	if dataset.DeviceID == nil || *dataset.DeviceID == "" {
		return false, nil
	}
	if dataset.DeviceManufacturers == nil || !app.StringsContainsAnyStrings(*dataset.DeviceManufacturers, AfterExpectedDeviceManufacturers) {
		return false, nil
	}

	return true, nil
}

func (a *AfterFactory) NewDeduplicator(logger log.Logger, dataStoreSession store.Session, dataset *upload.Upload) (data.Deduplicator, error) {
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
	if !app.StringsContainsAnyStrings(*dataset.DeviceManufacturers, AfterExpectedDeviceManufacturers) {
		return nil, app.Error("deduplicator", "dataset device manufacturers does not contain expected device manufacturer")
	}

	return &AfterDeduplicator{
		logger:           logger,
		dataStoreSession: dataStoreSession,
		dataset:          dataset,
	}, nil
}

func (a *AfterDeduplicator) InitializeDataset() error {
	a.dataset.SetDeduplicatorDescriptor(&data.DeduplicatorDescriptor{Name: AfterDeduplicatorName})

	if err := a.dataStoreSession.UpdateDataset(a.dataset); err != nil {
		return app.ExtError(err, "deduplicator", "unable to initialize dataset")
	}

	return nil
}

func (a *AfterDeduplicator) AddDataToDataset(datasetData []data.Datum) error {
	if datasetData == nil {
		return app.Error("deduplicator", "dataset data is missing")
	}

	if len(datasetData) == 0 {
		return nil
	}

	if err := a.dataStoreSession.CreateDatasetData(a.dataset, datasetData); err != nil {
		return app.ExtError(err, "deduplicator", "unable to add data to dataset")
	}

	return nil
}

func (a *AfterDeduplicator) FinalizeDataset() error {
	afterTime, err := a.dataStoreSession.FindEarliestDatasetDataTime(a.dataset)
	if err != nil {
		return app.ExtErrorf(err, "deduplicator", "unable to get earliest data in dataset with id %s", strconv.Quote(a.dataset.UploadID))
	}

	// TODO: Technically, ActivateDatasetData could succeed, but DeactivateOtherDatasetDataAfterTimestamp fail. This would
	// result in duplicate (and possible incorrect) data. Is there a way to resolve this? Would be nice to have transactions.

	if err = a.dataStoreSession.ActivateDatasetData(a.dataset); err != nil {
		return app.ExtErrorf(err, "deduplicator", "unable to activate data in dataset with id %s", strconv.Quote(a.dataset.UploadID))
	}
	if afterTime != "" {
		if err = a.dataStoreSession.DeactivateOtherDatasetDataAfterTime(a.dataset, afterTime); err != nil {
			return app.ExtErrorf(err, "deduplicator", "unable to remove all other data except dataset with id %s", strconv.Quote(a.dataset.UploadID))
		}
	}

	return nil
}
