package client

import (
	"context"
	"net/http"

	dataSource "github.com/tidepool-org/platform/data/source"
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

func (c *Client) List(ctx context.Context, userID string, filter *dataSource.Filter, pagination *page.Pagination) (dataSource.SourceArray, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	} else if !user.IsValidID(userID) {
		return nil, errors.New("user id is invalid")
	}
	if filter == nil {
		filter = dataSource.NewFilter()
	} else if err := structureValidator.New().Validate(filter); err != nil {
		return nil, errors.Wrap(err, "filter is invalid")
	}
	if pagination == nil {
		pagination = page.NewPagination()
	} else if err := structureValidator.New().Validate(pagination); err != nil {
		return nil, errors.Wrap(err, "pagination is invalid")
	}

	url := c.client.ConstructURL("v1", "users", userID, "data_sources")
	result := dataSource.SourceArray{}
	if err := c.client.RequestData(ctx, http.MethodGet, url, []request.RequestMutator{filter, pagination}, nil, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) Create(ctx context.Context, userID string, create *dataSource.Create) (*dataSource.Source, error) {
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

	url := c.client.ConstructURL("v1", "users", userID, "data_sources")
	result := &dataSource.Source{}
	if err := c.client.RequestData(ctx, http.MethodPost, url, nil, create, result); err != nil {
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

	url := c.client.ConstructURL("v1", "users", userID, "data_sources")
	return c.client.RequestData(ctx, http.MethodDelete, url, nil, nil, nil)
}

func (c *Client) Get(ctx context.Context, id string) (*dataSource.Source, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	} else if !dataSource.IsValidID(id) {
		return nil, errors.New("id is invalid")
	}

	url := c.client.ConstructURL("v1", "data_sources", id)
	result := &dataSource.Source{}
	if err := c.client.RequestData(ctx, http.MethodGet, url, nil, nil, result); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return result, nil
}

func (c *Client) Update(ctx context.Context, id string, condition *request.Condition, update *dataSource.Update) (*dataSource.Source, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	} else if !dataSource.IsValidID(id) {
		return nil, errors.New("id is invalid")
	}
	if condition == nil {
		condition = request.NewCondition()
	} else if err := structureValidator.New().Validate(condition); err != nil {
		return nil, errors.Wrap(err, "condition is invalid")
	}
	if update == nil {
		return nil, errors.New("update is missing")
	} else if err := structureValidator.New().Validate(update); err != nil {
		return nil, errors.Wrap(err, "update is invalid")
	}

	url := c.client.ConstructURL("v1", "data_sources", id)
	result := &dataSource.Source{}
	if err := c.client.RequestData(ctx, http.MethodPut, url, []request.RequestMutator{condition}, update, result); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, nil
		}
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
	} else if !dataSource.IsValidID(id) {
		return false, errors.New("id is invalid")
	}
	if condition == nil {
		condition = request.NewCondition()
	} else if err := structureValidator.New().Validate(condition); err != nil {
		return false, errors.Wrap(err, "condition is invalid")
	}

	url := c.client.ConstructURL("v1", "data_sources", id)
	if err := c.client.RequestData(ctx, http.MethodDelete, url, []request.RequestMutator{condition}, nil, nil); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
