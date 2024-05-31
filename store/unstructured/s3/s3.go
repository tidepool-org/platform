package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"

	awsSdkGoAwsAwserr "github.com/aws/aws-sdk-go/aws/awserr"
	awsSdkGoServiceS3 "github.com/aws/aws-sdk-go/service/s3"
	awsSdkGoServiceS3S3manager "github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/tidepool-org/platform/aws"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/pointer"
	storeUnstructured "github.com/tidepool-org/platform/store/unstructured"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const Type = "s3"

type Store struct {
	bucket string
	prefix string
	awsAPI aws.API
}

func NewStore(config *Config, awsAPI aws.API) (*Store, error) {
	if config == nil {
		return nil, errors.New("config is missing")
	} else if err := config.Validate(); err != nil {
		return nil, errors.Wrap(err, "config is invalid")
	}
	if awsAPI == nil {
		return nil, errors.New("aws api is missing")
	}

	return &Store{
		bucket: config.Bucket,
		prefix: config.Prefix,
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

	ctx, logger := log.ContextAndLoggerWithFields(ctx, log.Fields{"bucket": s.bucket, "prefix": s.prefix, "key": key})
	key = s.resolveKey(key)

	var exists bool
	input := &awsSdkGoServiceS3.HeadObjectInput{
		Bucket: pointer.FromString(s.bucket),
		Key:    pointer.FromString(key),
	}
	if _, err := s.awsAPI.S3().HeadObjectWithContext(ctx, input); err != nil {
		if awsErr, ok := err.(awsSdkGoAwsAwserr.Error); !ok || awsErr.Code() != "NotFound" {
			logger.WithError(err).Errorf("Unable to head object with key %q", key)
			return false, errors.Wrapf(err, "unable to head object with key %q", key)
		}
	} else {
		exists = true
	}

	logger.WithField("exists", exists).Debug("Exists")
	return exists, nil
}

func (s *Store) Put(ctx context.Context, key string, reader io.Reader, options *storeUnstructured.Options) error {
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
	if options == nil {
		options = storeUnstructured.NewOptions()
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(options); err != nil {
		return errors.Wrap(err, "options is invalid")
	}

	ctx, logger := log.ContextAndLoggerWithFields(ctx, log.Fields{"bucket": s.bucket, "prefix": s.prefix, "key": key, "options": options})
	key = s.resolveKey(key)

	input := &awsSdkGoServiceS3S3manager.UploadInput{
		Body:                 reader,
		Bucket:               pointer.FromString(s.bucket),
		ContentType:          options.MediaType,
		Key:                  pointer.FromString(key),
		ServerSideEncryption: pointer.FromString("AES256"),
	}
	if _, err := s.awsAPI.S3Manager().Uploader().UploadWithContext(ctx, input); err != nil {
		logger.WithError(err).Errorf("Unable to upload object with key %q, bucket %q", key, s.bucket)
		return errors.Wrapf(err, "unable to upload object with key %q, bucket %q", key, s.bucket)
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

	ctx, logger := log.ContextAndLoggerWithFields(ctx, log.Fields{"bucket": s.bucket, "prefix": s.prefix, "key": key})
	key = s.resolveKey(key)

	var reader io.ReadCloser
	input := &awsSdkGoServiceS3.GetObjectInput{
		Bucket: pointer.FromString(s.bucket),
		Key:    pointer.FromString(key),
	}
	output := aws.NewWriteAtBuffer(nil) // FUTURE: Uses memory - if large objects then need to use temporary file on disk
	if _, err := s.awsAPI.S3Manager().Downloader().DownloadWithContext(ctx, output, input); err != nil {
		if awsErr, ok := err.(awsSdkGoAwsAwserr.Error); !ok || awsErr.Code() != awsSdkGoServiceS3.ErrCodeNoSuchKey {
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

	ctx, logger := log.ContextAndLoggerWithFields(ctx, log.Fields{"bucket": s.bucket, "prefix": s.prefix, "key": key})
	key = s.resolveKey(key)

	var exists bool
	headObjectInput := &awsSdkGoServiceS3.HeadObjectInput{
		Bucket: pointer.FromString(s.bucket),
		Key:    pointer.FromString(key),
	}
	if _, err := s.awsAPI.S3().HeadObjectWithContext(ctx, headObjectInput); err != nil {
		if awsErr, ok := err.(awsSdkGoAwsAwserr.Error); !ok || awsErr.Code() != "NotFound" {
			logger.WithError(err).Errorf("Unable to head object with key %q", key)
			return false, errors.Wrapf(err, "unable to head object with key %q", key)
		}
	} else {
		exists = true
		deleteObjectInput := &awsSdkGoServiceS3.DeleteObjectInput{
			Bucket: pointer.FromString(s.bucket),
			Key:    pointer.FromString(key),
		}
		if _, err = s.awsAPI.S3().DeleteObjectWithContext(ctx, deleteObjectInput); err != nil {
			logger.WithError(err).Errorf("Unable to delete object with key %q", key)
			return false, errors.Wrapf(err, "unable to delete object with key %q", key)
		}
	}

	logger.WithField("exists", exists).Debug("Delete")
	return exists, nil
}

func (s *Store) DeleteDirectory(ctx context.Context, key string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if key == "" {
		return errors.New("key is missing")
	} else if !storeUnstructured.IsValidKey(key) {
		return errors.New("key is invalid")
	}

	ctx, logger := log.ContextAndLoggerWithFields(ctx, log.Fields{"bucket": s.bucket, "prefix": s.prefix, "key": key})
	key = s.resolveKey(key)

	batchDelete := s.awsAPI.S3Manager().NewBatchDeleteWithClient(func(batchDelete *awsSdkGoServiceS3S3manager.BatchDelete) {
		batchDelete.BatchSize = deleteDirectoryBatchSize
	})
	listObjectsInput := &awsSdkGoServiceS3.ListObjectsInput{
		Bucket:  pointer.FromString(s.bucket),
		Prefix:  pointer.FromString(key),
		MaxKeys: func(batchSize int64) *int64 { return &batchSize }(deleteDirectoryBatchSize),
	}
	if err := batchDelete.Delete(ctx, s.awsAPI.S3Manager().NewDeleteListIterator(listObjectsInput)); err != nil {
		logger.WithError(err).Errorf("Unable to delete all objects with key %q", key)
		return errors.Wrapf(err, "unable to delete all objects with key %q", key)
	}

	logger.Debug("DeleteDirectory")
	return nil
}

func (s *Store) resolveKey(key string) string {
	return fmt.Sprintf("%s/%s", s.prefix, key)
}

const deleteDirectoryBatchSize = 1000
