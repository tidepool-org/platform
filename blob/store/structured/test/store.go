package test

import blobStoreStructured "github.com/tidepool-org/platform/blob/store/structured"

type Store struct {
	NewRepositoryInvocations int
	NewRepositoryStub        func() blobStoreStructured.BlobRepository
	NewRepositoryOutputs     []blobStoreStructured.BlobRepository
	NewRepositoryOutput      *blobStoreStructured.BlobRepository
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

func (s *Store) AssertOutputsEmpty() {
	if len(s.NewRepositoryOutputs) > 0 {
		panic("NewRepositoryOutputs is not empty")
	}
}
