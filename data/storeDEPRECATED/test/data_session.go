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

type GetDatasetsForUserByIDInput struct {
	Context    context.Context
	UserID     string
	Filter     *dataStoreDEPRECATED.Filter
	Pagination *page.Pagination
}

type GetDatasetsForUserByIDOutput struct {
	Datasets []*upload.Upload
	Error    error
}

type GetDatasetByIDInput struct {
	Context   context.Context
	DatasetID string
}

type GetDatasetByIDOutput struct {
	Dataset *upload.Upload
	Error   error
}

type CreateDatasetInput struct {
	Context context.Context
	Dataset *upload.Upload
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

type DeleteDatasetInput struct {
	Context context.Context
	Dataset *upload.Upload
}

type CreateDatasetDataInput struct {
	Context     context.Context
	Dataset     *upload.Upload
	DatasetData []data.Datum
}

type ActivateDatasetDataInput struct {
	Context context.Context
	Dataset *upload.Upload
}

type ArchiveDeviceDataUsingHashesFromDatasetInput struct {
	Context context.Context
	Dataset *upload.Upload
}

type UnarchiveDeviceDataUsingHashesFromDatasetInput struct {
	Context context.Context
	Dataset *upload.Upload
}

type DeleteOtherDatasetDataInput struct {
	Context context.Context
	Dataset *upload.Upload
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
	GetDatasetsForUserByIDInvocations                    int
	GetDatasetsForUserByIDInputs                         []GetDatasetsForUserByIDInput
	GetDatasetsForUserByIDOutputs                        []GetDatasetsForUserByIDOutput
	GetDatasetByIDInvocations                            int
	GetDatasetByIDInputs                                 []GetDatasetByIDInput
	GetDatasetByIDOutputs                                []GetDatasetByIDOutput
	CreateDatasetInvocations                             int
	CreateDatasetInputs                                  []CreateDatasetInput
	CreateDatasetOutputs                                 []error
	UpdateDataSetInvocations                             int
	UpdateDataSetInputs                                  []UpdateDataSetInput
	UpdateDataSetOutputs                                 []UpdateDataSetOutput
	DeleteDatasetInvocations                             int
	DeleteDatasetInputs                                  []DeleteDatasetInput
	DeleteDatasetOutputs                                 []error
	CreateDatasetDataInvocations                         int
	CreateDatasetDataInputs                              []CreateDatasetDataInput
	CreateDatasetDataOutputs                             []error
	ActivateDatasetDataInvocations                       int
	ActivateDatasetDataInputs                            []ActivateDatasetDataInput
	ActivateDatasetDataOutputs                           []error
	ArchiveDeviceDataUsingHashesFromDatasetInvocations   int
	ArchiveDeviceDataUsingHashesFromDatasetInputs        []ArchiveDeviceDataUsingHashesFromDatasetInput
	ArchiveDeviceDataUsingHashesFromDatasetOutputs       []error
	UnarchiveDeviceDataUsingHashesFromDatasetInvocations int
	UnarchiveDeviceDataUsingHashesFromDatasetInputs      []UnarchiveDeviceDataUsingHashesFromDatasetInput
	UnarchiveDeviceDataUsingHashesFromDatasetOutputs     []error
	DeleteOtherDatasetDataInvocations                    int
	DeleteOtherDatasetDataInputs                         []DeleteOtherDatasetDataInput
	DeleteOtherDatasetDataOutputs                        []error
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

func (d *DataSession) GetDatasetsForUserByID(ctx context.Context, userID string, filter *dataStoreDEPRECATED.Filter, pagination *page.Pagination) ([]*upload.Upload, error) {
	d.GetDatasetsForUserByIDInvocations++

	d.GetDatasetsForUserByIDInputs = append(d.GetDatasetsForUserByIDInputs, GetDatasetsForUserByIDInput{Context: ctx, UserID: userID, Filter: filter, Pagination: pagination})

	gomega.Expect(d.GetDatasetsForUserByIDOutputs).ToNot(gomega.BeEmpty())

	output := d.GetDatasetsForUserByIDOutputs[0]
	d.GetDatasetsForUserByIDOutputs = d.GetDatasetsForUserByIDOutputs[1:]
	return output.Datasets, output.Error
}

func (d *DataSession) GetDatasetByID(ctx context.Context, datasetID string) (*upload.Upload, error) {
	d.GetDatasetByIDInvocations++

	d.GetDatasetByIDInputs = append(d.GetDatasetByIDInputs, GetDatasetByIDInput{Context: ctx, DatasetID: datasetID})

	gomega.Expect(d.GetDatasetByIDOutputs).ToNot(gomega.BeEmpty())

	output := d.GetDatasetByIDOutputs[0]
	d.GetDatasetByIDOutputs = d.GetDatasetByIDOutputs[1:]
	return output.Dataset, output.Error
}

func (d *DataSession) CreateDataset(ctx context.Context, dataset *upload.Upload) error {
	d.CreateDatasetInvocations++

	d.CreateDatasetInputs = append(d.CreateDatasetInputs, CreateDatasetInput{Context: ctx, Dataset: dataset})

	gomega.Expect(d.CreateDatasetOutputs).ToNot(gomega.BeEmpty())

	output := d.CreateDatasetOutputs[0]
	d.CreateDatasetOutputs = d.CreateDatasetOutputs[1:]
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

func (d *DataSession) DeleteDataset(ctx context.Context, dataset *upload.Upload) error {
	d.DeleteDatasetInvocations++

	d.DeleteDatasetInputs = append(d.DeleteDatasetInputs, DeleteDatasetInput{Context: ctx, Dataset: dataset})

	gomega.Expect(d.DeleteDatasetOutputs).ToNot(gomega.BeEmpty())

	output := d.DeleteDatasetOutputs[0]
	d.DeleteDatasetOutputs = d.DeleteDatasetOutputs[1:]
	return output
}

func (d *DataSession) CreateDatasetData(ctx context.Context, dataset *upload.Upload, datasetData []data.Datum) error {
	d.CreateDatasetDataInvocations++

	d.CreateDatasetDataInputs = append(d.CreateDatasetDataInputs, CreateDatasetDataInput{Context: ctx, Dataset: dataset, DatasetData: datasetData})

	gomega.Expect(d.CreateDatasetDataOutputs).ToNot(gomega.BeEmpty())

	output := d.CreateDatasetDataOutputs[0]
	d.CreateDatasetDataOutputs = d.CreateDatasetDataOutputs[1:]
	return output
}

func (d *DataSession) ActivateDatasetData(ctx context.Context, dataset *upload.Upload) error {
	d.ActivateDatasetDataInvocations++

	d.ActivateDatasetDataInputs = append(d.ActivateDatasetDataInputs, ActivateDatasetDataInput{Context: ctx, Dataset: dataset})

	gomega.Expect(d.ActivateDatasetDataOutputs).ToNot(gomega.BeEmpty())

	output := d.ActivateDatasetDataOutputs[0]
	d.ActivateDatasetDataOutputs = d.ActivateDatasetDataOutputs[1:]
	return output
}

func (d *DataSession) ArchiveDeviceDataUsingHashesFromDataset(ctx context.Context, dataset *upload.Upload) error {
	d.ArchiveDeviceDataUsingHashesFromDatasetInvocations++

	d.ArchiveDeviceDataUsingHashesFromDatasetInputs = append(d.ArchiveDeviceDataUsingHashesFromDatasetInputs, ArchiveDeviceDataUsingHashesFromDatasetInput{Context: ctx, Dataset: dataset})

	gomega.Expect(d.ArchiveDeviceDataUsingHashesFromDatasetOutputs).ToNot(gomega.BeEmpty())

	output := d.ArchiveDeviceDataUsingHashesFromDatasetOutputs[0]
	d.ArchiveDeviceDataUsingHashesFromDatasetOutputs = d.ArchiveDeviceDataUsingHashesFromDatasetOutputs[1:]
	return output
}

func (d *DataSession) UnarchiveDeviceDataUsingHashesFromDataset(ctx context.Context, dataset *upload.Upload) error {
	d.UnarchiveDeviceDataUsingHashesFromDatasetInvocations++

	d.UnarchiveDeviceDataUsingHashesFromDatasetInputs = append(d.UnarchiveDeviceDataUsingHashesFromDatasetInputs, UnarchiveDeviceDataUsingHashesFromDatasetInput{Context: ctx, Dataset: dataset})

	gomega.Expect(d.UnarchiveDeviceDataUsingHashesFromDatasetOutputs).ToNot(gomega.BeEmpty())

	output := d.UnarchiveDeviceDataUsingHashesFromDatasetOutputs[0]
	d.UnarchiveDeviceDataUsingHashesFromDatasetOutputs = d.UnarchiveDeviceDataUsingHashesFromDatasetOutputs[1:]
	return output
}

func (d *DataSession) DeleteOtherDatasetData(ctx context.Context, dataset *upload.Upload) error {
	d.DeleteOtherDatasetDataInvocations++

	d.DeleteOtherDatasetDataInputs = append(d.DeleteOtherDatasetDataInputs, DeleteOtherDatasetDataInput{Context: ctx, Dataset: dataset})

	gomega.Expect(d.DeleteOtherDatasetDataOutputs).ToNot(gomega.BeEmpty())

	output := d.DeleteOtherDatasetDataOutputs[0]
	d.DeleteOtherDatasetDataOutputs = d.DeleteOtherDatasetDataOutputs[1:]
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
	gomega.Expect(d.GetDatasetsForUserByIDOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.GetDatasetByIDOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.CreateDatasetOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.UpdateDataSetOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.DeleteDatasetOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.CreateDatasetDataOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.ActivateDatasetDataOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.ArchiveDeviceDataUsingHashesFromDatasetOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.UnarchiveDeviceDataUsingHashesFromDatasetOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.DeleteOtherDatasetDataOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.DestroyDataForUserByIDOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.ListUserDataSetsOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.GetDataSetOutputs).To(gomega.BeEmpty())
}
