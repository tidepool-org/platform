package test

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	dataStore "github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/test"
)

type GetDataSetsForUserByIDInput struct {
	Context    context.Context
	UserID     string
	Filter     *dataStore.Filter
	Pagination *page.Pagination
}

type GetDataSetsForUserByIDOutput struct {
	DataSets []*data.DataSet
	Error    error
}

type CreateDataSetInput struct {
	Context context.Context
	DataSet *data.DataSet
}

type CreateUserDataSetInput struct {
	Context context.Context
	UserID  string
	Create  *data.DataSetCreate
}

type CreateUserDataSetOutput struct {
	DataSet *data.DataSet
	Error   error
}

type UpdateDataSetInput struct {
	Context context.Context
	ID      string
	Update  *data.DataSetUpdate
}

type UpdateDataSetOutput struct {
	DataSet *data.DataSet
	Error   error
}

type DeleteDataSetInput struct {
	Context context.Context
	DataSet *data.DataSet
}

type CreateDataSetDataInput struct {
	Context     context.Context
	DataSet     *data.DataSet
	DataSetData []data.Datum
}

type ExistingDataSetDataInput struct {
	Context   context.Context
	DataSet   *data.DataSet
	Selectors *data.Selectors
}

type ExistingDataSetDataOutput struct {
	Selectors *data.Selectors
	Error     error
}

type ActivateDataSetDataInput struct {
	Context   context.Context
	DataSet   *data.DataSet
	Selectors *data.Selectors
}

type ArchiveDataSetDataInput struct {
	Context   context.Context
	DataSet   *data.DataSet
	Selectors *data.Selectors
}

type DeleteDataSetDataInput struct {
	Context   context.Context
	DataSet   *data.DataSet
	Selectors *data.Selectors
}

type DestroyDeletedDataSetDataInput struct {
	Context   context.Context
	DataSet   *data.DataSet
	Selectors *data.Selectors
}

type DestroyDataSetDataInput struct {
	Context   context.Context
	DataSet   *data.DataSet
	Selectors *data.Selectors
}

type ArchiveDeviceDataUsingHashesFromDataSetInput struct {
	Context context.Context
	DataSet *data.DataSet
}

type UnarchiveDeviceDataUsingHashesFromDataSetInput struct {
	Context context.Context
	DataSet *data.DataSet
}

type DeleteOtherDataSetDataInput struct {
	Context context.Context
	DataSet *data.DataSet
}

type DestroyDataForUserByIDInput struct {
	Context context.Context
	UserID  string
}

type GetDataSetInput struct {
	Context context.Context
	ID      string
}

type GetDataSetOutput struct {
	DataSet *data.DataSet
	Error   error
}

type ListUserDataSetsInput struct {
	Context    context.Context
	UserID     string
	Filter     *data.DataSetFilter
	Pagination *page.Pagination
}

type ListUserDataSetsOutput struct {
	DataSets data.DataSets
	Error    error
}

type GetLastUpdatedForUserInput struct {
	Context     context.Context
	UserID      string
	Typ         []string
	LastUpdated time.Time
}

type GetLastUpdatedForUserOutput struct {
	UserLastUpdated *data.UserDataStatus
	Error           error
}

type GetDataRangeInput struct {
	Context context.Context
	UserId  string
	Typ     []string
	Status  *data.UserDataStatus
}

type GetDataRangeOutput struct {
	Error  error
	Cursor *mongo.Cursor
}

type GetUsersWithBGDataSinceInput struct {
	Context     context.Context
	LastUpdated time.Time
}

type GetUsersWithBGDataSinceOutput struct {
	UserIDs []string
	Error   error
}

type GetAlertableDataInput struct {
	Context context.Context
	Params  dataStore.AlertableParams
}

type GetAlertableDataOutput struct {
	Response *dataStore.AlertableResponse
	Error    error
}

