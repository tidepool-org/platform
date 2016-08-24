package store

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
	"github.com/tidepool-org/platform/data/types/base/upload"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/store"
)

type Store interface {
	store.Store

	NewSession(logger log.Logger) (Session, error)
}

type Session interface {
	store.Session

	GetDatasetsForUser(userID string) ([]*upload.Upload, error)
	GetDataset(datasetID string) (*upload.Upload, error)
	CreateDataset(dataset *upload.Upload) error
	UpdateDataset(dataset *upload.Upload) error
	DeleteDataset(datasetID string) error
	CreateDatasetData(dataset *upload.Upload, datasetData []data.Datum) error
	ActivateAllDatasetData(dataset *upload.Upload) error
	DeleteAllOtherDatasetData(dataset *upload.Upload) error
}
