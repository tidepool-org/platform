package summary

import (
	"time"

	"github.com/pkg/errors"

	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/task"
)

type MinuteRange struct {
	Min int
	Max int
}

type TaskConfiguration struct {
	Interval MinuteRange
	Batch    *int `json:"batch,omitempty" bson:"batch,omitempty"`
}

func ValidateConfig(config TaskConfiguration, withBatch bool) error {
	if config.Interval.Min < 1 {
		return errors.New("Minimum Interval cannot be <1 minute")
	}
	if config.Interval.Max < config.Interval.Min {
		return errors.New("Maximum Interval cannot be less than Minimum Interval")
	}

	if withBatch == true {
		if config.Batch == nil {
			return errors.New("Batch is required but not provided")
		}
		if *config.Batch < 1 {
			return errors.New("Batch can not be <1")
		}
	} else {
		if config.Batch != nil {
			return errors.New("Batch is not required, but was provided")
		}
	}

	return nil
}

func NewDefaultBackfillConfig() TaskConfiguration {
	return TaskConfiguration{
		Interval: MinuteRange{
			int(DefaultBackfillAvailableAfterDurationMinimum.Minutes()),
			int(DefaultBackfillAvailableAfterDurationMaximum.Minutes()),
		},
	}
}

func NewDefaultBackfillTaskCreate() *task.TaskCreate {
	return &task.TaskCreate{
		Name:           pointer.FromAny(BackfillType),
		Type:           BackfillType,
		Priority:       5,
		AvailableTime:  pointer.FromAny(time.Now().UTC()),
		ExpirationTime: pointer.FromAny(time.Now().UTC().AddDate(1000, 0, 0)),
		Data: map[string]interface{}{
			"config": NewDefaultBackfillConfig(),
		},
	}
}

func NewDefaultUpdateConfig() TaskConfiguration {
	return TaskConfiguration{
		Interval: MinuteRange{
			int(DefaultUpdateAvailableAfterDurationMinimum.Minutes()),
			int(DefaultUpdateAvailableAfterDurationMaximum.Minutes())},
		Batch: pointer.FromAny(DefaultUpdateWorkerBatchSize),
	}
}

func NewDefaultUpdateTaskCreate() *task.TaskCreate {
	return &task.TaskCreate{
		Name:           pointer.FromAny(UpdateType),
		Type:           UpdateType,
		Priority:       5,
		AvailableTime:  pointer.FromAny(time.Now().UTC()),
		ExpirationTime: pointer.FromAny(time.Now().UTC().AddDate(1000, 0, 0)),
		Data: map[string]interface{}{
			"config": NewDefaultUpdateConfig(),
		},
	}
}

func NewDefaultMigrationConfig() TaskConfiguration {
	return TaskConfiguration{
		Interval: MinuteRange{
			int(DefaultMigrationAvailableAfterDurationMinimum.Minutes()),
			int(DefaultMigrationAvailableAfterDurationMaximum.Minutes())},
		Batch: pointer.FromAny(DefaultMigrationWorkerBatchSize),
	}
}

func NewDefaultMigrationTaskCreate() *task.TaskCreate {
	return &task.TaskCreate{
		Name:           pointer.FromAny(MigrationType),
		Type:           MigrationType,
		Priority:       5,
		AvailableTime:  pointer.FromAny(time.Now().UTC()),
		ExpirationTime: pointer.FromAny(time.Now().UTC().AddDate(1000, 0, 0)),
		Data: map[string]interface{}{
			"config": NewDefaultMigrationConfig(),
		},
	}
}
