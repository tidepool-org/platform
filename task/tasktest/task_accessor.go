package tasktest

import (
	"context"

	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/task"
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
	return &TaskAccessor{}
}

func (t *TaskAccessor) ListTasks(ctx context.Context, filter *task.TaskFilter, pagination *page.Pagination) (task.Tasks, error) {
	t.ListTasksInvocations++

	t.ListTasksInputs = append(t.ListTasksInputs, ListTasksInput{Context: ctx, Filter: filter, Pagination: pagination})

	gomega.Expect(t.ListTasksOutputs).ToNot(gomega.BeEmpty())

	output := t.ListTasksOutputs[0]
	t.ListTasksOutputs = t.ListTasksOutputs[1:]
	return output.Tasks, output.Error
}

func (t *TaskAccessor) CreateTask(ctx context.Context, create *task.TaskCreate) (*task.Task, error) {
	t.CreateTaskInvocations++

	t.CreateTaskInputs = append(t.CreateTaskInputs, CreateTaskInput{Context: ctx, Create: create})

	gomega.Expect(t.CreateTaskOutputs).ToNot(gomega.BeEmpty())

	output := t.CreateTaskOutputs[0]
	t.CreateTaskOutputs = t.CreateTaskOutputs[1:]
	return output.Task, output.Error
}

func (t *TaskAccessor) GetTask(ctx context.Context, id string) (*task.Task, error) {
	t.GetTaskInvocations++

	t.GetTaskInputs = append(t.GetTaskInputs, GetTaskInput{Context: ctx, ID: id})

	gomega.Expect(t.GetTaskOutputs).ToNot(gomega.BeEmpty())

	output := t.GetTaskOutputs[0]
	t.GetTaskOutputs = t.GetTaskOutputs[1:]
	return output.Task, output.Error
}

func (t *TaskAccessor) UpdateTask(ctx context.Context, id string, update *task.TaskUpdate) (*task.Task, error) {
	t.UpdateTaskInvocations++

	t.UpdateTaskInputs = append(t.UpdateTaskInputs, UpdateTaskInput{Context: ctx, ID: id, Update: update})

	gomega.Expect(t.UpdateTaskOutputs).ToNot(gomega.BeEmpty())

	output := t.UpdateTaskOutputs[0]
	t.UpdateTaskOutputs = t.UpdateTaskOutputs[1:]
	return output.Task, output.Error
}

func (t *TaskAccessor) DeleteTask(ctx context.Context, id string) error {
	t.DeleteTaskInvocations++

	t.DeleteTaskInputs = append(t.DeleteTaskInputs, DeleteTaskInput{Context: ctx, ID: id})

	gomega.Expect(t.DeleteTaskOutputs).ToNot(gomega.BeEmpty())

	output := t.DeleteTaskOutputs[0]
	t.DeleteTaskOutputs = t.DeleteTaskOutputs[1:]
	return output
}

func (t *TaskAccessor) Expectations() {
	gomega.Expect(t.ListTasksOutputs).To(gomega.BeEmpty())
	gomega.Expect(t.CreateTaskOutputs).To(gomega.BeEmpty())
	gomega.Expect(t.GetTaskOutputs).To(gomega.BeEmpty())
	gomega.Expect(t.UpdateTaskOutputs).To(gomega.BeEmpty())
}
