package test

import (
	"github.com/tidepool-org/platform/data"
	dataStore "github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/store"
	"github.com/tidepool-org/platform/test"
)

type GetDatasetsForUserByIDInput struct {
	UserID     string
	Filter     *dataStore.Filter
	Pagination *dataStore.Pagination
}

type GetDatasetsForUserByIDOutput struct {
	Datasets []*upload.Upload
	Error    error
}

type GetDatasetByIDOutput struct {
	Dataset *upload.Upload
	Error   error
}

type CreateDatasetDataInput struct {
	Dataset     *upload.Upload
	DatasetData []data.Datum
}

type DataSession struct {
	*test.Mock
	IsClosedInvocations                                  int
	IsClosedOutputs                                      []bool
	CloseInvocations                                     int
	LoggerInvocations                                    int
	LoggerImpl                                           log.Logger
	SetAgentInvocations                                  int
	SetAgentInputs                                       []store.Agent
	GetDatasetsForUserByIDInvocations                    int
	GetDatasetsForUserByIDInputs                         []GetDatasetsForUserByIDInput
	GetDatasetsForUserByIDOutputs                        []GetDatasetsForUserByIDOutput
	GetDatasetByIDInvocations                            int
	GetDatasetByIDInputs                                 []string
	GetDatasetByIDOutputs                                []GetDatasetByIDOutput
	CreateDatasetInvocations                             int
	CreateDatasetInputs                                  []*upload.Upload
	CreateDatasetOutputs                                 []error
	UpdateDatasetInvocations                             int
	UpdateDatasetInputs                                  []*upload.Upload
	UpdateDatasetOutputs                                 []error
	DeleteDatasetInvocations                             int
	DeleteDatasetInputs                                  []*upload.Upload
	DeleteDatasetOutputs                                 []error
	CreateDatasetDataInvocations                         int
	CreateDatasetDataInputs                              []CreateDatasetDataInput
	CreateDatasetDataOutputs                             []error
	ActivateDatasetDataInvocations                       int
	ActivateDatasetDataInputs                            []*upload.Upload
	ActivateDatasetDataOutputs                           []error
	ArchiveDeviceDataUsingHashesFromDatasetInvocations   int
	ArchiveDeviceDataUsingHashesFromDatasetInputs        []*upload.Upload
	ArchiveDeviceDataUsingHashesFromDatasetOutputs       []error
	UnarchiveDeviceDataUsingHashesFromDatasetInvocations int
	UnarchiveDeviceDataUsingHashesFromDatasetInputs      []*upload.Upload
	UnarchiveDeviceDataUsingHashesFromDatasetOutputs     []error
	DeleteOtherDatasetDataInvocations                    int
	DeleteOtherDatasetDataInputs                         []*upload.Upload
	DeleteOtherDatasetDataOutputs                        []error
	DestroyDataForUserByIDInvocations                    int
	DestroyDataForUserByIDInputs                         []string
	DestroyDataForUserByIDOutputs                        []error
}

func NewDataSession() *DataSession {
	return &DataSession{
		Mock:       test.NewMock(),
		LoggerImpl: null.NewLogger(),
	}
}

func (d *DataSession) IsClosed() bool {
	d.IsClosedInvocations++

	if len(d.IsClosedOutputs) == 0 {
		panic("Unexpected invocation of IsClosed on DataSession")
	}

	output := d.IsClosedOutputs[0]
	d.IsClosedOutputs = d.IsClosedOutputs[1:]
	return output
}

func (d *DataSession) Close() {
	d.CloseInvocations++
}

func (d *DataSession) Logger() log.Logger {
	d.LoggerInvocations++

	return d.LoggerImpl
}

func (d *DataSession) SetAgent(agent store.Agent) {
	d.SetAgentInvocations++

	d.SetAgentInputs = append(d.SetAgentInputs, agent)
}

func (d *DataSession) GetDatasetsForUserByID(userID string, filter *dataStore.Filter, pagination *dataStore.Pagination) ([]*upload.Upload, error) {
	d.GetDatasetsForUserByIDInvocations++

	d.GetDatasetsForUserByIDInputs = append(d.GetDatasetsForUserByIDInputs, GetDatasetsForUserByIDInput{userID, filter, pagination})

	if len(d.GetDatasetsForUserByIDOutputs) == 0 {
		panic("Unexpected invocation of GetDatasetsForUserByID on DataSession")
	}

	output := d.GetDatasetsForUserByIDOutputs[0]
	d.GetDatasetsForUserByIDOutputs = d.GetDatasetsForUserByIDOutputs[1:]
	return output.Datasets, output.Error
}

func (d *DataSession) GetDatasetByID(datasetID string) (*upload.Upload, error) {
	d.GetDatasetByIDInvocations++

	d.GetDatasetByIDInputs = append(d.GetDatasetByIDInputs, datasetID)

	if len(d.GetDatasetByIDOutputs) == 0 {
		panic("Unexpected invocation of GetDatasetByID on DataSession")
	}

	output := d.GetDatasetByIDOutputs[0]
	d.GetDatasetByIDOutputs = d.GetDatasetByIDOutputs[1:]
	return output.Dataset, output.Error
}

func (d *DataSession) CreateDataset(dataset *upload.Upload) error {
	d.CreateDatasetInvocations++

	d.CreateDatasetInputs = append(d.CreateDatasetInputs, dataset)

	if len(d.CreateDatasetOutputs) == 0 {
		panic("Unexpected invocation of CreateDataset on DataSession")
	}

	output := d.CreateDatasetOutputs[0]
	d.CreateDatasetOutputs = d.CreateDatasetOutputs[1:]
	return output
}

