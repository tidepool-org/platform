package deduplicator

import (
	"context"

	"github.com/tidepool-org/platform/data"
	dataStore "github.com/tidepool-org/platform/data/store"
	dataTypesUpload "github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
)

const (
	DeviceDeactivateHashVersionLegacy  = "0.0.0"
	DeviceDeactivateHashVersionCurrent = "1.1.0"
)

const DeviceDeactivateHashName = "org.tidepool.deduplicator.device.deactivate.hash"

var DeviceDeactivateHashDeviceManufacturerDeviceModels = map[string][]string{
	"Abbott":          {"FreeStyle Libre"},
	"LifeScan":        {"OneTouch Ultra 2", "OneTouch UltraMini", "Verio", "Verio Flex"},
	"Medtronic":       {"523", "523K", "551", "554", "723", "723K", "751", "754", "1510", "1510K", "1511", "1512", "1580", "1581", "1582", "1710", "1710K", "1711", "1712", "1714", "1714K", "1715", "1780", "1781", "1782"},
	"Trividia Health": {"TRUE METRIX", "TRUE METRIX AIR", "TRUE METRIX GO"},
}

var DeviceDeactivateLegacyHashDeviceManufacturerDeviceModels = map[string][]string{
	"Arkray":    {"GlucocardExpression"},
	"Bayer":     {"Contour Next Link", "Contour Next Link 2.4", "Contour Next", "Contour USB", "Contour Next USB", "Contour Next One", "Contour", "Contour Next EZ", "Contour Plus", "Contour Plus Blue"},
	"Dexcom":    {"G5 touchscreen receiver", "G6 touchscreen receiver"},
	"GlucoRx":   {"Nexus", "HCT", "Nexus Mini Ultra", "Go"},
	"i-SENS":    {"CareSens"},
	"MicroTech": {"Equil"},
	"Roche":     {"Aviva Connect", "Performa Connect", "Guide", "Instant (single-button)", "Guide Me", "Instant (two-button)", "Instant S (single-button)", "ReliOn Platinum"},

	"Insulet": {"Dash", "Eros", "OmniPod"},
	"Tandem":  {"1002717", "5602", "5448004", "5448003", "5448001", "5448", "4628003", "4628", "10037177", "1001357", "1000354", "1000096"},
}

type DeviceDeactivateHash struct {
	*Base
}

func NewDeviceDeactivateHash() (*DeviceDeactivateHash, error) {
	base, err := NewBase(DeviceDeactivateHashName, DeviceDeactivateHashVersionCurrent)
	if err != nil {
		return nil, err
	}

	return &DeviceDeactivateHash{
		Base: base,
	}, nil
}

func getDeduplicatorVersion(dataSet *dataTypesUpload.Upload) (string, bool) {
	if dataSet.DeviceManufacturers == nil || dataSet.DeviceModel == nil {
		return "", false
	}

	for _, deviceManufacturer := range *dataSet.DeviceManufacturers {
		if allowedDeviceModels, found := DeviceDeactivateLegacyHashDeviceManufacturerDeviceModels[deviceManufacturer]; found {
			for _, allowedDeviceModel := range allowedDeviceModels {
				if allowedDeviceModel == *dataSet.DeviceModel {
					return DeviceDeactivateHashVersionLegacy, true
				}
			}
		}

		if allowedDeviceModels, found := DeviceDeactivateHashDeviceManufacturerDeviceModels[deviceManufacturer]; found {
			for _, allowedDeviceModel := range allowedDeviceModels {
				if allowedDeviceModel == *dataSet.DeviceModel {
					return DeviceDeactivateHashVersionCurrent, true
				}
			}
		}
	}

	return "", false
}

func (d *DeviceDeactivateHash) New(ctx context.Context, dataSet *dataTypesUpload.Upload) (bool, error) {
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

	_, found := getDeduplicatorVersion(dataSet)
	return found, nil
}

func (d *DeviceDeactivateHash) Get(ctx context.Context, dataSet *dataTypesUpload.Upload) (bool, error) {
	if found, err := d.Base.Get(ctx, dataSet); err != nil || found {
		return found, err
	}

	return dataSet.HasDeduplicatorNameMatch("org.tidepool.hash-deactivate-old"), nil // TODO: DEPRECATED
}

func (d *DeviceDeactivateHash) Open(ctx context.Context, repository dataStore.DataRepository, dataSet *dataTypesUpload.Upload) (*dataTypesUpload.Upload, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if repository == nil {
		return nil, errors.New("repository is missing")
	}
	if dataSet == nil {
		return nil, errors.New("data set is missing")
	}

	version, found := getDeduplicatorVersion(dataSet)
	if !found {
		return nil, errors.New("deduplicator version not found")
	}

	dataSet.Deduplicator = data.NewDeduplicatorDescriptor()
	dataSet.Deduplicator.Name = pointer.FromString(DeviceDeactivateHashName)
	dataSet.Deduplicator.Version = pointer.FromString(version)

	return d.Base.Open(ctx, repository, dataSet)
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

	options := NewDefaultDeviceDeactivateHashOptions()
	if *dataSet.Deduplicator.Version == DeviceDeactivateHashVersionLegacy {
		filter := &data.DataSetFilter{LegacyOnly: pointer.FromBool(true), DeviceID: dataSet.DeviceID}
		pagination := &page.Pagination{Page: 1, Size: 1}

		uploads, err := repository.ListUserDataSets(ctx, *dataSet.UserID, filter, pagination)
		if err != nil {
			return errors.Wrap(err, "error getting datasets for user")
		}
		if len(uploads) != 0 {
			if uploads[0].LegacyGroupID == nil {
				return errors.New("missing required legacy groupId for the device deactivate hash legacy version")
			}
			options = NewLegacyHashOptions(*uploads[0].LegacyGroupID)
		}
	}

	if err := AssignDataSetDataIdentityHashes(dataSetData, options); err != nil {
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