type DataRepository struct {
	*test.Closer
	GetDataSetsForUserByIDInvocations                    int
	GetDataSetsForUserByIDInputs                         []GetDataSetsForUserByIDInput
	GetDataSetsForUserByIDOutputs                        []GetDataSetsForUserByIDOutput
	CreateDataSetInvocations                             int
	CreateDataSetInputs                                  []CreateDataSetInput
	CreateDataSetOutputs                                 []error
	CreateUserDataSetInvocations                         int
	CreateUserDataSetInputs                              []CreateUserDataSetInput
	CreateUserDataSetOutputs                             []CreateUserDataSetOutput
	UpdateDataSetInvocations                             int
	UpdateDataSetInputs                                  []UpdateDataSetInput
	UpdateDataSetOutputs                                 []UpdateDataSetOutput
	DeleteDataSetInvocations                             int
	DeleteDataSetInputs                                  []DeleteDataSetInput
	DeleteDataSetOutputs                                 []error
	CreateDataSetDataInvocations                         int
	CreateDataSetDataInputs                              []CreateDataSetDataInput
	CreateDataSetDataOutputs                             []error
	ExistingDataSetDataInvocations                       int
	EnsureAuthorizedInvocations                          int
	ExistingDataSetDataInputs                            []ExistingDataSetDataInput
	ExistingDataSetDataStub                              func(ctx context.Context, dataSet *data.DataSet, selectors *data.Selectors) (*data.Selectors, error)
	ExistingDataSetDataOutputs                           []ExistingDataSetDataOutput
	ActivateDataSetDataInvocations                       int
	ActivateDataSetDataInputs                            []ActivateDataSetDataInput
	ActivateDataSetDataOutputs                           []error
	ArchiveDataSetDataInvocations                        int
	ArchiveDataSetDataInputs                             []ArchiveDataSetDataInput
	ArchiveDataSetDataOutputs                            []error
	DeleteDataSetDataInvocations                         int
	DeleteDataSetDataInputs                              []DeleteDataSetDataInput
	DeleteDataSetDataOutputs                             []error
	DestroyDeletedDataSetDataInvocations                 int
	DestroyDeletedDataSetDataInputs                      []DestroyDeletedDataSetDataInput
	DestroyDeletedDataSetDataOutputs                     []error
	DestroyDataSetDataInvocations                        int
	DestroyDataSetDataInputs                             []DestroyDataSetDataInput
	DestroyDataSetDataOutputs                            []error
	ArchiveDeviceDataUsingHashesFromDataSetInvocations   int
	ArchiveDeviceDataUsingHashesFromDataSetInputs        []ArchiveDeviceDataUsingHashesFromDataSetInput
	ArchiveDeviceDataUsingHashesFromDataSetOutputs       []error
	UnarchiveDeviceDataUsingHashesFromDataSetInvocations int
	UnarchiveDeviceDataUsingHashesFromDataSetInputs      []UnarchiveDeviceDataUsingHashesFromDataSetInput
	UnarchiveDeviceDataUsingHashesFromDataSetOutputs     []error
	DeleteOtherDataSetDataInvocations                    int
	DeleteOtherDataSetDataInputs                         []DeleteOtherDataSetDataInput
	DeleteOtherDataSetDataOutputs                        []error
	DestroyDataForUserByIDInvocations                    int
	DestroyDataForUserByIDInputs                         []DestroyDataForUserByIDInput
	DestroyDataForUserByIDOutputs                        []error
	ListUserDataSetsInvocations                          int
	ListUserDataSetsInputs                               []ListUserDataSetsInput
	ListUserDataSetsOutputs                              []ListUserDataSetsOutput
	GetDataSetInvocations                                int
	GetDataSetInputs                                     []GetDataSetInput
	GetDataSetOutputs                                    []GetDataSetOutput

	GetDataRangeInvocations int
	GetDataRangeInputs      []GetDataRangeInput
	GetDataRangeOutputs     []GetDataRangeOutput

	GetLastUpdatedForUserInvocations int
	GetLastUpdatedForUserInputs      []GetLastUpdatedForUserInput
	GetLastUpdatedForUserOutputs     []GetLastUpdatedForUserOutput

	GetUsersWithBGDataSinceInvocations int
	GetUsersWithBGDataSinceInputs      []GetUsersWithBGDataSinceInput
	GetUsersWithBGDataSinceOutputs     []GetUsersWithBGDataSinceOutput

	GetAlertableDataInvocations int
	GetAlertableDataInputs      []GetAlertableDataInput
	GetAlertableDataOutputs     []GetAlertableDataOutput
}

