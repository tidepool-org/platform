package client

import (
	"context"
	"net/http"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/request"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/task"
)

type Client struct {
	client *platform.Client
}

func New(cfg *platform.Config, authorizeAs platform.AuthorizeAs) (*Client, error) {
	clnt, err := platform.NewClient(cfg, authorizeAs)
	if err != nil {
		return nil, err
	}

	return &Client{
		client: clnt,
	}, nil
}

func (c *Client) ListTasks(ctx context.Context, filter *task.TaskFilter, pagination *page.Pagination) (task.Tasks, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if filter == nil {
		filter = task.NewTaskFilter()
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(filter); err != nil {
		return nil, errors.Wrap(err, "filter is invalid")
	}
	if pagination == nil {
		pagination = page.NewPagination()
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(pagination); err != nil {
		return nil, errors.Wrap(err, "pagination is invalid")
	}

	url := c.client.ConstructURL("v1", "tasks")
	tsks := task.Tasks{}
	if err := c.client.RequestData(ctx, http.MethodGet, url, []request.RequestMutator{filter, pagination}, nil, &tsks); err != nil {
		return nil, err
	}

	return tsks, nil
}

func (c *Client) CreateTask(ctx context.Context, create *task.TaskCreate) (*task.Task, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if create == nil {
		return nil, errors.New("create is missing")
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(create); err != nil {
		return nil, errors.Wrap(err, "create is invalid")
	}

	url := c.client.ConstructURL("v1", "tasks")
	tsk := &task.Task{}
	if err := c.client.RequestData(ctx, http.MethodPost, url, nil, create, tsk); err != nil {
		return nil, err
	}

	return tsk, nil
}

func (c *Client) GetTask(ctx context.Context, id string, condition *request.Condition) (*task.Task, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	}
	if condition == nil {
		condition = request.NewCondition()
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(condition); err != nil {
		return nil, errors.Wrap(err, "condition is invalid")
	}

	url := c.client.ConstructURL("v1", "tasks", id)
	tsk := &task.Task{}
	if err := c.client.RequestData(ctx, http.MethodGet, url, []request.RequestMutator{condition}, nil, tsk); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return tsk, nil
}

func (c *Client) UpdateTask(ctx context.Context, id string, condition *request.Condition, update *task.TaskUpdate) (*task.Task, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	}
	if condition == nil {
		condition = request.NewCondition()
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(condition); err != nil {
		return nil, errors.Wrap(err, "condition is invalid")
	}
	if update == nil {
		return nil, errors.New("update is missing")
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(update); err != nil {
		return nil, errors.Wrap(err, "update is invalid")
	}

	url := c.client.ConstructURL("v1", "tasks", id)
	tsk := &task.Task{}
	if err := c.client.RequestData(ctx, http.MethodPut, url, []request.RequestMutator{condition}, update, tsk); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return tsk, nil
}

func (c *Client) DeleteTask(ctx context.Context, id string, condition *request.Condition) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if id == "" {
		return errors.New("id is missing")
	}
	if condition == nil {
		condition = request.NewCondition()
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(condition); err != nil {
		return errors.Wrap(err, "condition is invalid")
	}

	url := c.client.ConstructURL("v1", "tasks", id)
	return c.client.RequestData(ctx, http.MethodDelete, url, []request.RequestMutator{condition}, nil, nil)
}
