package store

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/data/types/blood/glucose"

	"github.com/tidepool-org/platform/data/summary"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/page"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	"github.com/tidepool-org/platform/structure"
)

type Store interface {
	Status(ctx context.Context) *storeStructuredMongo.Status

	NewDataRepository() DataRepository
	NewSummaryRepository() SummaryRepository
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

	GetDataRange(ctx context.Context, id string, t string, startTime time.Time, endTime time.Time) ([]*glucose.Glucose, error)
	GetLastUpdatedForUser(ctx context.Context, id string) (*summary.UserLastUpdated, error)
	DistinctUserIDs(ctx context.Context) ([]string, error)
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

type SummaryRepository interface {
	EnsureIndexes() error

	GetSummary(ctx context.Context, id string) (*summary.Summary, error)
	DeleteSummary(ctx context.Context, id string) error
	SetOutdated(ctx context.Context, id string, updates *data.SummaryTypeUpdates) (*summary.TypeOutdatedTimes, error)
	GetOutdatedUserIDs(ctx context.Context, page *page.Pagination) ([]string, error)
	UpdateSummary(ctx context.Context, summary *summary.Summary) (*summary.Summary, error)
	DistinctSummaryIDs(ctx context.Context) ([]string, error)
	CreateSummaries(ctx context.Context, summaries []*summary.Summary) (int, error)
}
