package mongo_test

import (
	"context"
	"strconv"
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
						"Key":    Equal(storeStructuredMongoTest.MakeKeySlice("id")),
						"Unique": Equal(true),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key":    Equal(storeStructuredMongoTest.MakeKeySlice("name")),
						"Unique": Equal(true),
						"Sparse": Equal(true),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key": Equal(storeStructuredMongoTest.MakeKeySlice("availableTime")),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key": Equal(storeStructuredMongoTest.MakeKeySlice("state")),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key": Equal(storeStructuredMongoTest.MakeKeySlice("type", "availableTime")),
						"PartialFilterExpression": Equal(bson.D{
							{Key: "state", Value: task.TaskStatePending},
						}),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key": Equal(storeStructuredMongoTest.MakeKeySlice("type", "deadlineTime")),
						"PartialFilterExpression": Equal(bson.D{
							{Key: "state", Value: task.TaskStateRunning},
						}),
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

			Context("StartTask", func() {
				It("embeds the claimed revision in the claim token", func() {
					pendingTask := insertTaskWithStateAndDeadlineTime(ctx, collection, task.TaskStatePending, nil)

					startedTask := test.Must(repository.StartTask(ctx, pendingTask.ID, pendingTask.Revision, time.Minute))
					Expect(startedTask).ToNot(BeNil())
					Expect(startedTask.ClaimToken).To(PointTo(HavePrefix(strconv.Itoa(pendingTask.Revision) + ":")))
				})

				It("generates a distinct claim token for successive claims of the same task", func() {
					pendingTask := insertTaskWithStateAndDeadlineTime(ctx, collection, task.TaskStatePending, nil)

					firstStartedTask := test.Must(repository.StartTask(ctx, pendingTask.ID, pendingTask.Revision, time.Minute))
					Expect(firstStartedTask).ToNot(BeNil())
					Expect(repository.StopTask(ctx, firstStartedTask.ID, firstStartedTask.Revision, firstStartedTask.ClaimToken, task.TaskStatePending, nil, nil)).To(Succeed())

					stoppedTask := &task.Task{}
					Expect(collection.FindOne(ctx, bson.M{"id": pendingTask.ID}).Decode(stoppedTask)).To(Succeed())

					secondStartedTask := test.Must(repository.StartTask(ctx, stoppedTask.ID, stoppedTask.Revision, time.Minute))
					Expect(secondStartedTask).ToNot(BeNil())
					Expect(secondStartedTask.ClaimToken).To(PointTo(Not(Equal(*firstStartedTask.ClaimToken))))
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

					Expect(repository.StopTask(ctx, startedTask.ID, startedTask.Revision, startedTask.ClaimToken, task.TaskStatePending, nil, nil)).To(Succeed())

					actualTask := &task.Task{}
					Expect(collection.FindOne(ctx, bson.M{"id": startedTask.ID}).Decode(actualTask)).To(Succeed())
					Expect(actualTask.State).To(Equal(task.TaskStatePending))
					Expect(actualTask.RunTime).To(BeNil())
					Expect(actualTask.Duration).To(BeNil())
					Expect(actualTask.ClaimToken).To(BeNil())
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
					Expect(repository.StopTask(ctx, startedTask.ID, startedTask.Revision, startedTask.ClaimToken, task.TaskStateCompleted, &duration, nil)).To(Succeed())

					actualTask := &task.Task{}
					Expect(collection.FindOne(ctx, bson.M{"id": startedTask.ID}).Decode(actualTask)).To(Succeed())
					Expect(actualTask.State).To(Equal(task.TaskStateCompleted))
					Expect(actualTask.RunTime).To(PointTo(BeTemporally("~", *startedTask.RunTime, time.Millisecond)))
					Expect(actualTask.Duration).To(PointTo(Equal(duration.Seconds())))
					Expect(actualTask.ClaimToken).To(BeNil())
					Expect(actualTask.DeadlineTime).To(BeNil())
					Expect(actualTask.AvailableTime).To(BeNil())
				})
			})

			Context("UnstickTasks", func() {
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

				It("returns an error when the context is missing", func() {
					unstuckTaskIDs, err := repository.UnstickTasks(context.Context(nil), 0)
					Expect(err).To(MatchError("context is missing"))
					Expect(unstuckTaskIDs).To(BeNil())
				})

				It("returns an error when the availability delay is negative", func() {
					unstuckTaskIDs, err := repository.UnstickTasks(ctx, -time.Second)
					Expect(err).To(MatchError("availability delay is invalid"))
					Expect(unstuckTaskIDs).To(BeNil())
				})

				It("returns no ids when there are no stuck tasks", func() {
					unstuckTaskIDs := test.Must(repository.UnstickTasks(ctx, 0))
					Expect(unstuckTaskIDs).To(BeEmpty())
				})

				It("unsticks a running task with an expired deadline", func() {
					stuckTask := insertTaskWithStateAndDeadlineTime(ctx, collection, task.TaskStateRunning, pointer.FromTime(test.Now().Add(-time.Minute)))

					unstuckTaskIDs := test.Must(repository.UnstickTasks(ctx, 0))
					Expect(unstuckTaskIDs).To(ConsistOf(stuckTask.ID))

					actualStuckTask := &task.Task{}
					Expect(collection.FindOne(ctx, bson.M{"id": stuckTask.ID}).Decode(actualStuckTask)).To(Succeed())
					Expect(actualStuckTask.State).To(Equal(task.TaskStatePending))
					Expect(actualStuckTask.AvailableTime).To(PointTo(BeTemporally("~", test.Now(), time.Second)))
					Expect(actualStuckTask.ModifiedTime).To(PointTo(BeTemporally("~", test.Now(), time.Second)))
					Expect(actualStuckTask.DeadlineTime).To(BeNil())
				})

				It("offsets the available time by the availability delay", func() {
					stuckTask := insertTaskWithStateAndDeadlineTime(ctx, collection, task.TaskStateRunning, pointer.FromTime(test.Now().Add(-time.Minute)))

					unstuckTaskIDs := test.Must(repository.UnstickTasks(ctx, time.Minute))
					Expect(unstuckTaskIDs).To(ConsistOf(stuckTask.ID))

					actualStuckTask := &task.Task{}
					Expect(collection.FindOne(ctx, bson.M{"id": stuckTask.ID}).Decode(actualStuckTask)).To(Succeed())
					Expect(actualStuckTask.State).To(Equal(task.TaskStatePending))
					Expect(actualStuckTask.AvailableTime).To(PointTo(BeTemporally("~", test.Now().Add(time.Minute), time.Second)))
				})

				It("does not unstick a running task with a deadline in the future", func() {
					notStuckTask := insertTaskWithStateAndDeadlineTime(ctx, collection, task.TaskStateRunning, pointer.FromTime(test.Now().Add(time.Hour)))

					unstuckTaskIDs := test.Must(repository.UnstickTasks(ctx, 0))
					Expect(unstuckTaskIDs).To(BeEmpty())

					actualNotStuckTask := &task.Task{}
					Expect(collection.FindOne(ctx, bson.M{"id": notStuckTask.ID}).Decode(actualNotStuckTask)).To(Succeed())
					Expect(actualNotStuckTask).To(Equal(notStuckTask))
				})

				It("does not unstick a task that is not running", func() {
					notStuckTask := insertTaskWithStateAndDeadlineTime(ctx, collection, task.TaskStatePending, pointer.FromTime(test.Now().Add(-time.Minute)))

					unstuckTaskIDs := test.Must(repository.UnstickTasks(ctx, 0))
					Expect(unstuckTaskIDs).To(BeEmpty())

					actualNotStuckTask := &task.Task{}
					Expect(collection.FindOne(ctx, bson.M{"id": notStuckTask.ID}).Decode(actualNotStuckTask)).To(Succeed())
					Expect(actualNotStuckTask).To(Equal(notStuckTask))
				})

				It("unsticks multiple running tasks with expired deadlines, ordered by deadline time", func() {
					laterStuckTask := insertTaskWithStateAndDeadlineTime(ctx, collection, task.TaskStateRunning, pointer.FromTime(test.Now().Add(-time.Minute)))
					earlierStuckTask := insertTaskWithStateAndDeadlineTime(ctx, collection, task.TaskStateRunning, pointer.FromTime(test.Now().Add(-time.Hour)))

					unstuckTaskIDs := test.Must(repository.UnstickTasks(ctx, 0))
					Expect(unstuckTaskIDs).To(Equal([]string{earlierStuckTask.ID, laterStuckTask.ID}))
				})

				It("only unsticks tasks matching the repository type filter", func() {
					stuckTask := insertTaskWithStateAndDeadlineTime(ctx, collection, task.TaskStateRunning, pointer.FromTime(test.Now().Add(-time.Minute)))
					otherTask := insertTaskWithStateAndDeadlineTime(ctx, collection, task.TaskStateRunning, pointer.FromTime(test.Now().Add(-time.Minute)))

					filteredRepository := str.WithTypeFilter(stuckTask.Type).NewTaskRepository()
					unstuckTaskIDs, err := filteredRepository.UnstickTasks(ctx, 0)
					Expect(err).ToNot(HaveOccurred())
					Expect(unstuckTaskIDs).To(ConsistOf(stuckTask.ID))

					actualOtherTask := &task.Task{}
					Expect(collection.FindOne(ctx, bson.M{"id": otherTask.ID}).Decode(actualOtherTask)).To(Succeed())
					Expect(actualOtherTask).To(Equal(otherTask))
				})
			})

			Context("GetTaskClaimTokens", func() {
				It("returns an error when the context is missing", func() {
					claimTokens, err := repository.GetTaskClaimTokens(context.Context(nil), []string{taskTest.RandomID()})
					Expect(err).To(MatchError("context is missing"))
					Expect(claimTokens).To(BeNil())
				})

				It("returns an error when the ids are missing", func() {
					claimTokens, err := repository.GetTaskClaimTokens(ctx, nil)
					Expect(err).To(MatchError("ids is missing"))
					Expect(claimTokens).To(BeNil())
				})

				It("omits a task that does not exist", func() {
					claimTokens, err := repository.GetTaskClaimTokens(ctx, []string{taskTest.RandomID()})
					Expect(err).ToNot(HaveOccurred())
					Expect(claimTokens).To(BeEmpty())
				})

				It("returns a nil claim token when the task exists, but is not claimed", func() {
					pendingTask := insertTaskWithStateAndDeadlineTime(ctx, collection, task.TaskStatePending, nil)
					Expect(pendingTask.ClaimToken).To(BeNil())

					claimTokens, err := repository.GetTaskClaimTokens(ctx, []string{pendingTask.ID})
					Expect(err).ToNot(HaveOccurred())
					Expect(claimTokens).To(HaveKeyWithValue(pendingTask.ID, BeNil()))
				})

				It("returns the claim token of a claimed task", func() {
					pendingTask := insertTaskWithStateAndDeadlineTime(ctx, collection, task.TaskStatePending, nil)

					startedTask := test.Must(repository.StartTask(ctx, pendingTask.ID, pendingTask.Revision, time.Minute))
					Expect(startedTask).ToNot(BeNil())
					Expect(startedTask.ClaimToken).ToNot(BeNil())

					claimTokens, err := repository.GetTaskClaimTokens(ctx, []string{startedTask.ID})
					Expect(err).ToNot(HaveOccurred())
					Expect(claimTokens).To(HaveKeyWithValue(startedTask.ID, PointTo(Equal(*startedTask.ClaimToken))))
				})

				It("returns the claim token of each of several tasks, omitting those that do not exist", func() {
					firstPendingTask := insertTaskWithStateAndDeadlineTime(ctx, collection, task.TaskStatePending, nil)
					secondPendingTask := insertTaskWithStateAndDeadlineTime(ctx, collection, task.TaskStatePending, nil)
					missingTaskID := taskTest.RandomID()

					firstStartedTask := test.Must(repository.StartTask(ctx, firstPendingTask.ID, firstPendingTask.Revision, time.Minute))
					Expect(firstStartedTask).ToNot(BeNil())
					secondStartedTask := test.Must(repository.StartTask(ctx, secondPendingTask.ID, secondPendingTask.Revision, time.Minute))
					Expect(secondStartedTask).ToNot(BeNil())

					claimTokens, err := repository.GetTaskClaimTokens(ctx, []string{firstStartedTask.ID, missingTaskID, secondStartedTask.ID})
					Expect(err).ToNot(HaveOccurred())
					Expect(claimTokens).To(HaveLen(2))
					Expect(claimTokens).To(HaveKeyWithValue(firstStartedTask.ID, PointTo(Equal(*firstStartedTask.ClaimToken))))
					Expect(claimTokens).To(HaveKeyWithValue(secondStartedTask.ID, PointTo(Equal(*secondStartedTask.ClaimToken))))
				})

				It("returns a nil claim token once the task is stopped", func() {
					pendingTask := insertTaskWithStateAndDeadlineTime(ctx, collection, task.TaskStatePending, nil)

					startedTask := test.Must(repository.StartTask(ctx, pendingTask.ID, pendingTask.Revision, time.Minute))
					Expect(startedTask).ToNot(BeNil())
					Expect(repository.StopTask(ctx, startedTask.ID, startedTask.Revision, startedTask.ClaimToken, task.TaskStateCompleted, nil, nil)).To(Succeed())

					claimTokens, err := repository.GetTaskClaimTokens(ctx, []string{startedTask.ID})
					Expect(err).ToNot(HaveOccurred())
					Expect(claimTokens).To(HaveKeyWithValue(startedTask.ID, BeNil()))
				})

				It("returns the new claim token once the task is re-claimed", func() {
					pendingTask := insertTaskWithStateAndDeadlineTime(ctx, collection, task.TaskStatePending, nil)

					firstStartedTask := test.Must(repository.StartTask(ctx, pendingTask.ID, pendingTask.Revision, time.Minute))
					Expect(firstStartedTask).ToNot(BeNil())
					Expect(repository.StopTask(ctx, firstStartedTask.ID, firstStartedTask.Revision, firstStartedTask.ClaimToken, task.TaskStatePending, nil, nil)).To(Succeed())

					stoppedTask := &task.Task{}
					Expect(collection.FindOne(ctx, bson.M{"id": pendingTask.ID}).Decode(stoppedTask)).To(Succeed())

					secondStartedTask := test.Must(repository.StartTask(ctx, stoppedTask.ID, stoppedTask.Revision, time.Minute))
					Expect(secondStartedTask).ToNot(BeNil())

					claimTokens, err := repository.GetTaskClaimTokens(ctx, []string{secondStartedTask.ID})
					Expect(err).ToNot(HaveOccurred())
					Expect(claimTokens).To(HaveKeyWithValue(secondStartedTask.ID, PointTo(Equal(*secondStartedTask.ClaimToken))))
					Expect(claimTokens).To(HaveKeyWithValue(secondStartedTask.ID, PointTo(Not(Equal(*firstStartedTask.ClaimToken)))))
				})

				It("does not modify the task", func() {
					pendingTask := insertTaskWithStateAndDeadlineTime(ctx, collection, task.TaskStatePending, nil)

					startedTask := test.Must(repository.StartTask(ctx, pendingTask.ID, pendingTask.Revision, time.Minute))
					Expect(startedTask).ToNot(BeNil())

					_, err := repository.GetTaskClaimTokens(ctx, []string{startedTask.ID})
					Expect(err).ToNot(HaveOccurred())

					actualTask := &task.Task{}
					Expect(collection.FindOne(ctx, bson.M{"id": startedTask.ID}).Decode(actualTask)).To(Succeed())
					Expect(actualTask).To(Equal(startedTask))
				})

				It("returns the claim token of a task matching the repository type filter", func() {
					pendingTask := insertTaskWithStateAndDeadlineTime(ctx, collection, task.TaskStatePending, nil)

					startedTask := test.Must(repository.StartTask(ctx, pendingTask.ID, pendingTask.Revision, time.Minute))
					Expect(startedTask).ToNot(BeNil())

					filteredRepository := str.WithTypeFilter(startedTask.Type).NewTaskRepository()
					claimTokens, err := filteredRepository.GetTaskClaimTokens(ctx, []string{startedTask.ID})
					Expect(err).ToNot(HaveOccurred())
					Expect(claimTokens).To(HaveKeyWithValue(startedTask.ID, PointTo(Equal(*startedTask.ClaimToken))))
				})

				It("omits a task that does not match the repository type filter", func() {
					pendingTask := insertTaskWithStateAndDeadlineTime(ctx, collection, task.TaskStatePending, nil)

					startedTask := test.Must(repository.StartTask(ctx, pendingTask.ID, pendingTask.Revision, time.Minute))
					Expect(startedTask).ToNot(BeNil())

					filteredRepository := str.WithTypeFilter(startedTask.Type + "-other").NewTaskRepository()
					claimTokens, err := filteredRepository.GetTaskClaimTokens(ctx, []string{startedTask.ID})
					Expect(err).ToNot(HaveOccurred())
					Expect(claimTokens).To(BeEmpty())
				})
			})

			Context("EnsureEHRReconcileTask", func() {
				BeforeEach(func() {
					taskStoreMongo.TypeStateTotal.Reset()
				})

				It("creates the task and increments the pending metric only on the initial insert", func() {
					repository := str.TaskRepository()
					Expect(repository).ToNot(BeNil())

					Expect(repository.EnsureEHRReconcileTask(ctx)).To(Succeed())
					Expect(testutil.ToFloat64(taskStoreMongo.TypeStateTotal)).To(Equal(1.0))

					Expect(repository.EnsureEHRReconcileTask(ctx)).To(Succeed())
					Expect(testutil.ToFloat64(taskStoreMongo.TypeStateTotal)).To(Equal(1.0))

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
