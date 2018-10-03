package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/tidepool-org/platform/aws"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	storeUnstructured "github.com/tidepool-org/platform/store/unstructured"
)

const Type = "s3"

type Store struct {
	bucket string
	prefix string
	awsAPI aws.API
}

func NewStore(cfg *Config, awsAPI aws.API) (*Store, error) {
	if cfg == nil {
		return nil, errors.New("config is missing")
	} else if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "config is invalid")
	}
	if awsAPI == nil {
		return nil, errors.New("aws api is missing")
	}

	return &Store{
		bucket: cfg.Bucket,
		prefix: cfg.Prefix,
		awsAPI: awsAPI,
	}, nil
}

func (s *Store) Exists(ctx context.Context, key string) (bool, error) {
	if ctx == nil {
		return false, errors.New("context is missing")
	}
	if key == "" {
		return false, errors.New("key is missing")
	} else if !storeUnstructured.IsValidKey(key) {
		return false, errors.New("key is invalid")
	}

	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"bucket": s.bucket, "prefix": s.prefix, "key": key})
	key = s.resolveKey(key)

	var exists bool
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}
	if _, err := s.awsAPI.S3().HeadObjectWithContext(ctx, input); err != nil {
		if awsErr, ok := err.(awserr.Error); !ok || awsErr.Code() != "NotFound" {
			logger.WithError(err).Errorf("Unable to head object with key %q", key)
			return false, errors.Wrapf(err, "unable to head object with key %q", key)
		}
	} else {
		exists = true
	}

	logger.WithField("exists", exists).Debug("Exists")
	return exists, nil
}

func (s *Store) Put(ctx context.Context, key string, reader io.Reader) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if key == "" {
		return errors.New("key is missing")
	} else if !storeUnstructured.IsValidKey(key) {
		return errors.New("key is invalid")
	}
	if reader == nil {
		return errors.New("reader is missing")
	}

	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"bucket": s.bucket, "prefix": s.prefix, "key": key})
	key = s.resolveKey(key)

	input := &s3manager.UploadInput{
		Body:                 reader,
		Bucket:               aws.String(s.bucket),
		Key:                  aws.String(key),
		ServerSideEncryption: aws.String("AES256"),
	}
	if _, err := s.awsAPI.S3ManagerUploader().UploadWithContext(ctx, input); err != nil {
		logger.WithError(err).Errorf("Unable to upload object with key %q", key)
		return errors.Wrapf(err, "unable to upload object with key %q", key)
	}

	logger.Debug("Put")
	return nil
}

func (s *Store) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if key == "" {
		return nil, errors.New("key is missing")
	} else if !storeUnstructured.IsValidKey(key) {
		return nil, errors.New("key is invalid")
	}

	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"bucket": s.bucket, "prefix": s.prefix, "key": key})
	key = s.resolveKey(key)

	var reader io.ReadCloser
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}
	output := aws.NewWriteAtBuffer(nil) // FUTURE: Uses memory - if large objects then need to use temporary file on disk
	if _, err := s.awsAPI.S3ManagerDownloader().DownloadWithContext(ctx, output, input); err != nil {
		if awsErr, ok := err.(awserr.Error); !ok || awsErr.Code() != s3.ErrCodeNoSuchKey {
			logger.WithError(err).Errorf("Unable to download object with key %q", key)
			return nil, errors.Wrapf(err, "unable to download object with key %q", key)
		}
	} else {
		reader = ioutil.NopCloser(bytes.NewReader(output.Bytes()))
	}

	logger.WithField("exists", reader != nil).Debug("Get")
	return reader, nil

}

func (s *Store) Delete(ctx context.Context, key string) (bool, error) {
	if ctx == nil {
		return false, errors.New("context is missing")
	}
	if key == "" {
		return false, errors.New("key is missing")
	} else if !storeUnstructured.IsValidKey(key) {
		return false, errors.New("key is invalid")
	}

	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"bucket": s.bucket, "prefix": s.prefix, "key": key})
	key = s.resolveKey(key)

	var exists bool
	headObjectInput := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}
	if _, err := s.awsAPI.S3().HeadObjectWithContext(ctx, headObjectInput); err != nil {
		if awsErr, ok := err.(awserr.Error); !ok || awsErr.Code() != "NotFound" {
			logger.WithError(err).Errorf("Unable to head object with key %q", key)
			return false, errors.Wrapf(err, "unable to head object with key %q", key)
		}
	} else {
		exists = true
		deleteObjectInput := &s3.DeleteObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(key),
		}
		if _, err = s.awsAPI.S3().DeleteObjectWithContext(ctx, deleteObjectInput); err != nil {
			logger.WithError(err).Errorf("Unable to delete object with key %q", key)
			return false, errors.Wrapf(err, "unable to delete object with key %q", key)
		}
	}

	logger.WithField("exists", exists).Debug("Delete")
	return exists, nil
}

func (s *Store) resolveKey(key string) string {
	return fmt.Sprintf("%s/%s", s.prefix, key)
}
