package client

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	stdlibImage "image"
	_ "image/jpeg" // Required for JPEG
	_ "image/png"  // Required for PNG
	"io"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/image"
	imageStoreStructured "github.com/tidepool-org/platform/image/store/structured"
	imageStoreUnstructured "github.com/tidepool-org/platform/image/store/unstructured"
	imageTransform "github.com/tidepool-org/platform/image/transform"
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
	ImageStructuredStore() imageStoreStructured.Store
	ImageUnstructuredStore() imageStoreUnstructured.Store
	ImageTransformer() imageTransform.Transformer
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

func (c *Client) List(ctx context.Context, userID string, filter *image.Filter, pagination *page.Pagination) (image.ImageArray, error) {
	ctx = log.ContextWithFields(ctx, log.Fields{"userId": userID, "filter": filter, "pagination": pagination})

	if _, err := c.AuthClient().EnsureAuthorizedUser(ctx, userID, permission.Read); err != nil {
		return nil, err
	}

	repository := c.ImageStructuredStore().NewImageRepository()
	return repository.List(ctx, userID, filter, pagination)
}

func (c *Client) Create(ctx context.Context, userID string, metadata *image.Metadata, contentIntent string, content *image.Content) (*image.Image, error) {
	ctx = log.ContextWithFields(ctx, log.Fields{"userId": userID, "metadata": metadata, "contentIntent": contentIntent, "content": content})

	if _, err := c.AuthClient().EnsureAuthorizedUser(ctx, userID, permission.Write); err != nil {
		return nil, err
	}

	return c.createWithMetadataAndContent(ctx, userID, metadata, contentIntent, content)
}

func (c *Client) CreateWithMetadata(ctx context.Context, userID string, metadata *image.Metadata) (*image.Image, error) {
	ctx = log.ContextWithFields(ctx, log.Fields{"userId": userID, "metadata": metadata})

	if _, err := c.AuthClient().EnsureAuthorizedUser(ctx, userID, permission.Write); err != nil {
		return nil, err
	}

	repository := c.ImageStructuredStore().NewImageRepository()
	return repository.Create(ctx, userID, metadata)
}

func (c *Client) CreateWithContent(ctx context.Context, userID string, contentIntent string, content *image.Content) (*image.Image, error) {
	ctx = log.ContextWithFields(ctx, log.Fields{"userId": userID, "contentIntent": contentIntent, "content": content})

	if _, err := c.AuthClient().EnsureAuthorizedUser(ctx, userID, permission.Write); err != nil {
		return nil, err
	}

	return c.createWithMetadataAndContent(ctx, userID, image.NewMetadata(), contentIntent, content)
}

func (c *Client) createWithMetadataAndContent(ctx context.Context, userID string, metadata *image.Metadata, contentIntent string, content *image.Content) (*image.Image, error) {
	ctx = log.ContextWithFields(ctx, log.Fields{"userId": userID, "metadata": metadata, "contentIntent": contentIntent, "content": content})

	if metadata == nil {
		return nil, errors.New("metadata is missing")
	} else if err := structureValidator.New().Validate(metadata); err != nil {
		return nil, errors.Wrap(err, "metadata is invalid")
	}
	if contentIntent == "" {
		return nil, errors.New("content intent is missing")
	} else if !image.IsValidContentIntent(contentIntent) {
		return nil, errors.New("content intent is invalid")
	}
	if content == nil {
		return nil, errors.New("content is missing")
	} else if err := structureValidator.New().Validate(content); err != nil {
		return nil, errors.Wrap(err, "content is invalid")
	}

	collecton := c.ImageStructuredStore().NewImageRepository()
	original, err := collecton.Create(ctx, userID, metadata)
	if err != nil {
		return nil, err
	}

	ctx, logger := log.ContextAndLoggerWithField(ctx, "id", *original.ID)

	updated, err := c.putContent(ctx, collecton, original, contentIntent, content)
	if err != nil {
		if _, destroyErr := collecton.Destroy(ctx, *original.ID, nil); destroyErr != nil {
			logger.WithError(destroyErr).Error("Unable to destroy image after failure to put image content")
		}
		return nil, err
	}

	return updated, nil
}

func (c *Client) DeleteAll(ctx context.Context, userID string) error {
	ctx = log.ContextWithField(ctx, "userId", userID)

	if err := c.AuthClient().EnsureAuthorizedService(ctx); err != nil {
		return err
	}

	repository := c.ImageStructuredStore().NewImageRepository()
	if deleted, err := repository.DeleteAll(ctx, userID); err != nil {
		return err
	} else if !deleted {
		return nil
	}

	if err := c.ImageUnstructuredStore().DeleteAll(ctx, userID); err != nil {
		return err
	}

	_, err := repository.DestroyAll(ctx, userID)
	return err
}

