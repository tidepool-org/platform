package test

import imageStoreStructured "github.com/tidepool-org/platform/image/store/structured"

type Store struct {
	NewImageRepositoryInvocations int
	NewImageRepositoryStub        func() imageStoreStructured.ImageRepository
	NewImageRepositoryOutputs     []imageStoreStructured.ImageRepository
	NewImageRepositoryOutput      *imageStoreStructured.ImageRepository
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) NewImageRepository() imageStoreStructured.ImageRepository {
	s.NewImageRepositoryInvocations++
	if s.NewImageRepositoryStub != nil {
		return s.NewImageRepositoryStub()
	}
	if len(s.NewImageRepositoryOutputs) > 0 {
		output := s.NewImageRepositoryOutputs[0]
		s.NewImageRepositoryOutputs = s.NewImageRepositoryOutputs[1:]
		return output
	}
	if s.NewImageRepositoryOutput != nil {
		return *s.NewImageRepositoryOutput
	}
	panic("NewImageRepository has no output")
}

func (s *Store) AssertOutputsEmpty() {
	if len(s.NewImageRepositoryOutputs) > 0 {
		panic("NewImageRepositoryOutputs is not empty")
	}
}
