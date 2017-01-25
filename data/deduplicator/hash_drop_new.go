package deduplicator

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

import (
	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/log"
)

type hashDropNewFactory struct {
	*BaseFactory
}

type hashDropNewDeduplicator struct {
	*BaseDeduplicator
}

const _HashDropNewDeduplicatorName = "hash-drop-new"

var _HashDropNewExpectedDeviceManufacturers = []string{"UNUSED"}

func NewHashDropNewFactory() (Factory, error) {
	baseFactory, err := NewBaseFactory(_HashDropNewDeduplicatorName)
	if err != nil {
		return nil, err
	}

	factory := &hashDropNewFactory{
		BaseFactory: baseFactory,
	}
	factory.Factory = factory

	return factory, nil
}

func (h *hashDropNewFactory) CanDeduplicateDataset(dataset *upload.Upload) (bool, error) {
	if can, err := h.BaseFactory.CanDeduplicateDataset(dataset); err != nil || !can {
		return can, err
	}

	if dataset.DeviceID == nil {
		return false, nil
	}
	if *dataset.DeviceID == "" {
		return false, nil
	}
	if dataset.DeviceManufacturers == nil {
		return false, nil
	}
	if !app.StringsContainsAnyStrings(*dataset.DeviceManufacturers, _HashDropNewExpectedDeviceManufacturers) {
		return false, nil
	}

	return true, nil
}

func (h *hashDropNewFactory) NewDeduplicatorForDataset(logger log.Logger, dataStoreSession store.Session, dataset *upload.Upload) (data.Deduplicator, error) {
	baseDeduplicator, err := NewBaseDeduplicator(h.name, logger, dataStoreSession, dataset)
	if err != nil {
		return nil, err
	}

	if dataset.DeviceID == nil {
		return nil, app.Error("deduplicator", "dataset device id is missing")
	}
	if *dataset.DeviceID == "" {
		return nil, app.Error("deduplicator", "dataset device id is empty")
	}
	if dataset.DeviceManufacturers == nil {
		return nil, app.Error("deduplicator", "dataset device manufacturers is missing")
	}
	if !app.StringsContainsAnyStrings(*dataset.DeviceManufacturers, _HashDropNewExpectedDeviceManufacturers) {
		return nil, app.Error("deduplicator", "dataset device manufacturers does not contain expected device manufacturers")
	}

	return &hashDropNewDeduplicator{
		BaseDeduplicator: baseDeduplicator,
	}, nil
}

func (h *hashDropNewDeduplicator) AddDatasetData(datasetData []data.Datum) error {
	hashes, err := AssignDatasetDataIdentityHashes(datasetData)
	if err != nil {
		return err
	} else if len(hashes) == 0 {
		return nil
	}

	hashes, err = h.dataStoreSession.FindAllDatasetDataDeduplicatorHashesForDevice(h.dataset.UserID, *h.dataset.DeviceID, hashes)
	if err != nil {
		return app.ExtError(err, "deduplicator", "unable to find all dataset data deduplicator hashes for device")
	}

	uniqueDatasetData := []data.Datum{}
	for _, datasetDatum := range datasetData {
		if !app.StringsContainsString(hashes, datasetDatum.DeduplicatorDescriptor().Hash) {
			uniqueDatasetData = append(uniqueDatasetData, datasetDatum)
		}
	}

	return h.BaseDeduplicator.AddDatasetData(uniqueDatasetData)
}
