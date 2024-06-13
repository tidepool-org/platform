package example

import (
	"context"
	"fmt"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/task"
)

const (
	Type                              = "org.tidepool.task.example"
	TaskDeadlineDuration              = 30 * time.Minute
	TaskTimeoutDuration               = 20 * time.Second
	TaskDurationMaximum               = 10 * time.Second
	TaskAvailableAfterDurationRepeat  = 30 * time.Second
	TaskAvailableAfterDurationInitial = 30 * time.Second
)

type Runner struct {
	taskClient task.Client
	logger     log.Logger
}

func NewRunner(taskClient task.Client, logger log.Logger) (*Runner, error) {
	runner := &Runner{
		taskClient: taskClient,
		logger:     logger,
	}

	if err := runner.setup(); err != nil {
		return nil, err
	}

	return runner, nil
}

func (r *Runner) GetRunnerType() string {
	return Type
}

func (r *Runner) GetRunnerDeadline() time.Time {
	return time.Now().Add(TaskDeadlineDuration)
}

func (r *Runner) GetRunnerTimeout() time.Duration {
	return TaskTimeoutDuration
}

func (r *Runner) GetRunnerDurationMaximum() time.Duration {
	return TaskDurationMaximum
}

func (r *Runner) Run(ctx context.Context, tsk *task.Task) {
	tsk.ClearError()

	r.execute(ctx, tsk)

	if !tsk.IsFailed() {
		tsk.RepeatAvailableAfter(TaskAvailableAfterDurationRepeat)
	}
}

// Add example code here
func (r *Runner) execute(ctx context.Context, tsk *task.Task) {
	r.logger.Infof("Executing example task '%s'", *tsk.Name)

	// Check for problem
	actualTask, err := r.taskClient.GetTask(ctx, tsk.ID)
	if err != nil {
		tsk.AppendError(err)
		tsk.SetFailed()
		return
	}

	if actualTask.AvailableTime != nil && tsk.AvailableTime != nil {
		if *actualTask.AvailableTime != *tsk.AvailableTime {
			r.failTask(tsk, errors.Newf("actual task available time (%v) different than task available time (%v)", *actualTask.AvailableTime, *tsk.AvailableTime))
			return
		}
	} else if actualTask.AvailableTime != nil || tsk.AvailableTime != nil {
		r.failTask(tsk, errors.Newf("actual task available time pointer (%v) different than task available time pointer (%v)", actualTask.AvailableTime, tsk.AvailableTime))
		return
	}

	var duration time.Duration
	if durationRaw, ok := tsk.Data["duration"]; !ok {
		r.failTask(tsk, errors.Newf("duration is missing"))
		return
	} else if durationInt64, ok := durationRaw.(int64); !ok {
		r.failTask(tsk, errors.Newf("duration is incorrect type"))
		return
	} else {
		duration = time.Duration(durationInt64)
	}

	r.logger.Infof("Sleeping %v", duration)

	select {
	case <-ctx.Done():
		r.logger.Infof("Cancelled example task '%s'", *tsk.Name)
	case <-time.After(duration):
		r.logger.Infof("Executed example task '%s'", *tsk.Name)
	}
}

func (r *Runner) failTask(tsk *task.Task, err error) {
	r.logger.WithError(err).Errorf("sleep task '%s' failed", *tsk.Name)
	tsk.AppendError(err)
	tsk.SetFailed()
}

func (r *Runner) setup() error {
	ctx := log.NewContextWithLogger(context.Background(), r.logger)

	for {
		filter := task.NewTaskFilter()
		filter.Type = pointer.FromString(Type)
		tsks, err := r.taskClient.ListTasks(ctx, filter, nil)
		if err != nil {
			return err
		} else if len(tsks) == 0 {
			break
		}

		for _, tsk := range tsks {
			if err := r.taskClient.DeleteTask(ctx, tsk.ID); err != nil {
				return err
			}
		}
	}

	availableTime := time.Now() // .Add(TaskAvailableAfterDurationInitial)

	durations := []time.Duration{9 * time.Second} // , 30 * time.Second}
	for _, duration := range durations {
		if err := r.createSleepTask(ctx, duration, &availableTime); err != nil {
			return err
		}
	}

	return nil
}

func (r *Runner) createSleepTask(ctx context.Context, duration time.Duration, availableTime *time.Time) error {
	create := &task.TaskCreate{
		Name: pointer.FromString(fmt.Sprintf("sleep(duration=%v)", duration)),
		Type: Type,
		Data: map[string]interface{}{
			"duration": duration,
		},
		AvailableTime: availableTime,
	}

	if _, err := r.taskClient.CreateTask(ctx, create); err != nil {
		return errors.Wrapf(err, "failure creating sleep task with duration %v", duration)
	}

	return nil
}
