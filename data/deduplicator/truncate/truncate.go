package truncate

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
	"strconv"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/deduplicator"
	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/types/base/upload"
	"github.com/tidepool-org/platform/log"
)

type factory struct {
}

type truncate struct {
	logger           log.Logger
	dataStoreSession store.Session
	dataset          *upload.Upload
}

const Name = "truncate"

func NewFactory() (deduplicator.Factory, error) {
	return &factory{}, nil
}

func (f *factory) CanDeduplicateDataset(dataset *upload.Upload) (bool, error) {
	if dataset == nil {
		return false, app.Error("truncate", "dataset is missing")
	}

	if dataset.UploadID == "" || dataset.UserID == "" || dataset.GroupID == "" {
		return false, nil
	}
	if dataset.DeviceID == nil || *dataset.DeviceID == "" {
		return false, nil
	}

	if dataset.Deduplicator != nil {
		return dataset.Deduplicator.Name == Name, nil
	}

	return true, nil
}

func (f *factory) NewDeduplicator(logger log.Logger, dataStoreSession store.Session, dataset *upload.Upload) (deduplicator.Deduplicator, error) {
	if logger == nil {
		return nil, app.Error("truncate", "logger is missing")
	}
	if dataStoreSession == nil {
		return nil, app.Error("truncate", "data store session is missing")
	}
	if dataset == nil {
		return nil, app.Error("truncate", "dataset is missing")
	}
	if dataset.UploadID == "" {
		return nil, app.Error("truncate", "dataset id is missing")
	}
	if dataset.UserID == "" {
		return nil, app.Error("truncate", "dataset user id is missing")
	}
	if dataset.GroupID == "" {
		return nil, app.Error("truncate", "dataset group id is missing")
	}
	if dataset.DeviceID == nil || *dataset.DeviceID == "" {
		return nil, app.Error("truncate", "dataset device id is missing")
	}

	return &truncate{
		logger:           logger,
		dataStoreSession: dataStoreSession,
		dataset:          dataset,
	}, nil
}

func (t *truncate) InitializeDataset() error {
	t.dataset.SetDeduplicator(&upload.Deduplicator{Name: Name})

	if err := t.dataStoreSession.UpdateDataset(t.dataset); err != nil {
		return app.ExtError(err, "truncate", "unable to initialize dataset")
	}

	return nil
}

func (t *truncate) AddDataToDataset(datasetData []data.Datum) error {
	if datasetData == nil {
		return app.Error("truncate", "dataset data is missing")
	}

	if err := t.dataStoreSession.CreateDatasetData(t.dataset, datasetData); err != nil {
		return app.ExtError(err, "truncate", "unable to add data to dataset")
	}

	return nil
}

func (t *truncate) FinalizeDataset() error {
	// TODO: Technically, ActivateDatasetData could succeed, but DeleteOtherDatasetData fail. This would
	// result in duplicate (and possible incorrect) data. Is there a way to resolve this? Would be nice to have transactions.

	if err := t.dataStoreSession.ActivateDatasetData(t.dataset); err != nil {
		return app.ExtErrorf(err, "truncate", "unable to activate data in dataset with id %s", strconv.Quote(t.dataset.UploadID))
	}
	if err := t.dataStoreSession.DeleteOtherDatasetData(t.dataset); err != nil {
		return app.ExtErrorf(err, "truncate", "unable to remove all other data except dataset with id %s", strconv.Quote(t.dataset.UploadID))
	}

	return nil
}
