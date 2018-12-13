package aws

import (
	awsSdkGoAws "github.com/aws/aws-sdk-go/aws"
	awsSdkGoServiceS3 "github.com/aws/aws-sdk-go/service/s3"
	awsSdkGoServiceS3S3iface "github.com/aws/aws-sdk-go/service/s3/s3iface"
	awsSdkGoServiceS3S3manager "github.com/aws/aws-sdk-go/service/s3/s3manager"
	awsSdkGoServiceS3S3managerS3manageriface "github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
)

type API interface {
	S3() awsSdkGoServiceS3S3iface.S3API
	S3Manager() S3Manager
}

type S3Manager interface {
	Downloader() awsSdkGoServiceS3S3managerS3manageriface.DownloaderAPI
	Uploader() awsSdkGoServiceS3S3managerS3manageriface.UploaderAPI

	NewBatchDeleteWithClient(options ...func(*awsSdkGoServiceS3S3manager.BatchDelete)) BatchDeleteWithClient
	NewDeleteListIterator(listObjectsInput *awsSdkGoServiceS3.ListObjectsInput, options ...func(*awsSdkGoServiceS3S3manager.DeleteListIterator)) awsSdkGoServiceS3S3manager.BatchDeleteIterator
}

type BatchDeleteWithClient interface {
	Delete(ctx awsSdkGoAws.Context, batchDeleteIterator awsSdkGoServiceS3S3manager.BatchDeleteIterator) error
}

func NewWriteAtBuffer(bytes []byte) *awsSdkGoAws.WriteAtBuffer {
	return awsSdkGoAws.NewWriteAtBuffer(bytes)
}
