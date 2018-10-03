package deduplicator

import (
	"context"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/storeDEPRECATED"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
)

type hashDeactivateOldFactory struct {
	*BaseFactory
}

type hashDeactivateOldDeduplicator struct {
	*BaseDeduplicator
}

const _HashDeactivateOldDeduplicatorName = "org.tidepool.hash-deactivate-old"
const _HashDeactivateOldDeduplicatorVersion = "1.1.0"

var _HashDeactivateOldExpectedDeviceManufacturerModels = map[string][]string{
	"Abbott":          {"FreeStyle Libre"},
	"LifeScan":        {"OneTouch Ultra 2", "OneTouch UltraMini", "Verio", "Verio Flex"},
	"Medtronic":       {"523", "523K", "551", "554", "723", "723K", "751", "754", "1510", "1510K", "1511", "1512", "1580", "1581", "1582", "1710", "1710K", "1711", "1712", "1714", "1714K", "1715", "1780", "1781", "1782"},
	"Trividia Health": {"TRUE METRIX", "TRUE METRIX AIR", "TRUE METRIX GO"},
}

func NewHashDeactivateOldFactory() (Factory, error) {
	baseFactory, err := NewBaseFactory(_HashDeactivateOldDeduplicatorName, _HashDeactivateOldDeduplicatorVersion)
	if err != nil {
		return nil, err
	}

	factory := &hashDeactivateOldFactory{
		BaseFactory: baseFactory,
	}
	factory.Factory = factory

	return factory, nil
}

func (h *hashDeactivateOldFactory) CanDeduplicateDataSet(dataSet *upload.Upload) (bool, error) {
	if can, err := h.BaseFactory.CanDeduplicateDataSet(dataSet); err != nil || !can {
		return can, err
	}

	if dataSet.DeviceID == nil {
		return false, nil
	}
	if *dataSet.DeviceID == "" {
		return false, nil
	}
	if dataSet.DeviceManufacturers == nil {
		return false, nil
	}
	if dataSet.DeviceModel == nil {
		return false, nil
	}

	return allowDeviceManufacturerModel(_HashDeactivateOldExpectedDeviceManufacturerModels, *dataSet.DeviceManufacturers, *dataSet.DeviceModel), nil
}

func (h *hashDeactivateOldFactory) NewDeduplicatorForDataSet(logger log.Logger, dataSession storeDEPRECATED.DataSession, dataSet *upload.Upload) (data.Deduplicator, error) {
	baseDeduplicator, err := NewBaseDeduplicator(h.name, h.version, logger, dataSession, dataSet)
	if err != nil {
		return nil, err
	}

	if dataSet.DeviceID == nil {
		return nil, errors.New("data set device id is missing")
	}
	if *dataSet.DeviceID == "" {
		return nil, errors.New("data set device id is empty")
	}
	if dataSet.DeviceManufacturers == nil {
		return nil, errors.New("data set device manufacturers is missing")
	}
	if dataSet.DeviceModel == nil {
		return nil, errors.New("data set device model is missing")
	}

	if !allowDeviceManufacturerModel(_HashDeactivateOldExpectedDeviceManufacturerModels, *dataSet.DeviceManufacturers, *dataSet.DeviceModel) {
		return nil, errors.New("data set device manufacturer and model does not contain expected device manufacturers and models")
	}

	return &hashDeactivateOldDeduplicator{
		BaseDeduplicator: baseDeduplicator,
	}, nil
}

func (h *hashDeactivateOldDeduplicator) AddDataSetData(ctx context.Context, dataSetData []data.Datum) error {
	hashes, err := AssignDataSetDataIdentityHashes(dataSetData)
	if err != nil {
		return err
	} else if len(hashes) == 0 {
		return nil
	}

	return h.BaseDeduplicator.AddDataSetData(ctx, dataSetData)
}

func (h *hashDeactivateOldDeduplicator) DeduplicateDataSet(ctx context.Context) error {
	if err := h.dataSession.ArchiveDeviceDataUsingHashesFromDataSet(ctx, h.dataSet); err != nil {
		return errors.Wrapf(err, "unable to archive device data using hashes from data set with id %q", *h.dataSet.UploadID)
	}

	return h.BaseDeduplicator.DeduplicateDataSet(ctx)
}

func (h *hashDeactivateOldDeduplicator) DeleteDataSet(ctx context.Context) error {
	if err := h.dataSession.UnarchiveDeviceDataUsingHashesFromDataSet(ctx, h.dataSet); err != nil {
		return errors.Wrapf(err, "unable to unarchive device data using hashes from data set with id %q", *h.dataSet.UploadID)
	}

	return h.BaseDeduplicator.DeleteDataSet(ctx)
}

func allowDeviceManufacturerModel(allowedDeviceManufacturerModels map[string][]string, deviceManufacturers []string, deviceModel string) bool {
	for _, deviceManufacturer := range deviceManufacturers {
		if allowedDeviceModels, found := allowedDeviceManufacturerModels[deviceManufacturer]; found {
			for _, allowedDeviceModel := range allowedDeviceModels {
				if deviceModel == allowedDeviceModel {
					return true
				}
			}
		}
	}

	return false
}
