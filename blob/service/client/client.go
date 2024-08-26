package client

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"io"

	"github.com/tidepool-org/platform/blob"
	blobStoreStructured "github.com/tidepool-org/platform/blob/store/structured"
	blobStoreUnstructured "github.com/tidepool-org/platform/blob/store/unstructured"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	storeUnstructured "github.com/tidepool-org/platform/store/unstructured"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type Provider interface {
	BlobStructuredStore() blobStoreStructured.Store
	BlobUnstructuredStore() blobStoreUnstructured.Store
	DeviceLogsUnstructuredStore() blobStoreUnstructured.Store
}

type Client struct {
	Provider
}

func New(provider Provider) (*Client, error) {
	if provider == nil {
		return nil, errors.New("provider is missing")
	}

	return &Client{
		Provider: provider,
	}, nil
}

// FUTURE: Return ErrorResourceNotFoundWithID(userID) if userID does not exist at all

func (c *Client) List(ctx context.Context, userID string, filter *blob.Filter, pagination *page.Pagination) (blob.BlobArray, error) {
	repository := c.BlobStructuredStore().NewBlobRepository()
	return repository.List(ctx, userID, filter, pagination)
}

func (c *Client) Create(ctx context.Context, userID string, content *blob.Content) (*blob.Blob, error) {
	if content == nil {
		return nil, errors.New("content is missing")
	} else if err := structureValidator.New().Validate(content); err != nil {
		return nil, errors.Wrap(err, "content is invalid")
	}

	repository := c.BlobStructuredStore().NewBlobRepository()

	structuredCreate := blobStoreStructured.NewCreate()
	structuredCreate.MediaType = pointer.CloneString(content.MediaType)
	result, err := repository.Create(ctx, userID, structuredCreate)
	if err != nil {
		return nil, err
	}

	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": userID, "id": *result.ID})

	hasher := md5.New()
	sizer := NewSizeWriter()
	options := storeUnstructured.NewOptions()
	options.MediaType = content.MediaType
	err = c.BlobUnstructuredStore().Put(ctx, userID, *result.ID, io.TeeReader(io.TeeReader(io.LimitReader(content.Body, blob.SizeMaximum+1), hasher), sizer), options)
	if err != nil {
		if _, destroyErr := repository.Destroy(ctx, *result.ID, nil); destroyErr != nil {
			logger.WithError(destroyErr).Error("Unable to destroy blob after failure to put blob content")
		}
		return nil, err
	}

	size := sizer.Size
	if size > blob.SizeMaximum {
		if _, deleteErr := c.BlobUnstructuredStore().Delete(ctx, userID, *result.ID); deleteErr != nil {
			logger.WithError(deleteErr).Error("Unable to delete blob content exceeding maximum size")
		}
		if _, destroyErr := repository.Destroy(ctx, *result.ID, nil); destroyErr != nil {
			logger.WithError(destroyErr).Error("Unable to destroy blob exceeding maximum size")
		}
		return nil, request.ErrorResourceTooLarge()
	}

	digestMD5 := base64.StdEncoding.EncodeToString(hasher.Sum(nil))
	if content.DigestMD5 != nil && *content.DigestMD5 != digestMD5 {
		if _, deleteErr := c.BlobUnstructuredStore().Delete(ctx, userID, *result.ID); deleteErr != nil {
			logger.WithError(deleteErr).Error("Unable to delete blob content with incorrect MD5 digest")
		}
		if _, destroyErr := repository.Destroy(ctx, *result.ID, nil); destroyErr != nil {
			logger.WithError(destroyErr).Error("Unable to destroy blob with incorrect MD5 digest")
		}
		return nil, errors.WithSource(request.ErrorDigestsNotEqual(*content.DigestMD5, digestMD5), structure.NewPointerSource().WithReference("digestMD5"))
	}

	update := blobStoreStructured.NewUpdate()
	update.DigestMD5 = pointer.FromString(digestMD5)
	update.Size = pointer.FromInt(size)
	update.Status = pointer.FromString(blob.StatusAvailable)
	return repository.Update(ctx, *result.ID, nil, update)
}

