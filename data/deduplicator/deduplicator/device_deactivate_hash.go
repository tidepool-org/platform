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

var DeviceDeactivateHashDeviceManufacturerDeviceModels = map[string][]string{
	"Abbott":          {"FreeStyle Libre"},
	"LifeScan":        {"OneTouch Ultra 2", "OneTouch UltraMini", "Verio", "Verio Flex"},
	"Medtronic":       {"523", "523K", "551", "554", "723", "723K", "751", "754", "1510", "1510K", "1511", "1512", "1580", "1581", "1582", "1710", "1710K", "1711", "1712", "1714", "1714K", "1715", "1780", "1781", "1782"},
	"Trividia Health": {"TRUE METRIX", "TRUE METRIX AIR", "TRUE METRIX GO"},
}

var DeviceDeactivateLegacyHashManufacturerDeviceModels = map[string][]string{
	"Arkray":    {"GlucocardExpression"},
	"Bayer":     {"Contour Next Link", "Contour Next Link 2.4", "Contour Next", "Contour USB", "Contour Next USB", "Contour Next One", "Contour", "Contour Next EZ", "Contour Plus", "Contour Plus Blue"},
	"Dexcom":    {"G5 touchscreen receiver", "G6 touchscreen receiver"},
	"GlucoRx":   {"Nexus", "HCT", "Nexus Mini Ultra", "Go"},
	"Insulet":   {"Dash", "Eros"},
	"i-SENS":    {"CareSens"},
	"MicroTech": {"Equil"},
	"Roche":     {"Aviva Connect", "Performa Connect", "Guide", "Instant (single-button)", "Guide Me", "Instant (two-button)", "Instant S (single-button)", "ReliOn Platinum"},
	"Tandem":    {"1002717"},
}

type DeviceDeactivateHash struct {
	*Base
}

func NewDeviceDeactivateLegacyHash() (*DeviceDeactivateHash, error) {
	base, err := NewBase(DeviceDeactivateHashName, "0.0.0")
	if err != nil {
		return nil, err
	}

	return &DeviceDeactivateHash{
		Base: base,
	}, nil
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

func isManufacturerDeviceModelMatch(dataSet *dataTypesUpload.Upload) (bool, HashType) {
	if dataSet.DeviceManufacturers != nil && dataSet.DeviceModel != nil {
		// legacy match first
		for _, deviceManufacturer := range *dataSet.DeviceManufacturers {
			if allowedDeviceModels, found := DeviceDeactivateLegacyHashManufacturerDeviceModels[deviceManufacturer]; found {
				for _, allowedDeviceModel := range allowedDeviceModels {
					if allowedDeviceModel == *dataSet.DeviceModel {
						return true, LegacyHash
					}
				}
			}
		}
		// fall back to current
		for _, deviceManufacturer := range *dataSet.DeviceManufacturers {
			if allowedDeviceModels, found := DeviceDeactivateHashDeviceManufacturerDeviceModels[deviceManufacturer]; found {
				for _, allowedDeviceModel := range allowedDeviceModels {
					if allowedDeviceModel == *dataSet.DeviceModel {
						return true, PlatformHash
					}
				}
			}
		}
	}
	return false, 0
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
	deviceMatch, _ := isManufacturerDeviceModelMatch(dataSet)
	return deviceMatch, nil
}

func (d *DeviceDeactivateHash) Get(dataSet *dataTypesUpload.Upload) (bool, error) {
	// NOTE: check legacy first then fallback to other matches
	if dataSet == nil {
		return false, errors.New("data set is missing")
	}

	if dataSet.HasDeduplicatorNameDeviceMatch(DeviceDeactivateHashName, DeviceDeactivateLegacyHashManufacturerDeviceModels) {
		return true, nil
	}

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

	if match, hasher := isManufacturerDeviceModelMatch(dataSet); !match {
		return errors.New("data set doesn't match")
	} else {
		if err := AssignDataSetDataIdentityHashes(dataSetData, hasher); err != nil {
			return err
		}
		return d.Base.AddData(ctx, repository, dataSet, dataSetData)
	}
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
