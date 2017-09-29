package deduplicator

import (
	"context"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/storeDEPRECATED"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
)

type continuousFactory struct {
	*BaseFactory
}

type continuousDeduplicator struct {
	*BaseDeduplicator
}

const _ContinuousDeduplicatorName = "org.tidepool.continuous"
const _ContinuousDeduplicatorVersion = "1.0.0"

func NewContinuousFactory() (Factory, error) {
	baseFactory, err := NewBaseFactory(_ContinuousDeduplicatorName, _ContinuousDeduplicatorVersion)
	if err != nil {
		return nil, err
	}

	factory := &continuousFactory{
		BaseFactory: baseFactory,
	}
	factory.Factory = factory

	return factory, nil
}

func (c *continuousFactory) CanDeduplicateDataset(dataset *upload.Upload) (bool, error) {
	if can, err := c.BaseFactory.CanDeduplicateDataset(dataset); err != nil || !can {
		return can, err
	}

	if dataset.DataSetType == nil {
		return false, nil
	}
	if *dataset.DataSetType != upload.DataSetTypeContinuous {
		return false, nil
	}

	return true, nil
}

func (c *continuousFactory) NewDeduplicatorForDataset(logger log.Logger, dataSession storeDEPRECATED.DataSession, dataset *upload.Upload) (data.Deduplicator, error) {
	baseDeduplicator, err := NewBaseDeduplicator(c.name, c.version, logger, dataSession, dataset)
	if err != nil {
		return nil, err
	}

	if dataset.DataSetType == nil {
		return nil, errors.New("dataset type is missing")
	}
	if *dataset.DataSetType != upload.DataSetTypeContinuous {
		return nil, errors.New("dataset type is not continuous")
	}

	return &continuousDeduplicator{
		BaseDeduplicator: baseDeduplicator,
	}, nil
}

func (c *continuousDeduplicator) RegisterDataset(ctx context.Context) error {
	c.dataset.SetActive(true)
	return c.BaseDeduplicator.RegisterDataset(ctx)
}

func (c *continuousDeduplicator) AddDatasetData(ctx context.Context, datasetData []data.Datum) error {
	c.logger.WithField("datasetDataLength", len(datasetData)).Debug("AddDatasetData")

	if len(datasetData) == 0 {
		return nil
	}

	for _, datasetDatum := range datasetData {
		datasetDatum.SetActive(true)
	}

	if err := c.dataSession.CreateDatasetData(ctx, c.dataset, datasetData); err != nil {
		return errors.Wrapf(err, "unable to create dataset data with id %q", c.dataset.UploadID)
	}

	return nil
}

func (c *continuousDeduplicator) DeduplicateDataset(ctx context.Context) error {
	return nil
}

func (c *continuousDeduplicator) DeleteDataset(ctx context.Context) error {
	return errors.Newf("unable to delete dataset with id %q", c.dataset.UploadID)
}
