package test

import (
	awsSdkGoServiceS3 "github.com/aws/aws-sdk-go/service/s3"
	awsSdkGoServiceS3S3manager "github.com/aws/aws-sdk-go/service/s3/s3manager"
	awsSdkGoServiceS3S3managerS3manageriface "github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"

	"github.com/tidepool-org/platform/aws"
)

type NewDeleteListIteratorInput struct {
	ListObjectsInput *awsSdkGoServiceS3.ListObjectsInput
	Options          []func(*awsSdkGoServiceS3S3manager.DeleteListIterator)
}

type S3Manager struct {
	DownloaderInvocations               int
	DownloaderStub                      func() awsSdkGoServiceS3S3managerS3manageriface.DownloaderAPI
	DownloaderOutputs                   []awsSdkGoServiceS3S3managerS3manageriface.DownloaderAPI
	DownloaderOutput                    *awsSdkGoServiceS3S3managerS3manageriface.DownloaderAPI
	UploaderInvocations                 int
	UploaderStub                        func() awsSdkGoServiceS3S3managerS3manageriface.UploaderAPI
	UploaderOutputs                     []awsSdkGoServiceS3S3managerS3manageriface.UploaderAPI
	UploaderOutput                      *awsSdkGoServiceS3S3managerS3manageriface.UploaderAPI
	NewBatchDeleteWithClientInvocations int
	NewBatchDeleteWithClientInputs      [][]func(*awsSdkGoServiceS3S3manager.BatchDelete)
	NewBatchDeleteWithClientStub        func(options ...func(*awsSdkGoServiceS3S3manager.BatchDelete)) *awsSdkGoServiceS3S3manager.BatchDelete
	NewBatchDeleteWithClientOutputs     []aws.BatchDeleteWithClient
	NewBatchDeleteWithClientOutput      *aws.BatchDeleteWithClient
	NewDeleteListIteratorInvocations    int
	NewDeleteListIteratorInputs         []NewDeleteListIteratorInput
	NewDeleteListIteratorStub           func(listObjectsInput *awsSdkGoServiceS3.ListObjectsInput, options ...func(*awsSdkGoServiceS3S3manager.DeleteListIterator)) awsSdkGoServiceS3S3manager.BatchDeleteIterator
	NewDeleteListIteratorOutputs        []awsSdkGoServiceS3S3manager.BatchDeleteIterator
	NewDeleteListIteratorOutput         *awsSdkGoServiceS3S3manager.BatchDeleteIterator
}

func NewS3Manager() *S3Manager {
	return &S3Manager{}
}

func (s *S3Manager) Downloader() awsSdkGoServiceS3S3managerS3manageriface.DownloaderAPI {
	s.DownloaderInvocations++
	if s.DownloaderStub != nil {
		return s.DownloaderStub()
	}
	if len(s.DownloaderOutputs) > 0 {
		output := s.DownloaderOutputs[0]
		s.DownloaderOutputs = s.DownloaderOutputs[1:]
		return output
	}
	if s.DownloaderOutput != nil {
		return *s.DownloaderOutput
	}
	panic("Downloader has no output")
}

func (s *S3Manager) Uploader() awsSdkGoServiceS3S3managerS3manageriface.UploaderAPI {
	s.UploaderInvocations++
	if s.UploaderStub != nil {
		return s.UploaderStub()
	}
	if len(s.UploaderOutputs) > 0 {
		output := s.UploaderOutputs[0]
		s.UploaderOutputs = s.UploaderOutputs[1:]
		return output
	}
	if s.UploaderOutput != nil {
		return *s.UploaderOutput
	}
	panic("Uploader has no output")
}

func (s *S3Manager) NewBatchDeleteWithClient(options ...func(*awsSdkGoServiceS3S3manager.BatchDelete)) aws.BatchDeleteWithClient {
	s.NewBatchDeleteWithClientInvocations++
	s.NewBatchDeleteWithClientInputs = append(s.NewBatchDeleteWithClientInputs, options)
	if s.NewBatchDeleteWithClientStub != nil {
		return s.NewBatchDeleteWithClientStub(options...)
	}
	if len(s.NewBatchDeleteWithClientOutputs) > 0 {
		output := s.NewBatchDeleteWithClientOutputs[0]
		s.NewBatchDeleteWithClientOutputs = s.NewBatchDeleteWithClientOutputs[1:]
		return output
	}
	if s.NewBatchDeleteWithClientOutput != nil {
		return *s.NewBatchDeleteWithClientOutput
	}
	panic("NewBatchDeleteWithClient has no output")
}

func (s *S3Manager) NewDeleteListIterator(listObjectsInput *awsSdkGoServiceS3.ListObjectsInput, options ...func(*awsSdkGoServiceS3S3manager.DeleteListIterator)) awsSdkGoServiceS3S3manager.BatchDeleteIterator {
	s.NewDeleteListIteratorInvocations++
	s.NewDeleteListIteratorInputs = append(s.NewDeleteListIteratorInputs, NewDeleteListIteratorInput{ListObjectsInput: listObjectsInput, Options: options})
	if s.NewDeleteListIteratorStub != nil {
		return s.NewDeleteListIteratorStub(listObjectsInput, options...)
	}
	if len(s.NewDeleteListIteratorOutputs) > 0 {
		output := s.NewDeleteListIteratorOutputs[0]
		s.NewDeleteListIteratorOutputs = s.NewDeleteListIteratorOutputs[1:]
		return output
	}
	if s.NewDeleteListIteratorOutput != nil {
		return *s.NewDeleteListIteratorOutput
	}
	panic("NewDeleteListIterator has no output")
}

func (s *S3Manager) SetDownloaderOutput(output awsSdkGoServiceS3S3managerS3manageriface.DownloaderAPI) {
	s.DownloaderOutput = &output
}

func (s *S3Manager) SetUploaderOutput(output awsSdkGoServiceS3S3managerS3manageriface.UploaderAPI) {
	s.UploaderOutput = &output
}

func (s *S3Manager) SetNewBatchDeleteWithClientOutput(output aws.BatchDeleteWithClient) {
	s.NewBatchDeleteWithClientOutput = &output
}

func (s *S3Manager) SetNewDeleteListIteratorOutput(output awsSdkGoServiceS3S3manager.BatchDeleteIterator) {
	s.NewDeleteListIteratorOutput = &output
}

func (s *S3Manager) AssertOutputsEmpty() {
	if len(s.DownloaderOutputs) > 0 {
		panic("DownloaderOutputs is not empty")
	}
	if len(s.UploaderOutputs) > 0 {
		panic("UploaderOutputs is not empty")
	}
	if len(s.NewBatchDeleteWithClientOutputs) > 0 {
		panic("NewBatchDeleteWithClientOutputs is not empty")
	}
	if len(s.NewDeleteListIteratorOutputs) > 0 {
		panic("NewDeleteListIteratorOutputs is not empty")
	}
}
