package test

import (
	"time"

	errorsTest "github.com/tidepool-org/platform/errors/test"
	metadataTest "github.com/tidepool-org/platform/metadata/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/task"
	"github.com/tidepool-org/platform/test"
)

func RandomID() string {
	return task.NewID()
}

func RandomName() string {
	return test.RandomString()
}

func RandomType() string {
	return test.RandomString()
}

func RandomData(options ...test.Option) map[string]any {
	return metadataTest.RandomOptionalMetadataMap(options...)
}

func RandomState() string {
	return test.RandomStringFromArray(task.TaskStates())
}

func RandomTaskCreate(options ...test.Option) *task.TaskCreate {
	now := test.Now()
	availableTime := test.RandomTimeAfter(now)
	return &task.TaskCreate{
		Name:          test.RandomOptional(RandomName, options...),
		Type:          RandomType(),
		Data:          RandomData(options...),
		AvailableTime: test.RandomOptional(test.Constant(availableTime), options...),
	}
}

func RandomTask(options ...test.Option) *task.Task {
	now := test.Now()

	tsk := &task.Task{
		ID:          RandomID(),
		Name:        test.RandomOptional(RandomName, options...),
		Type:        RandomType(),
		Data:        RandomData(options...),
		State:       RandomState(),
		CreatedTime: test.RandomTimeBefore(now),
	}

	switch tsk.State {
	case task.TaskStatePending:
		tsk.AvailableTime = pointer.From(test.RandomTimeAfterNow())
		tsk.DeadlineTime = nil
		tsk.ModifiedTime = test.RandomOptional(func() time.Time { return test.RandomTimeFromRange(tsk.CreatedTime, now) }, options...)
		if tsk.ModifiedTime != nil {
			tsk.Error = test.RandomOptionalPointer(errorsTest.RandomSerializable, options...)
			tsk.RunTime = tsk.ModifiedTime
			tsk.Duration = pointer.From(test.RandomFloat64FromRange(0, 10))
		}
	case task.TaskStateRunning:
		tsk.AvailableTime = pointer.From(test.RandomTimeFromRange(tsk.CreatedTime, now))
		tsk.DeadlineTime = pointer.From(test.RandomTimeAfterNow())
		tsk.ModifiedTime = pointer.From(test.RandomTimeFromRange(*tsk.AvailableTime, now))
		if test.RandomBool() {
			tsk.Error = test.RandomOptionalPointer(errorsTest.RandomSerializable, options...)
			tsk.RunTime = pointer.From(test.RandomTimeFromRange(tsk.CreatedTime, *tsk.AvailableTime))
			tsk.Duration = pointer.From(test.RandomFloat64FromRange(0, 10))
		}
	case task.TaskStateFailed:
		tsk.AvailableTime = pointer.From(test.RandomTimeFromRange(tsk.CreatedTime, now))
		tsk.DeadlineTime = nil
		tsk.ModifiedTime = pointer.From(test.RandomTimeFromRange(*tsk.AvailableTime, now))
		tsk.Error = errorsTest.RandomSerializable()
		tsk.RunTime = tsk.ModifiedTime
		tsk.Duration = pointer.From(test.RandomFloat64FromRange(0, 10))
	case task.TaskStateCompleted:
		tsk.AvailableTime = pointer.From(test.RandomTimeFromRange(tsk.CreatedTime, now))
		tsk.DeadlineTime = nil
		tsk.ModifiedTime = pointer.From(test.RandomTimeFromRange(*tsk.AvailableTime, now))
		tsk.Error = test.RandomOptionalPointer(errorsTest.RandomSerializable, options...)
		tsk.RunTime = tsk.ModifiedTime
		tsk.Duration = pointer.From(test.RandomFloat64FromRange(0, 10))
	}

	return tsk
}
