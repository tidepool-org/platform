package queue

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/task/store"
)

type failingIteratorStore struct {
	store.Store
}

func (f *failingIteratorStore) NewTaskRepository() store.TaskRepository {
	return &failingIteratorRepository{}
}

type failingIteratorRepository struct {
	store.TaskRepository
}

func (f *failingIteratorRepository) IteratePending(ctx context.Context) (
	*mongo.Cursor, error) {

	return mongo.NewCursorFromDocuments([]interface{}{bson.D{}},
		fmt.Errorf("batch fetch failed"), nil)
}

var _ = Describe("Queue", func() {
	Context("dispatchTasks", func() {
		It("logs an error when the pending iterator fails mid-iteration", func() {
			logger := logTest.NewLogger()
			q := &queue{
				logger:           logger,
				store:            &failingIteratorStore{},
				workersAvailable: 1,
				delay:            time.Minute,
			}

			Expect(q.dispatchTasks(context.Background())).To(Equal(time.Minute))
			logger.AssertError("Failure iterating pending tasks")
		})
	})
})
