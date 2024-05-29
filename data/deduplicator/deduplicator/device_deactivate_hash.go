package deduplicator

import (
	"context"

	"github.com/tidepool-org/platform/data"
	dataStore "github.com/tidepool-org/platform/data/store"
	dataTypesUpload "github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/errors"
)

type HashType int

const (
	_ HashType = iota
	PlatformHash
	LegacyHash
)

const DeviceDeactivateHashName = "org.tidepool.deduplicator.device.deactivate.hash"
const DeviceDeactivateLegacyHashName = "org.tidepool.deduplicator.device.deactivate.legacy.hash"

var DeviceDeactivateHashDeviceManufacturerDeviceModels = map[string][]string{
	"Abbott":          {"FreeStyle Libre"},
	"LifeScan":        {"OneTouch Ultra 2", "OneTouch UltraMini", "Verio", "Verio Flex"},
	"Medtronic":       {"523", "523K", "551", "554", "723", "723K", "751", "754", "1510", "1510K", "1511", "1512", "1580", "1581", "1582", "1710", "1710K", "1711", "1712", "1714", "1714K", "1715", "1780", "1781", "1782"},
	"Trividia Health": {"TRUE METRIX", "TRUE METRIX AIR", "TRUE METRIX GO"},
}

var DeviceDeactivateLegacyHasheManufacturerDeviceModels = map[string][]string{
	"Tandem":         {"T:Slim"},
	"InsuletOmniPod": {"Dash", "Eros"},
	//TODO: other devices here
}

type DeviceDeactivateHash struct {
	*Base
}

func NewDeviceDeactivateLegacyHash() (*DeviceDeactivateHash, error) {
	base, err := NewBase(DeviceDeactivateLegacyHashName, "0.0.0")
	if err != nil {
		return nil, err
	}

	return &DeviceDeactivateHash{
		Base: base,
	}, nil
}

func (d *DeviceDeactivateHash) getHashType() (HashType, error) {
	if d.name == DeviceDeactivateHashName {
		return PlatformHash, nil
	} else if d.name == DeviceDeactivateLegacyHashName {
		return LegacyHash, nil
	}
	return 0, errors.New("unknown hash type")
}

func NewDeviceDeactivateHash() (*DeviceDeactivateHash, error) {
	base, err := NewBase(DeviceDeactivateHashName, "1.1.0")
	if err != nil {
		return nil, err
	}

	return &DeviceDeactivateHash{
		Base: base,
	}, nil
}

func (d *DeviceDeactivateHash) New(dataSet *dataTypesUpload.Upload) (bool, error) {
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
		return d.Get(dataSet)
	}

	if dataSet.DeviceManufacturers == nil || dataSet.DeviceModel == nil {
		return false, nil
	}

	hasher, err := d.getHashType()

	if err != nil {
		return false, err
	}
	if hasher == PlatformHash {
		for _, deviceManufacturer := range *dataSet.DeviceManufacturers {
			if allowedDeviceModels, found := DeviceDeactivateHashDeviceManufacturerDeviceModels[deviceManufacturer]; found {
				for _, allowedDeviceModel := range allowedDeviceModels {
					if allowedDeviceModel == *dataSet.DeviceModel {
						return true, nil
					}
				}
			}
		}
	} else if hasher == LegacyHash {
		for _, deviceManufacturer := range *dataSet.DeviceManufacturers {
			if allowedDeviceModels, found := DeviceDeactivateLegacyHasheManufacturerDeviceModels[deviceManufacturer]; found {
				for _, allowedDeviceModel := range allowedDeviceModels {
					if allowedDeviceModel == *dataSet.DeviceModel {
						return true, nil
					}
				}
			}
		}
	}

	return false, nil
}

func (d *DeviceDeactivateHash) Get(dataSet *dataTypesUpload.Upload) (bool, error) {
	if found, err := d.Base.Get(dataSet); err != nil || found {
		return found, err
	}

	return dataSet.HasDeduplicatorNameMatch("org.tidepool.hash-deactivate-old"), nil // TODO: DEPRECATED
}

func (d *DeviceDeactivateHash) AddData(ctx context.Context, repository dataStore.DataRepository, dataSet *dataTypesUpload.Upload, dataSetData data.Data) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if repository == nil {
		return errors.New("repository is missing")
	}
	if dataSet == nil {
		return errors.New("data set is missing")
	}
	if dataSetData == nil {
		return errors.New("data set data is missing")
	}

	hasher, err := d.getHashType()

	if err != nil {
		return err
	}

	if err := AssignDataSetDataIdentityHashes(dataSetData, hasher); err != nil {
		return err
	}

	return d.Base.AddData(ctx, repository, dataSet, dataSetData)
}

func (d *DeviceDeactivateHash) Close(ctx context.Context, repository dataStore.DataRepository, dataSet *dataTypesUpload.Upload) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if repository == nil {
		return errors.New("repository is missing")
	}
	if dataSet == nil {
		return errors.New("data set is missing")
	}

	if err := repository.ArchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet); err != nil {
		return err
	}

	return d.Base.Close(ctx, repository, dataSet)
}

func (d *DeviceDeactivateHash) Delete(ctx context.Context, repository dataStore.DataRepository, dataSet *dataTypesUpload.Upload) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if repository == nil {
		return errors.New("repository is missing")
	}
	if dataSet == nil {
		return errors.New("data set is missing")
	}

	if err := repository.UnarchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet); err != nil {
		return err
	}

	return d.Base.Delete(ctx, repository, dataSet)
}
