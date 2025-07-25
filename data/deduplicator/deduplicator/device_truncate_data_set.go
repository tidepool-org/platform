package deduplicator

import (
	"context"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/errors"
)

const DeviceTruncateDataSetName = "org.tidepool.deduplicator.device.truncate.dataset"

var DeviceTruncateDataSetDeviceManufacturers = []string{"Animas"}

type DeviceTruncateDataSet struct {
	*Base
}

func NewDeviceTruncateDataSet(dependencies Dependencies) (*DeviceTruncateDataSet, error) {
	base, err := NewBase(dependencies, DeviceTruncateDataSetName, "1.1.0")
	if err != nil {
		return nil, err
	}

	return &DeviceTruncateDataSet{
		Base: base,
	}, nil
}

func (d *DeviceTruncateDataSet) New(ctx context.Context, dataSet *data.DataSet) (bool, error) {
	if dataSet == nil {
		return false, errors.New("data set is missing")
	}

	if !dataSet.HasDataSetTypeNormal() {
		return false, nil
	}
	if dataSet.DeviceID == nil {
		return false, nil
	}

	if dataSet.HasDeduplicatorName() {
		return d.Get(ctx, dataSet)
	}

	if dataSet.DeviceManufacturers == nil {
		return false, nil
	}

	for _, deviceManufacturer := range *dataSet.DeviceManufacturers {
		for _, allowedDeviceManufacturer := range DeviceTruncateDataSetDeviceManufacturers {
			if allowedDeviceManufacturer == deviceManufacturer {
				return true, nil
			}
		}
	}

	return false, nil
}

func (d *DeviceTruncateDataSet) Get(ctx context.Context, dataSet *data.DataSet) (bool, error) {
	if found, err := d.Base.Get(ctx, dataSet); err != nil || found {
		return found, err
	}

	return dataSet.HasDeduplicatorNameMatch("org.tidepool.truncate"), nil // TODO: DEPRECATED
}

func (d *DeviceTruncateDataSet) Close(ctx context.Context, dataSet *data.DataSet) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if dataSet == nil {
		return errors.New("data set is missing")
	}

	// TODO: Technically, DeleteOtherDataSetData could succeed, but Close fail. This would
	// temporarily result in missing data, which is better than the opposite (duplicate data).
	// If this fails, a subsequent successful upload will resolve.
	if err := d.DataStore.DeleteOtherDataSetData(ctx, dataSet); err != nil {
		return err
	}

	return d.Base.Close(ctx, dataSet)
}
