package sync

import (
	"context"
	"math/rand"
	"time"

	"github.com/tidepool-org/platform/clinics"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/task"
)

const (
	OnSuccessAvailableAfterDurationMaximum = OnSuccessAvailableAfterDurationMinimum + 1*time.Hour
	OnSuccessAvailableAfterDurationMinimum = 14*24*time.Hour - 30*time.Minute
	OnErrorAvailableAfterDurationMaximum   = OnErrorAvailableAfterDurationMinimum + 5*time.Minute
	OnErrorAvailableAfterDurationMinimum   = 1*time.Hour - 5*time.Minute
	TaskDurationMaximum                    = 5 * time.Minute
)

type Runner struct {
	clinicsClient clinics.Client
	logger        log.Logger
}

func NewRunner(clinicsClient clinics.Client, logger log.Logger) (*Runner, error) {
	return &Runner{
		clinicsClient: clinicsClient,
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

	if !tsk.IsFailed() {
		if tsk.HasError() {
			tsk.RepeatAvailableAfter(OnErrorAvailableAfterDurationMinimum + time.Duration(rand.Int63n(int64(OnErrorAvailableAfterDurationMaximum-OnErrorAvailableAfterDurationMinimum+1))))
		} else {
			tsk.RepeatAvailableAfter(OnSuccessAvailableAfterDurationMinimum + time.Duration(rand.Int63n(int64(OnSuccessAvailableAfterDurationMaximum-OnSuccessAvailableAfterDurationMinimum+1))))
		}
	}
}

func (r *Runner) doRun(ctx context.Context, tsk *task.Task) {
	clinicId, err := GetClinicId(tsk.Data)
	if err != nil {
		tsk.AppendError(errors.Wrap(err, "unable to get clinicId from task data"))
		// Unrecoverable condition, move the task to failed state so it won't be retried
		tsk.SetFailed()
		return
	}

	err = r.clinicsClient.SyncEHRData(ctx, clinicId)
	if err != nil {
		tsk.AppendError(errors.Wrap(err, "unable to sync ehr data"))
	}
}
