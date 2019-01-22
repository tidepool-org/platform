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
	storeUnstructured "github.com/tidepool-org/platform/store/unstructured"
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

// FUTURE: Return ErrorResourceNotFoundWithID(userID) if userID does not exist at all

func (c *Client) List(ctx context.Context, userID string, filter *blob.Filter, pagination *page.Pagination) (blob.BlobArray, error) {
	if err := c.AuthClient().EnsureAuthorizedService(ctx); err != nil {
		return nil, err
	}

	session := c.BlobStructuredStore().NewSession()
	defer session.Close()

	return session.List(ctx, userID, filter, pagination)
}

func (c *Client) Create(ctx context.Context, userID string, content *blob.Content) (*blob.Blob, error) {
	if _, err := c.AuthClient().EnsureAuthorizedUser(ctx, userID, permission.Write); err != nil {
		return nil, err
	}

	if content == nil {
		return nil, errors.New("content is missing")
	} else if err := structureValidator.New().Validate(content); err != nil {
		return nil, errors.Wrap(err, "content is invalid")
	}

	session := c.BlobStructuredStore().NewSession()
	defer session.Close()

	structuredCreate := blobStoreStructured.NewCreate()
	structuredCreate.MediaType = pointer.CloneString(content.MediaType)
	result, err := session.Create(ctx, userID, structuredCreate)
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
		if _, destroyErr := session.Destroy(ctx, *result.ID, nil); destroyErr != nil {
			logger.WithError(destroyErr).Error("Unable to destroy blob after failure to put blob content")
		}
		return nil, err
	}

	size := sizer.Size
	if size > blob.SizeMaximum {
		if _, deleteErr := c.BlobUnstructuredStore().Delete(ctx, userID, *result.ID); deleteErr != nil {
			logger.WithError(deleteErr).Error("Unable to delete blob content exceeding maximum size")
		}
		if _, destroyErr := session.Destroy(ctx, *result.ID, nil); destroyErr != nil {
			logger.WithError(destroyErr).Error("Unable to destroy blob exceeding maximum size")
		}
		return nil, request.ErrorResourceTooLarge()
	}

	digestMD5 := base64.StdEncoding.EncodeToString(hasher.Sum(nil))
	if content.DigestMD5 != nil && *content.DigestMD5 != digestMD5 {
		if _, deleteErr := c.BlobUnstructuredStore().Delete(ctx, userID, *result.ID); deleteErr != nil {
			logger.WithError(deleteErr).Error("Unable to delete blob content with incorrect MD5 digest")
		}
		if _, destroyErr := session.Destroy(ctx, *result.ID, nil); destroyErr != nil {
			logger.WithError(destroyErr).Error("Unable to destroy blob with incorrect MD5 digest")
		}
		return nil, errors.WithSource(request.ErrorDigestsNotEqual(*content.DigestMD5, digestMD5), structure.NewPointerSource().WithReference("digestMD5"))
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

	return session.Destroy(ctx, id, nil)
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
