package deduplicator

import (
	"context"

	"github.com/blang/semver"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/storeDEPRECATED"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/pointer"
)

type BaseFactory struct {
	Factory
	name    string
	version string
}

type BaseDeduplicator struct {
	name        string
	version     string
	logger      log.Logger
	dataSession storeDEPRECATED.DataSession
	dataset     *upload.Upload
}

func IsVersionValid(version string) bool {
	_, err := semver.Parse(version)
	return err == nil
}

func NewBaseFactory(name string, version string) (*BaseFactory, error) {
	if name == "" {
		return nil, errors.New("name is missing")
	}
	if version == "" {
		return nil, errors.New("version is missing")
	}
	if !IsVersionValid(version) {
		return nil, errors.New("version is invalid")
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
		return false, errors.New("dataset is missing")
	}

	if dataset.UploadID == nil || *dataset.UploadID == "" {
		return false, nil
	}
	if dataset.UserID == nil || *dataset.UserID == "" {
		return false, nil
	}

	return true, nil
}

func (b *BaseFactory) NewDeduplicatorForDataset(logger log.Logger, dataSession storeDEPRECATED.DataSession, dataset *upload.Upload) (data.Deduplicator, error) {
	return NewBaseDeduplicator(b.name, b.version, logger, dataSession, dataset)
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

func (b *BaseFactory) NewRegisteredDeduplicatorForDataset(logger log.Logger, dataSession storeDEPRECATED.DataSession, dataset *upload.Upload) (data.Deduplicator, error) {
	deduplicator, err := b.Factory.NewDeduplicatorForDataset(logger, dataSession, dataset)
	if err != nil {
		return nil, err
	}

	deduplicatorDescriptor := dataset.DeduplicatorDescriptor()
	if deduplicatorDescriptor == nil {
		return nil, errors.New("dataset deduplicator descriptor is missing")
	}
	if !deduplicatorDescriptor.IsRegisteredWithNamedDeduplicator(b.name) {
		return nil, errors.New("dataset deduplicator descriptor is not registered with expected deduplicator")
	}

	return deduplicator, nil
}

func NewBaseDeduplicator(name string, version string, logger log.Logger, dataSession storeDEPRECATED.DataSession, dataset *upload.Upload) (*BaseDeduplicator, error) {
	if name == "" {
		return nil, errors.New("name is missing")
	}
	if version == "" {
		return nil, errors.New("version is missing")
	}
	if !IsVersionValid(version) {
		return nil, errors.New("version is invalid")
	}
	if logger == nil {
		return nil, errors.New("logger is missing")
	}
	if dataSession == nil {
		return nil, errors.New("data store session is missing")
	}
	if dataset == nil {
		return nil, errors.New("dataset is missing")
	}
	if dataset.UploadID == nil {
		return nil, errors.New("dataset id is missing")
	}
	if *dataset.UploadID == "" {
		return nil, errors.New("dataset id is empty")
	}
	if dataset.UserID == nil {
		return nil, errors.New("dataset user id is missing")
	}
	if *dataset.UserID == "" {
		return nil, errors.New("dataset user id is empty")
	}

	logger = logger.WithFields(log.Fields{
		"deduplicatorName":    name,
		"deduplicatorVersion": version,
		"dataSetId":           dataset.UploadID,
	})

	return &BaseDeduplicator{
		name:        name,
		version:     version,
		logger:      logger,
		dataSession: dataSession,
		dataset:     dataset,
	}, nil
}

func (b *BaseDeduplicator) Name() string {
	return b.name
}

func (b *BaseDeduplicator) Version() string {
	return b.version
}

func (b *BaseDeduplicator) RegisterDataset(ctx context.Context) error {
	b.logger.Debug("RegisterDataset")

	deduplicatorDescriptor := b.dataset.DeduplicatorDescriptor()

	if deduplicatorDescriptor == nil {
		deduplicatorDescriptor = data.NewDeduplicatorDescriptor()
	} else if deduplicatorDescriptor.IsRegisteredWithAnyDeduplicator() {
		return errors.Newf("already registered dataset with id %q", b.dataset.UploadID)
	}
	deduplicatorDescriptor.RegisterWithDeduplicator(b)

	update := data.NewDataSetUpdate()
	update.Active = pointer.FromBool(b.dataset.Active)
	update.Deduplicator = deduplicatorDescriptor
	dataset, err := b.dataSession.UpdateDataSet(ctx, *b.dataset.UploadID, update)
	if err != nil {
		return errors.Wrapf(err, "unable to update dataset with id %q", b.dataset.UploadID)
	}
	b.dataset = dataset

	return nil
}

func (b *BaseDeduplicator) AddDatasetData(ctx context.Context, datasetData []data.Datum) error {
	b.logger.WithField("datasetDataLength", len(datasetData)).Debug("AddDatasetData")

	if len(datasetData) == 0 {
		return nil
	}

	if err := b.dataSession.CreateDatasetData(ctx, b.dataset, datasetData); err != nil {
		return errors.Wrapf(err, "unable to create dataset data with id %q", *b.dataset.UploadID)
	}

	return nil
}

func (b *BaseDeduplicator) DeduplicateDataset(ctx context.Context) error {
	b.logger.Debug("DeduplicateDataset")

	if err := b.dataSession.ActivateDatasetData(ctx, b.dataset); err != nil {
		return errors.Wrapf(err, "unable to activate dataset data with id %q", *b.dataset.UploadID)
	}

	return nil
}

func (b *BaseDeduplicator) DeleteDataset(ctx context.Context) error {
	b.logger.Debug("DeleteDataset")

	if err := b.dataSession.DeleteDataset(ctx, b.dataset); err != nil {
		return errors.Wrapf(err, "unable to delete dataset with id %q", *b.dataset.UploadID)
	}

	return nil
}
