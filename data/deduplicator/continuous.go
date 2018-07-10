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

func (c *continuousFactory) CanDeduplicateDataSet(dataSet *upload.Upload) (bool, error) {
	if can, err := c.BaseFactory.CanDeduplicateDataSet(dataSet); err != nil || !can {
		return can, err
	}

	if dataSet.DataSetType == nil {
		return false, nil
	}
	if *dataSet.DataSetType != upload.DataSetTypeContinuous {
		return false, nil
	}

	return true, nil
}

func (c *continuousFactory) NewDeduplicatorForDataSet(logger log.Logger, dataSession storeDEPRECATED.DataSession, dataSet *upload.Upload) (data.Deduplicator, error) {
	baseDeduplicator, err := NewBaseDeduplicator(c.name, c.version, logger, dataSession, dataSet)
	if err != nil {
		return nil, err
	}

	if dataSet.DataSetType == nil {
		return nil, errors.New("data set type is missing")
	}
	if *dataSet.DataSetType != upload.DataSetTypeContinuous {
		return nil, errors.New("data set type is not continuous")
	}

	return &continuousDeduplicator{
		BaseDeduplicator: baseDeduplicator,
	}, nil
}

func (c *continuousDeduplicator) RegisterDataSet(ctx context.Context) error {
	c.dataSet.SetActive(true)
	return c.BaseDeduplicator.RegisterDataSet(ctx)
}

func (c *continuousDeduplicator) AddDataSetData(ctx context.Context, dataSetData []data.Datum) error {
	c.logger.WithField("dataSetDataLength", len(dataSetData)).Debug("AddDataSetData")

	if len(dataSetData) == 0 {
		return nil
	}

	for _, dataSetDatum := range dataSetData {
		dataSetDatum.SetActive(true)
	}

	if err := c.dataSession.CreateDataSetData(ctx, c.dataSet, dataSetData); err != nil {
		return errors.Wrapf(err, "unable to create data set data with id %q", c.dataSet.UploadID)
	}

	return nil
}

func (c *continuousDeduplicator) DeduplicateDataSet(ctx context.Context) error {
	return nil
}

func (c *continuousDeduplicator) DeleteDataSet(ctx context.Context) error {
	return errors.Newf("unable to delete data set with id %q", c.dataSet.UploadID)
}