func (c *Client) CreateDeviceLogs(ctx context.Context, userID string, content *blob.DeviceLogsContent) (*blob.DeviceLogsBlob, error) {
	if content == nil {
		return nil, errors.New("content is missing")
	} else if err := structureValidator.New().Validate(content); err != nil {
		return nil, errors.Wrap(err, "content is invalid")
	}

	repository := c.BlobStructuredStore().NewDeviceLogsRepository()

	structuredCreate := blobStoreStructured.NewCreate()
	structuredCreate.MediaType = pointer.CloneString(content.MediaType)
	result, err := repository.Create(ctx, userID, structuredCreate)
	if err != nil {
		return nil, err
	}

	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": userID, "id": *result.ID})

	hasher := md5.New()
	sizer := NewSizeWriter()
	options := storeUnstructured.NewOptions()
	options.MediaType = content.MediaType
	err = c.DeviceLogsUnstructuredStore().Put(ctx, userID, *result.ID, io.TeeReader(io.TeeReader(io.LimitReader(content.Body, blob.SizeMaximum+1), hasher), sizer), options)
	if err != nil {
		if _, destroyErr := repository.Destroy(ctx, *result.ID, nil); destroyErr != nil {
			logger.WithError(destroyErr).Error("Unable to destroy blob after failure to put blob content")
		}
		return nil, err
	}

	size := sizer.Size
	if size > blob.SizeMaximum {
		if _, deleteErr := c.DeviceLogsUnstructuredStore().Delete(ctx, userID, *result.ID); deleteErr != nil {
			logger.WithError(deleteErr).Error("Unable to delete blob content exceeding maximum size")
		}
		if _, destroyErr := repository.Destroy(ctx, *result.ID, nil); destroyErr != nil {
			logger.WithError(destroyErr).Error("Unable to destroy blob exceeding maximum size")
		}
		return nil, request.ErrorResourceTooLarge()
	}

	digestMD5 := base64.StdEncoding.EncodeToString(hasher.Sum(nil))
	if content.DigestMD5 != nil && *content.DigestMD5 != digestMD5 {
		if _, deleteErr := c.DeviceLogsUnstructuredStore().Delete(ctx, userID, *result.ID); deleteErr != nil {
			logger.WithError(deleteErr).Error("Unable to delete blob content with incorrect MD5 digest")
		}
		if _, destroyErr := repository.Destroy(ctx, *result.ID, nil); destroyErr != nil {
			logger.WithError(destroyErr).Error("Unable to destroy blob with incorrect MD5 digest")
		}
		return nil, errors.WithSource(request.ErrorDigestsNotEqual(*content.DigestMD5, digestMD5), structure.NewPointerSource().WithReference("digestMD5"))
	}

	update := blobStoreStructured.NewDeviceLogsUpdate()
	update.DigestMD5 = pointer.FromString(digestMD5)
	update.Size = pointer.FromInt(size)
	update.StartAt = pointer.FromTime(*content.StartAt)
	update.EndAt = pointer.FromTime(*content.EndAt)
	return repository.Update(ctx, *result.ID, nil, update)
}

func (c *Client) ListDeviceLogs(ctx context.Context, userID string, filter *blob.DeviceLogsFilter, pagination *page.Pagination) (blob.DeviceLogsBlobArray, error) {
	repository := c.BlobStructuredStore().NewDeviceLogsRepository()
	return repository.List(ctx, userID, filter, pagination)
}

func (c *Client) GetDeviceLogsBlob(ctx context.Context, deviceLogID string) (*blob.DeviceLogsBlob, error) {
	repository := c.BlobStructuredStore().NewDeviceLogsRepository()
	return repository.Get(ctx, deviceLogID)
}

func (c *Client) GetDeviceLogsContent(ctx context.Context, deviceLogID string) (*blob.DeviceLogsContent, error) {
	store := c.DeviceLogsUnstructuredStore()

	logMetadata, err := c.GetDeviceLogsBlob(ctx, deviceLogID)
	if err != nil {
		return nil, err
	} else if logMetadata == nil {
		return nil, nil
	}

	reader, err := store.Get(ctx, *logMetadata.UserID, *logMetadata.ID)
	if err != nil {
		return nil, err
	}
	if reader == nil {
		return nil, request.ErrorResourceNotFoundWithID(*logMetadata.ID)
	}

	return &blob.DeviceLogsContent{
		Body:      reader,
		DigestMD5: logMetadata.DigestMD5,
		MediaType: logMetadata.MediaType,
		StartAt:   logMetadata.StartAtTime,
		EndAt:     logMetadata.EndAtTime,
	}, nil
}

func (c *Client) DeleteAll(ctx context.Context, userID string) error {
	ctx = log.ContextWithField(ctx, "userId", userID)
	repository := c.BlobStructuredStore().NewBlobRepository()

	if deleted, err := repository.DeleteAll(ctx, userID); err != nil {
		return err
	} else if !deleted {
		return nil
	}

	if err := c.BlobUnstructuredStore().DeleteAll(ctx, userID); err != nil {
		return err
	}

	_, err := repository.DestroyAll(ctx, userID)
	return err
}

func (c *Client) Get(ctx context.Context, id string) (*blob.Blob, error) {
	repository := c.BlobStructuredStore().NewBlobRepository()
	return repository.Get(ctx, id, nil)
}

func (c *Client) GetContent(ctx context.Context, id string) (*blob.Content, error) {
	repository := c.BlobStructuredStore().NewBlobRepository()

	result, err := repository.Get(ctx, id, nil)
	if err != nil {
		return nil, err
	} else if result == nil {
		return nil, nil
	}

	reader, err := c.BlobUnstructuredStore().Get(ctx, *result.UserID, *result.ID)
	if err != nil {
		return nil, err
	}

	return &blob.Content{
		Body:      reader,
		DigestMD5: result.DigestMD5,
		MediaType: result.MediaType,
	}, nil
}

func (c *Client) Delete(ctx context.Context, id string, condition *request.Condition) (bool, error) {
	repository := c.BlobStructuredStore().NewBlobRepository()

	result, err := repository.Get(ctx, id, condition)
	if err != nil {
		return false, err
	} else if result == nil {
		return false, nil
	}

	deleted, err := repository.Delete(ctx, id, condition)
	if err != nil {
		return false, err
	} else if !deleted {
		return false, nil
	}

	exists, err := c.BlobUnstructuredStore().Delete(ctx, *result.UserID, *result.ID)
	if err != nil {
		return false, err
	} else if !exists {
		log.LoggerFromContext(ctx).WithField("id", id).Error("Deleting blob with no content")
	}

	return repository.Destroy(ctx, id, nil)
}

type SizeWriter struct {
	Size int
}

func NewSizeWriter() *SizeWriter {
	return &SizeWriter{}
}

func (s *SizeWriter) Write(bites []byte) (int, error) {
	length := len(bites)
	s.Size += length
	return length, nil
}
