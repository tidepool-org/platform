package deduplicator

import (
	"context"

	dataStore "github.com/tidepool-org/platform/data/store"
	dataTypesUpload "github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/errors"
)

const DeviceTruncateDataSetName = "org.tidepool.deduplicator.device.truncate.dataset"

var DeviceTruncateDataSetDeviceManufacturers = []string{"Animas"}

type DeviceTruncateDataSet struct {
	*Base
}

func NewDeviceTruncateDataSet() (*DeviceTruncateDataSet, error) {
	base, err := NewBase(DeviceTruncateDataSetName, "1.1.0")
	if err != nil {
		return nil, err
	}

	return &DeviceTruncateDataSet{
		Base: base,
	}, nil
}

func (t *DeviceTruncateDataSet) New(dataSet *dataTypesUpload.Upload) (bool, error) {
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
		return t.Get(dataSet)
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

func (t *DeviceTruncateDataSet) Get(dataSet *dataTypesUpload.Upload) (bool, error) {
	if found, err := t.Base.Get(dataSet); err != nil || found {
		return found, err
	}

	return dataSet.HasDeduplicatorNameMatch("org.tidepool.truncate"), nil // TODO: DEPRECATED
}

func (t *DeviceTruncateDataSet) Close(ctx context.Context, repository dataStore.DataRepository, dataSet *dataTypesUpload.Upload) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if repository == nil {
		return errors.New("repository is missing")
	}
	if dataSet == nil {
		return errors.New("data set is missing")
	}

	// TODO: Technically, DeleteOtherDataSetData could succeed, but Close fail. This would
	// temporarily result in missing data, which is better than the opposite (duplicate data).
	// If this fails, a subsequent successful upload will resolve.
	if err := repository.DeleteOtherDataSetData(ctx, dataSet); err != nil {
		return err
	}

	return t.Base.Close(ctx, repository, dataSet)
}
