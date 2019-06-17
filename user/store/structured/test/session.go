package test

import (
	"context"

	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/user"
)

type GetInput struct {
	ID        string
	Condition *request.Condition
}

type GetOutput struct {
	User  *user.User
	Error error
}

type DeleteInput struct {
	ID        string
	Condition *request.Condition
}

type DeleteOutput struct {
	Deleted bool
	Error   error
}

type DestroyInput struct {
	ID        string
	Condition *request.Condition
}

type DestroyOutput struct {
	Destroyed bool
	Error     error
}

type Session struct {
	*test.Closer
	GetInvocations     int
	GetInputs          []GetInput
	GetStub            func(ctx context.Context, id string, condition *request.Condition) (*user.User, error)
	GetOutputs         []GetOutput
	GetOutput          *GetOutput
	DeleteInvocations  int
	DeleteInputs       []DeleteInput
	DeleteStub         func(ctx context.Context, id string, condition *request.Condition) (bool, error)
	DeleteOutputs      []DeleteOutput
	DeleteOutput       *DeleteOutput
	DestroyInvocations int
	DestroyInputs      []DestroyInput
	DestroyStub        func(ctx context.Context, id string, condition *request.Condition) (bool, error)
	DestroyOutputs     []DestroyOutput
	DestroyOutput      *DestroyOutput
}

func NewSession() *Session {
	return &Session{
		Closer: test.NewCloser(),
	}
}

func (s *Session) Get(ctx context.Context, id string, condition *request.Condition) (*user.User, error) {
	s.GetInvocations++
	s.GetInputs = append(s.GetInputs, GetInput{ID: id, Condition: condition})
	if s.GetStub != nil {
		return s.GetStub(ctx, id, condition)
	}
	if len(s.GetOutputs) > 0 {
		output := s.GetOutputs[0]
		s.GetOutputs = s.GetOutputs[1:]
		return output.User, output.Error
	}
	if s.GetOutput != nil {
		return s.GetOutput.User, s.GetOutput.Error
	}
	panic("Get has no output")
}

func (s *Session) Delete(ctx context.Context, id string, condition *request.Condition) (bool, error) {
	s.DeleteInvocations++
	s.DeleteInputs = append(s.DeleteInputs, DeleteInput{ID: id, Condition: condition})
	if s.DeleteStub != nil {
		return s.DeleteStub(ctx, id, condition)
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

func (s *Session) Destroy(ctx context.Context, id string, condition *request.Condition) (bool, error) {
	s.DestroyInvocations++
	s.DestroyInputs = append(s.DestroyInputs, DestroyInput{ID: id, Condition: condition})
	if s.DestroyStub != nil {
		return s.DestroyStub(ctx, id, condition)
	}
	if len(s.DestroyOutputs) > 0 {
		output := s.DestroyOutputs[0]
		s.DestroyOutputs = s.DestroyOutputs[1:]
		return output.Destroyed, output.Error
	}
	if s.DestroyOutput != nil {
		return s.DestroyOutput.Destroyed, s.DestroyOutput.Error
	}
	panic("Destroy has no output")
}

func (s *Session) AssertOutputsEmpty() {
	s.Closer.AssertOutputsEmpty()
	if len(s.GetOutputs) > 0 {
		panic("GetOutputs is not empty")
	}
	if len(s.DeleteOutputs) > 0 {
		panic("DeleteOutputs is not empty")
	}
	if len(s.DestroyOutputs) > 0 {
		panic("DestroyOutputs is not empty")
	}
}
