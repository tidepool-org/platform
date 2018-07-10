package test

import (
	"context"

	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	dataStoreDEPRECATED "github.com/tidepool-org/platform/data/storeDEPRECATED"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/test"
)

type GetDataSetsForUserByIDInput struct {
	Context    context.Context
	UserID     string
	Filter     *dataStoreDEPRECATED.Filter
	Pagination *page.Pagination
}

type GetDataSetsForUserByIDOutput struct {
	DataSets []*upload.Upload
	Error    error
}

type GetDataSetByIDInput struct {
	Context   context.Context
	DataSetID string
}

type GetDataSetByIDOutput struct {
	DataSet *upload.Upload
	Error   error
}

type CreateDataSetInput struct {
	Context context.Context
	DataSet *upload.Upload
}

type UpdateDataSetInput struct {
	Context context.Context
	ID      string
	Update  *data.DataSetUpdate
}

type UpdateDataSetOutput struct {
	DataSet *upload.Upload
	Error   error
}

type DeleteDataSetInput struct {
	Context context.Context
	DataSet *upload.Upload
}

type CreateDataSetDataInput struct {
	Context     context.Context
	DataSet     *upload.Upload
	DataSetData []data.Datum
}

type ActivateDataSetDataInput struct {
	Context context.Context
	DataSet *upload.Upload
}

type ArchiveDeviceDataUsingHashesFromDataSetInput struct {
	Context context.Context
	DataSet *upload.Upload
}

type UnarchiveDeviceDataUsingHashesFromDataSetInput struct {
	Context context.Context
	DataSet *upload.Upload
}

