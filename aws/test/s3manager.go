package test

import (
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
)

type DownloadWithContextInput struct {
	Context  aws.Context
	WriterAt io.WriterAt
	Input    *s3.GetObjectInput
	Options  []func(*s3manager.Downloader)
}

type DownloadWithContextOutput struct {
	BytesWritten int64
	Error        error
}

type UploadWithContextInput struct {
	Context aws.Context
	Input   *s3manager.UploadInput
	Options []func(*s3manager.Uploader)
}

type UploadWithContextOutput struct {
	Output *s3manager.UploadOutput
	Error  error
}

type S3Manager struct {
	s3manageriface.DownloaderAPI
	s3manageriface.UploaderAPI

	DownloadWithContextInvocations int
	DownloadWithContextInputs      []DownloadWithContextInput
	DownloadWithContextStub        func(ctx aws.Context, writerAt io.WriterAt, input *s3.GetObjectInput, options ...func(*s3manager.Downloader)) (int64, error)
	DownloadWithContextOutputs     []DownloadWithContextOutput
	DownloadWithContextOutput      *DownloadWithContextOutput
	UploadWithContextInvocations   int
	UploadWithContextInputs        []UploadWithContextInput
	UploadWithContextStub          func(ctx aws.Context, input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
	UploadWithContextOutputs       []UploadWithContextOutput
	UploadWithContextOutput        *UploadWithContextOutput
}

func NewS3Manager() *S3Manager {
	return &S3Manager{}
}

func (s *S3Manager) DownloadWithContext(ctx aws.Context, writerAt io.WriterAt, input *s3.GetObjectInput, options ...func(*s3manager.Downloader)) (int64, error) {
	s.DownloadWithContextInvocations++
	s.DownloadWithContextInputs = append(s.DownloadWithContextInputs, DownloadWithContextInput{Context: ctx, WriterAt: writerAt, Input: input, Options: options})
	if s.DownloadWithContextStub != nil {
		return s.DownloadWithContextStub(ctx, writerAt, input, options...)
	}
	if len(s.DownloadWithContextOutputs) > 0 {
		output := s.DownloadWithContextOutputs[0]
		s.DownloadWithContextOutputs = s.DownloadWithContextOutputs[1:]
		return output.BytesWritten, output.Error
	}
	if s.DownloadWithContextOutput != nil {
		return s.DownloadWithContextOutput.BytesWritten, s.DownloadWithContextOutput.Error
	}
	panic("DownloadWithContext has no output")
}

func (s *S3Manager) UploadWithContext(ctx aws.Context, input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
	s.UploadWithContextInvocations++
	s.UploadWithContextInputs = append(s.UploadWithContextInputs, UploadWithContextInput{Context: ctx, Input: input, Options: options})
	if s.UploadWithContextStub != nil {
		return s.UploadWithContextStub(ctx, input, options...)
	}
	if len(s.UploadWithContextOutputs) > 0 {
		output := s.UploadWithContextOutputs[0]
		s.UploadWithContextOutputs = s.UploadWithContextOutputs[1:]
		return output.Output, output.Error
	}
	if s.UploadWithContextOutput != nil {
		return s.UploadWithContextOutput.Output, s.UploadWithContextOutput.Error
	}
	panic("UploadWithContext has no output")
}

func (s *S3Manager) AssertOutputsEmpty() {
	if len(s.DownloadWithContextOutputs) > 0 {
		panic("DownloadWithContextOutputs is not empty")
	}
	if len(s.UploadWithContextOutputs) > 0 {
		panic("UploadWithContextOutputs is not empty")
	}
}
