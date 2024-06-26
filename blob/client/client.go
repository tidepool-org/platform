package client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/tidepool-org/platform/blob"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
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

func (c *Client) List(ctx context.Context, userID string, filter *blob.Filter, pagination *page.Pagination) (blob.BlobArray, error) {
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
	result := blob.BlobArray{}
	if err := c.client.RequestData(ctx, http.MethodGet, url, []request.RequestMutator{filter, pagination}, nil, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) Create(ctx context.Context, userID string, content *blob.Content) (*blob.Blob, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	} else if !user.IsValidID(userID) {
		return nil, errors.New("user id is invalid")
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

	url := c.client.ConstructURL("v1", "users", userID, "blobs")
	result := &blob.Blob{}
	if err := c.client.RequestData(ctx, http.MethodPost, url, mutators, content.Body, result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) ListDeviceLogs(ctx context.Context, userID string, filter *blob.DeviceLogsFilter, pagination *page.Pagination) (blob.DeviceLogsBlobArray, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	}
	if filter == nil {
		filter = blob.NewDeviceLogsFilter()
	} else if err := structureValidator.New().Validate(filter); err != nil {
		return nil, errors.Wrap(err, "filter is invalid")
	}
	if pagination == nil {
		pagination = page.NewPagination()
	} else if err := structureValidator.New().Validate(pagination); err != nil {
		return nil, errors.Wrap(err, "pagination is invalid")
	}

	url := c.client.ConstructURL("v1", "users", userID, "device_logs")
	var result blob.DeviceLogsBlobArray
	if err := c.client.RequestData(ctx, http.MethodGet, url, []request.RequestMutator{filter, pagination}, nil, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) CreateDeviceLogs(ctx context.Context, userID string, content *blob.DeviceLogsContent) (*blob.DeviceLogsBlob, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	} else if !user.IsValidID(userID) {
		return nil, errors.New("user id is invalid")
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
	if content.StartAt != nil {
		mutators = append(mutators, request.NewHeaderMutator("X-Logs-Start-At-Time", content.StartAt.Format(time.RFC3339)))
	}
	if content.EndAt != nil {
		mutators = append(mutators, request.NewHeaderMutator("X-Logs-End-At-Time", content.EndAt.Format(time.RFC3339)))
	}

	url := c.client.ConstructURL("v1", "users", userID, "device_logs")
	result := &blob.DeviceLogsBlob{}
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

	url := c.client.ConstructURL("v1", "users", userID, "blobs")
	return c.client.RequestData(ctx, http.MethodDelete, url, nil, nil, nil)
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

	url := c.client.ConstructURL("v1", "blobs", id, "content")
	headersInspector := request.NewHeadersInspector(log.LoggerFromContext(ctx))
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

	return &blob.Content{
		Body:      body,
		DigestMD5: digestMD5,
		MediaType: mediaType,
	}, nil
}

func (c *Client) GetDeviceLogsBlob(ctx context.Context, deviceLogID string) (*blob.DeviceLogsBlob, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if deviceLogID == "" {
		return nil, errors.New("deviceLogID is missing")
	} else if !blob.IsValidID(deviceLogID) {
		return nil, errors.New("deviceLogID is invalid")
	}

	url := c.client.ConstructURL("v1", "device_logs", deviceLogID)
	var result blob.DeviceLogsBlob
	if err := c.client.RequestData(ctx, http.MethodGet, url, nil, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetDeviceLogsContent(ctx context.Context, deviceLogID string) (*blob.DeviceLogsContent, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if deviceLogID == "" {
		return nil, errors.New("deviceLogID is missing")
	} else if !blob.IsValidID(deviceLogID) {
		return nil, errors.New("deviceLogID is invalid")
	}

	url := c.client.ConstructURL("v1", "device_logs", deviceLogID)

	headersInspector := request.NewHeadersInspector(log.LoggerFromContext(ctx))
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
	startAt, err := request.ParseTimeHeader(headersInspector.Headers, "Start-At", time.RFC3339Nano)
	if err != nil {
		return nil, err
	}
	endAt, err := request.ParseTimeHeader(headersInspector.Headers, "End-At", time.RFC3339Nano)
	if err != nil {
		return nil, err
	}

	return &blob.DeviceLogsContent{
		Body:      body,
		DigestMD5: digestMD5,
		MediaType: mediaType,
		StartAt:   startAt,
		EndAt:     endAt,
	}, nil

}

func (c *Client) Delete(ctx context.Context, id string, condition *request.Condition) (bool, error) {
	if ctx == nil {
		return false, errors.New("context is missing")
	}
	if id == "" {
		return false, errors.New("id is missing")
	} else if !blob.IsValidID(id) {
		return false, errors.New("id is invalid")
	}
	if condition == nil {
		condition = request.NewCondition()
	} else if err := structureValidator.New().Validate(condition); err != nil {
		return false, errors.Wrap(err, "condition is invalid")
	}

	url := c.client.ConstructURL("v1", "blobs", id)
	if err := c.client.RequestData(ctx, http.MethodDelete, url, []request.RequestMutator{condition}, nil, nil); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
