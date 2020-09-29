package store

import (
	"context"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/page"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	"github.com/tidepool-org/platform/structure"
)

type Store interface {
	Status(ctx context.Context) *storeStructuredMongo.Status

	NewDataRepository() DataRepository
}

type DataRepository interface {
	EnsureIndexes() error

	GetDataSetsForUserByID(ctx context.Context, userID string, filter *Filter, pagination *page.Pagination) ([]*upload.Upload, error)
	GetDataSetByID(ctx context.Context, dataSetID string) (*upload.Upload, error)
	CreateDataSet(ctx context.Context, dataSet *upload.Upload) error
	UpdateDataSet(ctx context.Context, id string, update *data.DataSetUpdate) (*upload.Upload, error)
	DeleteDataSet(ctx context.Context, dataSet *upload.Upload) error

	CreateDataSetData(ctx context.Context, dataSet *upload.Upload, dataSetData []data.Datum) error
	ActivateDataSetData(ctx context.Context, dataSet *upload.Upload, selectors *data.Selectors) error
	ArchiveDataSetData(ctx context.Context, dataSet *upload.Upload, selectors *data.Selectors) error
	DeleteDataSetData(ctx context.Context, dataSet *upload.Upload, selectors *data.Selectors) error
	DestroyDeletedDataSetData(ctx context.Context, dataSet *upload.Upload, selectors *data.Selectors) error
	DestroyDataSetData(ctx context.Context, dataSet *upload.Upload, selectors *data.Selectors) error

	ArchiveDeviceDataUsingHashesFromDataSet(ctx context.Context, dataSet *upload.Upload) error
	UnarchiveDeviceDataUsingHashesFromDataSet(ctx context.Context, dataSet *upload.Upload) error
	DeleteOtherDataSetData(ctx context.Context, dataSet *upload.Upload) error
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

func (f *Filter) Parse(parser structure.ObjectParser) {
	if deleted := parser.Bool("deleted"); deleted != nil {
		f.Deleted = *deleted
	}
}

func (f *Filter) Validate(validator structure.Validator) {}
