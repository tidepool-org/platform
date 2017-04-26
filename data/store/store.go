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

	NewSession(logger log.Logger) (Session, error)
}

type Session interface {
	store.Session

	GetDatasetsForUserByID(userID string, filter *Filter, pagination *Pagination) ([]*upload.Upload, error)
	GetDatasetByID(datasetID string) (*upload.Upload, error)
	FindPreviousActiveDatasetForDevice(dataset *upload.Upload) (*upload.Upload, error)
	CreateDataset(dataset *upload.Upload) error
	UpdateDataset(dataset *upload.Upload) error
	DeleteDataset(dataset *upload.Upload) error
	GetDatasetDataDeduplicatorHashes(dataset *upload.Upload, active bool) ([]string, error)
	FindAllDatasetDataDeduplicatorHashesForDevice(userID string, deviceID string, queryHashes []string) ([]string, error)
	CreateDatasetData(dataset *upload.Upload, datasetData []data.Datum) error
	FindEarliestDatasetDataTime(dataset *upload.Upload) (string, error)
	ActivateDatasetData(dataset *upload.Upload) error
	SetDatasetDataActiveUsingHashes(dataset *upload.Upload, queryHashes []string, active bool) error
	DeactivateOtherDatasetDataAfterTime(dataset *upload.Upload, time string) error
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
