package deduplicator

import (
	"strconv"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/store"
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

var _HashDeactivateOldExpectedDeviceManufacturers = []string{"Medtronic"}

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

func (h *hashDeactivateOldFactory) CanDeduplicateDataset(dataset *upload.Upload) (bool, error) {
	if can, err := h.BaseFactory.CanDeduplicateDataset(dataset); err != nil || !can {
		return can, err
	}

	if dataset.DeviceID == nil {
		return false, nil
	}
	if *dataset.DeviceID == "" {
		return false, nil
	}
	if !dataset.HasDeviceManufacturerOneOf(_HashDeactivateOldExpectedDeviceManufacturers) {
		return false, nil
	}

	return true, nil
}

func (h *hashDeactivateOldFactory) NewDeduplicatorForDataset(logger log.Logger, dataSession store.DataSession, dataset *upload.Upload) (data.Deduplicator, error) {
	baseDeduplicator, err := NewBaseDeduplicator(h.name, h.version, logger, dataSession, dataset)
	if err != nil {
		return nil, err
	}

	if dataset.DeviceID == nil {
		return nil, errors.New("deduplicator", "dataset device id is missing")
	}
	if *dataset.DeviceID == "" {
		return nil, errors.New("deduplicator", "dataset device id is empty")
	}
	if !dataset.HasDeviceManufacturerOneOf(_HashDeactivateOldExpectedDeviceManufacturers) {
		return nil, errors.New("deduplicator", "dataset device manufacturers does not contain expected device manufacturers")
	}

	return &hashDeactivateOldDeduplicator{
		BaseDeduplicator: baseDeduplicator,
	}, nil
}

func (h *hashDeactivateOldDeduplicator) AddDatasetData(datasetData []data.Datum) error {
	hashes, err := AssignDatasetDataIdentityHashes(datasetData)
	if err != nil {
		return err
	} else if len(hashes) == 0 {
		return nil
	}

	return h.BaseDeduplicator.AddDatasetData(datasetData)
}

func (h *hashDeactivateOldDeduplicator) DeduplicateDataset() error {
	if err := h.dataSession.ArchiveDeviceDataUsingHashesFromDataset(h.dataset); err != nil {
		return errors.Wrapf(err, "deduplicator", "unable to archive device data using hashes from dataset with id %s", strconv.Quote(h.dataset.UploadID))
	}

	return h.BaseDeduplicator.DeduplicateDataset()
}

func (h *hashDeactivateOldDeduplicator) DeleteDataset() error {
	if err := h.dataSession.UnarchiveDeviceDataUsingHashesFromDataset(h.dataset); err != nil {
		return errors.Wrapf(err, "deduplicator", "unable to unarchive device data using hashes from dataset with id %s", strconv.Quote(h.dataset.UploadID))
	}

	return h.BaseDeduplicator.DeleteDataset()
}
