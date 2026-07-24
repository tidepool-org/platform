package queue

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/task"
	taskStore "github.com/tidepool-org/platform/task/store"
	taskTest "github.com/tidepool-org/platform/task/test"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Queue", func() {
	Context("dispatchTasks", func() {
		It("logs an error when the pending iterator cannot be opened", func() {
			lgr := logTest.NewLogger()
			ctx := log.NewContextWithLogger(context.Background(), lgr)
			cfg := NewConfig()
			que := &Queue{
				name:             taskTest.RandomType(),
				config:           cfg,
				logger:           lgr,
				repository:       &failureOpeningIteratorRepository{},
				workersAvailable: 1,
			}
			que.dispatchTasks(ctx)
			lgr.AssertError("Unable to open task iterator")
		})

		It("logs an error when the pending iterator fails mid-iteration", func() {
			lgr := logTest.NewLogger()
			ctx := log.NewContextWithLogger(context.Background(), lgr)
			cfg := NewConfig()
			que := &Queue{
				name:             taskTest.RandomType(),
				config:           cfg,
				logger:           lgr,
				repository:       &failureIteratingIteratorRepository{},
				workersAvailable: 1,
			}
			que.dispatchTasks(ctx)
			lgr.AssertError("Unable to iterate tasks")
		})
	})

	Context("computeState", func() {
		It("clears the available time for a completed task", func() {
			completedTask := &task.Task{State: task.TaskStateCompleted, AvailableTime: pointer.FromTime(test.RandomTimeBeforeNow())}
			(&Queue{}).computeState(context.Background(), completedTask)
			Expect(completedTask.AvailableTime).To(BeNil())
		})

		It("clears the available time for a failed task", func() {
			failedTask := &task.Task{State: task.TaskStateFailed, AvailableTime: pointer.FromTime(test.RandomTimeBeforeNow())}
			(&Queue{}).computeState(context.Background(), failedTask)
			Expect(failedTask.AvailableTime).To(BeNil())
		})
	})
})

type failureOpeningIteratorRepository struct {
	taskStore.TaskRepository
}

func (f *failureOpeningIteratorRepository) IteratePending(ctx context.Context) (*mongo.Cursor, error) {
	return nil, fmt.Errorf("failure opening iterator")
}

type failureIteratingIteratorRepository struct {
	taskStore.TaskRepository
}

func (f *failureIteratingIteratorRepository) IteratePending(ctx context.Context) (*mongo.Cursor, error) {
	return mongo.NewCursorFromDocuments([]any{}, fmt.Errorf("failure iterating iterator"), nil)
}
