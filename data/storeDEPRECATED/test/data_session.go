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
	Context   context.Context
	DataSet   *upload.Upload
	Selectors *data.Selectors
}

type ArchiveDataSetDataInput struct {
	Context   context.Context
	DataSet   *upload.Upload
	Selectors *data.Selectors
}

type DeleteDataSetDataInput struct {
	Context   context.Context
	DataSet   *upload.Upload
	Selectors *data.Selectors
}

type DestroyDeletedDataSetDataInput struct {
	Context   context.Context
	DataSet   *upload.Upload
	Selectors *data.Selectors
}

type DestroyDataSetDataInput struct {
	Context   context.Context
	DataSet   *upload.Upload
	Selectors *data.Selectors
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
	*test.Closer
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
}

func NewDataSession() *DataSession {
	return &DataSession{
		Closer: test.NewCloser(),
	}
}

// EnsureIndexes required in order to implement the DataSession interface
func (s *DataSession) EnsureIndexes() error {
	return nil
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

func (d *DataSession) ActivateDataSetData(ctx context.Context, dataSet *upload.Upload, selectors *data.Selectors) error {
	d.ActivateDataSetDataInvocations++

	d.ActivateDataSetDataInputs = append(d.ActivateDataSetDataInputs, ActivateDataSetDataInput{Context: ctx, DataSet: dataSet, Selectors: selectors})

	gomega.Expect(d.ActivateDataSetDataOutputs).ToNot(gomega.BeEmpty())

	output := d.ActivateDataSetDataOutputs[0]
	d.ActivateDataSetDataOutputs = d.ActivateDataSetDataOutputs[1:]
	return output
}

func (d *DataSession) ArchiveDataSetData(ctx context.Context, dataSet *upload.Upload, selectors *data.Selectors) error {
	d.ArchiveDataSetDataInvocations++

	d.ArchiveDataSetDataInputs = append(d.ArchiveDataSetDataInputs, ArchiveDataSetDataInput{Context: ctx, DataSet: dataSet, Selectors: selectors})

	gomega.Expect(d.ArchiveDataSetDataOutputs).ToNot(gomega.BeEmpty())

	output := d.ArchiveDataSetDataOutputs[0]
	d.ArchiveDataSetDataOutputs = d.ArchiveDataSetDataOutputs[1:]
	return output
}

func (d *DataSession) DeleteDataSetData(ctx context.Context, dataSet *upload.Upload, selectors *data.Selectors) error {
	d.DeleteDataSetDataInvocations++

	d.DeleteDataSetDataInputs = append(d.DeleteDataSetDataInputs, DeleteDataSetDataInput{Context: ctx, DataSet: dataSet, Selectors: selectors})

	gomega.Expect(d.DeleteDataSetDataOutputs).ToNot(gomega.BeEmpty())

	output := d.DeleteDataSetDataOutputs[0]
	d.DeleteDataSetDataOutputs = d.DeleteDataSetDataOutputs[1:]
	return output
}

func (d *DataSession) DestroyDeletedDataSetData(ctx context.Context, dataSet *upload.Upload, selectors *data.Selectors) error {
	d.DestroyDeletedDataSetDataInvocations++

	d.DestroyDeletedDataSetDataInputs = append(d.DestroyDeletedDataSetDataInputs, DestroyDeletedDataSetDataInput{Context: ctx, DataSet: dataSet, Selectors: selectors})

	gomega.Expect(d.DestroyDeletedDataSetDataOutputs).ToNot(gomega.BeEmpty())

	output := d.DestroyDeletedDataSetDataOutputs[0]
	d.DestroyDeletedDataSetDataOutputs = d.DestroyDeletedDataSetDataOutputs[1:]
	return output
}

func (d *DataSession) DestroyDataSetData(ctx context.Context, dataSet *upload.Upload, selectors *data.Selectors) error {
	d.DestroyDataSetDataInvocations++

	d.DestroyDataSetDataInputs = append(d.DestroyDataSetDataInputs, DestroyDataSetDataInput{Context: ctx, DataSet: dataSet, Selectors: selectors})

	gomega.Expect(d.DestroyDataSetDataOutputs).ToNot(gomega.BeEmpty())

	output := d.DestroyDataSetDataOutputs[0]
	d.DestroyDataSetDataOutputs = d.DestroyDataSetDataOutputs[1:]
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
	d.Closer.AssertOutputsEmpty()
	gomega.Expect(d.GetDataSetsForUserByIDOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.GetDataSetByIDOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.CreateDataSetOutputs).To(gomega.BeEmpty())
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
}
