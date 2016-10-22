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
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/types/base/upload"
	"github.com/tidepool-org/platform/log"
)

type Deduplicator interface {
	InitializeDataset() error
	AddDataToDataset(datasetData []data.Datum) error
	FinalizeDataset() error
}

type Factory interface {
	CanDeduplicateDataset(dataset *upload.Upload) (bool, error)
	NewDeduplicator(logger log.Logger, dataStoreSession store.Session, dataset *upload.Upload) (Deduplicator, error)
}
