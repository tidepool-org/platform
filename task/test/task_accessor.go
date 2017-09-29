package test

import (
	"context"

	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/task"
	"github.com/tidepool-org/platform/test"
)

type ListTasksInput struct {
	Context    context.Context
	Filter     *task.TaskFilter
	Pagination *page.Pagination
}

type ListTasksOutput struct {
	Tasks task.Tasks
	Error error
}

type CreateTaskInput struct {
	Context context.Context
	Create  *task.TaskCreate
}

type CreateTaskOutput struct {
	Task  *task.Task
	Error error
}

type GetTaskInput struct {
	Context context.Context
	ID      string
}

type GetTaskOutput struct {
	Task  *task.Task
	Error error
}

type UpdateTaskInput struct {
	Context context.Context
	ID      string
	Update  *task.TaskUpdate
}

type UpdateTaskOutput struct {
	Task  *task.Task
	Error error
}

type DeleteTaskInput struct {
	Context context.Context
	ID      string
}

type TaskAccessor struct {
	*test.Mock
	ListTasksInvocations  int
	ListTasksInputs       []ListTasksInput
	ListTasksOutputs      []ListTasksOutput
	CreateTaskInvocations int
	CreateTaskInputs      []CreateTaskInput
	CreateTaskOutputs     []CreateTaskOutput
	GetTaskInvocations    int
	GetTaskInputs         []GetTaskInput
	GetTaskOutputs        []GetTaskOutput
	UpdateTaskInvocations int
	UpdateTaskInputs      []UpdateTaskInput
	UpdateTaskOutputs     []UpdateTaskOutput
	DeleteTaskInvocations int
	DeleteTaskInputs      []DeleteTaskInput
	DeleteTaskOutputs     []error
}

func NewTaskAccessor() *TaskAccessor {
	return &TaskAccessor{
		Mock: test.NewMock(),
	}
}

func (r *TaskAccessor) ListTasks(ctx context.Context, filter *task.TaskFilter, pagination *page.Pagination) (task.Tasks, error) {
	r.ListTasksInvocations++

	r.ListTasksInputs = append(r.ListTasksInputs, ListTasksInput{Context: ctx, Filter: filter, Pagination: pagination})

	gomega.Expect(r.ListTasksOutputs).ToNot(gomega.BeEmpty())

	output := r.ListTasksOutputs[0]
	r.ListTasksOutputs = r.ListTasksOutputs[1:]
	return output.Tasks, output.Error
}

func (r *TaskAccessor) CreateTask(ctx context.Context, create *task.TaskCreate) (*task.Task, error) {
	r.CreateTaskInvocations++

	r.CreateTaskInputs = append(r.CreateTaskInputs, CreateTaskInput{Context: ctx, Create: create})

	gomega.Expect(r.CreateTaskOutputs).ToNot(gomega.BeEmpty())

	output := r.CreateTaskOutputs[0]
	r.CreateTaskOutputs = r.CreateTaskOutputs[1:]
	return output.Task, output.Error
}

func (r *TaskAccessor) GetTask(ctx context.Context, id string) (*task.Task, error) {
	r.GetTaskInvocations++

	r.GetTaskInputs = append(r.GetTaskInputs, GetTaskInput{Context: ctx, ID: id})

	gomega.Expect(r.GetTaskOutputs).ToNot(gomega.BeEmpty())

	output := r.GetTaskOutputs[0]
	r.GetTaskOutputs = r.GetTaskOutputs[1:]
	return output.Task, output.Error
}

func (r *TaskAccessor) UpdateTask(ctx context.Context, id string, update *task.TaskUpdate) (*task.Task, error) {
	r.UpdateTaskInvocations++

	r.UpdateTaskInputs = append(r.UpdateTaskInputs, UpdateTaskInput{Context: ctx, ID: id, Update: update})

	gomega.Expect(r.UpdateTaskOutputs).ToNot(gomega.BeEmpty())

	output := r.UpdateTaskOutputs[0]
	r.UpdateTaskOutputs = r.UpdateTaskOutputs[1:]
	return output.Task, output.Error
}

func (r *TaskAccessor) DeleteTask(ctx context.Context, id string) error {
	r.DeleteTaskInvocations++

	r.DeleteTaskInputs = append(r.DeleteTaskInputs, DeleteTaskInput{Context: ctx, ID: id})

	gomega.Expect(r.DeleteTaskOutputs).ToNot(gomega.BeEmpty())

	output := r.DeleteTaskOutputs[0]
	r.DeleteTaskOutputs = r.DeleteTaskOutputs[1:]
	return output
}

func (r *TaskAccessor) Expectations() {
	r.Mock.Expectations()
	gomega.Expect(r.ListTasksOutputs).To(gomega.BeEmpty())
	gomega.Expect(r.CreateTaskOutputs).To(gomega.BeEmpty())
	gomega.Expect(r.GetTaskOutputs).To(gomega.BeEmpty())
	gomega.Expect(r.UpdateTaskOutputs).To(gomega.BeEmpty())
}
