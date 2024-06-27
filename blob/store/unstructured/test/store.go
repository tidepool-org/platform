package test

import (
	"context"
	"io"

	storeUnstructured "github.com/tidepool-org/platform/store/unstructured"
)

type ExistsInput struct {
	UserID string
	ID     string
}

type ExistsOutput struct {
	Exists bool
	Error  error
}

type PutInput struct {
	UserID  string
	ID      string
	Reader  io.Reader
	Options *storeUnstructured.Options
}

type GetInput struct {
	UserID string
	ID     string
}

type GetOutput struct {
	Reader io.ReadCloser
	Error  error
}

type DeleteInput struct {
	UserID string
	ID     string
}

type DeleteOutput struct {
	Deleted bool
	Error   error
}

type Store struct {
	ExistsInvocations    int
	ExistsInputs         []ExistsInput
	ExistsStub           func(ctx context.Context, userID string, id string) (bool, error)
	ExistsOutputs        []ExistsOutput
	ExistsOutput         *ExistsOutput
	PutInvocations       int
	PutInputs            []PutInput
	PutStub              func(ctx context.Context, userID string, id string, reader io.Reader, options *storeUnstructured.Options) error
	PutOutputs           []error
	PutOutput            *error
	GetInvocations       int
	GetInputs            []GetInput
	GetStub              func(ctx context.Context, userID string, id string) (io.ReadCloser, error)
	GetOutputs           []GetOutput
	GetOutput            *GetOutput
	DeleteInvocations    int
	DeleteInputs         []DeleteInput
	DeleteStub           func(ctx context.Context, userID string, id string) (bool, error)
	DeleteOutputs        []DeleteOutput
	DeleteOutput         *DeleteOutput
	DeleteAllInvocations int
	DeleteAllInputs      []string
	DeleteAllStub        func(ctx context.Context, userID string) error
	DeleteAllOutputs     []error
	DeleteAllOutput      *error
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) Exists(ctx context.Context, userID string, id string) (bool, error) {
	s.ExistsInvocations++
	s.ExistsInputs = append(s.ExistsInputs, ExistsInput{UserID: userID, ID: id})
	if s.ExistsStub != nil {
		return s.ExistsStub(ctx, userID, id)
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

func (s *Store) Put(ctx context.Context, userID string, id string, reader io.Reader, options *storeUnstructured.Options) error {
	s.PutInvocations++
	s.PutInputs = append(s.PutInputs, PutInput{UserID: userID, ID: id, Reader: reader, Options: options})
	if s.PutStub != nil {
		return s.PutStub(ctx, userID, id, reader, options)
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

func (s *Store) Get(ctx context.Context, userID string, id string) (io.ReadCloser, error) {
	s.GetInvocations++
	s.GetInputs = append(s.GetInputs, GetInput{UserID: userID, ID: id})
	if s.GetStub != nil {
		return s.GetStub(ctx, userID, id)
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

func (s *Store) Delete(ctx context.Context, userID string, id string) (bool, error) {
	s.DeleteInvocations++
	s.DeleteInputs = append(s.DeleteInputs, DeleteInput{UserID: userID, ID: id})
	if s.DeleteStub != nil {
		return s.DeleteStub(ctx, userID, id)
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

func (s *Store) DeleteAll(ctx context.Context, userID string) error {
	s.DeleteAllInvocations++
	s.DeleteAllInputs = append(s.DeleteAllInputs, userID)
	if s.DeleteAllStub != nil {
		return s.DeleteAllStub(ctx, userID)
	}
	if len(s.DeleteAllOutputs) > 0 {
		output := s.DeleteAllOutputs[0]
		s.DeleteAllOutputs = s.DeleteAllOutputs[1:]
		return output
	}
	if s.DeleteAllOutput != nil {
		return *s.DeleteAllOutput
	}
	panic("DeleteAll has no output")
}

func (s *Store) GetMany(ctx context.Context, userID string, ids ...string) ([]io.ReadCloser, error) {
	return nil, nil
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
	if len(s.DeleteAllOutputs) > 0 {
		panic("DeleteAllOutputs is not empty")
	}
}
