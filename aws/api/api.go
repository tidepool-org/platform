package api

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
)

type API struct {
	awsSession *session.Session
}

func New(awsSession *session.Session) (*API, error) {
	if awsSession == nil {
		return nil, errors.New("aws session is missing")
	}
	return &API{
		awsSession: awsSession,
	}, nil
}

func (a *API) S3() s3iface.S3API {
	return s3.New(a.awsSession)
}

func (a *API) S3ManagerDownloader() s3manageriface.DownloaderAPI {
	return s3manager.NewDownloader(a.awsSession)
}

func (a *API) S3ManagerUploader() s3manageriface.UploaderAPI {
	return s3manager.NewUploader(a.awsSession)
}
