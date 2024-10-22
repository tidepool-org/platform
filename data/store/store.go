package store

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/tidepool-org/platform/alerts"
	"github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/data/types/dosingdecision"

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
	NewBucketsRepository() BucketsRepository
	NewAlertsRepository() alerts.Repository
}

// DataSetRepository is the interface for interacting and modifying
// the "parent" datum document, formerly the documents where type
// = "upload".
type DataSetRepository interface {
	EnsureIndexes() error

	GetDataSetsForUserByID(ctx context.Context, userID string, filter *Filter, pagination *page.Pagination) ([]*upload.Upload, error)
	GetDataSetByID(ctx context.Context, dataSetID string) (*upload.Upload, error)
	CreateDataSet(ctx context.Context, dataSet *upload.Upload) error
	UpdateDataSet(ctx context.Context, id string, update *data.DataSetUpdate) (*upload.Upload, error)
	DeleteDataSet(ctx context.Context, dataSet *upload.Upload) error
	DestroyDataForUserByID(ctx context.Context, userID string) error

	ListUserDataSets(ctx context.Context, userID string, filter *data.DataSetFilter, pagination *page.Pagination) (data.DataSets, error)
	GetDataSet(ctx context.Context, dataSetID string) (*data.DataSet, error)
}

// DatumRepository is the interface for interacting and modifying
// the "children" data documents, documents where type != "upload" and
// whose "parent" is the datum whose type = "upload". It can be thought of as
// the DataSet's data.
type DatumRepository interface {
	EnsureIndexes() error

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

	GetDataRange(ctx context.Context, userId string, typ []string, status *data.UserDataStatus) (*mongo.Cursor, error)
	GetLastUpdatedForUser(ctx context.Context, userId string, typ []string, lastUpdated time.Time) (*data.UserDataStatus, error)
	DistinctUserIDs(ctx context.Context, typ []string) ([]string, error)

	// GetAlertableData queries for the data used to evaluate alerts configurations.
	GetAlertableData(ctx context.Context, params AlertableParams) (*AlertableResponse, error)
}

// DataRepository is the combined interface of DataSetRepository and
// DatumRepository.
type DataRepository interface {
	DataSetRepository
	DatumRepository
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

	GetStore() *storeStructuredMongo.Repository
}

type BucketsRepository interface {
	EnsureIndexes() error

	GetStore() *storeStructuredMongo.Repository
}

type AlertableParams struct {
	// UserID of the user that owns the data.
	UserID string
	// UploadID of the device data set to query.
	UploadID string
	// Start limits the data to those recorded after this time.
	Start time.Time
	// End limits the data to those recorded before this time.
	End time.Time
}

type AlertableResponse struct {
	Glucose         []*glucose.Glucose
	DosingDecisions []*dosingdecision.DosingDecision
}
