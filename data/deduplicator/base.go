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
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/log"
)

type BaseFactory struct {
	Factory
	name string
}

type BaseDeduplicator struct {
	name             string
	logger           log.Logger
	dataStoreSession store.Session
	dataset          *upload.Upload
}

func NewBaseFactory(name string) (*BaseFactory, error) {
	if name == "" {
		return nil, app.Error("deduplicator", "name is missing")
	}

	factory := &BaseFactory{
		name: name,
	}
	factory.Factory = factory

	return factory, nil
}

func (b *BaseFactory) CanDeduplicateDataset(dataset *upload.Upload) (bool, error) {
	if dataset == nil {
		return false, app.Error("deduplicator", "dataset is missing")
	}

	if dataset.UploadID == "" {
		return false, nil
	}
	if dataset.UserID == "" {
		return false, nil
	}
	if dataset.GroupID == "" {
		return false, nil
	}

	return true, nil
}

func (b *BaseFactory) NewDeduplicatorForDataset(logger log.Logger, dataStoreSession store.Session, dataset *upload.Upload) (data.Deduplicator, error) {
	return NewBaseDeduplicator(b.name, logger, dataStoreSession, dataset)
}

func (b *BaseFactory) IsRegisteredWithDataset(dataset *upload.Upload) (bool, error) {
	if can, err := b.Factory.CanDeduplicateDataset(dataset); err != nil || !can {
		return can, err
	}

	deduplicatorDescriptor := dataset.DeduplicatorDescriptor()
	if deduplicatorDescriptor == nil {
		return false, nil
	}
	if !deduplicatorDescriptor.IsRegisteredWithNamedDeduplicator(b.name) {
		return false, nil
	}

	return true, nil
}

func (b *BaseFactory) NewRegisteredDeduplicatorForDataset(logger log.Logger, dataStoreSession store.Session, dataset *upload.Upload) (data.Deduplicator, error) {
	deduplicator, err := b.Factory.NewDeduplicatorForDataset(logger, dataStoreSession, dataset)
	if err != nil {
		return nil, err
	}

	deduplicatorDescriptor := dataset.DeduplicatorDescriptor()
	if deduplicatorDescriptor == nil {
		return nil, app.Error("deduplicator", "dataset deduplicator descriptor is missing")
	}
	if !deduplicatorDescriptor.IsRegisteredWithNamedDeduplicator(b.name) {
		return nil, app.Error("deduplicator", "dataset deduplicator descriptor is not registered with expected deduplicator")
	}

	return deduplicator, nil
}

func NewBaseDeduplicator(name string, logger log.Logger, dataStoreSession store.Session, dataset *upload.Upload) (*BaseDeduplicator, error) {
	if name == "" {
		return nil, app.Error("deduplicator", "name is missing")
	}
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

	logger = logger.WithFields(log.Fields{
		"deduplicatorName": name,
		"datasetId":        dataset.UploadID,
	})

	return &BaseDeduplicator{
		name:             name,
		logger:           logger,
		dataStoreSession: dataStoreSession,
		dataset:          dataset,
	}, nil
}

func (b *BaseDeduplicator) Name() string {
	return b.name
}

func (b *BaseDeduplicator) RegisterDataset() error {
	b.logger.Debug("RegisterDataset")

	deduplicatorDescriptor := b.dataset.DeduplicatorDescriptor()

	if deduplicatorDescriptor == nil {
		deduplicatorDescriptor = data.NewDeduplicatorDescriptor()
	} else if deduplicatorDescriptor.IsRegisteredWithAnyDeduplicator() {
		return app.Errorf("deduplicator", "already registered dataset with id %s", strconv.Quote(b.dataset.UploadID))
	}
	deduplicatorDescriptor.RegisterWithNamedDeduplicator(b.name)

	b.dataset.SetDeduplicatorDescriptor(deduplicatorDescriptor)

	if err := b.dataStoreSession.UpdateDataset(b.dataset); err != nil {
		return app.ExtErrorf(err, "deduplicator", "unable to update dataset with id %s", strconv.Quote(b.dataset.UploadID))
	}

	return nil
}

func (b *BaseDeduplicator) AddDatasetData(datasetData []data.Datum) error {
	b.logger.WithField("datasetDataLength", len(datasetData)).Debug("AddDatasetData")

	if len(datasetData) == 0 {
		return nil
	}

	if err := b.dataStoreSession.CreateDatasetData(b.dataset, datasetData); err != nil {
		return app.ExtErrorf(err, "deduplicator", "unable to create dataset data with id %s", strconv.Quote(b.dataset.UploadID))
	}

	return nil
}

func (b *BaseDeduplicator) DeduplicateDataset() error {
	b.logger.Debug("DeduplicateDataset")

	if err := b.dataStoreSession.ActivateDatasetData(b.dataset); err != nil {
		return app.ExtErrorf(err, "deduplicator", "unable to activate dataset data with id %s", strconv.Quote(b.dataset.UploadID))
	}

	return nil
}

func (b *BaseDeduplicator) DeleteDataset() error {
	b.logger.Debug("DeleteDataset")

	if err := b.dataStoreSession.DeleteDataset(b.dataset); err != nil {
		return app.ExtErrorf(err, "deduplicator", "unable to delete dataset with id %s", strconv.Quote(b.dataset.UploadID))
	}

	return nil
}
