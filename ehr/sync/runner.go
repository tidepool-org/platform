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
	AvailableAfterDurationMaximum = AvailableAfterDurationMinimum + 1*time.Hour
	AvailableAfterDurationMinimum = 14*24*time.Hour - 30*time.Minute
	TaskDurationMaximum           = 5 * time.Minute
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

func (r *Runner) GetRunnerMaximumDuration() time.Duration {
	return TaskDurationMaximum
}

func (r *Runner) Run(ctx context.Context, tsk *task.Task) bool {
	now := time.Now()
	tsk.ClearError()

	clinicId, err := GetClinicId(tsk.Data)
	if err != nil {
		tsk.AppendError(errors.Wrap(err, "unable to get clinicId from task data"))
		return true
	}

	err = r.clinicsClient.SyncEHRData(ctx, clinicId)
	if err != nil {
		tsk.AppendError(errors.Wrap(err, "unable to sync ehr data"))
	}

	if !tsk.IsFailed() {
		tsk.RepeatAvailableAfter(AvailableAfterDurationMinimum + time.Duration(rand.Int63n(int64(AvailableAfterDurationMaximum-AvailableAfterDurationMinimum+1))))
	}

	if taskDuration := time.Since(now); taskDuration > TaskDurationMaximum {
		r.logger.WithField("taskDuration", taskDuration.Truncate(time.Millisecond).Seconds()).Warn("Task duration exceeds maximum")
	}

	return true
}
