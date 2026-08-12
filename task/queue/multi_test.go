package queue_test

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/tidepool-org/platform/log"
	logNull "github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/page"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	"github.com/tidepool-org/platform/task"
	taskQueue "github.com/tidepool-org/platform/task/queue"
	taskQueueTest "github.com/tidepool-org/platform/task/queue/test"
	taskStoreMongo "github.com/tidepool-org/platform/task/store/mongo"
)

var (
	types        = []string{"first", "second"}
	tasksPerType = 50
)

var _ = Describe("multi queue", func() {
	var config *storeStructuredMongo.Config
	var queueConfig *taskQueue.Config
	var lgr log.Logger
	var str *taskStoreMongo.Store
	var multi *taskQueue.MultiQueue

	BeforeEach(func() {
		config = storeStructuredMongoTest.NewConfig()
		var err error
		str, err = taskStoreMongo.NewStore(config)
		Expect(err).ToNot(HaveOccurred())
		Expect(str).ToNot(BeNil())
		lgr = logNull.NewLogger()
		queueConfig = taskQueue.NewConfig()
		queueConfig.Workers = 10
		queueConfig.StartManagerDelay = time.Millisecond
		queueConfig.DispatchTasksDelay = time.Millisecond
		multi = nil
	})

	AfterEach(func() {
		if multi != nil {
			multi.Stop()
		}
		Expect(str.Terminate(context.Background())).To(Succeed())
	})

	Describe("NewMultiQueue", func() {
		It("creates a new queue for each runner type", func() {
			runners := make([]taskQueue.Runner, 0, len(types))
			for _, typ := range types {
				runners = append(runners, taskQueueTest.NewCountingRunner(typ))
			}

			var err error
			multi, err = taskQueue.NewMultiQueue(queueConfig, lgr, str, runners...)
			Expect(err).ToNot(HaveOccurred())
			Expect(multi).ToNot(BeNil())

			queues := multi.GetQueues()
			Expect(queues).To(HaveLen(len(types)))
			for _, typ := range types {
				Expect(queues).To(HaveKey(typ))
			}
		})

		It("returns an error when the config is missing", func() {
			invalidMulti, err := taskQueue.NewMultiQueue(nil, lgr, str)
			Expect(err).To(MatchError("config is missing"))
			Expect(invalidMulti).To(BeNil())
		})

		It("returns an error when the logger is missing", func() {
			invalidMulti, err := taskQueue.NewMultiQueue(queueConfig, nil, str)
			Expect(err).To(MatchError("logger is missing"))
			Expect(invalidMulti).To(BeNil())
		})

		It("returns an error when the store is missing", func() {
			invalidMulti, err := taskQueue.NewMultiQueue(queueConfig, lgr, nil)
			Expect(err).To(MatchError("store is missing"))
			Expect(invalidMulti).To(BeNil())
		})

		It("returns an error when a runner is missing", func() {
			invalidMulti, err := taskQueue.NewMultiQueue(queueConfig, lgr, str, nil)
			Expect(err).To(MatchError("runner is missing"))
			Expect(invalidMulti).To(BeNil())
		})

		It("returns an error when two runners have the same type", func() {
			invalidMulti, err := taskQueue.NewMultiQueue(queueConfig, lgr, str, taskQueueTest.NewCountingRunner(types[0]), taskQueueTest.NewCountingRunner(types[0]))
			Expect(err).To(MatchError("runner type already registered"))
			Expect(invalidMulti).To(BeNil())
		})

		It("returns an error when a runner has invalid durations", func() {
			runner := taskQueueTest.NewSleepRunner(types[0], 2*time.Minute, 2*time.Minute, time.Minute, 0)
			invalidMulti, err := taskQueue.NewMultiQueue(queueConfig, lgr, str, runner)
			Expect(err).To(MatchError("runner deadline is invalid"))
			Expect(invalidMulti).To(BeNil())
		})
	})

	Describe("Tasks", func() {
		BeforeEach(func() {
			_, err := str.GetRepository("tasks").DeleteMany(context.Background(), bson.M{})
			Expect(err).To(Succeed())
		})

		It("Are partitioned correctly", func() {
			ctx := log.NewContextWithLogger(context.Background(), lgr)
			creates := make([]*task.TaskCreate, 0, len(types)*tasksPerType)
			countingRunners := make([]*taskQueueTest.CountingRunner, 0, len(types))
			runners := make([]taskQueue.Runner, 0, len(types))
			now := time.Now()

			// Create tasks and runners for each task type
			for _, typ := range types {
				runner := taskQueueTest.NewCountingRunner(typ)
				countingRunners = append(countingRunners, runner)
				runners = append(runners, runner)

				for index := 0; index < tasksPerType; index++ {
					name := fmt.Sprintf("%v:%v", typ, index)
					creates = append(creates, &task.TaskCreate{
						Name:          &name,
						Type:          typ,
						AvailableTime: &now,
					})
				}
			}

			// Insert tasks in the database
			rand.Shuffle(len(creates), func(i, j int) { creates[i], creates[j] = creates[j], creates[i] })
			for _, create := range creates {
				tsk, err := str.NewTaskRepository().CreateTask(ctx, create)
				Expect(err).ToNot(HaveOccurred())
				Expect(tsk).ToNot(BeNil())
			}

			var err error
			multi, err = taskQueue.NewMultiQueue(queueConfig, lgr, str, runners...)
			Expect(err).ToNot(HaveOccurred())
			Expect(multi).ToNot(BeNil())

			multi.Start()

			nonTerminalStates := []string{task.TaskStatePending, task.TaskStateRunning}

			// Wait until completion, within limits. On my local laptop, this typically takes < 15 seconds when run via
			// Gingko (no parallel), but under Go test (parallel via package) it takes around 35 seconds. Who knows how
			// long it would take running in parallel on a CI host. So give it plenty of time.
			tCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
			defer cancel()

			ticker := time.NewTicker(200 * time.Millisecond)
			defer ticker.Stop()

		loop:
			for {
				select {
				case <-tCtx.Done():
					Fail("the test did not fail; it ran out of time, give it more time")
					break loop
				case <-ticker.C:
					nonTerminatedTasks := 0
					for _, state := range nonTerminalStates {
						pending, err := str.NewTaskRepository().ListTasks(ctx, &task.TaskFilter{
							State: &state,
						}, &page.Pagination{
							Page: 0,
							Size: 10,
						})
						Expect(err).ToNot(HaveOccurred())
						nonTerminatedTasks += len(pending)
					}
					if nonTerminatedTasks == 0 {
						break loop
					}
				}
			}

			expected := map[string]int{}
			for _, typ := range types {
				expected[typ] = tasksPerType
			}
			results := map[string]int{}
			for _, runner := range countingRunners {
				results[runner.GetRunnerType()] = runner.GetCount()
			}

			Expect(results).To(Equal(expected))
		})
	})
})