func (c *Client) Get(ctx context.Context, id string) (*image.Image, error) {
	ctx = log.ContextWithField(ctx, "id", id)

	repository := c.ImageStructuredStore().NewImageRepository()
	result, err := repository.Get(ctx, id, nil)
	if err != nil {
		return nil, err
	} else if result == nil {
		return nil, nil
	}

	if _, err = c.AuthClient().EnsureAuthorizedUser(ctx, *result.UserID, permission.Write); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) GetMetadata(ctx context.Context, id string) (*image.Metadata, error) {
	ctx = log.ContextWithField(ctx, "id", id)

	repository := c.ImageStructuredStore().NewImageRepository()

	result, err := repository.Get(ctx, id, nil)
	if err != nil {
		return nil, err
	} else if result == nil {
		return nil, nil
	}

	if _, err = c.AuthClient().EnsureAuthorizedUser(ctx, *result.UserID, permission.Write); err != nil {
		return nil, err
	}

	metadata := result.Metadata
	if metadata == nil {
		metadata = image.NewMetadata()
	}

	return metadata, nil
}

func (c *Client) GetContent(ctx context.Context, id string, mediaType *string) (*image.Content, error) {
	ctx = log.ContextWithFields(ctx, log.Fields{"id": id, "mediaType": mediaType})

	repository := c.ImageStructuredStore().NewImageRepository()

	result, err := repository.Get(ctx, id, nil)
	if err != nil {
		return nil, err
	} else if result == nil {
		return nil, nil
	}

	if _, err = c.AuthClient().EnsureAuthorizedUser(ctx, *result.UserID, permission.Read); err != nil {
		return nil, err
	}

	if !result.HasContent() {
		return nil, nil
	} else if mediaType != nil && *mediaType != *result.ContentAttributes.MediaType {
		return nil, nil
	}

	reader, err := c.ImageUnstructuredStore().GetContent(ctx, *result.UserID, *result.ID, *result.ContentID, *result.ContentIntent)
	if err != nil {
		return nil, err
	}

	return &image.Content{
		Body:      reader,
		DigestMD5: result.ContentAttributes.DigestMD5,
		MediaType: result.ContentAttributes.MediaType,
	}, nil
}

func (c *Client) GetRenditionContent(ctx context.Context, id string, rendition *image.Rendition) (*image.Content, error) {
	ctx = log.ContextWithFields(ctx, log.Fields{"id": id, "rendition": rendition})

	repository := c.ImageStructuredStore().NewImageRepository()

	result, err := repository.Get(ctx, id, nil)
	if err != nil {
		return nil, err
	} else if result == nil {
		return nil, nil
	}

	if _, err = c.AuthClient().EnsureAuthorizedUser(ctx, *result.UserID, permission.Read); err != nil {
		return nil, err
	}

	ctx = log.ContextWithField(ctx, "contentId", *result.ContentID)

	if !result.HasContent() {
		return nil, nil
	}

	transform, err := c.ImageTransformer().CalculateTransform(result.ContentAttributes, rendition)
	if err != nil {
		return nil, err
	}

	ctx = log.ContextWithField(ctx, "transform", transform)

	renditionsID := result.RenditionsID
	renditionString := transform.Rendition.String()

	if !result.HasRendition(transform.Rendition) {
		var contentReader io.ReadCloser
		contentReader, err = c.ImageUnstructuredStore().GetContent(ctx, *result.UserID, *result.ID, *result.ContentID, *result.ContentIntent)
		if err != nil {
			return nil, err
		}
		defer contentReader.Close()

		var renditionReader io.ReadCloser
		renditionReader, err = c.ImageTransformer().TransformContent(contentReader, transform)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to transform image content")
		}
		defer renditionReader.Close()

		if renditionsID == nil || (result.Renditions != nil && len(*result.Renditions) >= image.RenditionsLengthMaximum) {
			renditionsID = pointer.FromString(image.NewRenditionsID())
		}

		var logger log.Logger
		ctx, logger = log.ContextAndLoggerWithField(ctx, "renditionsId", *renditionsID)

		options := storeUnstructured.NewOptions()
		options.MediaType = transform.Rendition.MediaType
		if err = c.ImageUnstructuredStore().PutRenditionContent(ctx, *result.UserID, *result.ID, *result.ContentID, *renditionsID, renditionString, renditionReader, options); err != nil {
			return nil, err
		}

		condition := request.NewCondition()
		condition.Revision = result.Revision
		update := imageStoreStructured.NewUpdate()
		if result.RenditionsID == nil || *result.RenditionsID != *renditionsID {
			update.RenditionsID = renditionsID
		}
		update.Rendition = pointer.FromString(renditionString)
		if updated, updateErr := repository.Update(ctx, *result.ID, condition, update); updateErr != nil || updated == nil {
			logger.WithError(updateErr).Error("Unable to update image with rendition; orphaned rendition")
		} else if result.RenditionsID != nil && *result.RenditionsID != *renditionsID {
			logger.Error("Deleting excess image rendition content")
			if deleteErr := c.ImageUnstructuredStore().DeleteRenditionContent(ctx, *result.UserID, *result.ID, *result.ContentID, *result.RenditionsID); deleteErr != nil {
				logger.WithError(deleteErr).Error("Unable to delete excess image rendition content")
			}
		}
	}

	ctx = log.ContextWithField(ctx, "renditionsId", *renditionsID)

	reader, err := c.ImageUnstructuredStore().GetRenditionContent(ctx, *result.UserID, *result.ID, *result.ContentID, *renditionsID, renditionString)
	if err != nil {
		return nil, err
	}

	return &image.Content{
		Body:      reader,
		MediaType: rendition.MediaType,
	}, nil
}

