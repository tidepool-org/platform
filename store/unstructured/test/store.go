package test

import (
	"context"
	"io"

	storeUnstructured "github.com/tidepool-org/platform/store/unstructured"
)

type ExistsOutput struct {
	Exists bool
	Error  error
}

type PutInput struct {
	Key     string
	Reader  io.Reader
	Options *storeUnstructured.Options
}

type GetOutput struct {
	Reader io.ReadCloser
	Error  error
}

type DeleteOutput struct {
	Deleted bool
	Error   error
}

type Store struct {
	ExistsInvocations          int
	ExistsInputs               []string
	ExistsStub                 func(ctx context.Context, key string) (bool, error)
	ExistsOutputs              []ExistsOutput
	ExistsOutput               *ExistsOutput
	PutInvocations             int
	PutInputs                  []PutInput
	PutStub                    func(ctx context.Context, key string, reader io.Reader, options *storeUnstructured.Options) error
	PutOutputs                 []error
	PutOutput                  *error
	GetInvocations             int
	GetInputs                  []string
	GetStub                    func(ctx context.Context, key string) (io.ReadCloser, error)
	GetOutputs                 []GetOutput
	GetOutput                  *GetOutput
	DeleteInvocations          int
	DeleteInputs               []string
	DeleteStub                 func(ctx context.Context, key string) (bool, error)
	DeleteOutputs              []DeleteOutput
	DeleteOutput               *DeleteOutput
	DeleteDirectoryInvocations int
	DeleteDirectoryInputs      []string
	DeleteDirectoryStub        func(ctx context.Context, key string) error
	DeleteDirectoryOutputs     []error
	DeleteDirectoryOutput      *error
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) Exists(ctx context.Context, key string) (bool, error) {
	s.ExistsInvocations++
	s.ExistsInputs = append(s.ExistsInputs, key)
	if s.ExistsStub != nil {
		return s.ExistsStub(ctx, key)
	}
	if len(s.ExistsOutputs) > 0 {
		output := s.ExistsOutputs[0]
		s.ExistsOutputs = s.ExistsOutputs[1:]
		return output.Exists, output.Error
	}
	if s.ExistsOutput != nil {
		return s.ExistsOutput.Exists, s.ExistsOutput.Error
	}
	panic("Exists has no output")
}

func (s *Store) Put(ctx context.Context, key string, reader io.Reader, options *storeUnstructured.Options) error {
	s.PutInvocations++
	s.PutInputs = append(s.PutInputs, PutInput{Key: key, Reader: reader, Options: options})
	if s.PutStub != nil {
		return s.PutStub(ctx, key, reader, options)
	}
	if len(s.PutOutputs) > 0 {
		output := s.PutOutputs[0]
		s.PutOutputs = s.PutOutputs[1:]
		return output
	}
	if s.PutOutput != nil {
		return *s.PutOutput
	}
	panic("Put has no output")
}

func (s *Store) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	s.GetInvocations++
	s.GetInputs = append(s.GetInputs, key)
	if s.GetStub != nil {
		return s.GetStub(ctx, key)
	}
	if len(s.GetOutputs) > 0 {
		output := s.GetOutputs[0]
		s.GetOutputs = s.GetOutputs[1:]
		return output.Reader, output.Error
	}
	if s.GetOutput != nil {
		return s.GetOutput.Reader, s.GetOutput.Error
	}
	panic("Get has no output")
}

func (s *Store) Delete(ctx context.Context, key string) (bool, error) {
	s.DeleteInvocations++
	s.DeleteInputs = append(s.DeleteInputs, key)
	if s.DeleteStub != nil {
		return s.DeleteStub(ctx, key)
	}
	if len(s.DeleteOutputs) > 0 {
		output := s.DeleteOutputs[0]
		s.DeleteOutputs = s.DeleteOutputs[1:]
		return output.Deleted, output.Error
	}
	if s.DeleteOutput != nil {
		return s.DeleteOutput.Deleted, s.DeleteOutput.Error
	}
	panic("Delete has no output")
}

func (s *Store) DeleteDirectory(ctx context.Context, key string) error {
	s.DeleteDirectoryInvocations++
	s.DeleteDirectoryInputs = append(s.DeleteDirectoryInputs, key)
	if s.DeleteDirectoryStub != nil {
		return s.DeleteDirectoryStub(ctx, key)
	}
	if len(s.DeleteDirectoryOutputs) > 0 {
		output := s.DeleteDirectoryOutputs[0]
		s.DeleteDirectoryOutputs = s.DeleteDirectoryOutputs[1:]
		return output
	}
	if s.DeleteDirectoryOutput != nil {
		return *s.DeleteDirectoryOutput
	}
	panic("DeleteDirectory has no output")
}

func (s *Store) AssertOutputsEmpty() {
	if len(s.ExistsOutputs) > 0 {
		panic("ExistsOutputs is not empty")
	}
	if len(s.PutOutputs) > 0 {
		panic("PutOutputs is not empty")
	}
	if len(s.GetOutputs) > 0 {
		panic("GetOutputs is not empty")
	}
	if len(s.DeleteOutputs) > 0 {
		panic("DeleteOutputs is not empty")
	}
	if len(s.DeleteDirectoryOutputs) > 0 {
		panic("DeleteDirectoryOutputs is not empty")
	}
}
