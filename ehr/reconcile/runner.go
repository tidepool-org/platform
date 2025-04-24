package reconcile

import (
	"context"
	"math/rand"
	"time"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/clinics"
	"github.com/tidepool-org/platform/ehr/sync"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/task"
)

const (
	AvailableAfterDurationMaximum = 75 * time.Minute
	AvailableAfterDurationMinimum = 45 * time.Minute
	TaskDurationMaximum           = 5 * time.Minute
)

type Runner struct {
	authClient    auth.Client
	clinicsClient clinics.Client
	taskClient    task.Client
	logger        log.Logger
}

func NewRunner(authClient auth.Client, clinicsClient clinics.Client, taskClient task.Client, logger log.Logger) (*Runner, error) {
	return &Runner{
		authClient:    authClient,
		clinicsClient: clinicsClient,
		taskClient:    taskClient,
		logger:        logger,
	}, nil
}

func (r *Runner) GetRunnerType() string {
	return Type
}

func (r *Runner) GetRunnerDeadline() time.Time {
	return time.Now().Add(TaskDurationMaximum * 3)
}

func (r *Runner) GetRunnerTimeout() time.Duration {
	return TaskDurationMaximum * 2
}

func (r *Runner) GetRunnerDurationMaximum() time.Duration {
	return TaskDurationMaximum
}

func (r *Runner) Run(ctx context.Context, tsk *task.Task) {
	tsk.ClearError()

	r.doRun(ctx, tsk)
	tsk.RepeatAvailableAfter(AvailableAfterDurationMinimum + time.Duration(rand.Int63n(int64(AvailableAfterDurationMaximum-AvailableAfterDurationMinimum+1))))
}

func (r *Runner) doRun(ctx context.Context, tsk *task.Task) {
	ctx = auth.NewContextWithServerSessionTokenProvider(ctx, r.authClient)

	// Get the list of all existing EHR sync tasks
	syncTasks, err := r.getSyncTasks(ctx)
	if err != nil {
		tsk.AppendError(errors.Wrap(err, "unable to get sync tasks"))
		return
	}

	planner := NewPlanner(r.clinicsClient, r.logger)
	plan, err := planner.GetReconciliationPlan(ctx, syncTasks)
	if err != nil {
		tsk.AppendError(errors.Wrap(err, "unable to create reconciliation plan"))
		return
	}

	r.reconcileTasks(ctx, tsk, *plan)
}

func (r *Runner) getSyncTasks(ctx context.Context) (map[string]task.Task, error) {
	filter := task.TaskFilter{
		Type: pointer.FromString(sync.Type),
	}
	pagination := page.Pagination{
		Page: 0,
		Size: 1000,
	}

	tasksByClinicId := make(map[string]task.Task)
	for {
		tasks, err := r.taskClient.ListTasks(ctx, &filter, &pagination)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list tasks")
		}

		for _, tsk := range tasks {
			tsk := tsk
			clinicId, err := sync.GetClinicId(tsk.Data)
			if err != nil {
				r.logger.Errorf("unable to get clinicId from task data (taskId %v): %v", tsk.ID, err)
				continue
			}
			tasksByClinicId[clinicId] = *tsk
		}
		if len(tasks) < pagination.Size {
			break
		} else {
			pagination.Page++
		}
	}

	return tasksByClinicId, nil
}

func (r *Runner) reconcileTasks(ctx context.Context, tsk *task.Task, plan ReconciliationPlan) {
	for _, t := range plan.ToDelete {
		if err := r.taskClient.DeleteTask(ctx, t.ID); err != nil {
			tsk.AppendError(errors.Wrap(err, "unable to delete task"))
		}
	}
	for _, t := range plan.ToCreate {
		if _, err := r.taskClient.CreateTask(ctx, &t); err != nil {
			tsk.AppendError(errors.Wrap(err, "unable to create task"))
		}
	}
	for id, update := range plan.ToUpdate {
		if _, err := r.taskClient.UpdateTask(ctx, id, update); err != nil {
			tsk.AppendError(errors.Wrap(err, "unable to update task"))
		}
	}
}

type ReconciliationPlan struct {
	ToCreate []task.TaskCreate
	ToDelete []task.Task
	ToUpdate map[string]*task.TaskUpdate
}
