package store

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/store"
)

type Store interface {
	store.Store

	NewDataSession(logger log.Logger) DataSession
}

type DataSession interface {
	store.Session

	GetDatasetsForUserByID(userID string, filter *Filter, pagination *Pagination) ([]*upload.Upload, error)
	GetDatasetByID(datasetID string) (*upload.Upload, error)
	CreateDataset(dataset *upload.Upload) error
	UpdateDataset(dataset *upload.Upload) error
	DeleteDataset(dataset *upload.Upload) error
	CreateDatasetData(dataset *upload.Upload, datasetData []data.Datum) error
	ActivateDatasetData(dataset *upload.Upload) error
	ArchiveDeviceDataUsingHashesFromDataset(dataset *upload.Upload) error
	UnarchiveDeviceDataUsingHashesFromDataset(dataset *upload.Upload) error
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
		return errors.New("store", "page is invalid")
	}
	if p.Size < PaginationSizeMinimum || p.Size > PaginationSizeMaximum {
		return errors.New("store", "size is invalid")
	}
	return nil
}
