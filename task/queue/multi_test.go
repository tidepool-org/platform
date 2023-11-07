package queue_test

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
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

		var q queue.Queue
		q, err = queue.NewMultiQueue(
			&queue.Config{
				Workers: 10,
				Delay:   1,
			},
			lgr,
			str,
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(q).ToNot(BeNil())

		var ok bool
		multi, ok = q.(*queue.MultiQueue)
		Expect(ok).To(BeTrue())
	})

	AfterEach(func() {
		Expect(str.Terminate(context.Background())).To(Succeed())
	})

	Describe("Register Runner", func() {
		It("Creates a new queue for each runner type", func() {
			for _, t := range types {
				runner := test.NewCountingRunner(t)
				Expect(multi.RegisterRunner(runner)).To(Succeed())
				Expect(multi.GetQueues()).To(HaveKey(t))
			}
			queues := multi.GetQueues()
			Expect(queues).To(HaveLen(len(types)))
		})
	})

	Describe("Tasks", func() {
		BeforeEach(func() {
			_, err := str.GetRepository("tasks").DeleteMany(context.Background(), bson.M{})
			Expect(err).To(Succeed())
		})

		It("Are partitioned correctly", func() {
			ctx := log.NewContextWithLogger(context.Background(), lgr)
			tasks := make([]*task.Task, 0, len(types)*tasksPerType)
			creates := make([]*task.TaskCreate, 0, len(types)*tasksPerType)
			runners := make([]*test.CountingRunner, 0, len(types))
			now := time.Now()

			// Create tasks and runners for each task type
			for _, t := range types {
				runner := test.NewCountingRunner(t)
				runners = append(runners, runner)
				Expect(multi.RegisterRunner(runner)).To(Succeed())

				for i := 0; i < tasksPerType; i++ {
					name := fmt.Sprintf("%v:%v", t, i)
					creates = append(creates, &task.TaskCreate{
						Name:          &name,
						Type:          t,
						AvailableTime: &now,
					})
				}
			}

			// Insert tasks in the database
			rand.Shuffle(len(creates), func(i, j int) { creates[i], creates[j] = creates[j], creates[i] })
			for _, create := range creates {
				create := create
				tsk, err := str.NewTaskRepository().CreateTask(ctx, create)
				Expect(err).ToNot(HaveOccurred())
				Expect(tsk).ToNot(BeNil())
				tasks = append(tasks, tsk)
			}

			// Register runners from all types in the underlying queue
			// To make sure they are empty when all work is processed
			expectedNoopRunners := make([]*test.CountingRunner, 0)
			for typ, q := range multi.GetQueues() {
				for _, t := range types {
					if typ != t {
						runner := test.NewCountingRunner(t)
						expectedNoopRunners = append(expectedNoopRunners, runner)
						Expect(q.RegisterRunner(runner)).To(Succeed())
					}
				}

			}

			multi.Start()

			// Wait until completion (up to 10 seconds)
			tCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
			ticker := time.NewTicker(200 * time.Millisecond)
			wg := &sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer ticker.Stop()
				defer cancel()

				for {
					select {
					case <-tCtx.Done():
						return
					case <-ticker.C:
						state := task.TaskStatePending
						pending, err := str.NewTaskRepository().ListTasks(ctx, &task.TaskFilter{
							State: &state,
						}, &page.Pagination{
							Page: 0,
							Size: 10,
						})
						if err == nil && len(pending) == 0 {
							return
						}
						Expect(err).ToNot(HaveOccurred())
					}
				}
			}()
			wg.Wait()

			for _, runner := range runners {
				// Check all tasks were processed
				Expect(runner.GetCount()).To(Equal(tasksPerType))
			}
			for _, runner := range expectedNoopRunners {
				// Check extra runners didn't do any work
				Expect(runner.GetCount()).To(Equal(0))
			}
		})
	})
})