func NewDataRepository() *DataRepository {
	return &DataRepository{
		Closer: test.NewCloser(),
	}
}

// EnsureIndexes required in order to implement the DataRepository interface
func (d *DataRepository) EnsureIndexes() error {
	return nil
}

func (d *DataRepository) GetDataSetsForUserByID(ctx context.Context, userID string, filter *dataStore.Filter, pagination *page.Pagination) ([]*data.DataSet, error) {
	d.GetDataSetsForUserByIDInvocations++

	d.GetDataSetsForUserByIDInputs = append(d.GetDataSetsForUserByIDInputs, GetDataSetsForUserByIDInput{Context: ctx, UserID: userID, Filter: filter, Pagination: pagination})

	gomega.Expect(d.GetDataSetsForUserByIDOutputs).ToNot(gomega.BeEmpty())

	output := d.GetDataSetsForUserByIDOutputs[0]
	d.GetDataSetsForUserByIDOutputs = d.GetDataSetsForUserByIDOutputs[1:]
	return output.DataSets, output.Error
}

func (d *DataRepository) CreateDataSet(ctx context.Context, dataSet *data.DataSet) error {
	d.CreateDataSetInvocations++

	d.CreateDataSetInputs = append(d.CreateDataSetInputs, CreateDataSetInput{Context: ctx, DataSet: dataSet})

	gomega.Expect(d.CreateDataSetOutputs).ToNot(gomega.BeEmpty())

	output := d.CreateDataSetOutputs[0]
	d.CreateDataSetOutputs = d.CreateDataSetOutputs[1:]
	return output
}

func (d *DataRepository) CreateUserDataSet(ctx context.Context, userID string, create *data.DataSetCreate) (*data.DataSet, error) {
	d.CreateUserDataSetInvocations++

	d.CreateUserDataSetInputs = append(d.CreateUserDataSetInputs, CreateUserDataSetInput{Context: ctx, UserID: userID, Create: create})

	gomega.Expect(d.CreateUserDataSetOutputs).ToNot(gomega.BeEmpty())

	output := d.CreateUserDataSetOutputs[0]
	d.CreateUserDataSetOutputs = d.CreateUserDataSetOutputs[1:]
	return output.DataSet, output.Error
}

func (d *DataRepository) UpdateDataSet(ctx context.Context, id string, update *data.DataSetUpdate) (*data.DataSet, error) {
	d.UpdateDataSetInvocations++

	d.UpdateDataSetInputs = append(d.UpdateDataSetInputs, UpdateDataSetInput{Context: ctx, ID: id, Update: update})

	gomega.Expect(d.UpdateDataSetOutputs).ToNot(gomega.BeEmpty())

	output := d.UpdateDataSetOutputs[0]
	d.UpdateDataSetOutputs = d.UpdateDataSetOutputs[1:]
	return output.DataSet, output.Error
}

func (d *DataRepository) DeleteDataSet(ctx context.Context, dataSet *data.DataSet) error {
	d.DeleteDataSetInvocations++

	d.DeleteDataSetInputs = append(d.DeleteDataSetInputs, DeleteDataSetInput{Context: ctx, DataSet: dataSet})

	gomega.Expect(d.DeleteDataSetOutputs).ToNot(gomega.BeEmpty())

	output := d.DeleteDataSetOutputs[0]
	d.DeleteDataSetOutputs = d.DeleteDataSetOutputs[1:]
	return output
}

func (d *DataRepository) CreateDataSetData(ctx context.Context, dataSet *data.DataSet, dataSetData []data.Datum) error {
	d.CreateDataSetDataInvocations++

	d.CreateDataSetDataInputs = append(d.CreateDataSetDataInputs, CreateDataSetDataInput{Context: ctx, DataSet: dataSet, DataSetData: dataSetData})

	gomega.Expect(d.CreateDataSetDataOutputs).ToNot(gomega.BeEmpty())

	output := d.CreateDataSetDataOutputs[0]
	d.CreateDataSetDataOutputs = d.CreateDataSetDataOutputs[1:]
	return output
}

