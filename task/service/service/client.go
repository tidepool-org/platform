package service

import (
	"context"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/page"
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
	ssn := c.taskStore.NewTaskSession()
	defer ssn.Close()

	return ssn.ListTasks(ctx, filter, pagination)
}

func (c *Client) CreateTask(ctx context.Context, create *task.TaskCreate) (*task.Task, error) {
	ssn := c.taskStore.NewTaskSession()
	defer ssn.Close()

	return ssn.CreateTask(ctx, create)
}

func (c *Client) GetTask(ctx context.Context, id string) (*task.Task, error) {
	ssn := c.taskStore.NewTaskSession()
	defer ssn.Close()

	return ssn.GetTask(ctx, id)
}

func (c *Client) UpdateTask(ctx context.Context, id string, update *task.TaskUpdate) (*task.Task, error) {
	ssn := c.taskStore.NewTaskSession()
	defer ssn.Close()

	return ssn.UpdateTask(ctx, id, update)
}

func (c *Client) DeleteTask(ctx context.Context, id string) error {
	ssn := c.taskStore.NewTaskSession()
	defer ssn.Close()

	return ssn.DeleteTask(ctx, id)
}