func (c *Client) PutMetadata(ctx context.Context, id string, condition *request.Condition, metadata *image.Metadata) (*image.Image, error) {
	ctx = log.ContextWithFields(ctx, log.Fields{"id": id, "condition": condition, "metadata": metadata})

	repository := c.ImageStructuredStore().NewImageRepository()

	result, err := repository.Get(ctx, id, condition)
	if err != nil {
		return nil, err
	} else if result == nil {
		return nil, nil
	}

	if _, err = c.AuthClient().EnsureAuthorizedUser(ctx, *result.UserID, permission.Write); err != nil {
		return nil, err
	}

	if metadata == nil {
		return nil, errors.New("metadata is missing")
	} else if err = structureValidator.New().Validate(metadata); err != nil {
		return nil, errors.Wrap(err, "metadata is invalid")
	}

	update := imageStoreStructured.NewUpdate()
	update.Metadata = metadata

	return repository.Update(ctx, id, condition, update)
}

func (c *Client) PutContent(ctx context.Context, id string, condition *request.Condition, contentIntent string, content *image.Content) (*image.Image, error) {
	ctx = log.ContextWithFields(ctx, log.Fields{"id": id, "condition": condition, "contentIntent": contentIntent, "content": content})

	repository := c.ImageStructuredStore().NewImageRepository()

	original, err := repository.Get(ctx, id, condition)
	if err != nil {
		return nil, err
	} else if original == nil {
		return nil, nil
	}

	if _, err = c.AuthClient().EnsureAuthorizedUser(ctx, *original.UserID, permission.Write); err != nil {
		return nil, err
	}

	if contentIntent == "" {
		return nil, errors.New("content intent is missing")
	} else if !image.IsValidContentIntent(contentIntent) {
		return nil, errors.New("content intent is invalid")
	} else if original.HasContent() && (*original.ContentIntent == image.ContentIntentOriginal || (*original.ContentIntent == image.ContentIntentAlternate && contentIntent == image.ContentIntentAlternate)) {
		return nil, image.ErrorImageContentIntentUnexpected(contentIntent)
	}
	if content == nil {
		return nil, errors.New("content is missing")
	} else if err = structureValidator.New().Validate(content); err != nil {
		return nil, errors.Wrap(err, "content is invalid")
	}

	return c.putContent(ctx, repository, original, contentIntent, content)
}

