package api

import (
	"errors"

	awsSdkGoAwsSession "github.com/aws/aws-sdk-go/aws/session"
	awsSdkGoServiceS3 "github.com/aws/aws-sdk-go/service/s3"
	awsSdkGoServiceS3S3iface "github.com/aws/aws-sdk-go/service/s3/s3iface"
	awsSdkGoServiceS3S3manager "github.com/aws/aws-sdk-go/service/s3/s3manager"
	awsSdkGoServiceS3S3managerS3manageriface "github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"

	"github.com/tidepool-org/platform/aws"
)

type API struct {
	session *awsSdkGoAwsSession.Session
}

func New(session *awsSdkGoAwsSession.Session) (*API, error) {
	if session == nil {
		return nil, errors.New("aws session is missing")
	}
	return &API{
		session: session,
	}, nil
}

func (a *API) S3() awsSdkGoServiceS3S3iface.S3API {
	return awsSdkGoServiceS3.New(a.session)
}

func (a *API) S3Manager() aws.S3Manager {
	return &S3Manager{
		session: a.session,
	}
}

type S3Manager struct {
	session *awsSdkGoAwsSession.Session
}

func (s *S3Manager) Downloader() awsSdkGoServiceS3S3managerS3manageriface.DownloaderAPI {
	return awsSdkGoServiceS3S3manager.NewDownloader(s.session)
}

func (s *S3Manager) Uploader() awsSdkGoServiceS3S3managerS3manageriface.UploaderAPI {
	return awsSdkGoServiceS3S3manager.NewUploader(s.session)
}

func (s *S3Manager) NewBatchDeleteWithClient(options ...func(*awsSdkGoServiceS3S3manager.BatchDelete)) aws.BatchDeleteWithClient {
	return awsSdkGoServiceS3S3manager.NewBatchDeleteWithClient(awsSdkGoServiceS3.New(s.session), options...)
}

func (s *S3Manager) NewDeleteListIterator(input *awsSdkGoServiceS3.ListObjectsInput, options ...func(*awsSdkGoServiceS3S3manager.DeleteListIterator)) awsSdkGoServiceS3S3manager.BatchDeleteIterator {
	return awsSdkGoServiceS3S3manager.NewDeleteListIterator(awsSdkGoServiceS3.New(s.session), input, options...)
}
