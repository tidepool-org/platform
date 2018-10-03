package test

import (
	"context"
	"io"
)

type ExistsInput struct {
	Context context.Context
	Key     string
}

type ExistsOutput struct {
	Exists bool
	Error  error
}

type PutInput struct {
	Context context.Context
	Key     string
	Reader  io.Reader
}

type GetInput struct {
	Context context.Context
	Key     string
}

type GetOutput struct {
	Reader io.ReadCloser
	Error  error
}

type DeleteInput struct {
	Context context.Context
	Key     string
}

type DeleteOutput struct {
	Deleted bool
	Error   error
}

type Store struct {
	ExistsInvocations int
	ExistsInputs      []ExistsInput
	ExistsStub        func(ctx context.Context, key string) (bool, error)
	ExistsOutputs     []ExistsOutput
	ExistsOutput      *ExistsOutput
	PutInvocations    int
	PutInputs         []PutInput
	PutStub           func(ctx context.Context, key string, reader io.Reader) error
	PutOutputs        []error
	PutOutput         *error
	GetInvocations    int
	GetInputs         []GetInput
	GetStub           func(ctx context.Context, key string) (io.ReadCloser, error)
	GetOutputs        []GetOutput
	GetOutput         *GetOutput
	DeleteInvocations int
	DeleteInputs      []DeleteInput
	DeleteStub        func(ctx context.Context, key string) (bool, error)
	DeleteOutputs     []DeleteOutput
	DeleteOutput      *DeleteOutput
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) Exists(ctx context.Context, key string) (bool, error) {
	s.ExistsInvocations++
	s.ExistsInputs = append(s.ExistsInputs, ExistsInput{Context: ctx, Key: key})
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

func (s *Store) Put(ctx context.Context, key string, reader io.Reader) error {
	s.PutInvocations++
	s.PutInputs = append(s.PutInputs, PutInput{Context: ctx, Key: key, Reader: reader})
	if s.PutStub != nil {
		return s.PutStub(ctx, key, reader)
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
	s.GetInputs = append(s.GetInputs, GetInput{Context: ctx, Key: key})
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
	s.DeleteInputs = append(s.DeleteInputs, DeleteInput{Context: ctx, Key: key})
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
}