type DeleteOtherDataSetDataInput struct {
	Context context.Context
	DataSet *upload.Upload
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

type DataSession struct {
	*test.Mock
	*test.Closer
	IsClosedInvocations                                  int
	IsClosedOutputs                                      []bool
	EnsureIndexesInvocations                             int
	EnsureIndexesOutputs                                 []error
	GetDataSetsForUserByIDInvocations                    int
	GetDataSetsForUserByIDInputs                         []GetDataSetsForUserByIDInput
	GetDataSetsForUserByIDOutputs                        []GetDataSetsForUserByIDOutput
	GetDataSetByIDInvocations                            int
	GetDataSetByIDInputs                                 []GetDataSetByIDInput
	GetDataSetByIDOutputs                                []GetDataSetByIDOutput
	CreateDataSetInvocations                             int
	CreateDataSetInputs                                  []CreateDataSetInput
	CreateDataSetOutputs                                 []error
	UpdateDataSetInvocations                             int
	UpdateDataSetInputs                                  []UpdateDataSetInput
	UpdateDataSetOutputs                                 []UpdateDataSetOutput
	DeleteDataSetInvocations                             int
	DeleteDataSetInputs                                  []DeleteDataSetInput
	DeleteDataSetOutputs                                 []error
	CreateDataSetDataInvocations                         int
	CreateDataSetDataInputs                              []CreateDataSetDataInput
	CreateDataSetDataOutputs                             []error
	ActivateDataSetDataInvocations                       int
	ActivateDataSetDataInputs                            []ActivateDataSetDataInput
	ActivateDataSetDataOutputs                           []error
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
}

func NewDataSession() *DataSession {
	return &DataSession{
		Mock:   test.NewMock(),
		Closer: test.NewCloser(),
	}
}

func (d *DataSession) IsClosed() bool {
	d.IsClosedInvocations++

	gomega.Expect(d.IsClosedOutputs).ToNot(gomega.BeEmpty())

	output := d.IsClosedOutputs[0]
	d.IsClosedOutputs = d.IsClosedOutputs[1:]
	return output
}

func (d *DataSession) EnsureIndexes() error {
	d.EnsureIndexesInvocations++

	gomega.Expect(d.EnsureIndexesOutputs).ToNot(gomega.BeEmpty())

	output := d.EnsureIndexesOutputs[0]
	d.EnsureIndexesOutputs = d.EnsureIndexesOutputs[1:]
	return output
}

func (d *DataSession) GetDataSetsForUserByID(ctx context.Context, userID string, filter *dataStoreDEPRECATED.Filter, pagination *page.Pagination) ([]*upload.Upload, error) {
	d.GetDataSetsForUserByIDInvocations++

	d.GetDataSetsForUserByIDInputs = append(d.GetDataSetsForUserByIDInputs, GetDataSetsForUserByIDInput{Context: ctx, UserID: userID, Filter: filter, Pagination: pagination})

	gomega.Expect(d.GetDataSetsForUserByIDOutputs).ToNot(gomega.BeEmpty())

	output := d.GetDataSetsForUserByIDOutputs[0]
	d.GetDataSetsForUserByIDOutputs = d.GetDataSetsForUserByIDOutputs[1:]
	return output.DataSets, output.Error
}

func (d *DataSession) GetDataSetByID(ctx context.Context, dataSetID string) (*upload.Upload, error) {
	d.GetDataSetByIDInvocations++

	d.GetDataSetByIDInputs = append(d.GetDataSetByIDInputs, GetDataSetByIDInput{Context: ctx, DataSetID: dataSetID})

	gomega.Expect(d.GetDataSetByIDOutputs).ToNot(gomega.BeEmpty())

	output := d.GetDataSetByIDOutputs[0]
	d.GetDataSetByIDOutputs = d.GetDataSetByIDOutputs[1:]
	return output.DataSet, output.Error
}

func (d *DataSession) CreateDataSet(ctx context.Context, dataSet *upload.Upload) error {
	d.CreateDataSetInvocations++

	d.CreateDataSetInputs = append(d.CreateDataSetInputs, CreateDataSetInput{Context: ctx, DataSet: dataSet})

	gomega.Expect(d.CreateDataSetOutputs).ToNot(gomega.BeEmpty())

	output := d.CreateDataSetOutputs[0]
	d.CreateDataSetOutputs = d.CreateDataSetOutputs[1:]
	return output
}

func (d *DataSession) UpdateDataSet(ctx context.Context, id string, update *data.DataSetUpdate) (*upload.Upload, error) {
	d.UpdateDataSetInvocations++

	d.UpdateDataSetInputs = append(d.UpdateDataSetInputs, UpdateDataSetInput{Context: ctx, ID: id, Update: update})

	gomega.Expect(d.UpdateDataSetOutputs).ToNot(gomega.BeEmpty())

	output := d.UpdateDataSetOutputs[0]
	d.UpdateDataSetOutputs = d.UpdateDataSetOutputs[1:]
	return output.DataSet, output.Error
}

func (d *DataSession) DeleteDataSet(ctx context.Context, dataSet *upload.Upload) error {
	d.DeleteDataSetInvocations++

	d.DeleteDataSetInputs = append(d.DeleteDataSetInputs, DeleteDataSetInput{Context: ctx, DataSet: dataSet})

	gomega.Expect(d.DeleteDataSetOutputs).ToNot(gomega.BeEmpty())

	output := d.DeleteDataSetOutputs[0]
	d.DeleteDataSetOutputs = d.DeleteDataSetOutputs[1:]
	return output
}

func (d *DataSession) CreateDataSetData(ctx context.Context, dataSet *upload.Upload, dataSetData []data.Datum) error {
	d.CreateDataSetDataInvocations++

	d.CreateDataSetDataInputs = append(d.CreateDataSetDataInputs, CreateDataSetDataInput{Context: ctx, DataSet: dataSet, DataSetData: dataSetData})

	gomega.Expect(d.CreateDataSetDataOutputs).ToNot(gomega.BeEmpty())

	output := d.CreateDataSetDataOutputs[0]
	d.CreateDataSetDataOutputs = d.CreateDataSetDataOutputs[1:]
	return output
}

func (d *DataSession) ActivateDataSetData(ctx context.Context, dataSet *upload.Upload) error {
	d.ActivateDataSetDataInvocations++

	d.ActivateDataSetDataInputs = append(d.ActivateDataSetDataInputs, ActivateDataSetDataInput{Context: ctx, DataSet: dataSet})

	gomega.Expect(d.ActivateDataSetDataOutputs).ToNot(gomega.BeEmpty())

	output := d.ActivateDataSetDataOutputs[0]
	d.ActivateDataSetDataOutputs = d.ActivateDataSetDataOutputs[1:]
	return output
}

func (d *DataSession) ArchiveDeviceDataUsingHashesFromDataSet(ctx context.Context, dataSet *upload.Upload) error {
	d.ArchiveDeviceDataUsingHashesFromDataSetInvocations++

	d.ArchiveDeviceDataUsingHashesFromDataSetInputs = append(d.ArchiveDeviceDataUsingHashesFromDataSetInputs, ArchiveDeviceDataUsingHashesFromDataSetInput{Context: ctx, DataSet: dataSet})

	gomega.Expect(d.ArchiveDeviceDataUsingHashesFromDataSetOutputs).ToNot(gomega.BeEmpty())

	output := d.ArchiveDeviceDataUsingHashesFromDataSetOutputs[0]
	d.ArchiveDeviceDataUsingHashesFromDataSetOutputs = d.ArchiveDeviceDataUsingHashesFromDataSetOutputs[1:]
	return output
}

func (d *DataSession) UnarchiveDeviceDataUsingHashesFromDataSet(ctx context.Context, dataSet *upload.Upload) error {
	d.UnarchiveDeviceDataUsingHashesFromDataSetInvocations++

	d.UnarchiveDeviceDataUsingHashesFromDataSetInputs = append(d.UnarchiveDeviceDataUsingHashesFromDataSetInputs, UnarchiveDeviceDataUsingHashesFromDataSetInput{Context: ctx, DataSet: dataSet})

	gomega.Expect(d.UnarchiveDeviceDataUsingHashesFromDataSetOutputs).ToNot(gomega.BeEmpty())

	output := d.UnarchiveDeviceDataUsingHashesFromDataSetOutputs[0]
	d.UnarchiveDeviceDataUsingHashesFromDataSetOutputs = d.UnarchiveDeviceDataUsingHashesFromDataSetOutputs[1:]
	return output
}

func (d *DataSession) DeleteOtherDataSetData(ctx context.Context, dataSet *upload.Upload) error {
	d.DeleteOtherDataSetDataInvocations++

	d.DeleteOtherDataSetDataInputs = append(d.DeleteOtherDataSetDataInputs, DeleteOtherDataSetDataInput{Context: ctx, DataSet: dataSet})

	gomega.Expect(d.DeleteOtherDataSetDataOutputs).ToNot(gomega.BeEmpty())

	output := d.DeleteOtherDataSetDataOutputs[0]
	d.DeleteOtherDataSetDataOutputs = d.DeleteOtherDataSetDataOutputs[1:]
	return output
}

func (d *DataSession) DestroyDataForUserByID(ctx context.Context, userID string) error {
	d.DestroyDataForUserByIDInvocations++

	d.DestroyDataForUserByIDInputs = append(d.DestroyDataForUserByIDInputs, DestroyDataForUserByIDInput{Context: ctx, UserID: userID})

	gomega.Expect(d.DestroyDataForUserByIDOutputs).ToNot(gomega.BeEmpty())

	output := d.DestroyDataForUserByIDOutputs[0]
	d.DestroyDataForUserByIDOutputs = d.DestroyDataForUserByIDOutputs[1:]
	return output
}

func (d *DataSession) ListUserDataSets(ctx context.Context, userID string, filter *data.DataSetFilter, pagination *page.Pagination) (data.DataSets, error) {
	d.ListUserDataSetsInvocations++

	d.ListUserDataSetsInputs = append(d.ListUserDataSetsInputs, ListUserDataSetsInput{Context: ctx, UserID: userID, Filter: filter, Pagination: pagination})

	gomega.Expect(d.ListUserDataSetsOutputs).ToNot(gomega.BeEmpty())

	output := d.ListUserDataSetsOutputs[0]
	d.ListUserDataSetsOutputs = d.ListUserDataSetsOutputs[1:]
	return output.DataSets, output.Error
}

func (d *DataSession) GetDataSet(ctx context.Context, id string) (*data.DataSet, error) {
	d.GetDataSetInvocations++

	d.GetDataSetInputs = append(d.GetDataSetInputs, GetDataSetInput{Context: ctx, ID: id})

	gomega.Expect(d.GetDataSetOutputs).ToNot(gomega.BeEmpty())

	output := d.GetDataSetOutputs[0]
	d.GetDataSetOutputs = d.GetDataSetOutputs[1:]
	return output.DataSet, output.Error
}

func (d *DataSession) Expectations() {
	d.Mock.Expectations()
	d.Closer.AssertOutputsEmpty()
	gomega.Expect(d.IsClosedOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.EnsureIndexesOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.GetDataSetsForUserByIDOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.GetDataSetByIDOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.CreateDataSetOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.UpdateDataSetOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.DeleteDataSetOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.CreateDataSetDataOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.ActivateDataSetDataOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.ArchiveDeviceDataUsingHashesFromDataSetOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.UnarchiveDeviceDataUsingHashesFromDataSetOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.DeleteOtherDataSetDataOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.DestroyDataForUserByIDOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.ListUserDataSetsOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.GetDataSetOutputs).To(gomega.BeEmpty())
}
