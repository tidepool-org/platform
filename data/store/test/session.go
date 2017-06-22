package test

import (
	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/types/upload"
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

type FindPreviousActiveDatasetForDeviceOutput struct {
	Dataset *upload.Upload
	Error   error
}

type GetDatasetDataDeduplicatorHashesInput struct {
	Dataset *upload.Upload
	Active  bool
}

type GetDatasetDataDeduplicatorHashesOutput struct {
	Hashes []string
	Error  error
}

type FindAllDatasetDataDeduplicatorHashesForDeviceInput struct {
	UserID   string
	DeviceID string
	Hashes   []string
}

type FindAllDatasetDataDeduplicatorHashesForDeviceOutput struct {
	Hashes []string
	Error  error
}

type CreateDatasetDataInput struct {
	Dataset     *upload.Upload
	DatasetData []data.Datum
}

type FindEarliestDatasetDataTimeOutput struct {
	Time  string
	Error error
}

type SetDatasetDataActiveUsingHashesInput struct {
	Dataset *upload.Upload
	Hashes  []string
	Active  bool
}

type SetDeviceDataActiveUsingHashesInput struct {
	Dataset *upload.Upload
	Hashes  []string
	Active  bool
}

type DeactivateOtherDatasetDataAfterTimeInput struct {
	Dataset *upload.Upload
	Time    string
}

type Session struct {
	ID                                                       string
	IsClosedInvocations                                      int
	IsClosedOutputs                                          []bool
	CloseInvocations                                         int
	SetAgentInvocations                                      int
	SetAgentInputs                                           []commonStore.Agent
	GetDatasetsForUserByIDInvocations                        int
	GetDatasetsForUserByIDInputs                             []GetDatasetsForUserByIDInput
	GetDatasetsForUserByIDOutputs                            []GetDatasetsForUserByIDOutput
	GetDatasetByIDInvocations                                int
	GetDatasetByIDInputs                                     []string
	GetDatasetByIDOutputs                                    []GetDatasetByIDOutput
	FindPreviousActiveDatasetForDeviceInvocations            int
	FindPreviousActiveDatasetForDeviceInputs                 []*upload.Upload
	FindPreviousActiveDatasetForDeviceOutputs                []FindPreviousActiveDatasetForDeviceOutput
	CreateDatasetInvocations                                 int
	CreateDatasetInputs                                      []*upload.Upload
	CreateDatasetOutputs                                     []error
	UpdateDatasetInvocations                                 int
	UpdateDatasetInputs                                      []*upload.Upload
	UpdateDatasetOutputs                                     []error
	DeleteDatasetInvocations                                 int
	DeleteDatasetInputs                                      []*upload.Upload
	DeleteDatasetOutputs                                     []error
	GetDatasetDataDeduplicatorHashesInvocations              int
	GetDatasetDataDeduplicatorHashesInputs                   []GetDatasetDataDeduplicatorHashesInput
	GetDatasetDataDeduplicatorHashesOutputs                  []GetDatasetDataDeduplicatorHashesOutput
	FindAllDatasetDataDeduplicatorHashesForDeviceInvocations int
	FindAllDatasetDataDeduplicatorHashesForDeviceInputs      []FindAllDatasetDataDeduplicatorHashesForDeviceInput
	FindAllDatasetDataDeduplicatorHashesForDeviceOutputs     []FindAllDatasetDataDeduplicatorHashesForDeviceOutput
	CreateDatasetDataInvocations                             int
	CreateDatasetDataInputs                                  []CreateDatasetDataInput
	CreateDatasetDataOutputs                                 []error
	FindEarliestDatasetDataTimeInvocations                   int
	FindEarliestDatasetDataTimeInputs                        []*upload.Upload
	FindEarliestDatasetDataTimeOutputs                       []FindEarliestDatasetDataTimeOutput
	ActivateDatasetDataInvocations                           int
	ActivateDatasetDataInputs                                []*upload.Upload
	ActivateDatasetDataOutputs                               []error
	SetDatasetDataActiveUsingHashesInvocations               int
	SetDatasetDataActiveUsingHashesInputs                    []SetDatasetDataActiveUsingHashesInput
	SetDatasetDataActiveUsingHashesOutputs                   []error
	SetDeviceDataActiveUsingHashesInvocations                int
	SetDeviceDataActiveUsingHashesInputs                     []SetDeviceDataActiveUsingHashesInput
	SetDeviceDataActiveUsingHashesOutputs                    []error
	DeactivateOtherDatasetDataAfterTimeInvocations           int
	DeactivateOtherDatasetDataAfterTimeInputs                []DeactivateOtherDatasetDataAfterTimeInput
	DeactivateOtherDatasetDataAfterTimeOutputs               []error
	DeleteOtherDatasetDataInvocations                        int
	DeleteOtherDatasetDataInputs                             []*upload.Upload
	DeleteOtherDatasetDataOutputs                            []error
	DestroyDataForUserByIDInvocations                        int
	DestroyDataForUserByIDInputs                             []string
	DestroyDataForUserByIDOutputs                            []error
}

func NewSession() *Session {
	return &Session{
		ID: app.NewID(),
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

func (s *Session) FindPreviousActiveDatasetForDevice(dataset *upload.Upload) (*upload.Upload, error) {
	s.FindPreviousActiveDatasetForDeviceInvocations++

	s.FindPreviousActiveDatasetForDeviceInputs = append(s.FindPreviousActiveDatasetForDeviceInputs, dataset)

	if len(s.FindPreviousActiveDatasetForDeviceOutputs) == 0 {
		panic("Unexpected invocation of FindPreviousActiveDatasetForDevice on Session")
	}

	output := s.FindPreviousActiveDatasetForDeviceOutputs[0]
	s.FindPreviousActiveDatasetForDeviceOutputs = s.FindPreviousActiveDatasetForDeviceOutputs[1:]
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

func (s *Session) GetDatasetDataDeduplicatorHashes(dataset *upload.Upload, active bool) ([]string, error) {
	s.GetDatasetDataDeduplicatorHashesInvocations++

	s.GetDatasetDataDeduplicatorHashesInputs = append(s.GetDatasetDataDeduplicatorHashesInputs, GetDatasetDataDeduplicatorHashesInput{dataset, active})

	if len(s.GetDatasetDataDeduplicatorHashesOutputs) == 0 {
		panic("Unexpected invocation of GetDatasetDataDeduplicatorHashes on Session")
	}

	output := s.GetDatasetDataDeduplicatorHashesOutputs[0]
	s.GetDatasetDataDeduplicatorHashesOutputs = s.GetDatasetDataDeduplicatorHashesOutputs[1:]
	return output.Hashes, output.Error
}

func (s *Session) FindAllDatasetDataDeduplicatorHashesForDevice(userID string, deviceID string, queryHashes []string) ([]string, error) {
	s.FindAllDatasetDataDeduplicatorHashesForDeviceInvocations++

	s.FindAllDatasetDataDeduplicatorHashesForDeviceInputs = append(s.FindAllDatasetDataDeduplicatorHashesForDeviceInputs, FindAllDatasetDataDeduplicatorHashesForDeviceInput{userID, deviceID, queryHashes})

	if len(s.FindAllDatasetDataDeduplicatorHashesForDeviceOutputs) == 0 {
		panic("Unexpected invocation of FindAllDatasetDataDeduplicatorHashesForDevice on Session")
	}

	output := s.FindAllDatasetDataDeduplicatorHashesForDeviceOutputs[0]
	s.FindAllDatasetDataDeduplicatorHashesForDeviceOutputs = s.FindAllDatasetDataDeduplicatorHashesForDeviceOutputs[1:]
	return output.Hashes, output.Error
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

func (s *Session) FindEarliestDatasetDataTime(dataset *upload.Upload) (string, error) {
	s.FindEarliestDatasetDataTimeInvocations++

	s.FindEarliestDatasetDataTimeInputs = append(s.FindEarliestDatasetDataTimeInputs, dataset)

	if len(s.FindEarliestDatasetDataTimeOutputs) == 0 {
		panic("Unexpected invocation of FindEarliestDatasetDataTime on Session")
	}

	output := s.FindEarliestDatasetDataTimeOutputs[0]
	s.FindEarliestDatasetDataTimeOutputs = s.FindEarliestDatasetDataTimeOutputs[1:]
	return output.Time, output.Error
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

func (s *Session) SetDatasetDataActiveUsingHashes(dataset *upload.Upload, queryHashes []string, active bool) error {
	s.SetDatasetDataActiveUsingHashesInvocations++

	s.SetDatasetDataActiveUsingHashesInputs = append(s.SetDatasetDataActiveUsingHashesInputs, SetDatasetDataActiveUsingHashesInput{dataset, queryHashes, active})

	if len(s.SetDatasetDataActiveUsingHashesOutputs) == 0 {
		panic("Unexpected invocation of SetDatasetDataActiveUsingHashes on Session")
	}

	output := s.SetDatasetDataActiveUsingHashesOutputs[0]
	s.SetDatasetDataActiveUsingHashesOutputs = s.SetDatasetDataActiveUsingHashesOutputs[1:]
	return output
}

func (s *Session) SetDeviceDataActiveUsingHashes(dataset *upload.Upload, queryHashes []string, active bool) error {
	s.SetDeviceDataActiveUsingHashesInvocations++

	s.SetDeviceDataActiveUsingHashesInputs = append(s.SetDeviceDataActiveUsingHashesInputs, SetDeviceDataActiveUsingHashesInput{dataset, queryHashes, active})

	if len(s.SetDeviceDataActiveUsingHashesOutputs) == 0 {
		panic("Unexpected invocation of SetDeviceDataActiveUsingHashes on Session")
	}

	output := s.SetDeviceDataActiveUsingHashesOutputs[0]
	s.SetDeviceDataActiveUsingHashesOutputs = s.SetDeviceDataActiveUsingHashesOutputs[1:]
	return output
}

func (s *Session) DeactivateOtherDatasetDataAfterTime(dataset *upload.Upload, time string) error {
	s.DeactivateOtherDatasetDataAfterTimeInvocations++

	s.DeactivateOtherDatasetDataAfterTimeInputs = append(s.DeactivateOtherDatasetDataAfterTimeInputs, DeactivateOtherDatasetDataAfterTimeInput{dataset, time})

	if len(s.DeactivateOtherDatasetDataAfterTimeOutputs) == 0 {
		panic("Unexpected invocation of DeactivateOtherDatasetDataAfterTime on Session")
	}

	output := s.DeactivateOtherDatasetDataAfterTimeOutputs[0]
	s.DeactivateOtherDatasetDataAfterTimeOutputs = s.DeactivateOtherDatasetDataAfterTimeOutputs[1:]
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
		len(s.FindPreviousActiveDatasetForDeviceOutputs) +
		len(s.CreateDatasetOutputs) +
		len(s.UpdateDatasetOutputs) +
		len(s.DeleteDatasetOutputs) +
		len(s.GetDatasetDataDeduplicatorHashesOutputs) +
		len(s.FindAllDatasetDataDeduplicatorHashesForDeviceOutputs) +
		len(s.CreateDatasetDataOutputs) +
		len(s.FindEarliestDatasetDataTimeOutputs) +
		len(s.ActivateDatasetDataOutputs) +
		len(s.SetDatasetDataActiveUsingHashesOutputs) +
		len(s.DeactivateOtherDatasetDataAfterTimeOutputs) +
		len(s.DeleteOtherDatasetDataOutputs) +
		len(s.DestroyDataForUserByIDOutputs)
}
