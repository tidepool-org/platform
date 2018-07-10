package deduplicator

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/storeDEPRECATED"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/log"
)

type Factory interface {
	CanDeduplicateDataSet(dataSet *upload.Upload) (bool, error)
	NewDeduplicatorForDataSet(logger log.Logger, dataSession storeDEPRECATED.DataSession, dataSet *upload.Upload) (data.Deduplicator, error)

	IsRegisteredWithDataSet(dataSet *upload.Upload) (bool, error)
	NewRegisteredDeduplicatorForDataSet(logger log.Logger, dataSession storeDEPRECATED.DataSession, dataSet *upload.Upload) (data.Deduplicator, error)
}
