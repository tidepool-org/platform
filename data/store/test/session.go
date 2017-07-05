package test

import (
	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/log"
	commonStore "github.com/tidepool-org/platform/store"
)

type GetDatasetsForUserByIDInput struct {
	UserID     string
	Filter     *store.Filter
	Pagination *store.Pagination
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

type Session struct {
	ID                                                   string
	IsClosedInvocations                                  int
	IsClosedOutputs                                      []bool
	CloseInvocations                                     int
	LoggerInvocations                                    int
	LoggerImpl                                           log.Logger
	SetAgentInvocations                                  int
	SetAgentInputs                                       []commonStore.Agent
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

func NewSession() *Session {
	return &Session{
		ID:         app.NewID(),
		LoggerImpl: log.NewNull(),
	}
}

func (s *Session) IsClosed() bool {
	s.IsClosedInvocations++

	if len(s.IsClosedOutputs) == 0 {
		panic("Unexpected invocation of IsClosed on Session")
	}

	output := s.IsClosedOutputs[0]
	s.IsClosedOutputs = s.IsClosedOutputs[1:]
	return output
}

func (s *Session) Close() {
	s.CloseInvocations++
}

func (s *Session) Logger() log.Logger {
	s.LoggerInvocations++

	return s.LoggerImpl
}

func (s *Session) SetAgent(agent commonStore.Agent) {
	s.SetAgentInvocations++

	s.SetAgentInputs = append(s.SetAgentInputs, agent)
}

func (s *Session) GetDatasetsForUserByID(userID string, filter *store.Filter, pagination *store.Pagination) ([]*upload.Upload, error) {
	s.GetDatasetsForUserByIDInvocations++

	s.GetDatasetsForUserByIDInputs = append(s.GetDatasetsForUserByIDInputs, GetDatasetsForUserByIDInput{userID, filter, pagination})

	if len(s.GetDatasetsForUserByIDOutputs) == 0 {
		panic("Unexpected invocation of GetDatasetsForUserByID on Session")
	}

	output := s.GetDatasetsForUserByIDOutputs[0]
	s.GetDatasetsForUserByIDOutputs = s.GetDatasetsForUserByIDOutputs[1:]
	return output.Datasets, output.Error
}

func (s *Session) GetDatasetByID(datasetID string) (*upload.Upload, error) {
	s.GetDatasetByIDInvocations++

	s.GetDatasetByIDInputs = append(s.GetDatasetByIDInputs, datasetID)

	if len(s.GetDatasetByIDOutputs) == 0 {
		panic("Unexpected invocation of GetDatasetByID on Session")
	}

	output := s.GetDatasetByIDOutputs[0]
	s.GetDatasetByIDOutputs = s.GetDatasetByIDOutputs[1:]
	return output.Dataset, output.Error
}

func (s *Session) CreateDataset(dataset *upload.Upload) error {
	s.CreateDatasetInvocations++

	s.CreateDatasetInputs = append(s.CreateDatasetInputs, dataset)

	if len(s.CreateDatasetOutputs) == 0 {
		panic("Unexpected invocation of CreateDataset on Session")
	}

	output := s.CreateDatasetOutputs[0]
	s.CreateDatasetOutputs = s.CreateDatasetOutputs[1:]
	return output
}

func (s *Session) UpdateDataset(dataset *upload.Upload) error {
	s.UpdateDatasetInvocations++

	s.UpdateDatasetInputs = append(s.UpdateDatasetInputs, dataset)

	if len(s.UpdateDatasetOutputs) == 0 {
		panic("Unexpected invocation of UpdateDataset on Session")
	}

	output := s.UpdateDatasetOutputs[0]
	s.UpdateDatasetOutputs = s.UpdateDatasetOutputs[1:]
	return output
}

func (s *Session) DeleteDataset(dataset *upload.Upload) error {
	s.DeleteDatasetInvocations++

	s.DeleteDatasetInputs = append(s.DeleteDatasetInputs, dataset)

	if len(s.DeleteDatasetOutputs) == 0 {
		panic("Unexpected invocation of DeleteDataset on Session")
	}

	output := s.DeleteDatasetOutputs[0]
	s.DeleteDatasetOutputs = s.DeleteDatasetOutputs[1:]
	return output
}

func (s *Session) CreateDatasetData(dataset *upload.Upload, datasetData []data.Datum) error {
	s.CreateDatasetDataInvocations++

	s.CreateDatasetDataInputs = append(s.CreateDatasetDataInputs, CreateDatasetDataInput{dataset, datasetData})

	if len(s.CreateDatasetDataOutputs) == 0 {
		panic("Unexpected invocation of CreateDatasetData on Session")
	}

	output := s.CreateDatasetDataOutputs[0]
	s.CreateDatasetDataOutputs = s.CreateDatasetDataOutputs[1:]
	return output
}

func (s *Session) ActivateDatasetData(dataset *upload.Upload) error {
	s.ActivateDatasetDataInvocations++

	s.ActivateDatasetDataInputs = append(s.ActivateDatasetDataInputs, dataset)

	if len(s.ActivateDatasetDataOutputs) == 0 {
		panic("Unexpected invocation of ActivateDatasetData on Session")
	}

	output := s.ActivateDatasetDataOutputs[0]
	s.ActivateDatasetDataOutputs = s.ActivateDatasetDataOutputs[1:]
	return output
}

func (s *Session) ArchiveDeviceDataUsingHashesFromDataset(dataset *upload.Upload) error {
	s.ArchiveDeviceDataUsingHashesFromDatasetInvocations++

	s.ArchiveDeviceDataUsingHashesFromDatasetInputs = append(s.ArchiveDeviceDataUsingHashesFromDatasetInputs, dataset)

	if len(s.ArchiveDeviceDataUsingHashesFromDatasetOutputs) == 0 {
		panic("Unexpected invocation of ArchiveDeviceDataUsingHashesFromDataset on Session")
	}

	output := s.ArchiveDeviceDataUsingHashesFromDatasetOutputs[0]
	s.ArchiveDeviceDataUsingHashesFromDatasetOutputs = s.ArchiveDeviceDataUsingHashesFromDatasetOutputs[1:]
	return output
}

func (s *Session) UnarchiveDeviceDataUsingHashesFromDataset(dataset *upload.Upload) error {
	s.UnarchiveDeviceDataUsingHashesFromDatasetInvocations++

	s.UnarchiveDeviceDataUsingHashesFromDatasetInputs = append(s.UnarchiveDeviceDataUsingHashesFromDatasetInputs, dataset)

	if len(s.UnarchiveDeviceDataUsingHashesFromDatasetOutputs) == 0 {
		panic("Unexpected invocation of UnarchiveDeviceDataUsingHashesFromDataset on Session")
	}

	output := s.UnarchiveDeviceDataUsingHashesFromDatasetOutputs[0]
	s.UnarchiveDeviceDataUsingHashesFromDatasetOutputs = s.UnarchiveDeviceDataUsingHashesFromDatasetOutputs[1:]
	return output
}

func (s *Session) DeleteOtherDatasetData(dataset *upload.Upload) error {
	s.DeleteOtherDatasetDataInvocations++

	s.DeleteOtherDatasetDataInputs = append(s.DeleteOtherDatasetDataInputs, dataset)

	if len(s.DeleteOtherDatasetDataOutputs) == 0 {
		panic("Unexpected invocation of DeleteOtherDatasetData on Session")
	}

	output := s.DeleteOtherDatasetDataOutputs[0]
	s.DeleteOtherDatasetDataOutputs = s.DeleteOtherDatasetDataOutputs[1:]
	return output
}

func (s *Session) DestroyDataForUserByID(userID string) error {
	s.DestroyDataForUserByIDInvocations++

	s.DestroyDataForUserByIDInputs = append(s.DestroyDataForUserByIDInputs, userID)

	if len(s.DestroyDataForUserByIDOutputs) == 0 {
		panic("Unexpected invocation of DestroyDataForUserByID on Session")
	}

	output := s.DestroyDataForUserByIDOutputs[0]
	s.DestroyDataForUserByIDOutputs = s.DestroyDataForUserByIDOutputs[1:]
	return output
}

func (s *Session) UnusedOutputsCount() int {
	return len(s.IsClosedOutputs) +
		len(s.GetDatasetsForUserByIDOutputs) +
		len(s.GetDatasetByIDOutputs) +
		len(s.CreateDatasetOutputs) +
		len(s.UpdateDatasetOutputs) +
		len(s.DeleteDatasetOutputs) +
		len(s.CreateDatasetDataOutputs) +
		len(s.ActivateDatasetDataOutputs) +
		len(s.ArchiveDeviceDataUsingHashesFromDatasetOutputs) +
		len(s.UnarchiveDeviceDataUsingHashesFromDatasetOutputs) +
		len(s.DeleteOtherDatasetDataOutputs) +
		len(s.DestroyDataForUserByIDOutputs)
}
