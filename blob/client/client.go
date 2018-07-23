package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/tidepool-org/platform/blob"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/request"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/user"
)

type Client struct {
	client *platform.Client
}

func New(config *platform.Config, authorizeAs platform.AuthorizeAs) (*Client, error) {
	client, err := platform.NewClient(config, authorizeAs)
	if err != nil {
		return nil, err
	}

	return &Client{
		client: client,
	}, nil
}

func (c *Client) List(ctx context.Context, userID string, filter *blob.Filter, pagination *page.Pagination) (blob.Blobs, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	} else if !user.IsValidID(userID) {
		return nil, errors.New("user id is invalid")
	}
	if filter == nil {
		filter = blob.NewFilter()
	} else if err := structureValidator.New().Validate(filter); err != nil {
		return nil, errors.Wrap(err, "filter is invalid")
	}
	if pagination == nil {
		pagination = page.NewPagination()
	} else if err := structureValidator.New().Validate(pagination); err != nil {
		return nil, errors.Wrap(err, "pagination is invalid")
	}

	url := c.client.ConstructURL("v1", "users", userID, "blobs")
	result := blob.Blobs{}
	if err := c.client.RequestData(ctx, http.MethodGet, url, []request.RequestMutator{filter, pagination}, nil, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) Create(ctx context.Context, userID string, create *blob.Create) (*blob.Blob, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	} else if !user.IsValidID(userID) {
		return nil, errors.New("user id is invalid")
	}
	if create == nil {
		return nil, errors.New("create is missing")
	} else if err := structureValidator.New().Validate(create); err != nil {
		return nil, errors.Wrap(err, "create is invalid")
	}

	var mutators []request.RequestMutator
	if create.DigestMD5 != nil {
		mutators = append(mutators, request.NewHeaderMutator("Digest", fmt.Sprintf("MD5=%s", *create.DigestMD5)))
	}
	if create.MediaType != nil {
		mutators = append(mutators, request.NewHeaderMutator("Content-Type", *create.MediaType))
	}

	url := c.client.ConstructURL("v1", "users", userID, "blobs")
	result := &blob.Blob{}
	if err := c.client.RequestData(ctx, http.MethodPost, url, mutators, create.Body, result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) Get(ctx context.Context, id string) (*blob.Blob, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	} else if !blob.IsValidID(id) {
		return nil, errors.New("id is invalid")
	}

	url := c.client.ConstructURL("v1", "blobs", id)
	result := &blob.Blob{}
	if err := c.client.RequestData(ctx, http.MethodGet, url, nil, nil, result); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return result, nil
}

func (c *Client) GetContent(ctx context.Context, id string) (*blob.Content, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	} else if !blob.IsValidID(id) {
		return nil, errors.New("id is invalid")
	}

	headersInspector := request.NewHeadersInspector()
	url := c.client.ConstructURL("v1", "blobs", id, "content")
	body, err := c.client.RequestStream(ctx, http.MethodGet, url, nil, nil, headersInspector)
	if err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	digestMD5, err := request.ParseDigestMD5Header(headersInspector.Headers, "Digest")
	if err != nil {
		return nil, err
	}
	mediaType, err := request.ParseMediaTypeHeader(headersInspector.Headers, "Content-Type")
	if err != nil {
		return nil, err
	}
	size, err := request.ParseIntHeader(headersInspector.Headers, "Content-Length")
	if err != nil {
		return nil, err
	}

	return &blob.Content{
		Body:      body,
		DigestMD5: digestMD5,
		MediaType: mediaType,
		Size:      size,
	}, nil
}

func (c *Client) Delete(ctx context.Context, id string) (bool, error) {
	if ctx == nil {
		return false, errors.New("context is missing")
	}
	if id == "" {
		return false, errors.New("id is missing")
	} else if !blob.IsValidID(id) {
		return false, errors.New("id is invalid")
	}

	url := c.client.ConstructURL("v1", "blobs", id)
	if err := c.client.RequestData(ctx, http.MethodDelete, url, nil, nil, nil); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
