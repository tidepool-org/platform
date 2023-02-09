package mongo_test

import (
	"context"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/task"

	"go.mongodb.org/mongo-driver/mongo"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	taskStore "github.com/tidepool-org/platform/task/store"
	taskStoreMongo "github.com/tidepool-org/platform/task/store/mongo"
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
			str.Terminate(context.Background())
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
						"Key":        Equal(storeStructuredMongoTest.MakeKeySlice("priority")),
						"Background": Equal(true),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key":        Equal(storeStructuredMongoTest.MakeKeySlice("availableTime")),
						"Background": Equal(true),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key":        Equal(storeStructuredMongoTest.MakeKeySlice("expirationTime")),
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
					tsk, err = task.NewTask(&task.TaskCreate{
						Name:           pointer.FromString("test"),
						Type:           "fetch",
						Priority:       0,
						Data:           nil,
						AvailableTime:  pointer.FromTime(time.Now()),
						ExpirationTime: pointer.FromTime(time.Now().Add(5 * time.Minute)),
					})
					Expect(err).ToNot(HaveOccurred())
					tsk.State = task.TaskStateRunning
					_, err = collection.InsertOne(ctx, tsk)
					Expect(err).ToNot(HaveOccurred())
				})

				Context("UpdateFromState", func() {
					var updated *task.Task

					const defaultPrometheusOutput = `
							# HELP tidepool_task_tasks_state_total The total number of tasks sorted by state and type
							# TYPE tidepool_task_tasks_state_total counter
							tidepool_task_tasks_state_total{ state = "<state>", type = "<type>" } 1
						`
					const metricName = "tidepool_task_tasks_state_total"

					BeforeEach(func() {
						taskStoreMongo.TasksStateTotal.Reset()
						var err error
						updated, err = task.NewTask(&task.TaskCreate{
							Name:           pointer.FromString("updated"),
							Type:           "fetch",
							Priority:       0,
							Data:           nil,
							AvailableTime:  pointer.FromTime(time.Now()),
							ExpirationTime: pointer.FromTime(time.Now().Add(5 * time.Minute)),
						})
						Expect(err).ToNot(HaveOccurred())
						updated.ID = tsk.ID
					})

					It("returns an error when the context is missing", func() {
						ctx = nil
						result, err := repository.UpdateFromState(ctx, updated, tsk.State)
						errorsTest.ExpectEqual(err, errors.New("context is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the updated task is missing", func() {
						updated = nil
						result, err := repository.UpdateFromState(ctx, updated, tsk.State)
						errorsTest.ExpectEqual(err, errors.New("task is missing"))
						Expect(result).To(BeNil())
					})

					It("successfully fails the task with multiple errors", func() {
						updated.State = task.TaskStateFailed
						updated.AppendError(errors.New("first error"))
						updated.AppendError(errors.New("second error"))
						_, err := repository.UpdateFromState(ctx, updated, tsk.State)
						Expect(err).ToNot(HaveOccurred())

						result := task.Task{}
						err = collection.FindOne(ctx, bson.M{"id": tsk.ID}).Decode(&result)
						Expect(err).ToNot(HaveOccurred())
						Expect(result.State).To(Equal(updated.State))
						Expect(result.Error).To(Equal(updated.Error))
					})

					It("returns error if task is updated from the same state multiple times", func() {
						updated.State = task.TaskStatePending
						_, err := repository.UpdateFromState(ctx, updated, tsk.State)
						Expect(err).ToNot(HaveOccurred())

						_, err = repository.UpdateFromState(ctx, updated, tsk.State)
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("Task has already been claimed or is now unavailable."))
					})

					It("records metrics of completed tasks", func() {
						updated.State = task.TaskStateCompleted
						completedTask, err := repository.UpdateFromState(ctx, updated, tsk.State)

						Expect(err).ToNot(HaveOccurred())

						prometheusState := strings.ReplaceAll(defaultPrometheusOutput, "<state>", task.TaskStateCompleted)
						expectedOutput := strings.ReplaceAll(prometheusState, "<type>", completedTask.Type)

						prometheusErr := testutil.
							CollectAndCompare(taskStoreMongo.TasksStateTotal, strings.NewReader(expectedOutput), metricName)
						Expect(prometheusErr).ToNot(HaveOccurred())
					})

					It("records metrics of failed tasks", func() {
						updated.State = task.TaskStateFailed
						failedTask, err := repository.UpdateFromState(ctx, updated, tsk.State)

						Expect(err).ToNot(HaveOccurred())

						prometheusState := strings.ReplaceAll(defaultPrometheusOutput, "<state>", task.TaskStateFailed)
						expectedOutput := strings.ReplaceAll(prometheusState, "<type>", failedTask.Type)

						prometheusErr := testutil.
							CollectAndCompare(taskStoreMongo.TasksStateTotal, strings.NewReader(expectedOutput), metricName)
						Expect(prometheusErr).ToNot(HaveOccurred())
					})

					It("records metrics of running tasks", func() {
						updated.State = task.TaskStateRunning
						runningTask, err := repository.UpdateFromState(ctx, updated, tsk.State)

						Expect(err).ToNot(HaveOccurred())

						prometheusState := strings.ReplaceAll(defaultPrometheusOutput, "<state>", task.TaskStateRunning)
						expectedOutput := strings.ReplaceAll(prometheusState, "<type>", runningTask.Type)

						prometheusErr := testutil.
							CollectAndCompare(taskStoreMongo.TasksStateTotal, strings.NewReader(expectedOutput), metricName)
						Expect(prometheusErr).ToNot(HaveOccurred())
					})

					It("records metrics of pending tasks", func() {
						tskCreate := &task.TaskCreate{
							Name:           pointer.FromString("test"),
							Type:           "fetch",
							Priority:       0,
							Data:           nil,
							AvailableTime:  pointer.FromTime(time.Now()),
							ExpirationTime: pointer.FromTime(time.Now().Add(5 * time.Minute)),
						}
						_, err := repository.CreateTask(ctx, tskCreate)
						Expect(err).ToNot(HaveOccurred())

						prometheusState := strings.ReplaceAll(defaultPrometheusOutput, "<state>", task.TaskStatePending)
						expectedOutput := strings.ReplaceAll(prometheusState, "<type>", tsk.Type)

						prometheusErr := testutil.
							CollectAndCompare(taskStoreMongo.TasksStateTotal, strings.NewReader(expectedOutput), metricName)
						Expect(prometheusErr).ToNot(HaveOccurred())
					})

				})
			})
		})
	})
})
