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
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/page"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	"github.com/tidepool-org/platform/task"
	"github.com/tidepool-org/platform/task/queue"
	"github.com/tidepool-org/platform/task/queue/test"
	"github.com/tidepool-org/platform/task/store/mongo"
)

var (
	types        = []string{"first", "second"}
	tasksPerType = 50
)

var _ = Describe("multi queue", func() {
	var config *storeStructuredMongo.Config
	var queueConfig *queue.Config
	var lgr log.Logger
	var str *mongo.Store
	var multi *queue.MultiQueue

	BeforeEach(func() {
		config = storeStructuredMongoTest.NewConfig()
		var err error
		str, err = mongo.NewStore(config)
		Expect(err).ToNot(HaveOccurred())
		Expect(str).ToNot(BeNil())
		lgr = null.NewLogger()
		queueConfig = &queue.Config{Workers: 10, Delay: time.Millisecond, DelayInitial: time.Millisecond, DelayUnstick: queue.DelayUnstickDefault, StopWaitTimeout: queue.StopWaitTimeoutDefault, RunnerWatchdogGracePeriod: queue.RunnerWatchdogGracePeriodDefault}
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
			runners := make([]queue.Runner, 0, len(types))
			for _, typ := range types {
				runners = append(runners, test.NewCountingRunner(typ))
			}

			var err error
			multi, err = queue.NewMultiQueue(queueConfig, lgr, str, runners...)
			Expect(err).ToNot(HaveOccurred())
			Expect(multi).ToNot(BeNil())

			queues := multi.GetQueues()
			Expect(queues).To(HaveLen(len(types)))
			for _, typ := range types {
				Expect(queues).To(HaveKey(typ))
			}
		})

		It("returns an error when the config is missing", func() {
			invalidMulti, err := queue.NewMultiQueue(nil, lgr, str)
			Expect(err).To(MatchError("config is missing"))
			Expect(invalidMulti).To(BeNil())
		})

		It("returns an error when the logger is missing", func() {
			invalidMulti, err := queue.NewMultiQueue(queueConfig, nil, str)
			Expect(err).To(MatchError("logger is missing"))
			Expect(invalidMulti).To(BeNil())
		})

		It("returns an error when the store is missing", func() {
			invalidMulti, err := queue.NewMultiQueue(queueConfig, lgr, nil)
			Expect(err).To(MatchError("store is missing"))
			Expect(invalidMulti).To(BeNil())
		})

		It("returns an error when a runner is missing", func() {
			invalidMulti, err := queue.NewMultiQueue(queueConfig, lgr, str, nil)
			Expect(err).To(MatchError("runner is missing"))
			Expect(invalidMulti).To(BeNil())
		})

		It("returns an error when two runners have the same type", func() {
			invalidMulti, err := queue.NewMultiQueue(queueConfig, lgr, str, test.NewCountingRunner(types[0]), test.NewCountingRunner(types[0]))
			Expect(err).To(MatchError("runner type already registered"))
			Expect(invalidMulti).To(BeNil())
		})

		It("returns an error when a runner has invalid durations", func() {
			runner := test.NewSleepRunner(types[0], 2*time.Minute, 2*time.Minute, time.Minute, 0)
			invalidMulti, err := queue.NewMultiQueue(queueConfig, lgr, str, runner)
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
			countingRunners := make([]*test.CountingRunner, 0, len(types))
			runners := make([]queue.Runner, 0, len(types))
			now := time.Now()

			// Create tasks and runners for each task type
			for _, typ := range types {
				runner := test.NewCountingRunner(typ)
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
			multi, err = queue.NewMultiQueue(queueConfig, lgr, str, runners...)
			Expect(err).ToNot(HaveOccurred())
			Expect(multi).ToNot(BeNil())

			multi.Start()

			nonTerminalStates := []string{task.TaskStatePending, task.TaskStateRunning}

			// Wait until completion, within limits. On my local laptop, this typically
			// takes < 15 seconds when run via Gingko (no parallel), but under Go test
			// (parallel via package) it takes around 35 seconds. Who knows how long it
			// would take running in parallel on a CI host. So give it plenty of time.
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
