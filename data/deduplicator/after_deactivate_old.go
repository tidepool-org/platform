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
	"strconv"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/log"
)

type afterDeactivateOldFactory struct {
	*BaseFactory
}

type afterDeactivateOldDeduplicator struct {
	*BaseDeduplicator
}

const _AfterDeactivateOldDeduplicatorName = "after-deactivate-old"

var _AfterDeactivateOldExpectedDeviceManufacturers = []string{"UNUSED"}

func NewAfterDeactivateOldFactory() (Factory, error) {
	baseFactory, err := NewBaseFactory(_AfterDeactivateOldDeduplicatorName)
	if err != nil {
		return nil, err
	}

	factory := &afterDeactivateOldFactory{
		BaseFactory: baseFactory,
	}
	factory.Factory = factory

	return factory, nil
}

func (a *afterDeactivateOldFactory) CanDeduplicateDataset(dataset *upload.Upload) (bool, error) {
	if can, err := a.BaseFactory.CanDeduplicateDataset(dataset); err != nil || !can {
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
	if !app.StringsContainsAnyStrings(*dataset.DeviceManufacturers, _AfterDeactivateOldExpectedDeviceManufacturers) {
		return false, nil
	}

	return true, nil
}

func (a *afterDeactivateOldFactory) NewDeduplicatorForDataset(logger log.Logger, dataStoreSession store.Session, dataset *upload.Upload) (data.Deduplicator, error) {
	baseDeduplicator, err := NewBaseDeduplicator(a.name, logger, dataStoreSession, dataset)
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
	if !app.StringsContainsAnyStrings(*dataset.DeviceManufacturers, _AfterDeactivateOldExpectedDeviceManufacturers) {
		return nil, app.Error("deduplicator", "dataset device manufacturers does not contain expected device manufacturers")
	}

	return &afterDeactivateOldDeduplicator{
		BaseDeduplicator: baseDeduplicator,
	}, nil
}

func (a *afterDeactivateOldDeduplicator) DeduplicateDataset() error {
	afterTime, err := a.dataStoreSession.FindEarliestDatasetDataTime(a.dataset)
	if err != nil {
		return app.ExtErrorf(err, "deduplicator", "unable to get earliest data in dataset with id %s", strconv.Quote(a.dataset.UploadID))
	}

	// TODO: Technically, ActivateDatasetData could succeed, but DeactivateOtherDatasetDataAfterTime fail. This would
	// result in duplicate (and possible incorrect) data. Is there a way to resolve this? Would be nice to have transactions.

	if err = a.BaseDeduplicator.DeduplicateDataset(); err != nil {
		return err
	}

	if afterTime != "" {
		if err = a.dataStoreSession.DeactivateOtherDatasetDataAfterTime(a.dataset, afterTime); err != nil {
			return app.ExtErrorf(err, "deduplicator", "unable to remove all other data except dataset with id %s", strconv.Quote(a.dataset.UploadID))
		}
	}

	return nil
}
