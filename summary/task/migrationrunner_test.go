package task_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	. "github.com/tidepool-org/platform/summary/task"
	"github.com/tidepool-org/platform/task"
	taskStore "github.com/tidepool-org/platform/task/store"
	taskStoreMongo "github.com/tidepool-org/platform/task/store/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

var _ = Describe("migrate runner tasks", func() {
	var err error
	var ctx context.Context
	var cfg *storeStructuredMongo.Config
	var logger *logTest.Logger
	var store *taskStoreMongo.Store
	var repository taskStore.TaskRepository

	BeforeEach(func() {
		cfg = storeStructuredMongoTest.NewConfig()
		logger = logTest.NewLogger()
		ctx = log.NewContextWithLogger(context.Background(), logger)

		store, err = taskStoreMongo.NewStore(cfg)
		Expect(err).ToNot(HaveOccurred())
		Expect(store).ToNot(BeNil())

		repository = store.NewTaskRepository()
		Expect(repository).ToNot(BeNil())
	})

	AfterEach(func() {
		if store != nil {
			store.TaskRepository().DeleteMany(ctx, bson.D{})
			store.Terminate(context.Background())
		}
	})

	Context("Task create and read", func() {
		// int/int32/int64 is inconsistent between the db written type and the db read type
		// for our uses, the mongo driver should return int32 into our config interface
		// so we standardize all config handling to int32

		It("create task, validate contents are filled correctly", func() {
			t := NewDefaultMigrationTaskCreate("cgm")
			Expect(t.Name).ToNot(BeNil())
			Expect(*t.Name).To(Equal("org.tidepool.summary.migrate.cgm"))
			Expect(t.Type).To(Equal("org.tidepool.summary.migrate.cgm"))
			Expect(t.AvailableTime).ToNot(BeNil())
			Expect(t.AvailableTime.IsZero()).ToNot(BeTrue())

			batch, ok := t.Data[ConfigBatch].(int32)
			Expect(ok).To(BeTrue())
			Expect(batch).To(Equal(int32(500)))

			minInterval, ok := t.Data[ConfigMinInterval].(int32)
			Expect(ok).To(BeTrue())
			Expect(minInterval).To(Equal(int32(300)))

			maxInterval, ok := t.Data[ConfigMaxInterval].(int32)
			Expect(ok).To(BeTrue())
			Expect(maxInterval).To(Equal(int32(300)))
		})

		It("create task, write to db and confirm contents", func() {
			taskToWrite := NewDefaultMigrationTaskCreate("cgm")
			_, err := repository.CreateTask(ctx, taskToWrite)
			Expect(err).ToNot(HaveOccurred())

			tasks, err := repository.ListTasks(ctx, &task.TaskFilter{
				State: pointer.FromAny(task.TaskStatePending),
			}, &page.Pagination{
				Page: 0,
				Size: 10,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(len(tasks)).To(Equal(1))

			t := tasks[0]

			Expect(t.Name).ToNot(BeNil())
			Expect(*t.Name).To(Equal("org.tidepool.summary.migrate.cgm"))
			Expect(t.Type).To(Equal("org.tidepool.summary.migrate.cgm"))
			Expect(t.AvailableTime).ToNot(BeNil())
			Expect(t.AvailableTime.IsZero()).ToNot(BeTrue())

			batch, ok := t.Data[ConfigBatch].(int32)
			Expect(ok).To(BeTrue())
			Expect(batch).To(Equal(int32(500)))

			minInterval, ok := t.Data[ConfigMinInterval].(int32)
			Expect(ok).To(BeTrue())
			Expect(minInterval).To(Equal(int32(300)))

			maxInterval, ok := t.Data[ConfigMaxInterval].(int32)
			Expect(ok).To(BeTrue())
			Expect(maxInterval).To(Equal(int32(300)))
		})

		It("create task, write to db and check getconfig functions with custom config", func() {
			taskToWrite := NewDefaultMigrationTaskCreate("cgm")
			taskToWrite.Data[ConfigBatch] = 300
			taskToWrite.Data[ConfigMinInterval] = 100
			taskToWrite.Data[ConfigMaxInterval] = 200
			_, err := repository.CreateTask(ctx, taskToWrite)
			Expect(err).ToNot(HaveOccurred())

			tasks, err := repository.ListTasks(ctx, &task.TaskFilter{
				State: pointer.FromAny(task.TaskStatePending),
			}, &page.Pagination{
				Page: 0,
				Size: 10,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(len(tasks)).To(Equal(1))

			t := tasks[0]

			runner := MigrationTaskRunner{Task: t}
			Expect(runner.GetBatch()).To(Equal(300))

			minSeconds, maxSeconds := runner.GetIntervalRange()
			Expect(minSeconds).To(Equal(100))
			Expect(maxSeconds).To(Equal(200))
		})

		It("create task, write to db and check getconfig functions with invalid config types", func() {
			taskToWrite := NewDefaultMigrationTaskCreate("cgm")
			taskToWrite.Data[ConfigBatch] = "300"
			taskToWrite.Data[ConfigMinInterval] = "100"
			taskToWrite.Data[ConfigMaxInterval] = "200"
			_, err := repository.CreateTask(ctx, taskToWrite)
			Expect(err).ToNot(HaveOccurred())

			tasks, err := repository.ListTasks(ctx, &task.TaskFilter{
				State: pointer.FromAny(task.TaskStatePending),
			}, &page.Pagination{
				Page: 0,
				Size: 10,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(len(tasks)).To(Equal(1))

			t := tasks[0]

			runner := MigrationTaskRunner{Task: t}
			Expect(runner.GetBatch()).To(Equal(500))

			minSeconds, maxSeconds := runner.GetIntervalRange()
			Expect(minSeconds).To(Equal(300))
			Expect(maxSeconds).To(Equal(300))
		})

		It("create task, write to db and check getconfig functions with invalid config values", func() {
			taskToWrite := NewDefaultMigrationTaskCreate("cgm")
			taskToWrite.Data[ConfigBatch] = 0
			taskToWrite.Data[ConfigMinInterval] = 0
			taskToWrite.Data[ConfigMaxInterval] = 0
			_, err := repository.CreateTask(ctx, taskToWrite)
			Expect(err).ToNot(HaveOccurred())

			tasks, err := repository.ListTasks(ctx, &task.TaskFilter{
				State: pointer.FromAny(task.TaskStatePending),
			}, &page.Pagination{
				Page: 0,
				Size: 10,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(len(tasks)).To(Equal(1))

			t := tasks[0]

			runner := MigrationTaskRunner{Task: t}
			Expect(runner.GetBatch()).To(Equal(500))

			minSeconds, maxSeconds := runner.GetIntervalRange()
			Expect(minSeconds).To(Equal(300))
			Expect(maxSeconds).To(Equal(300))
		})

		It("create task, write to db and check getconfig functions with flipped interval config", func() {
			taskToWrite := NewDefaultMigrationTaskCreate("cgm")
			taskToWrite.Data[ConfigBatch] = 0
			taskToWrite.Data[ConfigMinInterval] = 200
			taskToWrite.Data[ConfigMaxInterval] = 100
			_, err := repository.CreateTask(ctx, taskToWrite)
			Expect(err).ToNot(HaveOccurred())

			tasks, err := repository.ListTasks(ctx, &task.TaskFilter{
				State: pointer.FromAny(task.TaskStatePending),
			}, &page.Pagination{
				Page: 0,
				Size: 10,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(len(tasks)).To(Equal(1))

			t := tasks[0]

			runner := MigrationTaskRunner{Task: t}
			Expect(runner.GetBatch()).To(Equal(500))

			minSeconds, maxSeconds := runner.GetIntervalRange()
			Expect(minSeconds).To(Equal(300))
			Expect(maxSeconds).To(Equal(300))
		})
	})

})
