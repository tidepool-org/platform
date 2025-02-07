package summary

import (
	"errors"
	"time"

	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/task"
)

const ConfigVersion = 1

type MinuteRange struct {
	Min int
	Max int
}

type TaskConfiguration struct {
	Interval MinuteRange `json:"interval" bson:"interval"`
	Batch    *int        `json:"batch,omitempty" bson:"batch,omitempty"`
	Version  int         `json:"version" bson:"version"`
}

func ValidateConfig(config TaskConfiguration, withBatch bool) error {
	if config.Version != ConfigVersion {
		return errors.New("old version number, must be remade")
	}
	if config.Interval.Min < 1 {
		return errors.New("minimum Interval cannot be <1 minute")
	}
	if config.Interval.Max < config.Interval.Min {
		return errors.New("maximum Interval cannot be less than Minimum Interval")
	}

	if withBatch == true {
		if config.Batch == nil {
			return errors.New("batch is required but not provided")
		}
		if *config.Batch < 1 {
			return errors.New("batch can not be <1")
		}
	} else {
		if config.Batch != nil {
			return errors.New("batch is not required, but was provided")
		}
	}

	return nil
}

func NewDefaultUpdateConfig() TaskConfiguration {
	return TaskConfiguration{
		Interval: MinuteRange{
			int(DefaultUpdateAvailableAfterDurationMinimum.Seconds()),
			int(DefaultUpdateAvailableAfterDurationMaximum.Seconds())},
		Batch:   pointer.FromAny(DefaultUpdateWorkerBatchSize),
		Version: ConfigVersion,
	}
}

func NewDefaultUpdateTaskCreate() *task.TaskCreate {
	return &task.TaskCreate{
		Name:          pointer.FromAny(UpdateType),
		Type:          UpdateType,
		Priority:      5,
		AvailableTime: pointer.FromAny(time.Now().UTC()),
		Data: map[string]interface{}{
			"config": NewDefaultUpdateConfig(),
		},
	}
}

func NewDefaultMigrationConfig() TaskConfiguration {
	return TaskConfiguration{
		Interval: MinuteRange{
			int(DefaultMigrationAvailableAfterDurationMinimum.Seconds()),
			int(DefaultMigrationAvailableAfterDurationMaximum.Seconds())},
		Batch:   pointer.FromAny(DefaultMigrationWorkerBatchSize),
		Version: ConfigVersion,
	}
}

func NewDefaultMigrationTaskCreate() *task.TaskCreate {
	return &task.TaskCreate{
		Name:          pointer.FromAny(MigrationType),
		Type:          MigrationType,
		Priority:      5,
		AvailableTime: pointer.FromAny(time.Now().UTC()),
		Data: map[string]interface{}{
			"config": NewDefaultMigrationConfig(),
		},
	}
}
