package test

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

type HeadObjectWithContextInput struct {
	Context aws.Context
	Input   *s3.HeadObjectInput
	Options []request.Option
}

type HeadObjectWithContextOutput struct {
	Output *s3.HeadObjectOutput
	Error  error
}

type DeleteObjectWithContextInput struct {
	Context aws.Context
	Input   *s3.DeleteObjectInput
	Options []request.Option
}

type DeleteObjectWithContextOutput struct {
	Output *s3.DeleteObjectOutput
	Error  error
}

type S3 struct {
	s3iface.S3API

	HeadObjectWithContextInvocations   int
	HeadObjectWithContextInputs        []HeadObjectWithContextInput
	HeadObjectWithContextStub          func(ctx aws.Context, input *s3.HeadObjectInput, options ...request.Option) (*s3.HeadObjectOutput, error)
	HeadObjectWithContextOutputs       []HeadObjectWithContextOutput
	HeadObjectWithContextOutput        *HeadObjectWithContextOutput
	DeleteObjectWithContextInvocations int
	DeleteObjectWithContextInputs      []DeleteObjectWithContextInput
	DeleteObjectWithContextStub        func(ctx aws.Context, input *s3.DeleteObjectInput, options ...request.Option) (*s3.DeleteObjectOutput, error)
	DeleteObjectWithContextOutputs     []DeleteObjectWithContextOutput
	DeleteObjectWithContextOutput      *DeleteObjectWithContextOutput
}

func NewS3() *S3 {
	return &S3{}
}

func (s *S3) HeadObjectWithContext(ctx aws.Context, input *s3.HeadObjectInput, options ...request.Option) (*s3.HeadObjectOutput, error) {
	s.HeadObjectWithContextInvocations++
	s.HeadObjectWithContextInputs = append(s.HeadObjectWithContextInputs, HeadObjectWithContextInput{Context: ctx, Input: input, Options: options})
	if s.HeadObjectWithContextStub != nil {
		return s.HeadObjectWithContextStub(ctx, input, options...)
	}
	if len(s.HeadObjectWithContextOutputs) > 0 {
		output := s.HeadObjectWithContextOutputs[0]
		s.HeadObjectWithContextOutputs = s.HeadObjectWithContextOutputs[1:]
		return output.Output, output.Error
	}
	if s.HeadObjectWithContextOutput != nil {
		return s.HeadObjectWithContextOutput.Output, s.HeadObjectWithContextOutput.Error
	}
	panic("HeadObjectWithContext has no output")
}

func (s *S3) DeleteObjectWithContext(ctx aws.Context, input *s3.DeleteObjectInput, options ...request.Option) (*s3.DeleteObjectOutput, error) {
	s.DeleteObjectWithContextInvocations++
	s.DeleteObjectWithContextInputs = append(s.DeleteObjectWithContextInputs, DeleteObjectWithContextInput{Context: ctx, Input: input, Options: options})
	if s.DeleteObjectWithContextStub != nil {
		return s.DeleteObjectWithContextStub(ctx, input, options...)
	}
	if len(s.DeleteObjectWithContextOutputs) > 0 {
		output := s.DeleteObjectWithContextOutputs[0]
		s.DeleteObjectWithContextOutputs = s.DeleteObjectWithContextOutputs[1:]
		return output.Output, output.Error
	}
	if s.DeleteObjectWithContextOutput != nil {
		return s.DeleteObjectWithContextOutput.Output, s.DeleteObjectWithContextOutput.Error
	}
	panic("DeleteObjectWithContext has no output")
}

func (s *S3) AssertOutputsEmpty() {
	if len(s.HeadObjectWithContextOutputs) > 0 {
		panic("HeadObjectWithContextOutputs is not empty")
	}
	if len(s.DeleteObjectWithContextOutputs) > 0 {
		panic("DeleteObjectWithContextOutputs is not empty")
	}
}
