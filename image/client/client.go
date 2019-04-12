package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/image"
	imageMultipart "github.com/tidepool-org/platform/image/multipart"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/request"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/user"
)

type Client struct {
	client      *platform.Client
	formEncoder imageMultipart.FormEncoder
}

func New(config *platform.Config, authorizeAs platform.AuthorizeAs, formEncoder imageMultipart.FormEncoder) (*Client, error) {
	client, err := platform.NewClient(config, authorizeAs)
	if err != nil {
		return nil, err
	}

	if formEncoder == nil {
		return nil, errors.New("form encoder is missing")
	}

	return &Client{
		client:      client,
		formEncoder: formEncoder,
	}, nil
}

func (c *Client) List(ctx context.Context, userID string, filter *image.Filter, pagination *page.Pagination) (image.ImageArray, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	} else if !user.IsValidID(userID) {
		return nil, errors.New("user id is invalid")
	}
	if filter == nil {
		filter = image.NewFilter()
	} else if err := structureValidator.New().Validate(filter); err != nil {
		return nil, errors.Wrap(err, "filter is invalid")
	}
	if pagination == nil {
		pagination = page.NewPagination()
	} else if err := structureValidator.New().Validate(pagination); err != nil {
		return nil, errors.Wrap(err, "pagination is invalid")
	}

	url := c.client.ConstructURL("v1", "users", userID, "images")
	result := image.ImageArray{}
	if err := c.client.RequestData(ctx, http.MethodGet, url, []request.RequestMutator{filter, pagination}, nil, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) Create(ctx context.Context, userID string, metadata *image.Metadata, contentIntent string, content *image.Content) (*image.Image, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	} else if !user.IsValidID(userID) {
		return nil, errors.New("user id is invalid")
	}
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

	reader, contentType := c.formEncoder.EncodeForm(metadata, contentIntent, content)
	if reader == nil {
		return nil, errors.New("multipart reader is missing")
	}
	defer reader.Close()
	if contentType == "" {
		return nil, errors.New("multipart content type is missing")
	}

	mutators := []request.RequestMutator{request.NewHeaderMutator("Content-Type", contentType)}

	url := c.client.ConstructURL("v1", "users", userID, "images")
	result := &image.Image{}
	if err := c.client.RequestData(ctx, http.MethodPost, url, mutators, reader, result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) CreateWithMetadata(ctx context.Context, userID string, metadata *image.Metadata) (*image.Image, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	} else if !user.IsValidID(userID) {
		return nil, errors.New("user id is invalid")
	}
	if metadata == nil {
		return nil, errors.New("metadata is missing")
	} else if err := structureValidator.New().Validate(metadata); err != nil {
		return nil, errors.Wrap(err, "metadata is invalid")
	}

	url := c.client.ConstructURL("v1", "users", userID, "images", "metadata")
	result := &image.Image{}
	if err := c.client.RequestData(ctx, http.MethodPost, url, nil, metadata, result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) CreateWithContent(ctx context.Context, userID string, contentIntent string, content *image.Content) (*image.Image, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	} else if !user.IsValidID(userID) {
		return nil, errors.New("user id is invalid")
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

	var mutators []request.RequestMutator
	if content.DigestMD5 != nil {
		mutators = append(mutators, request.NewHeaderMutator("Digest", fmt.Sprintf("MD5=%s", *content.DigestMD5)))
	}
	if content.MediaType != nil {
		mutators = append(mutators, request.NewHeaderMutator("Content-Type", *content.MediaType))
	}

	url := c.client.ConstructURL("v1", "users", userID, "images", "content", contentIntent)
	result := &image.Image{}
	if err := c.client.RequestData(ctx, http.MethodPost, url, mutators, content.Body, result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) DeleteAll(ctx context.Context, userID string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if userID == "" {
		return errors.New("user id is missing")
	} else if !user.IsValidID(userID) {
		return errors.New("user id is invalid")
	}

	url := c.client.ConstructURL("v1", "users", userID, "images")
	return c.client.RequestData(ctx, http.MethodDelete, url, nil, nil, nil)
}

func (c *Client) Get(ctx context.Context, id string) (*image.Image, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	} else if !image.IsValidID(id) {
		return nil, errors.New("id is invalid")
	}

	url := c.client.ConstructURL("v1", "images", id)
	result := &image.Image{}
	if err := c.client.RequestData(ctx, http.MethodGet, url, nil, nil, result); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return result, nil
}

func (c *Client) GetMetadata(ctx context.Context, id string) (*image.Metadata, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	} else if !image.IsValidID(id) {
		return nil, errors.New("id is invalid")
	}

	url := c.client.ConstructURL("v1", "images", id, "metadata")
	result := &image.Metadata{}
	if err := c.client.RequestData(ctx, http.MethodGet, url, nil, nil, result); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return result, nil
}

