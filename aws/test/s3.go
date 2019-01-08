package test

import (
	awsSdkGoAws "github.com/aws/aws-sdk-go/aws"
	awsSdkGoAwsRequest "github.com/aws/aws-sdk-go/aws/request"
	awsSdkGoServiceS3 "github.com/aws/aws-sdk-go/service/s3"
	awsSdkGoServiceS3S3iface "github.com/aws/aws-sdk-go/service/s3/s3iface"
)

type HeadObjectWithContextInput struct {
	Input   *awsSdkGoServiceS3.HeadObjectInput
	Options []awsSdkGoAwsRequest.Option
}

type HeadObjectWithContextOutput struct {
	Output *awsSdkGoServiceS3.HeadObjectOutput
	Error  error
}

type DeleteObjectWithContextInput struct {
	Input   *awsSdkGoServiceS3.DeleteObjectInput
	Options []awsSdkGoAwsRequest.Option
}

type DeleteObjectWithContextOutput struct {
	Output *awsSdkGoServiceS3.DeleteObjectOutput
	Error  error
}

type S3 struct {
	awsSdkGoServiceS3S3iface.S3API

	HeadObjectWithContextInvocations   int
	HeadObjectWithContextInputs        []HeadObjectWithContextInput
	HeadObjectWithContextStub          func(ctx awsSdkGoAws.Context, input *awsSdkGoServiceS3.HeadObjectInput, options ...awsSdkGoAwsRequest.Option) (*awsSdkGoServiceS3.HeadObjectOutput, error)
	HeadObjectWithContextOutputs       []HeadObjectWithContextOutput
	HeadObjectWithContextOutput        *HeadObjectWithContextOutput
	DeleteObjectWithContextInvocations int
	DeleteObjectWithContextInputs      []DeleteObjectWithContextInput
	DeleteObjectWithContextStub        func(ctx awsSdkGoAws.Context, input *awsSdkGoServiceS3.DeleteObjectInput, options ...awsSdkGoAwsRequest.Option) (*awsSdkGoServiceS3.DeleteObjectOutput, error)
	DeleteObjectWithContextOutputs     []DeleteObjectWithContextOutput
	DeleteObjectWithContextOutput      *DeleteObjectWithContextOutput
}

func NewS3() *S3 {
	return &S3{}
}

func (s *S3) HeadObjectWithContext(ctx awsSdkGoAws.Context, input *awsSdkGoServiceS3.HeadObjectInput, options ...awsSdkGoAwsRequest.Option) (*awsSdkGoServiceS3.HeadObjectOutput, error) {
	s.HeadObjectWithContextInvocations++
	s.HeadObjectWithContextInputs = append(s.HeadObjectWithContextInputs, HeadObjectWithContextInput{Input: input, Options: options})
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

func (s *S3) DeleteObjectWithContext(ctx awsSdkGoAws.Context, input *awsSdkGoServiceS3.DeleteObjectInput, options ...awsSdkGoAwsRequest.Option) (*awsSdkGoServiceS3.DeleteObjectOutput, error) {
	s.DeleteObjectWithContextInvocations++
	s.DeleteObjectWithContextInputs = append(s.DeleteObjectWithContextInputs, DeleteObjectWithContextInput{Input: input, Options: options})
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

func (s *S3) SetHeadObjectWithContextOutput(output HeadObjectWithContextOutput) {
	s.HeadObjectWithContextOutput = &output
}

func (s *S3) SetDeleteObjectWithContextOutput(output DeleteObjectWithContextOutput) {
	s.DeleteObjectWithContextOutput = &output
}

func (s *S3) AssertOutputsEmpty() {
	if len(s.HeadObjectWithContextOutputs) > 0 {
		panic("HeadObjectWithContextOutputs is not empty")
	}
	if len(s.DeleteObjectWithContextOutputs) > 0 {
		panic("DeleteObjectWithContextOutputs is not empty")
	}
}
