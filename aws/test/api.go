package test

import (
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
)

type API struct {
	S3Invocations                  int
	S3Stub                         func() s3iface.S3API
	S3Outputs                      []s3iface.S3API
	S3Output                       *s3iface.S3API
	S3ManagerDownloaderInvocations int
	S3ManagerDownloaderStub        func() s3manageriface.DownloaderAPI
	S3ManagerDownloaderOutputs     []s3manageriface.DownloaderAPI
	S3ManagerDownloaderOutput      *s3manageriface.DownloaderAPI
	S3ManagerUploaderInvocations   int
	S3ManagerUploaderStub          func() s3manageriface.UploaderAPI
	S3ManagerUploaderOutputs       []s3manageriface.UploaderAPI
	S3ManagerUploaderOutput        *s3manageriface.UploaderAPI
}

func NewAPI() *API {
	return &API{}
}

func (a *API) S3() s3iface.S3API {
	a.S3Invocations++
	if a.S3Stub != nil {
		return a.S3Stub()
	}
	if len(a.S3Outputs) > 0 {
		output := a.S3Outputs[0]
		a.S3Outputs = a.S3Outputs[1:]
		return output
	}
	if a.S3Output != nil {
		return *a.S3Output
	}
	panic("S3API has no output")
}

func (a *API) S3ManagerDownloader() s3manageriface.DownloaderAPI {
	a.S3ManagerDownloaderInvocations++
	if a.S3ManagerDownloaderStub != nil {
		return a.S3ManagerDownloaderStub()
	}
	if len(a.S3ManagerDownloaderOutputs) > 0 {
		output := a.S3ManagerDownloaderOutputs[0]
		a.S3ManagerDownloaderOutputs = a.S3ManagerDownloaderOutputs[1:]
		return output
	}
	if a.S3ManagerDownloaderOutput != nil {
		return *a.S3ManagerDownloaderOutput
	}
	panic("S3ManagerDownloader has no output")
}

func (a *API) S3ManagerUploader() s3manageriface.UploaderAPI {
	a.S3ManagerUploaderInvocations++
	if a.S3ManagerUploaderStub != nil {
		return a.S3ManagerUploaderStub()
	}
	if len(a.S3ManagerUploaderOutputs) > 0 {
		output := a.S3ManagerUploaderOutputs[0]
		a.S3ManagerUploaderOutputs = a.S3ManagerUploaderOutputs[1:]
		return output
	}
	if a.S3ManagerUploaderOutput != nil {
		return *a.S3ManagerUploaderOutput
	}
	panic("S3ManagerUploader has no output")
}

func (a *API) AssertOutputsEmpty() {
	if len(a.S3Outputs) > 0 {
		panic("S3Outputs is not empty")
	}
	if len(a.S3ManagerDownloaderOutputs) > 0 {
		panic("S3ManagerDownloaderOutputs is not empty")
	}
	if len(a.S3ManagerUploaderOutputs) > 0 {
		panic("S3ManagerUploaderOutputs is not empty")
	}
}
