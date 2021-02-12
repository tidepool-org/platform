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
							# HELP tidepool_task_tasks_<state>_total The total number of <state> tasks
							# TYPE tidepool_task_tasks_<state>_total counter
							tidepool_task_tasks_<state>_total{ task_type = "<task_type>" } 1
						`

					BeforeEach(func() {
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

					It("records metrics of completed tasks", func() {
						completedTask, err := repository.UpdateFromState(ctx, updated, task.TaskStateCompleted)

						Expect(err).ToNot(HaveOccurred())

						prometheusState := strings.ReplaceAll(defaultPrometheusOutput, "<state>", task.TaskStateCompleted)
						expectedOutput := strings.ReplaceAll(prometheusState, "<task_type>", completedTask.Type)

						prometheusErr := testutil.
							CollectAndCompare(taskStoreMongo.TasksCompletedTotal, strings.NewReader(expectedOutput), "tidepool_task_tasks_completed_total")
						Expect(prometheusErr).ToNot(HaveOccurred())
					})

					It("records metrics of failed tasks", func() {
						failedTask, err := repository.UpdateFromState(ctx, updated, task.TaskStateFailed)

						Expect(err).ToNot(HaveOccurred())

						prometheusState := strings.ReplaceAll(defaultPrometheusOutput, "<state>", task.TaskStateFailed)
						expectedOutput := strings.ReplaceAll(prometheusState, "<task_type>", failedTask.Type)

						prometheusErr := testutil.
							CollectAndCompare(taskStoreMongo.TasksFailedTotal, strings.NewReader(expectedOutput), "tidepool_task_tasks_failed_total")
						Expect(prometheusErr).ToNot(HaveOccurred())
					})

					It("records metrics of created tasks", func() {
						taskStoreMongo.TasksCreatedTotal.Reset()
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

						prometheusState := strings.ReplaceAll(defaultPrometheusOutput, "<state>", "created")
						expectedOutput := strings.ReplaceAll(prometheusState, "<task_type>", tsk.Type)

						prometheusErr := testutil.
							CollectAndCompare(taskStoreMongo.TasksCreatedTotal, strings.NewReader(expectedOutput), "tidepool_task_tasks_created_total")
						Expect(prometheusErr).ToNot(HaveOccurred())
					})

				})
			})
		})
	})
})