func (d *DataSession) UpdateDataset(dataset *upload.Upload) error {
	d.UpdateDatasetInvocations++

	d.UpdateDatasetInputs = append(d.UpdateDatasetInputs, dataset)

	if len(d.UpdateDatasetOutputs) == 0 {
		panic("Unexpected invocation of UpdateDataset on DataSession")
	}

	output := d.UpdateDatasetOutputs[0]
	d.UpdateDatasetOutputs = d.UpdateDatasetOutputs[1:]
	return output
}

func (d *DataSession) DeleteDataset(dataset *upload.Upload) error {
	d.DeleteDatasetInvocations++

	d.DeleteDatasetInputs = append(d.DeleteDatasetInputs, dataset)

	if len(d.DeleteDatasetOutputs) == 0 {
		panic("Unexpected invocation of DeleteDataset on DataSession")
	}

	output := d.DeleteDatasetOutputs[0]
	d.DeleteDatasetOutputs = d.DeleteDatasetOutputs[1:]
	return output
}

func (d *DataSession) CreateDatasetData(dataset *upload.Upload, datasetData []data.Datum) error {
	d.CreateDatasetDataInvocations++

	d.CreateDatasetDataInputs = append(d.CreateDatasetDataInputs, CreateDatasetDataInput{dataset, datasetData})

	if len(d.CreateDatasetDataOutputs) == 0 {
		panic("Unexpected invocation of CreateDatasetData on DataSession")
	}

	output := d.CreateDatasetDataOutputs[0]
	d.CreateDatasetDataOutputs = d.CreateDatasetDataOutputs[1:]
	return output
}

func (d *DataSession) ActivateDatasetData(dataset *upload.Upload) error {
	d.ActivateDatasetDataInvocations++

	d.ActivateDatasetDataInputs = append(d.ActivateDatasetDataInputs, dataset)

	if len(d.ActivateDatasetDataOutputs) == 0 {
		panic("Unexpected invocation of ActivateDatasetData on DataSession")
	}

	output := d.ActivateDatasetDataOutputs[0]
	d.ActivateDatasetDataOutputs = d.ActivateDatasetDataOutputs[1:]
	return output
}

func (d *DataSession) ArchiveDeviceDataUsingHashesFromDataset(dataset *upload.Upload) error {
	d.ArchiveDeviceDataUsingHashesFromDatasetInvocations++

	d.ArchiveDeviceDataUsingHashesFromDatasetInputs = append(d.ArchiveDeviceDataUsingHashesFromDatasetInputs, dataset)

	if len(d.ArchiveDeviceDataUsingHashesFromDatasetOutputs) == 0 {
		panic("Unexpected invocation of ArchiveDeviceDataUsingHashesFromDataset on DataSession")
	}

	output := d.ArchiveDeviceDataUsingHashesFromDatasetOutputs[0]
	d.ArchiveDeviceDataUsingHashesFromDatasetOutputs = d.ArchiveDeviceDataUsingHashesFromDatasetOutputs[1:]
	return output
}

func (d *DataSession) UnarchiveDeviceDataUsingHashesFromDataset(dataset *upload.Upload) error {
	d.UnarchiveDeviceDataUsingHashesFromDatasetInvocations++

	d.UnarchiveDeviceDataUsingHashesFromDatasetInputs = append(d.UnarchiveDeviceDataUsingHashesFromDatasetInputs, dataset)

	if len(d.UnarchiveDeviceDataUsingHashesFromDatasetOutputs) == 0 {
		panic("Unexpected invocation of UnarchiveDeviceDataUsingHashesFromDataset on DataSession")
	}

	output := d.UnarchiveDeviceDataUsingHashesFromDatasetOutputs[0]
	d.UnarchiveDeviceDataUsingHashesFromDatasetOutputs = d.UnarchiveDeviceDataUsingHashesFromDatasetOutputs[1:]
	return output
}

func (d *DataSession) DeleteOtherDatasetData(dataset *upload.Upload) error {
	d.DeleteOtherDatasetDataInvocations++

	d.DeleteOtherDatasetDataInputs = append(d.DeleteOtherDatasetDataInputs, dataset)

	if len(d.DeleteOtherDatasetDataOutputs) == 0 {
		panic("Unexpected invocation of DeleteOtherDatasetData on DataSession")
	}

	output := d.DeleteOtherDatasetDataOutputs[0]
	d.DeleteOtherDatasetDataOutputs = d.DeleteOtherDatasetDataOutputs[1:]
	return output
}

func (d *DataSession) DestroyDataForUserByID(userID string) error {
	d.DestroyDataForUserByIDInvocations++

	d.DestroyDataForUserByIDInputs = append(d.DestroyDataForUserByIDInputs, userID)

	if len(d.DestroyDataForUserByIDOutputs) == 0 {
		panic("Unexpected invocation of DestroyDataForUserByID on DataSession")
	}

	output := d.DestroyDataForUserByIDOutputs[0]
	d.DestroyDataForUserByIDOutputs = d.DestroyDataForUserByIDOutputs[1:]
	return output
}

func (d *DataSession) UnusedOutputsCount() int {
	return len(d.IsClosedOutputs) +
		len(d.GetDatasetsForUserByIDOutputs) +
		len(d.GetDatasetByIDOutputs) +
		len(d.CreateDatasetOutputs) +
		len(d.UpdateDatasetOutputs) +
		len(d.DeleteDatasetOutputs) +
		len(d.CreateDatasetDataOutputs) +
		len(d.ActivateDatasetDataOutputs) +
		len(d.ArchiveDeviceDataUsingHashesFromDatasetOutputs) +
		len(d.UnarchiveDeviceDataUsingHashesFromDatasetOutputs) +
		len(d.DeleteOtherDatasetDataOutputs) +
		len(d.DestroyDataForUserByIDOutputs)
}
