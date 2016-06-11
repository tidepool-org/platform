package store

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
	"github.com/tidepool-org/platform/data/types/base/upload"
	"github.com/tidepool-org/platform/log"
)

type Store interface {
	IsClosed() bool
	Close()
	GetStatus() interface{}
	NewSession(logger log.Logger) (Session, error)
}

type Session interface {
	IsClosed() bool
	Close()
	GetDataset(datasetID string) (*upload.Upload, error)
	CreateDataset(dataset *upload.Upload) error
	UpdateDataset(dataset *upload.Upload) error
	CreateDatasetData(dataset *upload.Upload, datasetData []data.Datum) error
	ActivateAllDatasetData(dataset *upload.Upload) error
	RemoveAllOtherDatasetData(dataset *upload.Upload) error
}
