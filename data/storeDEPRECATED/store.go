package storeDEPRECATED

import (
	"context"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/store"
)

type Store interface {
	store.Store

	NewDataSession() DataSession
}

type DataSession interface {
	store.Session

	GetDatasetsForUserByID(ctx context.Context, userID string, filter *Filter, pagination *page.Pagination) ([]*upload.Upload, error)
	GetDatasetByID(ctx context.Context, datasetID string) (*upload.Upload, error)
	CreateDataset(ctx context.Context, dataset *upload.Upload) error
	UpdateDataSet(ctx context.Context, id string, update *data.DataSetUpdate) (*upload.Upload, error)
	DeleteDataset(ctx context.Context, dataset *upload.Upload) error
	CreateDatasetData(ctx context.Context, dataset *upload.Upload, datasetData []data.Datum) error
	ActivateDatasetData(ctx context.Context, dataset *upload.Upload) error
	ArchiveDeviceDataUsingHashesFromDataset(ctx context.Context, dataset *upload.Upload) error
	UnarchiveDeviceDataUsingHashesFromDataset(ctx context.Context, dataset *upload.Upload) error
	DeleteOtherDatasetData(ctx context.Context, dataset *upload.Upload) error
	DestroyDataForUserByID(ctx context.Context, userID string) error

	ListUserDataSets(ctx context.Context, userID string, filter *data.DataSetFilter, pagination *page.Pagination) (data.DataSets, error)
	GetDataSet(ctx context.Context, id string) (*data.DataSet, error)
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
