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
	"github.com/tidepool-org/platform/app"
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

	GetDatasetsForUserByID(userID string, filter *Filter, pagination *Pagination) ([]*upload.Upload, error)
	GetDatasetByID(datasetID string) (*upload.Upload, error)
	CreateDataset(dataset *upload.Upload) error
	UpdateDataset(dataset *upload.Upload) error
	DeleteDataset(dataset *upload.Upload) error
	CreateDatasetData(dataset *upload.Upload, datasetData []data.Datum) error
	ActivateDatasetData(dataset *upload.Upload) error
	DeleteOtherDatasetData(dataset *upload.Upload) error
	DestroyDataForUserByID(userID string) error
}

type Filter struct {
	Deleted bool
}

func NewFilter() *Filter {
	return &Filter{}
}

func (f *Filter) Validate() error {
	return nil
}

const (
	PaginationPageMinimum = 0
	PaginationSizeMinimum = 1
	PaginationSizeMaximum = 100
)

type Pagination struct {
	Page int
	Size int
}

func NewPagination() *Pagination {
	return &Pagination{
		Size: PaginationSizeMaximum,
	}
}

func (p *Pagination) Validate() error {
	if p.Page < PaginationPageMinimum {
		return app.Error("store", "page is invalid")
	}
	if p.Size < PaginationSizeMinimum || p.Size > PaginationSizeMaximum {
		return app.Error("store", "size is invalid")
	}
	return nil
}