func (d *DataRepository) ExistingDataSetData(ctx context.Context, dataSet *data.DataSet, selectors *data.Selectors) (*data.Selectors, error) {
	d.ExistingDataSetDataInvocations++

	d.ExistingDataSetDataInputs = append(d.ExistingDataSetDataInputs, ExistingDataSetDataInput{Context: ctx, DataSet: dataSet, Selectors: selectors})

	if d.ExistingDataSetDataStub != nil {
		return d.ExistingDataSetDataStub(ctx, dataSet, selectors)
	}

	gomega.Expect(d.ExistingDataSetDataOutputs).ToNot(gomega.BeEmpty())

	output := d.ExistingDataSetDataOutputs[0]
	d.ExistingDataSetDataOutputs = d.ExistingDataSetDataOutputs[1:]
	return output.Selectors, output.Error
}

func (d *DataRepository) ActivateDataSetData(ctx context.Context, dataSet *data.DataSet, selectors *data.Selectors) error {
	d.ActivateDataSetDataInvocations++

	d.ActivateDataSetDataInputs = append(d.ActivateDataSetDataInputs, ActivateDataSetDataInput{Context: ctx, DataSet: dataSet, Selectors: selectors})

	gomega.Expect(d.ActivateDataSetDataOutputs).ToNot(gomega.BeEmpty())

	output := d.ActivateDataSetDataOutputs[0]
	d.ActivateDataSetDataOutputs = d.ActivateDataSetDataOutputs[1:]
	return output
}

func (d *DataRepository) ArchiveDataSetData(ctx context.Context, dataSet *data.DataSet, selectors *data.Selectors) error {
	d.ArchiveDataSetDataInvocations++

	d.ArchiveDataSetDataInputs = append(d.ArchiveDataSetDataInputs, ArchiveDataSetDataInput{Context: ctx, DataSet: dataSet, Selectors: selectors})

	gomega.Expect(d.ArchiveDataSetDataOutputs).ToNot(gomega.BeEmpty())

	output := d.ArchiveDataSetDataOutputs[0]
	d.ArchiveDataSetDataOutputs = d.ArchiveDataSetDataOutputs[1:]
	return output
}

func (d *DataRepository) DeleteDataSetData(ctx context.Context, dataSet *data.DataSet, selectors *data.Selectors) error {
	d.DeleteDataSetDataInvocations++

	d.DeleteDataSetDataInputs = append(d.DeleteDataSetDataInputs, DeleteDataSetDataInput{Context: ctx, DataSet: dataSet, Selectors: selectors})

	gomega.Expect(d.DeleteDataSetDataOutputs).ToNot(gomega.BeEmpty())

	output := d.DeleteDataSetDataOutputs[0]
	d.DeleteDataSetDataOutputs = d.DeleteDataSetDataOutputs[1:]
	return output
}

func (d *DataRepository) DestroyDeletedDataSetData(ctx context.Context, dataSet *data.DataSet, selectors *data.Selectors) error {
	d.DestroyDeletedDataSetDataInvocations++

	d.DestroyDeletedDataSetDataInputs = append(d.DestroyDeletedDataSetDataInputs, DestroyDeletedDataSetDataInput{Context: ctx, DataSet: dataSet, Selectors: selectors})

	gomega.Expect(d.DestroyDeletedDataSetDataOutputs).ToNot(gomega.BeEmpty())

	output := d.DestroyDeletedDataSetDataOutputs[0]
	d.DestroyDeletedDataSetDataOutputs = d.DestroyDeletedDataSetDataOutputs[1:]
	return output
}

func (d *DataRepository) DestroyDataSetData(ctx context.Context, dataSet *data.DataSet, selectors *data.Selectors) error {
	d.DestroyDataSetDataInvocations++

	d.DestroyDataSetDataInputs = append(d.DestroyDataSetDataInputs, DestroyDataSetDataInput{Context: ctx, DataSet: dataSet, Selectors: selectors})

	gomega.Expect(d.DestroyDataSetDataOutputs).ToNot(gomega.BeEmpty())

	output := d.DestroyDataSetDataOutputs[0]
	d.DestroyDataSetDataOutputs = d.DestroyDataSetDataOutputs[1:]
	return output
}

