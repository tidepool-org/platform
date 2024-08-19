package test

import blobStoreStructured "github.com/tidepool-org/platform/blob/store/structured"

type Store struct {
	NewRepositoryInvocations int
	NewRepositoryStub        func() blobStoreStructured.BlobRepository
	NewRepositoryOutputs     []blobStoreStructured.BlobRepository
	NewRepositoryOutput      *blobStoreStructured.BlobRepository

	NewDeviceLogsRepositoryInvocations int
	NewDeviceLogsRepositoryStub        func() blobStoreStructured.DeviceLogsRepository
	NewDeviceLogsRepositoryOutputs     []blobStoreStructured.DeviceLogsRepository
	NewDeviceLogsRepositoryOutput      *blobStoreStructured.DeviceLogsRepository
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) NewBlobRepository() blobStoreStructured.BlobRepository {
	s.NewRepositoryInvocations++
	if s.NewRepositoryStub != nil {
		return s.NewRepositoryStub()
	}
	if len(s.NewRepositoryOutputs) > 0 {
		output := s.NewRepositoryOutputs[0]
		s.NewRepositoryOutputs = s.NewRepositoryOutputs[1:]
		return output
	}
	if s.NewRepositoryOutput != nil {
		return *s.NewRepositoryOutput
	}
	panic("NewBlobRepository has no output")
}

func (s *Store) NewDeviceLogsRepository() blobStoreStructured.DeviceLogsRepository {
	s.NewDeviceLogsRepositoryInvocations++
	if s.NewDeviceLogsRepositoryStub != nil {
		return s.NewDeviceLogsRepositoryStub()
	}
	if len(s.NewDeviceLogsRepositoryOutputs) > 0 {
		output := s.NewDeviceLogsRepositoryOutputs[0]
		s.NewDeviceLogsRepositoryOutputs = s.NewDeviceLogsRepositoryOutputs[1:]
		return output
	}
	if s.NewDeviceLogsRepositoryOutput != nil {
		return *s.NewDeviceLogsRepositoryOutput
	}
	panic("NewDeviceLogsRepository has no output")
}

func (s *Store) AssertOutputsEmpty() {
	if len(s.NewRepositoryOutputs) > 0 {
		panic("NewRepositoryOutputs is not empty")
	}
	if len(s.NewDeviceLogsRepositoryOutputs) > 0 {
		panic("NewDeviceLogsRepositoryOutputs is not empty")
	}
}