func (c *Client) GetContent(ctx context.Context, id string, mediaType *string) (*image.Content, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	} else if !image.IsValidID(id) {
		return nil, errors.New("id is invalid")
	}
	if mediaType != nil && !image.IsValidMediaType(*mediaType) {
		return nil, errors.New("media type is invalid")
	}

	var url string
	if mediaType != nil {
		extension, _ := image.ExtensionFromMediaType(*mediaType)
		url = c.client.ConstructURL("v1", "images", id, "content", fmt.Sprintf("content.%s", extension))
	} else {
		url = c.client.ConstructURL("v1", "images", id, "content")
	}

	headersInspector := request.NewHeadersInspector()
	body, err := c.client.RequestStream(ctx, http.MethodGet, url, nil, nil, headersInspector)
	if err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	digestMD5, err := request.ParseDigestMD5Header(headersInspector.Headers, "Digest")
	if err != nil {
		body.Close()
		return nil, err
	}
	mediaType, err = request.ParseMediaTypeHeader(headersInspector.Headers, "Content-Type")
	if err != nil {
		body.Close()
		return nil, err
	}

	return &image.Content{
		Body:      body,
		DigestMD5: digestMD5,
		MediaType: mediaType,
	}, nil
}

func (c *Client) GetRenditionContent(ctx context.Context, id string, rendition *image.Rendition) (*image.Content, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	} else if !image.IsValidID(id) {
		return nil, errors.New("id is invalid")
	}
	if rendition == nil {
		return nil, errors.New("rendition is missing")
	} else if err := structureValidator.New().Validate(rendition); err != nil {
		return nil, errors.Wrap(err, "rendition is invalid")
	}

	headersInspector := request.NewHeadersInspector()
	url := c.client.ConstructURL("v1", "images", id, "rendition", "content", rendition.String())
	body, err := c.client.RequestStream(ctx, http.MethodGet, url, nil, nil, headersInspector)
	if err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	digestMD5, err := request.ParseDigestMD5Header(headersInspector.Headers, "Digest")
	if err != nil {
		body.Close()
		return nil, err
	}
	mediaType, err := request.ParseMediaTypeHeader(headersInspector.Headers, "Content-Type")
	if err != nil {
		body.Close()
		return nil, err
	}

	return &image.Content{
		Body:      body,
		DigestMD5: digestMD5,
		MediaType: mediaType,
	}, nil
}

func (c *Client) PutMetadata(ctx context.Context, id string, condition *request.Condition, metadata *image.Metadata) (*image.Image, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	} else if !image.IsValidID(id) {
		return nil, errors.New("id is invalid")
	}
	if condition == nil {
		condition = request.NewCondition()
	} else if err := structureValidator.New().Validate(condition); err != nil {
		return nil, errors.Wrap(err, "condition is invalid")
	}
	if metadata == nil {
		return nil, errors.New("metadata is missing")
	} else if err := structureValidator.New().Validate(metadata); err != nil {
		return nil, errors.Wrap(err, "metadata is invalid")
	}

	url := c.client.ConstructURL("v1", "images", id, "metadata")
	result := &image.Image{}
	if err := c.client.RequestData(ctx, http.MethodPut, url, []request.RequestMutator{condition}, metadata, result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) PutContent(ctx context.Context, id string, condition *request.Condition, contentIntent string, content *image.Content) (*image.Image, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	} else if !image.IsValidID(id) {
		return nil, errors.New("id is invalid")
	}
	if condition == nil {
		condition = request.NewCondition()
	} else if err := structureValidator.New().Validate(condition); err != nil {
		return nil, errors.Wrap(err, "condition is invalid")
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

	mutators := []request.RequestMutator{condition}
	if content.DigestMD5 != nil {
		mutators = append(mutators, request.NewHeaderMutator("Digest", fmt.Sprintf("MD5=%s", *content.DigestMD5)))
	}
	if content.MediaType != nil {
		mutators = append(mutators, request.NewHeaderMutator("Content-Type", *content.MediaType))
	}

	url := c.client.ConstructURL("v1", "images", id, "content", contentIntent)
	result := &image.Image{}
	if err := c.client.RequestData(ctx, http.MethodPut, url, mutators, content.Body, result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) Delete(ctx context.Context, id string, condition *request.Condition) (bool, error) {
	if ctx == nil {
		return false, errors.New("context is missing")
	}
	if id == "" {
		return false, errors.New("id is missing")
	} else if !image.IsValidID(id) {
		return false, errors.New("id is invalid")
	}
	if condition == nil {
		condition = request.NewCondition()
	} else if err := structureValidator.New().Validate(condition); err != nil {
		return false, errors.Wrap(err, "condition is invalid")
	}

	url := c.client.ConstructURL("v1", "images", id)
	if err := c.client.RequestData(ctx, http.MethodDelete, url, []request.RequestMutator{condition}, nil, nil); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
