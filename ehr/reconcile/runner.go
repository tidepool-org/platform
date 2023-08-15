package reconcile

import (
	"context"
	"math/rand"
	"time"

	api "github.com/tidepool-org/clinic/client"

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

func (r *Runner) GetRunnerMaximumDuration() time.Duration {
	return TaskDurationMaximum
}

func (r *Runner) Run(ctx context.Context, tsk *task.Task) bool {
	now := time.Now()
	tsk.ClearError()

	serverSessionToken, err := r.authClient.ServerSessionToken()
	if err != nil {
		tsk.AppendError(errors.Wrap(err, "unable to get server session token"))
		return true
	}

	ctx = auth.NewContextWithServerSessionToken(ctx, serverSessionToken)

	// Get the list of all existing EHR sync tasks
	syncTasks, err := r.getSyncTasks(ctx)
	if err != nil {
		tsk.AppendError(errors.Wrap(err, "unable to get sync tasks"))
	}

	// Get the list of all EHR enabled clinics
	clinicsList, err := r.clinicsClient.ListEHREnabledClinics(ctx)
	if err != nil {
		tsk.AppendError(errors.Wrap(err, "unable to list clinics"))
	}

	plan := GetReconciliationPlan(syncTasks, clinicsList)
	r.reconcileTasks(ctx, tsk, plan)

	if !tsk.IsFailed() {
		tsk.RepeatAvailableAfter(AvailableAfterDurationMinimum + time.Duration(rand.Int63n(int64(AvailableAfterDurationMaximum-AvailableAfterDurationMinimum+1))))
	}

	if taskDuration := time.Since(now); taskDuration > TaskDurationMaximum {
		r.logger.WithField("taskDuration", taskDuration.Truncate(time.Millisecond).Seconds()).Warn("Task duration exceeds maximum")
	}

	return true
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

func (r *Runner) reconcileTasks(ctx context.Context, task *task.Task, plan ReconciliationPlan) {
	for _, t := range plan.ToDelete {
		if err := r.taskClient.DeleteTask(ctx, t.ID); err != nil {
			task.AppendError(errors.Wrap(err, "unable to delete task"))
		}
	}
	for _, t := range plan.ToCreate {
		if _, err := r.taskClient.CreateTask(ctx, &t); err != nil {
			task.AppendError(errors.Wrap(err, "unable to create task"))
		}
	}
}

type ReconciliationPlan struct {
	ToCreate []task.TaskCreate
	ToDelete []task.Task
}

func GetReconciliationPlan(syncTasks map[string]task.Task, clinics []api.Clinic) ReconciliationPlan {
	toDelete := make([]task.Task, 0)
	toCreate := make([]task.TaskCreate, 0)

	// At the end of the loop syncTasks will contain only the tasks that need to be deleted,
	// and toCreate will contain tasks for new clinics that need to be synced.
	for _, clinic := range clinics {
		clinicId := *clinic.Id
		_, exists := syncTasks[clinicId]

		if exists {
			delete(syncTasks, clinicId)
		} else {
			create := sync.NewTaskCreate(clinicId)
			toCreate = append(toCreate, *create)
		}
	}
	for _, tsk := range syncTasks {
		toDelete = append(toDelete, tsk)
	}
	return ReconciliationPlan{
		ToCreate: toCreate,
		ToDelete: toDelete,
	}
}
