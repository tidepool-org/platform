package deduplicator

import (
	"context"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/storeDEPRECATED"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
)

type truncateFactory struct {
	*BaseFactory
}

type truncateDeduplicator struct {
	*BaseDeduplicator
}

const _TruncateDeduplicatorName = "org.tidepool.truncate"
const _TruncateDeduplicatorVersion = "1.0.0"

var _TruncateExpectedDeviceManufacturers = []string{"Animas"}

func NewTruncateFactory() (Factory, error) {
	baseFactory, err := NewBaseFactory(_TruncateDeduplicatorName, _TruncateDeduplicatorVersion)
	if err != nil {
		return nil, err
	}

	factory := &truncateFactory{
		BaseFactory: baseFactory,
	}
	factory.Factory = factory

	return factory, nil
}

func (t *truncateFactory) CanDeduplicateDataSet(dataSet *upload.Upload) (bool, error) {
	if can, err := t.BaseFactory.CanDeduplicateDataSet(dataSet); err != nil || !can {
		return can, err
	}

	if dataSet.DeviceID == nil {
		return false, nil
	}
	if *dataSet.DeviceID == "" {
		return false, nil
	}
	if !dataSet.HasDeviceManufacturerOneOf(_TruncateExpectedDeviceManufacturers) {
		return false, nil
	}

	return true, nil
}

func (t *truncateFactory) NewDeduplicatorForDataSet(logger log.Logger, dataSession storeDEPRECATED.DataSession, dataSet *upload.Upload) (data.Deduplicator, error) {
	baseDeduplicator, err := NewBaseDeduplicator(t.name, t.version, logger, dataSession, dataSet)
	if err != nil {
		return nil, err
	}

	if dataSet.DeviceID == nil {
		return nil, errors.New("data set device id is missing")
	}
	if *dataSet.DeviceID == "" {
		return nil, errors.New("data set device id is empty")
	}
	if !dataSet.HasDeviceManufacturerOneOf(_TruncateExpectedDeviceManufacturers) {
		return nil, errors.New("data set device manufacturers does not contain expected device manufacturers")
	}

	return &truncateDeduplicator{
		BaseDeduplicator: baseDeduplicator,
	}, nil
}

func (t *truncateDeduplicator) DeduplicateDataSet(ctx context.Context) error {
	// TODO: Technically, ActivateDataSetData could succeed, but DeleteOtherDataSetData fail. This would
	// result in duplicate (and possible incorrect) data. Is there a way to resolve this? Would be nice to have transactions.

	if err := t.BaseDeduplicator.DeduplicateDataSet(ctx); err != nil {
		return err
	}

	if err := t.dataSession.DeleteOtherDataSetData(ctx, t.dataSet); err != nil {
		return errors.Wrapf(err, "unable to remove all other data except data set with id %q", *t.dataSet.UploadID)
	}

	return nil
}
