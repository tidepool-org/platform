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
	GetDataSetsForUserByID(ctx context.Context, userID string, filter *Filter, pagination *page.Pagination) ([]*upload.Upload, error)
	GetDataSetByID(ctx context.Context, dataSetID string) (*upload.Upload, error)
	CreateDataSet(ctx context.Context, dataSet *upload.Upload) error
	UpdateDataSet(ctx context.Context, id string, update *data.DataSetUpdate) (*upload.Upload, error)
	DeleteDataSet(ctx context.Context, dataSet *upload.Upload, doPurge bool) error

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

// Filter available on HTTP query
type Filter struct {
	Deleted     bool
	State       *string // State: open, closed
	DataSetType *string // DataSetType: continuous, normal
}

// NewFilter for HTTP query URL
func NewFilter() *Filter {
	return &Filter{}
}

// Parse HTTP query URL parameters
func (f *Filter) Parse(parser structure.ObjectParser) {
	if deleted := parser.Bool("deleted"); deleted != nil {
		f.Deleted = *deleted
	}
	if state := parser.String("state"); state != nil {
		f.State = state
	}
	if dataSetType := parser.String("dataSetType"); dataSetType != nil {
		f.DataSetType = dataSetType
	}
}

// Validate HTTP query URL parameters
func (f *Filter) Validate(validator structure.Validator) {
	if f.State != nil {
		validator.String("state", f.State).OneOf(upload.States()...)
	}
	if f.DataSetType != nil {
		validator.String("dataSetType", f.DataSetType).OneOf(upload.DataSetTypes()...)
	}
}
