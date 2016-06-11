package deduplicator

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
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
	NewDeduplicator(logger log.Logger, storeSession store.Session, dataset *upload.Upload) (Deduplicator, error)
}