func (d *DataRepository) ArchiveDeviceDataUsingHashesFromDataSet(ctx context.Context, dataSet *data.DataSet) error {
	d.ArchiveDeviceDataUsingHashesFromDataSetInvocations++

	d.ArchiveDeviceDataUsingHashesFromDataSetInputs = append(d.ArchiveDeviceDataUsingHashesFromDataSetInputs, ArchiveDeviceDataUsingHashesFromDataSetInput{Context: ctx, DataSet: dataSet})

	gomega.Expect(d.ArchiveDeviceDataUsingHashesFromDataSetOutputs).ToNot(gomega.BeEmpty())

	output := d.ArchiveDeviceDataUsingHashesFromDataSetOutputs[0]
	d.ArchiveDeviceDataUsingHashesFromDataSetOutputs = d.ArchiveDeviceDataUsingHashesFromDataSetOutputs[1:]
	return output
}

func (d *DataRepository) UnarchiveDeviceDataUsingHashesFromDataSet(ctx context.Context, dataSet *data.DataSet) error {
	d.UnarchiveDeviceDataUsingHashesFromDataSetInvocations++

	d.UnarchiveDeviceDataUsingHashesFromDataSetInputs = append(d.UnarchiveDeviceDataUsingHashesFromDataSetInputs, UnarchiveDeviceDataUsingHashesFromDataSetInput{Context: ctx, DataSet: dataSet})

	gomega.Expect(d.UnarchiveDeviceDataUsingHashesFromDataSetOutputs).ToNot(gomega.BeEmpty())

	output := d.UnarchiveDeviceDataUsingHashesFromDataSetOutputs[0]
	d.UnarchiveDeviceDataUsingHashesFromDataSetOutputs = d.UnarchiveDeviceDataUsingHashesFromDataSetOutputs[1:]
	return output
}

func (d *DataRepository) DeleteOtherDataSetData(ctx context.Context, dataSet *data.DataSet) error {
	d.DeleteOtherDataSetDataInvocations++

	d.DeleteOtherDataSetDataInputs = append(d.DeleteOtherDataSetDataInputs, DeleteOtherDataSetDataInput{Context: ctx, DataSet: dataSet})

	gomega.Expect(d.DeleteOtherDataSetDataOutputs).ToNot(gomega.BeEmpty())

	output := d.DeleteOtherDataSetDataOutputs[0]
	d.DeleteOtherDataSetDataOutputs = d.DeleteOtherDataSetDataOutputs[1:]
	return output
}

func (d *DataRepository) DestroyDataForUserByID(ctx context.Context, userID string) error {
	d.DestroyDataForUserByIDInvocations++

	d.DestroyDataForUserByIDInputs = append(d.DestroyDataForUserByIDInputs, DestroyDataForUserByIDInput{Context: ctx, UserID: userID})

	gomega.Expect(d.DestroyDataForUserByIDOutputs).ToNot(gomega.BeEmpty())

	output := d.DestroyDataForUserByIDOutputs[0]
	d.DestroyDataForUserByIDOutputs = d.DestroyDataForUserByIDOutputs[1:]
	return output
}

func (d *DataRepository) ListUserDataSets(ctx context.Context, userID string, filter *data.DataSetFilter, pagination *page.Pagination) (data.DataSets, error) {
	d.ListUserDataSetsInvocations++

	d.ListUserDataSetsInputs = append(d.ListUserDataSetsInputs, ListUserDataSetsInput{Context: ctx, UserID: userID, Filter: filter, Pagination: pagination})

	gomega.Expect(d.ListUserDataSetsOutputs).ToNot(gomega.BeEmpty())

	output := d.ListUserDataSetsOutputs[0]
	d.ListUserDataSetsOutputs = d.ListUserDataSetsOutputs[1:]
	return output.DataSets, output.Error
}

