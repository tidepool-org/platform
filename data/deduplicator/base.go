package deduplicator

import (
	"strconv"

	"github.com/blang/semver"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
)

type BaseFactory struct {
	Factory
	name    string
	version string
}

type BaseDeduplicator struct {
	name             string
	version          string
	logger           log.Logger
	dataStoreSession store.Session
	dataset          *upload.Upload
}

func IsVersionValid(version string) bool {
	_, err := semver.Parse(version)
	return err == nil
}

func NewBaseFactory(name string, version string) (*BaseFactory, error) {
	if name == "" {
		return nil, errors.New("deduplicator", "name is missing")
	}
	if version == "" {
		return nil, errors.New("deduplicator", "version is missing")
	}
	if !IsVersionValid(version) {
		return nil, errors.New("deduplicator", "version is invalid")
	}

	factory := &BaseFactory{
		name:    name,
		version: version,
	}
	factory.Factory = factory

	return factory, nil
}

func (b *BaseFactory) CanDeduplicateDataset(dataset *upload.Upload) (bool, error) {
	if dataset == nil {
		return false, errors.New("deduplicator", "dataset is missing")
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
	return NewBaseDeduplicator(b.name, b.version, logger, dataStoreSession, dataset)
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
		return nil, errors.New("deduplicator", "dataset deduplicator descriptor is missing")
	}
	if !deduplicatorDescriptor.IsRegisteredWithNamedDeduplicator(b.name) {
		return nil, errors.New("deduplicator", "dataset deduplicator descriptor is not registered with expected deduplicator")
	}

	return deduplicator, nil
}

func NewBaseDeduplicator(name string, version string, logger log.Logger, dataStoreSession store.Session, dataset *upload.Upload) (*BaseDeduplicator, error) {
	if name == "" {
		return nil, errors.New("deduplicator", "name is missing")
	}
	if version == "" {
		return nil, errors.New("deduplicator", "version is missing")
	}
	if !IsVersionValid(version) {
		return nil, errors.New("deduplicator", "version is invalid")
	}
	if logger == nil {
		return nil, errors.New("deduplicator", "logger is missing")
	}
	if dataStoreSession == nil {
		return nil, errors.New("deduplicator", "data store session is missing")
	}
	if dataset == nil {
		return nil, errors.New("deduplicator", "dataset is missing")
	}
	if dataset.UploadID == "" {
		return nil, errors.New("deduplicator", "dataset id is missing")
	}
	if dataset.UserID == "" {
		return nil, errors.New("deduplicator", "dataset user id is missing")
	}
	if dataset.GroupID == "" {
		return nil, errors.New("deduplicator", "dataset group id is missing")
	}

	logger = logger.WithFields(log.Fields{
		"deduplicatorName":    name,
		"deduplicatorVersion": version,
		"datasetId":           dataset.UploadID,
	})

	return &BaseDeduplicator{
		name:             name,
		version:          version,
		logger:           logger,
		dataStoreSession: dataStoreSession,
		dataset:          dataset,
	}, nil
}

func (b *BaseDeduplicator) Name() string {
	return b.name
}

func (b *BaseDeduplicator) Version() string {
	return b.version
}

func (b *BaseDeduplicator) RegisterDataset() error {
	b.logger.Debug("RegisterDataset")

	deduplicatorDescriptor := b.dataset.DeduplicatorDescriptor()

	if deduplicatorDescriptor == nil {
		deduplicatorDescriptor = data.NewDeduplicatorDescriptor()
	} else if deduplicatorDescriptor.IsRegisteredWithAnyDeduplicator() {
		return errors.Newf("deduplicator", "already registered dataset with id %s", strconv.Quote(b.dataset.UploadID))
	}
	deduplicatorDescriptor.RegisterWithDeduplicator(b)

	b.dataset.SetDeduplicatorDescriptor(deduplicatorDescriptor)

	if err := b.dataStoreSession.UpdateDataset(b.dataset); err != nil {
		return errors.Wrapf(err, "deduplicator", "unable to update dataset with id %s", strconv.Quote(b.dataset.UploadID))
	}

	return nil
}

func (b *BaseDeduplicator) AddDatasetData(datasetData []data.Datum) error {
	b.logger.WithField("datasetDataLength", len(datasetData)).Debug("AddDatasetData")

	if len(datasetData) == 0 {
		return nil
	}

	if err := b.dataStoreSession.CreateDatasetData(b.dataset, datasetData); err != nil {
		return errors.Wrapf(err, "deduplicator", "unable to create dataset data with id %s", strconv.Quote(b.dataset.UploadID))
	}

	return nil
}

func (b *BaseDeduplicator) DeduplicateDataset() error {
	b.logger.Debug("DeduplicateDataset")

	if err := b.dataStoreSession.ActivateDatasetData(b.dataset); err != nil {
		return errors.Wrapf(err, "deduplicator", "unable to activate dataset data with id %s", strconv.Quote(b.dataset.UploadID))
	}

	return nil
}

func (b *BaseDeduplicator) DeleteDataset() error {
	b.logger.Debug("DeleteDataset")

	if err := b.dataStoreSession.DeleteDataset(b.dataset); err != nil {
		return errors.Wrapf(err, "deduplicator", "unable to delete dataset with id %s", strconv.Quote(b.dataset.UploadID))
	}

	return nil
}
