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
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/store"
)

type Deduplicator interface {
	InitializeDataset() error
	AddDataToDataset(datumArray data.BuiltDatumArray) error
	FinalizeDataset() error
}

type Factory interface {
	CanDeduplicateDataset(datasetUpload *upload.Upload) (bool, error)
	NewDeduplicator(datasetUpload *upload.Upload, storeSession store.Session, logger log.Logger) (Deduplicator, error)
}