func (d *DataRepository) GetDataSet(ctx context.Context, id string) (*data.DataSet, error) {
	d.GetDataSetInvocations++

	d.GetDataSetInputs = append(d.GetDataSetInputs, GetDataSetInput{Context: ctx, ID: id})

	gomega.Expect(d.GetDataSetOutputs).ToNot(gomega.BeEmpty())

	output := d.GetDataSetOutputs[0]
	d.GetDataSetOutputs = d.GetDataSetOutputs[1:]
	return output.DataSet, output.Error
}

func (d *DataRepository) GetLastUpdatedForUser(ctx context.Context, userId string, typ []string, lastUpdated time.Time) (*data.UserDataStatus, error) {
	d.GetLastUpdatedForUserInvocations++

	d.GetLastUpdatedForUserInputs = append(d.GetLastUpdatedForUserInputs, GetLastUpdatedForUserInput{Context: ctx, UserID: userId, Typ: typ, LastUpdated: lastUpdated})

	gomega.Expect(d.GetLastUpdatedForUserOutputs).ToNot(gomega.BeEmpty())

	output := d.GetLastUpdatedForUserOutputs[0]
	d.GetLastUpdatedForUserOutputs = d.GetLastUpdatedForUserOutputs[1:]
	return output.UserLastUpdated, output.Error
}

func (d *DataRepository) GetDataRange(ctx context.Context, userId string, typ []string, status *data.UserDataStatus) (*mongo.Cursor, error) {
	d.GetDataRangeInvocations++

	d.GetDataRangeInputs = append(d.GetDataRangeInputs, GetDataRangeInput{Context: ctx, UserId: userId, Typ: typ, Status: status})

	gomega.Expect(d.GetDataRangeOutputs).ToNot(gomega.BeEmpty())

	output := d.GetDataRangeOutputs[0]
	d.GetDataRangeOutputs = d.GetDataRangeOutputs[1:]
	return output.Cursor, output.Error
}

func (d *DataRepository) GetUsersWithBGDataSince(ctx context.Context, lastUpdated time.Time) ([]string, error) {
	d.GetUsersWithBGDataSinceInvocations++

	d.GetUsersWithBGDataSinceInputs = append(d.GetUsersWithBGDataSinceInputs, GetUsersWithBGDataSinceInput{Context: ctx, LastUpdated: lastUpdated})

	gomega.Expect(d.GetUsersWithBGDataSinceOutputs).ToNot(gomega.BeEmpty())

	output := d.GetUsersWithBGDataSinceOutputs[0]
	d.GetUsersWithBGDataSinceOutputs = d.GetUsersWithBGDataSinceOutputs[1:]
	return output.UserIDs, output.Error
}

func (d *DataRepository) GetAlertableData(ctx context.Context, params dataStore.AlertableParams) (*dataStore.AlertableResponse, error) {
	d.GetAlertableDataInvocations++

	d.GetAlertableDataInputs = append(d.GetAlertableDataInputs, GetAlertableDataInput{Context: ctx, Params: params})

	gomega.Expect(d.GetAlertableDataOutputs).ToNot(gomega.BeEmpty())

	output := d.GetAlertableDataOutputs[0]
	d.GetAlertableDataOutputs = d.GetAlertableDataOutputs[1:]
	return output.Response, output.Error
}

func (d *DataRepository) Expectations() {
	d.Closer.AssertOutputsEmpty()
	gomega.Expect(d.GetDataSetsForUserByIDOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.CreateDataSetOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.CreateUserDataSetOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.UpdateDataSetOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.DeleteDataSetOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.CreateDataSetDataOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.ActivateDataSetDataOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.ArchiveDataSetDataOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.DeleteDataSetDataOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.DestroyDeletedDataSetDataOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.DestroyDataSetDataOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.ArchiveDeviceDataUsingHashesFromDataSetOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.UnarchiveDeviceDataUsingHashesFromDataSetOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.DeleteOtherDataSetDataOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.DestroyDataForUserByIDOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.ListUserDataSetsOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.GetDataSetOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.GetLastUpdatedForUserOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.GetUsersWithBGDataSinceOutputs).To(gomega.BeEmpty())
}
