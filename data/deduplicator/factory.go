package deduplicator

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/storeDEPRECATED"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/log"
)

type Factory interface {
	CanDeduplicateDataset(dataset *upload.Upload) (bool, error)
	NewDeduplicatorForDataset(logger log.Logger, dataSession storeDEPRECATED.DataSession, dataset *upload.Upload) (data.Deduplicator, error)

	IsRegisteredWithDataset(dataset *upload.Upload) (bool, error)
	NewRegisteredDeduplicatorForDataset(logger log.Logger, dataSession storeDEPRECATED.DataSession, dataset *upload.Upload) (data.Deduplicator, error)
}
