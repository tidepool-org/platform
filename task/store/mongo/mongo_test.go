package mongo_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/tidepool-org/platform/ehr/reconcile"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	"github.com/tidepool-org/platform/task"
	taskStore "github.com/tidepool-org/platform/task/store"
	taskStoreMongo "github.com/tidepool-org/platform/task/store/mongo"
	taskTest "github.com/tidepool-org/platform/task/test"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Mongo", func() {
	var cfg *storeStructuredMongo.Config
	var logger *logTest.Logger
	var str *taskStoreMongo.Store
	var repository taskStore.TaskRepository

	BeforeEach(func() {
		cfg = storeStructuredMongoTest.NewConfig()
		logger = logTest.NewLogger()
	})

	AfterEach(func() {
		if str != nil {
			_ = str.Terminate(context.Background())
		}
	})

	Context("New", func() {
		It("returns an error if unsuccessful", func() {
			var err error
			str, err = taskStoreMongo.NewStore(nil)
			Expect(err).To(HaveOccurred())
			Expect(str).To(BeNil())
		})

		It("returns a new store and no error if successful", func() {
			var err error
			str, err = taskStoreMongo.NewStore(cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(str).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		var collection *mongo.Collection

		BeforeEach(func() {
			var err error
			str, err = taskStoreMongo.NewStore(cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(str).ToNot(BeNil())
			collection = str.GetCollection("tasks")
		})

		Context("EnsureIndexes", func() {
			It("returns successfully", func() {
				Expect(str.EnsureIndexes()).To(Succeed())
				cursor, err := collection.Indexes().List(context.Background())
				Expect(err).ToNot(HaveOccurred())
				Expect(cursor).ToNot(BeNil())
				var indexes []storeStructuredMongoTest.MongoIndex
				err = cursor.All(context.Background(), &indexes)
				Expect(err).ToNot(HaveOccurred())

				Expect(indexes).To(ConsistOf(
					MatchFields(IgnoreExtras, Fields{
						"Key": Equal(storeStructuredMongoTest.MakeKeySlice("_id")),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key":        Equal(storeStructuredMongoTest.MakeKeySlice("id")),
						"Background": Equal(true),
						"Unique":     Equal(true),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key":        Equal(storeStructuredMongoTest.MakeKeySlice("name")),
						"Background": Equal(true),
						"Unique":     Equal(true),
						"Sparse":     Equal(true),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key":        Equal(storeStructuredMongoTest.MakeKeySlice("availableTime")),
						"Background": Equal(true),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key":        Equal(storeStructuredMongoTest.MakeKeySlice("state")),
						"Background": Equal(true),
					}),
				))
			})
		})

		Context("NewTaskRepository", func() {
			It("returns a new repository", func() {
				repository = str.NewTaskRepository()
				Expect(repository).ToNot(BeNil())
			})
		})

		Context("with a new repository", func() {
			var ctx context.Context

			BeforeEach(func() {
				repository = str.NewTaskRepository()
				Expect(repository).ToNot(BeNil())
				ctx = log.NewContextWithLogger(context.Background(), logger)
			})

			Context("with an existing task", func() {
				var tsk *task.Task

				BeforeEach(func() {
					var err error
					tsk, err = task.NewTask(context.Background(), &task.TaskCreate{
						Name:          pointer.FromString("test"),
						Type:          "fetch",
						Data:          nil,
						AvailableTime: pointer.FromTime(time.Now()),
					})
					Expect(err).ToNot(HaveOccurred())
					tsk.State = task.TaskStateRunning
					_, err = collection.InsertOne(ctx, tsk)
					Expect(err).ToNot(HaveOccurred())
				})

				Context("UnstickTasks", func() {
					It("returns an error when the context is missing", func() {
						unstuckTaskIDs, err := repository.UnstickTasks(context.Context(nil))
						Expect(err).To(MatchError("context is missing"))
						Expect(unstuckTaskIDs).To(BeNil())
					})

					It("returns no ids when there are no stuck tasks", func() {
						unstuckTaskIDs := test.Must(repository.UnstickTasks(ctx))
						Expect(unstuckTaskIDs).To(BeEmpty())
					})

					It("unsticks a running task with an expired deadline", func() {
						stuckTask := insertTaskWithStateAndDeadlineTime(ctx, collection, task.TaskStateRunning, pointer.FromTime(test.Now().Add(-time.Minute)))

						unstuckTaskIDs := test.Must(repository.UnstickTasks(ctx))
						Expect(unstuckTaskIDs).To(ConsistOf(stuckTask.ID))

						actualStuckTask := &task.Task{}
						Expect(collection.FindOne(ctx, bson.M{"id": stuckTask.ID}).Decode(actualStuckTask)).To(Succeed())
						Expect(actualStuckTask.State).To(Equal(task.TaskStatePending))
						Expect(actualStuckTask.AvailableTime).To(PointTo(BeTemporally("~", test.Now(), time.Second)))
						Expect(actualStuckTask.ModifiedTime).To(PointTo(BeTemporally("~", test.Now(), time.Second)))
						Expect(actualStuckTask.DeadlineTime).To(BeNil())
					})

					It("does not unstick a running task with a deadline in the future", func() {
						notStuckTask := insertTaskWithStateAndDeadlineTime(ctx, collection, task.TaskStateRunning, pointer.FromTime(test.Now().Add(time.Hour)))

						unstuckTaskIDs := test.Must(repository.UnstickTasks(ctx))
						Expect(unstuckTaskIDs).To(BeEmpty())

						actualNotStuckTask := &task.Task{}
						Expect(collection.FindOne(ctx, bson.M{"id": notStuckTask.ID}).Decode(actualNotStuckTask)).To(Succeed())
						Expect(actualNotStuckTask).To(Equal(notStuckTask))
					})

					It("does not unstick a task that is not running", func() {
						notStuckTask := insertTaskWithStateAndDeadlineTime(ctx, collection, task.TaskStatePending, pointer.FromTime(test.Now().Add(-time.Minute)))

						unstuckTaskIDs := test.Must(repository.UnstickTasks(ctx))
						Expect(unstuckTaskIDs).To(BeEmpty())

						actualNotStuckTask := &task.Task{}
						Expect(collection.FindOne(ctx, bson.M{"id": notStuckTask.ID}).Decode(actualNotStuckTask)).To(Succeed())
						Expect(actualNotStuckTask).To(Equal(notStuckTask))
					})

					It("unsticks multiple running tasks with expired deadlines, ordered by deadline time", func() {
						laterStuckTask := insertTaskWithStateAndDeadlineTime(ctx, collection, task.TaskStateRunning, pointer.FromTime(test.Now().Add(-time.Minute)))
						earlierStuckTask := insertTaskWithStateAndDeadlineTime(ctx, collection, task.TaskStateRunning, pointer.FromTime(test.Now().Add(-time.Hour)))

						unstuckTaskIDs := test.Must(repository.UnstickTasks(ctx))
						Expect(unstuckTaskIDs).To(Equal([]string{earlierStuckTask.ID, laterStuckTask.ID}))
					})

					It("only unsticks tasks matching the repository type filter", func() {
						stuckTask := insertTaskWithStateAndDeadlineTime(ctx, collection, task.TaskStateRunning, pointer.FromTime(test.Now().Add(-time.Minute)))
						otherTask := insertTaskWithStateAndDeadlineTime(ctx, collection, task.TaskStateRunning, pointer.FromTime(test.Now().Add(-time.Minute)))

						filteredRepository := str.WithTypeFilter(stuckTask.Type).NewTaskRepository()
						unstuckTaskIDs, err := filteredRepository.UnstickTasks(ctx)
						Expect(err).ToNot(HaveOccurred())
						Expect(unstuckTaskIDs).To(ConsistOf(stuckTask.ID))

						actualOtherTask := &task.Task{}
						Expect(collection.FindOne(ctx, bson.M{"id": otherTask.ID}).Decode(actualOtherTask)).To(Succeed())
						Expect(actualOtherTask).To(Equal(otherTask))
					})
				})
			})

			Context("StopTask", func() {
				It("clears the run time and duration when stopping without a duration", func() {
					pendingTask := insertTaskWithStateAndDeadlineTime(ctx, collection, task.TaskStatePending, nil)
					Expect(pendingTask.AvailableTime).ToNot(BeNil())

					startedTask := test.Must(repository.StartTask(ctx, pendingTask.ID, pendingTask.Revision, time.Minute))
					Expect(startedTask).ToNot(BeNil())
					Expect(startedTask.RunTime).ToNot(BeNil())
					Expect(startedTask.AvailableTime).To(BeNil())

					Expect(repository.StopTask(ctx, startedTask.ID, startedTask.StateLock, task.TaskStatePending, nil, nil)).To(Succeed())

					actualTask := &task.Task{}
					Expect(collection.FindOne(ctx, bson.M{"id": startedTask.ID}).Decode(actualTask)).To(Succeed())
					Expect(actualTask.State).To(Equal(task.TaskStatePending))
					Expect(actualTask.RunTime).To(BeNil())
					Expect(actualTask.Duration).To(BeNil())
					Expect(actualTask.StateLock).To(BeNil())
					Expect(actualTask.DeadlineTime).To(BeNil())
					Expect(actualTask.AvailableTime).To(BeNil())
				})

				It("retains the run time when stopping with a duration", func() {
					pendingTask := insertTaskWithStateAndDeadlineTime(ctx, collection, task.TaskStatePending, nil)
					Expect(pendingTask.AvailableTime).ToNot(BeNil())

					startedTask := test.Must(repository.StartTask(ctx, pendingTask.ID, pendingTask.Revision, time.Minute))
					Expect(startedTask).ToNot(BeNil())
					Expect(startedTask.RunTime).ToNot(BeNil())

					duration := time.Second
					Expect(repository.StopTask(ctx, startedTask.ID, startedTask.StateLock, task.TaskStateCompleted, &duration, nil)).To(Succeed())

					actualTask := &task.Task{}
					Expect(collection.FindOne(ctx, bson.M{"id": startedTask.ID}).Decode(actualTask)).To(Succeed())
					Expect(actualTask.State).To(Equal(task.TaskStateCompleted))
					Expect(actualTask.RunTime).To(PointTo(BeTemporally("~", *startedTask.RunTime, time.Millisecond)))
					Expect(actualTask.Duration).To(PointTo(Equal(duration.Seconds())))
					Expect(actualTask.StateLock).To(BeNil())
					Expect(actualTask.DeadlineTime).To(BeNil())
					Expect(actualTask.AvailableTime).To(BeNil())
				})
			})

			Context("EnsureEHRReconcileTask", func() {
				BeforeEach(func() {
					taskStoreMongo.TasksStateTotal.Reset()
				})

				It("creates the task and increments the pending metric only on the initial insert", func() {
					repository := str.TaskRepository()
					Expect(repository).ToNot(BeNil())

					Expect(repository.EnsureEHRReconcileTask(ctx)).To(Succeed())
					Expect(testutil.ToFloat64(taskStoreMongo.TasksStateTotal)).To(Equal(1.0))

					Expect(repository.EnsureEHRReconcileTask(ctx)).To(Succeed())
					Expect(testutil.ToFloat64(taskStoreMongo.TasksStateTotal)).To(Equal(1.0))

					count := test.Must(collection.CountDocuments(context.Background(), bson.M{"type": reconcile.Type}))
					Expect(count).To(Equal(int64(1)))
				})
			})
		})
	})
})

func insertTaskWithStateAndDeadlineTime(ctx context.Context, collection *mongo.Collection, state string, deadlineTime *time.Time) *task.Task {
	tsk := test.Must(task.NewTask(ctx, taskTest.RandomTaskCreate()))
	tsk.State = state
	tsk.DeadlineTime = deadlineTime
	result := test.Must(collection.InsertOne(ctx, tsk))
	Expect(collection.FindOne(ctx, bson.M{"_id": result.InsertedID}).Decode(tsk)).To(Succeed())
	return tsk
}
