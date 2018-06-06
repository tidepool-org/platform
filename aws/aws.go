package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
)

type API interface {
	S3() s3iface.S3API
	S3ManagerDownloader() s3manageriface.DownloaderAPI
	S3ManagerUploader() s3manageriface.UploaderAPI
}

func String(value string) *string {
	return aws.String(value)
}

func NewWriteAtBuffer(bytes []byte) *aws.WriteAtBuffer {
	return aws.NewWriteAtBuffer(bytes)
}
