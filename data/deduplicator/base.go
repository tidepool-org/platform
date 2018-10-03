package deduplicator

import (
	"context"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/storeDEPRECATED"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/net"
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
	dataSet     *upload.Upload
}

func NewBaseFactory(name string, version string) (*BaseFactory, error) {
	if name == "" {
		return nil, errors.New("name is missing")
	}
	if version == "" {
		return nil, errors.New("version is missing")
	}
	if !net.IsValidSemanticVersion(version) {
		return nil, errors.New("version is invalid")
	}

	factory := &BaseFactory{
		name:    name,
		version: version,
	}
	factory.Factory = factory

	return factory, nil
}

func (b *BaseFactory) CanDeduplicateDataSet(dataSet *upload.Upload) (bool, error) {
	if dataSet == nil {
		return false, errors.New("data set is missing")
	}

	if dataSet.UploadID == nil || *dataSet.UploadID == "" {
		return false, nil
	}
	if dataSet.UserID == nil || *dataSet.UserID == "" {
		return false, nil
	}

	return true, nil
}

func (b *BaseFactory) NewDeduplicatorForDataSet(logger log.Logger, dataSession storeDEPRECATED.DataSession, dataSet *upload.Upload) (data.Deduplicator, error) {
	return NewBaseDeduplicator(b.name, b.version, logger, dataSession, dataSet)
}

func (b *BaseFactory) IsRegisteredWithDataSet(dataSet *upload.Upload) (bool, error) {
	if can, err := b.Factory.CanDeduplicateDataSet(dataSet); err != nil || !can {
		return can, err
	}

	deduplicatorDescriptor := dataSet.DeduplicatorDescriptor()
	if deduplicatorDescriptor == nil {
		return false, nil
	}
	if !deduplicatorDescriptor.IsRegisteredWithNamedDeduplicator(b.name) {
		return false, nil
	}

	return true, nil
}

func (b *BaseFactory) NewRegisteredDeduplicatorForDataSet(logger log.Logger, dataSession storeDEPRECATED.DataSession, dataSet *upload.Upload) (data.Deduplicator, error) {
	deduplicator, err := b.Factory.NewDeduplicatorForDataSet(logger, dataSession, dataSet)
	if err != nil {
		return nil, err
	}

	deduplicatorDescriptor := dataSet.DeduplicatorDescriptor()
	if deduplicatorDescriptor == nil {
		return nil, errors.New("data set deduplicator descriptor is missing")
	}
	if !deduplicatorDescriptor.IsRegisteredWithNamedDeduplicator(b.name) {
		return nil, errors.New("data set deduplicator descriptor is not registered with expected deduplicator")
	}

	return deduplicator, nil
}

func NewBaseDeduplicator(name string, version string, logger log.Logger, dataSession storeDEPRECATED.DataSession, dataSet *upload.Upload) (*BaseDeduplicator, error) {
	if name == "" {
		return nil, errors.New("name is missing")
	}
	if version == "" {
		return nil, errors.New("version is missing")
	}
	if !net.IsValidSemanticVersion(version) {
		return nil, errors.New("version is invalid")
	}
	if logger == nil {
		return nil, errors.New("logger is missing")
	}
	if dataSession == nil {
		return nil, errors.New("data store session is missing")
	}
	if dataSet == nil {
		return nil, errors.New("data set is missing")
	}
	if dataSet.UploadID == nil {
		return nil, errors.New("data set id is missing")
	}
	if *dataSet.UploadID == "" {
		return nil, errors.New("data set id is empty")
	}
	if dataSet.UserID == nil {
		return nil, errors.New("data set user id is missing")
	}
	if *dataSet.UserID == "" {
		return nil, errors.New("data set user id is empty")
	}

	logger = logger.WithFields(log.Fields{
		"deduplicatorName":    name,
		"deduplicatorVersion": version,
		"dataSetId":           dataSet.UploadID,
	})

	return &BaseDeduplicator{
		name:        name,
		version:     version,
		logger:      logger,
		dataSession: dataSession,
		dataSet:     dataSet,
	}, nil
}

func (b *BaseDeduplicator) Name() string {
	return b.name
}

func (b *BaseDeduplicator) Version() string {
	return b.version
}

func (b *BaseDeduplicator) RegisterDataSet(ctx context.Context) error {
	b.logger.Debug("RegisterDataSet")

	deduplicatorDescriptor := b.dataSet.DeduplicatorDescriptor()

	if deduplicatorDescriptor == nil {
		deduplicatorDescriptor = data.NewDeduplicatorDescriptor()
	} else if deduplicatorDescriptor.IsRegisteredWithAnyDeduplicator() {
		return errors.Newf("already registered data set with id %q", b.dataSet.UploadID)
	}
	deduplicatorDescriptor.RegisterWithDeduplicator(b)

	update := data.NewDataSetUpdate()
	update.Active = pointer.FromBool(b.dataSet.Active)
	update.Deduplicator = deduplicatorDescriptor
	dataSet, err := b.dataSession.UpdateDataSet(ctx, *b.dataSet.UploadID, update)
	if err != nil {
		return errors.Wrapf(err, "unable to update data set with id %q", b.dataSet.UploadID)
	}
	b.dataSet = dataSet

	return nil
}

func (b *BaseDeduplicator) AddDataSetData(ctx context.Context, dataSetData []data.Datum) error {
	b.logger.WithField("dataSetDataLength", len(dataSetData)).Debug("AddDataSetData")

	if len(dataSetData) == 0 {
		return nil
	}

	if err := b.dataSession.CreateDataSetData(ctx, b.dataSet, dataSetData); err != nil {
		return errors.Wrapf(err, "unable to create data set data with id %q", *b.dataSet.UploadID)
	}

	return nil
}

func (b *BaseDeduplicator) DeduplicateDataSet(ctx context.Context) error {
	b.logger.Debug("DeduplicateDataSet")

	if err := b.dataSession.ActivateDataSetData(ctx, b.dataSet); err != nil {
		return errors.Wrapf(err, "unable to activate data set data with id %q", *b.dataSet.UploadID)
	}

	return nil
}

func (b *BaseDeduplicator) DeleteDataSet(ctx context.Context) error {
	b.logger.Debug("DeleteDataSet")

	if err := b.dataSession.DeleteDataSet(ctx, b.dataSet); err != nil {
		return errors.Wrapf(err, "unable to delete data set with id %q", *b.dataSet.UploadID)
	}

	return nil
}
