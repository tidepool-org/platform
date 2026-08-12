package service

import (
	"context"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/request"
	storeStructured "github.com/tidepool-org/platform/store/structured"
	"github.com/tidepool-org/platform/task"
	taskStore "github.com/tidepool-org/platform/task/store"
)

type Client struct {
	taskStore taskStore.Store
}

func NewClient(str taskStore.Store) (*Client, error) {
	if str == nil {
		return nil, errors.New("task store is missing")
	}

	return &Client{
		taskStore: str,
	}, nil
}

func (c *Client) ListTasks(ctx context.Context, filter *task.TaskFilter, pagination *page.Pagination) (task.Tasks, error) {
	repository := c.taskStore.NewTaskRepository()
	return repository.ListTasks(ctx, filter, pagination)
}

func (c *Client) CreateTask(ctx context.Context, create *task.TaskCreate) (*task.Task, error) {
	repository := c.taskStore.NewTaskRepository()
	return repository.CreateTask(ctx, create)
}

func (c *Client) GetTask(ctx context.Context, id string, condition *request.Condition) (*task.Task, error) {
	repository := c.taskStore.NewTaskRepository()
	return repository.GetTask(ctx, id, storeStructured.MapCondition(condition))
}

func (c *Client) UpdateTask(ctx context.Context, id string, condition *request.Condition, update *task.TaskUpdate) (*task.Task, error) {
	repository := c.taskStore.NewTaskRepository()
	return repository.UpdateTask(ctx, id, storeStructured.MapCondition(condition), update)
}

func (c *Client) DeleteTask(ctx context.Context, id string, condition *request.Condition) error {
	repository := c.taskStore.NewTaskRepository()
	return repository.DeleteTask(ctx, id, storeStructured.MapCondition(condition))
}