func (c *Client) putContent(ctx context.Context, repository imageStoreStructured.ImageRepository, original *image.Image, contentIntent string, content *image.Content) (*image.Image, error) {
	ctx = log.ContextWithFields(ctx, log.Fields{"userId": *original.UserID, "id": *original.ID, "contentIntent": contentIntent, "content": content})

	contentID := image.NewContentID()
	ctx = log.ContextWithField(ctx, "contentId", contentID)

	bodyBuffer := &bytes.Buffer{}
	body := io.TeeReader(io.LimitReader(content.Body, image.SizeMaximum+1), bodyBuffer)

	mediaType, width, height, err := c.decodeConfig(body)
	if err != nil {
		return nil, err
	} else if mediaType != *content.MediaType {
		return nil, image.ErrorImageMalformed("header does not match media type")
	}

	ctx, logger := log.ContextAndLoggerWithFields(ctx, log.Fields{"mediaType": mediaType, "width": width, "height": height})

	hasher := md5.New()
	sizer := NewSizeWriter()
	options := storeUnstructured.NewOptions()
	options.MediaType = content.MediaType
	err = c.ImageUnstructuredStore().PutContent(ctx, *original.UserID, *original.ID, contentID, contentIntent, io.TeeReader(io.TeeReader(io.MultiReader(bodyBuffer, body), hasher), sizer), options)
	if err != nil {
		return nil, err
	}

	size := sizer.Size
	if size > image.SizeMaximum {
		if deleteErr := c.ImageUnstructuredStore().DeleteContent(ctx, *original.UserID, *original.ID, contentID); deleteErr != nil {
			logger.WithError(deleteErr).Error("Unable to delete image content exceeding maximum size")
		}
		return nil, request.ErrorResourceTooLarge()
	}

	ctx, logger = log.ContextAndLoggerWithField(ctx, "size", size)

	digestMD5 := base64.StdEncoding.EncodeToString(hasher.Sum(nil))
	if content.DigestMD5 != nil && *content.DigestMD5 != digestMD5 {
		if deleteErr := c.ImageUnstructuredStore().DeleteContent(ctx, *original.UserID, *original.ID, contentID); deleteErr != nil {
			logger.WithError(deleteErr).Error("Unable to delete image content with incorrect MD5 digest")
		}
		return nil, errors.WithSource(request.ErrorDigestsNotEqual(*content.DigestMD5, digestMD5), structure.NewPointerSource().WithReference("digestMD5"))
	}

	ctx, logger = log.ContextAndLoggerWithField(ctx, "digestMD5", digestMD5)

	condition := request.NewCondition()
	condition.Revision = original.Revision
	update := imageStoreStructured.NewUpdate()
	update.ContentID = pointer.FromString(contentID)
	update.ContentIntent = pointer.FromString(contentIntent)
	update.ContentAttributes = imageStoreStructured.NewContentAttributes()
	update.ContentAttributes.DigestMD5 = pointer.FromString(digestMD5)
	update.ContentAttributes.MediaType = pointer.CloneString(content.MediaType)
	update.ContentAttributes.Width = pointer.FromInt(width)
	update.ContentAttributes.Height = pointer.FromInt(height)
	update.ContentAttributes.Size = pointer.FromInt(size)
	updated, err := repository.Update(ctx, *original.ID, condition, update)
	if err != nil {
		if deleteErr := c.ImageUnstructuredStore().DeleteContent(ctx, *original.UserID, *original.ID, contentID); deleteErr != nil {
			logger.WithError(deleteErr).Error("Unable to delete image content for failed update")
		}
		return nil, err
	}

	if original.HasContent() {
		if deleteErr := c.ImageUnstructuredStore().DeleteContent(ctx, *original.UserID, *original.ID, *original.ContentID); deleteErr != nil {
			logger.WithError(deleteErr).Error("Unable to delete image content for previous content intent")
		}
	}

	return updated, nil
}

func (c *Client) Delete(ctx context.Context, id string, condition *request.Condition) (bool, error) {
	ctx = log.ContextWithFields(ctx, log.Fields{"id": id, "condition": condition})

	repository := c.ImageStructuredStore().NewImageRepository()

	result, err := repository.Get(ctx, id, condition)
	if err != nil {
		return false, err
	} else if result == nil {
		return false, nil
	}

	if _, err = c.AuthClient().EnsureAuthorizedUser(ctx, *result.UserID, permission.Write); err != nil {
		return false, err
	}

	deleted, err := repository.Delete(ctx, id, condition)
	if err != nil {
		return false, err
	} else if !deleted {
		return false, nil
	}

	if err = c.ImageUnstructuredStore().Delete(ctx, *result.UserID, *result.ID); err != nil {
		return false, err
	}

	return repository.Destroy(ctx, id, nil)
}

func (c *Client) decodeConfig(reader io.Reader) (string, int, int, error) {
	img, format, err := stdlibImage.Decode(reader)
	if err != nil {
		return "", 0, 0, image.ErrorImageMalformed(fmt.Sprintf("unable to decode image; %s", err))
	}
	size := img.Bounds().Size()
	return fmt.Sprintf("image/%s", format), size.X, size.Y, nil
}

type peeker interface {
	Peek(n int) ([]byte, error)
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
