package test

import dataSourceStoreStructured "github.com/tidepool-org/platform/data/source/store/structured"

type Store struct {
	NewDataSourcesInvocations int
	NewDataSourcesStub        func() dataSourceStoreStructured.DataSourcesRepository
	NewDataSourcesOutputs     []dataSourceStoreStructured.DataSourcesRepository
	NewDataSourcesOutput      *dataSourceStoreStructured.DataSourcesRepository
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) NewDataSourcesRepository() dataSourceStoreStructured.DataSourcesRepository {
	s.NewDataSourcesInvocations++
	if s.NewDataSourcesStub != nil {
		return s.NewDataSourcesStub()
	}
	if len(s.NewDataSourcesOutputs) > 0 {
		output := s.NewDataSourcesOutputs[0]
		s.NewDataSourcesOutputs = s.NewDataSourcesOutputs[1:]
		return output
	}
	if s.NewDataSourcesOutput != nil {
		return *s.NewDataSourcesOutput
	}
	panic("NewDataSourcesRepository has no output")
}

func (s *Store) AssertOutputsEmpty() {
	if len(s.NewDataSourcesOutputs) > 0 {
		panic("NewDataSourcesOutputs is not empty")
	}
}
