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

func (t *truncateFactory) CanDeduplicateDataset(dataset *upload.Upload) (bool, error) {
	if can, err := t.BaseFactory.CanDeduplicateDataset(dataset); err != nil || !can {
		return can, err
	}

	if dataset.DeviceID == nil {
		return false, nil
	}
	if *dataset.DeviceID == "" {
		return false, nil
	}
	if !dataset.HasDeviceManufacturerOneOf(_TruncateExpectedDeviceManufacturers) {
		return false, nil
	}

	return true, nil
}

func (t *truncateFactory) NewDeduplicatorForDataset(logger log.Logger, dataSession storeDEPRECATED.DataSession, dataset *upload.Upload) (data.Deduplicator, error) {
	baseDeduplicator, err := NewBaseDeduplicator(t.name, t.version, logger, dataSession, dataset)
	if err != nil {
		return nil, err
	}

	if dataset.DeviceID == nil {
		return nil, errors.New("dataset device id is missing")
	}
	if *dataset.DeviceID == "" {
		return nil, errors.New("dataset device id is empty")
	}
	if !dataset.HasDeviceManufacturerOneOf(_TruncateExpectedDeviceManufacturers) {
		return nil, errors.New("dataset device manufacturers does not contain expected device manufacturers")
	}

	return &truncateDeduplicator{
		BaseDeduplicator: baseDeduplicator,
	}, nil
}

func (t *truncateDeduplicator) DeduplicateDataset(ctx context.Context) error {
	// TODO: Technically, ActivateDatasetData could succeed, but DeleteOtherDatasetData fail. This would
	// result in duplicate (and possible incorrect) data. Is there a way to resolve this? Would be nice to have transactions.

	if err := t.BaseDeduplicator.DeduplicateDataset(ctx); err != nil {
		return err
	}

	if err := t.dataSession.DeleteOtherDatasetData(ctx, t.dataset); err != nil {
		return errors.Wrapf(err, "unable to remove all other data except dataset with id %q", *t.dataset.UploadID)
	}

	return nil
}
