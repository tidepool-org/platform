package deduplicator

import (
	"context"

	"github.com/tidepool-org/platform/data"
	dataStoreDEPRECATED "github.com/tidepool-org/platform/data/storeDEPRECATED"
	dataTypes "github.com/tidepool-org/platform/data/types"
	dataTypesUpload "github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
)

type continousOriginFactory struct {
	*BaseFactory
}

type continuousOriginDeduplicator struct {
	*BaseDeduplicator
}

const _ContinuousOriginDeduplicatorName = "org.tidepool.continuous.origin"
const _ContinuousOriginDeduplicatorVersion = "1.0.0"

func NewContinuousOriginFactory() (Factory, error) {
	baseFactory, err := NewBaseFactory(_ContinuousOriginDeduplicatorName, _ContinuousOriginDeduplicatorVersion)
	if err != nil {
		return nil, err
	}

	factory := &continousOriginFactory{
		BaseFactory: baseFactory,
	}
	factory.Factory = factory

	return factory, nil
}

func (c *continousOriginFactory) CanDeduplicateDataSet(dataSet *dataTypesUpload.Upload) (bool, error) {
	if can, err := c.BaseFactory.CanDeduplicateDataSet(dataSet); err != nil || !can {
		return can, err
	}

	if dataSet.Deduplicator == nil {
		return false, nil
	}
	if !dataSet.Deduplicator.IsRegisteredWithNamedDeduplicator(_ContinuousOriginDeduplicatorName) {
		return false, nil
	}
	if dataSet.DataSetType == nil {
		return false, nil
	}
	if *dataSet.DataSetType != dataTypesUpload.DataSetTypeContinuous {
		return false, nil
	}

	return true, nil
}

func (c *continousOriginFactory) NewDeduplicatorForDataSet(logger log.Logger, dataSession dataStoreDEPRECATED.DataSession, dataSet *dataTypesUpload.Upload) (data.Deduplicator, error) {
	baseDeduplicator, err := NewBaseDeduplicator(c.name, c.version, logger, dataSession, dataSet)
	if err != nil {
		return nil, err
	}

	if dataSet.Deduplicator == nil {
		return nil, errors.New("data set deduplicator is missing")
	}
	if !dataSet.Deduplicator.IsRegisteredWithNamedDeduplicator(_ContinuousOriginDeduplicatorName) {
		return nil, errors.New("data set is not registered with deduplicator")
	}
	if dataSet.DataSetType == nil {
		return nil, errors.New("data set type is missing")
	}
	if *dataSet.DataSetType != dataTypesUpload.DataSetTypeContinuous {
		return nil, errors.New("data set type is not continuous")
	}

	return &continuousOriginDeduplicator{
		BaseDeduplicator: baseDeduplicator,
	}, nil
}

func (c *continuousOriginDeduplicator) AddDataSetData(ctx context.Context, dataSetData []data.Datum) error {
	if len(dataSetData) == 0 {
		return nil
	}

	var originIDs []string
	for _, datum := range dataSetData {
		datum.SetActive(true)
		if base, ok := datum.(*dataTypes.Base); !ok {
			return errors.New("data set data invalid")
		} else if base.Origin != nil && base.Origin.ID != nil {
			originIDs = append(originIDs, *base.Origin.ID)
		}
	}

	if err := c.dataSession.ArchiveDataSetDataUsingOriginIDs(ctx, c.dataSet, originIDs); err != nil {
		return errors.Wrapf(err, "unable to archive device data using origin from data set with id %q", *c.dataSet.UploadID)
	}

	if err := c.BaseDeduplicator.AddDataSetData(ctx, dataSetData); err != nil {
		return err
	}

	if err := c.dataSession.DeleteArchivedDataSetData(ctx, c.dataSet); err != nil {
		return errors.Wrapf(err, "unable to delete archived device data from data set with id %q", *c.dataSet.UploadID)
	}

	return nil
}
