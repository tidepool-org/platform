package client

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"io"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/blob"
	blobStoreStructured "github.com/tidepool-org/platform/blob/store/structured"
	blobStoreUnstructured "github.com/tidepool-org/platform/blob/store/unstructured"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type Provider interface {
	AuthClient() auth.Client
	BlobStructuredStore() blobStoreStructured.Store
	BlobUnstructuredStore() blobStoreUnstructured.Store
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

func (c *Client) List(ctx context.Context, userID string, filter *blob.Filter, pagination *page.Pagination) (blob.Blobs, error) {
	if err := c.AuthClient().EnsureAuthorizedService(ctx); err != nil {
		return nil, err
	}

	session := c.BlobStructuredStore().NewSession()
	defer session.Close()

	return session.List(ctx, userID, filter, pagination)
}

func (c *Client) Create(ctx context.Context, userID string, create *blob.Create) (*blob.Blob, error) {
	if _, err := c.AuthClient().EnsureAuthorizedUser(ctx, userID, permission.Write); err != nil {
		return nil, err
	}

	if create == nil {
		return nil, errors.New("create is missing")
	} else if err := structureValidator.New().Validate(create); err != nil {
		return nil, errors.Wrap(err, "create is invalid")
	}

	session := c.BlobStructuredStore().NewSession()
	defer session.Close()

	structuredCreate := blobStoreStructured.NewCreate()
	structuredCreate.MediaType = pointer.CloneString(create.MediaType)
	result, err := session.Create(ctx, userID, structuredCreate)
	if err != nil {
		return nil, err
	}

	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": userID, "id": *result.ID})

	hasher := md5.New()
	sizer := NewSizeWriter()
	err = c.BlobUnstructuredStore().Put(ctx, userID, *result.ID, io.TeeReader(io.TeeReader(create.Body, hasher), sizer))
	if err != nil {
		if _, deleteErr := session.Delete(ctx, *result.ID, nil); deleteErr != nil {
			logger.WithError(deleteErr).Error("Unable to delete blob after failure to put blob content")
		}
		return nil, err
	}

	// FUTURE: Consider Digest struct that pulls apart and manages digest

	digestMD5 := base64.StdEncoding.EncodeToString(hasher.Sum(nil))
	if create.DigestMD5 != nil && *create.DigestMD5 != digestMD5 {
		if _, deleteErr := c.BlobUnstructuredStore().Delete(ctx, userID, *result.ID); deleteErr != nil {
			logger.WithError(deleteErr).Error("Unable to delete blob content with incorrect MD5 digest")
		}
		if _, deleteErr := session.Delete(ctx, *result.ID, nil); deleteErr != nil {
			logger.WithError(deleteErr).Error("Unable to delete blob with incorrect MD5 digest")
		}
		return nil, errors.WithSource(blob.ErrorDigestsNotEqual(*create.DigestMD5, digestMD5), structure.NewPointerSource().WithReference("digestMD5"))
	}

	size := sizer.Size
	if size > blob.SizeMaximum {
		if _, deleteErr := c.BlobUnstructuredStore().Delete(ctx, userID, *result.ID); deleteErr != nil {
			logger.WithError(deleteErr).Error("Unable to delete blob content exceeding maximum size")
		}
		if _, deleteErr := session.Delete(ctx, *result.ID, nil); deleteErr != nil {
			logger.WithError(deleteErr).Error("Unable to delete blob exceeding maximum size")
		}
		return nil, errors.WithSource(structureValidator.ErrorValueNotLessThanOrEqualTo(size, blob.SizeMaximum), structure.NewPointerSource().WithReference("size"))
	}

	update := blobStoreStructured.NewUpdate()
	update.DigestMD5 = pointer.FromString(digestMD5)
	update.Size = pointer.FromInt(size)
	update.Status = pointer.FromString(blob.StatusAvailable)
	return session.Update(ctx, *result.ID, nil, update)
}

func (c *Client) Get(ctx context.Context, id string) (*blob.Blob, error) {
	if err := c.AuthClient().EnsureAuthorizedService(ctx); err != nil {
		return nil, err
	}

	session := c.BlobStructuredStore().NewSession()
	defer session.Close()

	return session.Get(ctx, id)
}

func (c *Client) GetContent(ctx context.Context, id string) (*blob.Content, error) {
	if err := c.AuthClient().EnsureAuthorizedService(ctx); err != nil {
		return nil, err
	}

	session := c.BlobStructuredStore().NewSession()
	defer session.Close()

	result, err := session.Get(ctx, id)
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
		Size:      result.Size,
	}, nil
}

func (c *Client) Delete(ctx context.Context, id string, condition *request.Condition) (bool, error) {
	if err := c.AuthClient().EnsureAuthorizedService(ctx); err != nil {
		return false, err
	}

	session := c.BlobStructuredStore().NewSession()
	defer session.Close()

	result, err := session.Get(ctx, id)
	if err != nil {
		return false, err
	} else if result == nil {
		return false, nil
	} else if condition != nil && condition.Revision != nil && *condition.Revision != *result.Revision {
		return false, nil
	}

	exists, err := c.BlobUnstructuredStore().Delete(ctx, *result.UserID, *result.ID)
	if err != nil {
		return false, err
	} else if !exists {
		log.LoggerFromContext(ctx).WithField("id", id).Error("Deleting blob with no content")
	}

	return session.Delete(ctx, id, nil)
}

type SizeWriter struct {
	Size int
}

func NewSizeWriter() *SizeWriter {
	return &SizeWriter{}
}

func (s *SizeWriter) Write(bytes []byte) (int, error) {
	length := len(bytes)
	s.Size += length
	return length, nil
}
