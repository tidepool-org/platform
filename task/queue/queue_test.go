package queue_test

import (
	"context"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"go.mongodb.org/mongo-driver/bson"

	configTest "github.com/tidepool-org/platform/config/test"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	metadataTest "github.com/tidepool-org/platform/metadata/test"
	"github.com/tidepool-org/platform/pointer"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	"github.com/tidepool-org/platform/task"
	taskQueue "github.com/tidepool-org/platform/task/queue"
	taskQueueTest "github.com/tidepool-org/platform/task/queue/test"
	taskStore "github.com/tidepool-org/platform/task/store"
	taskStoreMongo "github.com/tidepool-org/platform/task/store/mongo"
	taskStoreTest "github.com/tidepool-org/platform/task/store/test"
	taskTest "github.com/tidepool-org/platform/task/test"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Queue", func() {
	It("WorkersDefault is expected", func() {
		Expect(taskQueue.WorkersDefault).To(Equal(5))
	})

	It("StartManagerDelayDefault is expected", func() {
		Expect(taskQueue.StartManagerDelayDefault).To(Equal(1 * time.Minute))
	})

	It("DispatchTasksDelayDefault is expected", func() {
		Expect(taskQueue.DispatchTasksDelayDefault).To(Equal(1 * time.Minute))
	})

	It("MonitorTaskDelayDefault is expected", func() {
		Expect(taskQueue.MonitorTaskDelayDefault).To(Equal(1 * time.Minute))
	})

	It("RunnerWatchdogGracePeriodDefault is expected", func() {
		Expect(taskQueue.RunnerWatchdogGracePeriodDefault).To(Equal(5 * time.Second))
	})

	It("UnstickTasksDelayDefault is expected", func() {
		Expect(taskQueue.UnstickTasksDelayDefault).To(Equal(5 * time.Minute))
	})

	It("UnstickTasksAvailableGracePeriodDefault is expected", func() {
		Expect(taskQueue.UnstickTasksAvailableGracePeriodDefault).To(Equal(1 * time.Minute))
	})

	It("StopWaitTimeoutDefault is expected", func() {
		Expect(taskQueue.StopWaitTimeoutDefault).To(Equal(10 * time.Second))
	})

	It("TaskDeadlineDefault is expected", func() {
		Expect(taskQueue.TaskDeadlineDefault).To(Equal(1 * time.Minute))
	})

	It("DurationJitterFactor is expected", func() {
		Expect(taskQueue.DurationJitterFactor).To(Equal(0.2))
	})

	Context("Config", func() {
		Context("NewConfig", func() {
			It("returns successfully", func() {
				cfg := taskQueue.NewConfig()
				Expect(cfg).ToNot(BeNil())
			})

			It("returns default values", func() {
				cfg := taskQueue.NewConfig()
				Expect(cfg).ToNot(BeNil())
				Expect(cfg.Workers).To(Equal(taskQueue.WorkersDefault))
				Expect(cfg.StartManagerDelay).To(Equal(taskQueue.StartManagerDelayDefault))
				Expect(cfg.DispatchTasksDelay).To(Equal(taskQueue.DispatchTasksDelayDefault))
				Expect(cfg.MonitorTaskDelay).To(Equal(taskQueue.MonitorTaskDelayDefault))
				Expect(cfg.UnstickTasksDelay).To(Equal(taskQueue.UnstickTasksDelayDefault))
				Expect(cfg.UnstickTasksAvailableGracePeriod).To(Equal(taskQueue.UnstickTasksAvailableGracePeriodDefault))
				Expect(cfg.RunnerWatchdogGracePeriod).To(Equal(taskQueue.RunnerWatchdogGracePeriodDefault))
				Expect(cfg.StopWaitTimeout).To(Equal(taskQueue.StopWaitTimeoutDefault))
			})
		})

		Context("with new config", func() {
			var cfg *taskQueue.Config

			BeforeEach(func() {
				cfg = taskQueue.NewConfig()
			})

			Context("Load", func() {
				var configReporter *configTest.Reporter

				BeforeEach(func() {
					configReporter = configTest.NewReporter()
				})

				It("returns an error when the config reporter is missing", func() {
					Expect(cfg.Load(nil)).To(MatchError("config reporter is missing"))
				})

				It("returns an error when workers is not parsable", func() {
					configReporter.Config["workers"] = test.RandomStringFromCharset(test.CharsetAlpha)
					Expect(cfg.Load(configReporter)).To(MatchError("workers is invalid"))
				})

				It("returns an error when start manager delay is not parsable", func() {
					configReporter.Config["start_manager_delay"] = test.RandomStringFromCharset(test.CharsetAlpha)
					Expect(cfg.Load(configReporter)).To(MatchError("start manager delay is invalid"))
				})

				It("returns an error when dispatch tasks delay is not parsable", func() {
					configReporter.Config["dispatch_tasks_delay"] = test.RandomStringFromCharset(test.CharsetAlpha)
					Expect(cfg.Load(configReporter)).To(MatchError("dispatch tasks delay is invalid"))
				})

				It("returns an error when monitor task delay is not parsable", func() {
					configReporter.Config["monitor_task_delay"] = test.RandomStringFromCharset(test.CharsetAlpha)
					Expect(cfg.Load(configReporter)).To(MatchError("monitor task delay is invalid"))
				})

				It("returns an error when runner watchdog grace period is not parsable", func() {
					configReporter.Config["runner_watchdog_grace_period"] = test.RandomStringFromCharset(test.CharsetAlpha)
					Expect(cfg.Load(configReporter)).To(MatchError("runner watchdog grace period is invalid"))
				})

				It("returns an error when unstick tasks delay is not parsable", func() {
					configReporter.Config["unstick_tasks_delay"] = test.RandomStringFromCharset(test.CharsetAlpha)
					Expect(cfg.Load(configReporter)).To(MatchError("unstick tasks delay is invalid"))
				})

				It("returns an error when unstick tasks available grace period is not parsable", func() {
					configReporter.Config["unstick_tasks_available_grace_period"] = test.RandomStringFromCharset(test.CharsetAlpha)
					Expect(cfg.Load(configReporter)).To(MatchError("unstick tasks available grace period is invalid"))
				})

				It("returns an error when stop wait timeout is not parsable", func() {
					configReporter.Config["stop_wait_timeout"] = test.RandomStringFromCharset(test.CharsetAlpha)
					Expect(cfg.Load(configReporter)).To(MatchError("stop wait timeout is invalid"))
				})

				It("uses existing workers if not set", func() {
					Expect(cfg.Load(configReporter)).To(Succeed())
					Expect(cfg.Workers).To(Equal(taskQueue.WorkersDefault))
				})

				It("uses existing start manager delay if not set", func() {
					Expect(cfg.Load(configReporter)).To(Succeed())
					Expect(cfg.StartManagerDelay).To(Equal(taskQueue.StartManagerDelayDefault))
				})

				It("uses existing dispatch tasks delay if not set", func() {
					Expect(cfg.Load(configReporter)).To(Succeed())
					Expect(cfg.DispatchTasksDelay).To(Equal(taskQueue.DispatchTasksDelayDefault))
				})

				It("uses existing monitor task delay if not set", func() {
					Expect(cfg.Load(configReporter)).To(Succeed())
					Expect(cfg.MonitorTaskDelay).To(Equal(taskQueue.MonitorTaskDelayDefault))
				})

				It("uses existing runner watchdog grace period if not set", func() {
					Expect(cfg.Load(configReporter)).To(Succeed())
					Expect(cfg.RunnerWatchdogGracePeriod).To(Equal(taskQueue.RunnerWatchdogGracePeriodDefault))
				})

				It("uses existing unstick tasks delay if not set", func() {
					Expect(cfg.Load(configReporter)).To(Succeed())
					Expect(cfg.UnstickTasksDelay).To(Equal(taskQueue.UnstickTasksDelayDefault))
				})

				It("uses existing unstick tasks available grace period if not set", func() {
					Expect(cfg.Load(configReporter)).To(Succeed())
					Expect(cfg.UnstickTasksAvailableGracePeriod).To(Equal(taskQueue.UnstickTasksAvailableGracePeriodDefault))
				})

				It("uses existing stop wait timeout if not set", func() {
					Expect(cfg.Load(configReporter)).To(Succeed())
					Expect(cfg.StopWaitTimeout).To(Equal(taskQueue.StopWaitTimeoutDefault))
				})

				It("returns successfully and uses values from the config reporter", func() {
					configReporter.Config["workers"] = "5"
					configReporter.Config["start_manager_delay"] = "45"
					configReporter.Config["dispatch_tasks_delay"] = "30"
					configReporter.Config["monitor_task_delay"] = "75"
					configReporter.Config["runner_watchdog_grace_period"] = "20"
					configReporter.Config["unstick_tasks_delay"] = "60"
					configReporter.Config["unstick_tasks_available_grace_period"] = "10"
					configReporter.Config["stop_wait_timeout"] = "15"
					Expect(cfg.Load(configReporter)).To(Succeed())
					Expect(cfg.Workers).To(Equal(5))
					Expect(cfg.StartManagerDelay).To(Equal(45 * time.Second))
					Expect(cfg.DispatchTasksDelay).To(Equal(30 * time.Second))
					Expect(cfg.MonitorTaskDelay).To(Equal(75 * time.Second))
					Expect(cfg.RunnerWatchdogGracePeriod).To(Equal(20 * time.Second))
					Expect(cfg.UnstickTasksDelay).To(Equal(60 * time.Second))
					Expect(cfg.UnstickTasksAvailableGracePeriod).To(Equal(10 * time.Second))
					Expect(cfg.StopWaitTimeout).To(Equal(15 * time.Second))
				})
			})

			Context("Validate", func() {
				It("returns an error when workers is less than 1", func() {
					cfg.Workers = 0
					Expect(cfg.Validate()).To(MatchError("workers is invalid"))
				})

				It("returns an error when start manager delay is invalid", func() {
					cfg.StartManagerDelay = 0
					Expect(cfg.Validate()).To(MatchError("start manager delay is invalid"))
				})

				It("returns an error when dispatch tasks delay is invalid", func() {
					cfg.DispatchTasksDelay = 0
					Expect(cfg.Validate()).To(MatchError("dispatch tasks delay is invalid"))
				})

				It("returns an error when monitor task delay is invalid", func() {
					cfg.MonitorTaskDelay = 0
					Expect(cfg.Validate()).To(MatchError("monitor task delay is invalid"))
				})

				It("returns an error when runner watchdog grace period is invalid", func() {
					cfg.RunnerWatchdogGracePeriod = 0
					Expect(cfg.Validate()).To(MatchError("runner watchdog grace period is invalid"))
				})

				It("returns an error when unstick tasks delay is invalid", func() {
					cfg.UnstickTasksDelay = 0
					Expect(cfg.Validate()).To(MatchError("unstick tasks delay is invalid"))
				})

				It("returns an error when unstick tasks available grace period is invalid", func() {
					cfg.UnstickTasksAvailableGracePeriod = 0
					Expect(cfg.Validate()).To(MatchError("unstick tasks available grace period is invalid"))
				})

				It("returns an error when stop wait timeout is invalid", func() {
					cfg.StopWaitTimeout = 0
					Expect(cfg.Validate()).To(MatchError("stop wait timeout is invalid"))
				})

				It("returns successfully", func() {
					Expect(cfg.Validate()).To(Succeed())
				})
			})
		})
	})

	Context("New", func() {
		var cfg *taskQueue.Config
		var lgr *logTest.Logger
		var str *taskStoreTest.Store

		BeforeEach(func() {
			cfg = taskQueue.NewConfig()
			lgr = logTest.NewLogger()
			str = taskStoreTest.NewStore()
		})

		It("returns an error when name is missing", func() {
			que, err := taskQueue.New("", cfg, lgr, str)
			Expect(err).To(MatchError("name is missing"))
			Expect(que).To(BeNil())
		})

		It("returns an error when config is missing", func() {
			que, err := taskQueue.New(taskTest.RandomType(), nil, lgr, str)
			Expect(err).To(MatchError("config is missing"))
			Expect(que).To(BeNil())
		})

		It("returns an error when logger is missing", func() {
			que, err := taskQueue.New(taskTest.RandomType(), cfg, nil, str)
			Expect(err).To(MatchError("logger is missing"))
			Expect(que).To(BeNil())
		})

		It("returns an error when store is missing", func() {
			que, err := taskQueue.New(taskTest.RandomType(), cfg, lgr, nil)
			Expect(err).To(MatchError("store is missing"))
			Expect(que).To(BeNil())
		})

		It("returns an error when config is invalid", func() {
			cfg.Workers = 0
			que, err := taskQueue.New(taskTest.RandomType(), cfg, lgr, str)
			Expect(err).To(MatchError("config is invalid; workers is invalid"))
			Expect(que).To(BeNil())
		})

		It("returns an error when a runner is missing", func() {
			que, err := taskQueue.New(taskTest.RandomType(), cfg, lgr, str, nil)
			Expect(err).To(MatchError("runner is missing"))
			Expect(que).To(BeNil())
		})

		It("returns an error when multiple runners have the same type", func() {
			typ := taskTest.RandomType()
			que, err := taskQueue.New(taskTest.RandomType(), cfg, lgr, str, taskQueueTest.NewCountingRunner(typ), taskQueueTest.NewCountingRunner(typ))
			Expect(err).To(MatchError("runner type already registered"))
			Expect(que).To(BeNil())
		})

		It("returns an error when a runner duration maximum is not positive", func() {
			runner := taskQueueTest.NewStubRunner(taskTest.RandomType()).
				WithDeadline(3 * time.Minute).
				WithTimeout(2 * time.Minute).
				WithDurationMaximum(0)
			que, err := taskQueue.New(taskTest.RandomType(), cfg, lgr, str, runner)
			Expect(err).To(MatchError("runner duration maximum is invalid"))
			Expect(que).To(BeNil())
		})

		It("returns an error when a runner timeout does not exceed its duration maximum", func() {
			runner := taskQueueTest.NewStubRunner(taskTest.RandomType()).
				WithDeadline(3 * time.Minute).
				WithTimeout(time.Minute).
				WithDurationMaximum(time.Minute)
			que, err := taskQueue.New(taskTest.RandomType(), cfg, lgr, str, runner)
			Expect(err).To(MatchError("runner timeout is invalid"))
			Expect(que).To(BeNil())
		})

		It("returns an error when a runner deadline does not exceed its timeout", func() {
			runner := taskQueueTest.NewStubRunner(taskTest.RandomType()).
				WithDeadline(2 * time.Minute).
				WithTimeout(2 * time.Minute).
				WithDurationMaximum(time.Minute)
			que, err := taskQueue.New(taskTest.RandomType(), cfg, lgr, str, runner)
			Expect(err).To(MatchError("runner deadline is invalid"))
			Expect(que).To(BeNil())
		})

		It("returns successfully", func() {
			str.NewTaskRepositoryOutputs = []taskStore.TaskRepository{nil}
			que := test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str))
			Expect(que).ToNot(BeNil())
		})

		It("returns successfully with multiple runners of different types", func() {
			str.NewTaskRepositoryOutputs = []taskStore.TaskRepository{nil}
			que := test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, taskQueueTest.NewCountingRunner(taskTest.RandomType()), taskQueueTest.NewCountingRunner(taskTest.RandomType())))
			Expect(que).ToNot(BeNil())
		})
	})

	Context("with a new queue", func() {
		var que *taskQueue.Queue

		BeforeEach(func() {
			cfg := taskQueue.NewConfig()
			lgr := logTest.NewLogger()
			str := taskStoreTest.NewStore()
			str.NewTaskRepositoryOutputs = []taskStore.TaskRepository{nil}
			que = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, taskQueueTest.NewCountingRunner(taskTest.RandomType())))
		})

		Context("Start", func() {
			It("does nothing when called after Stop, since the queue is single-use", func() {
				que.Stop()
				Expect(func() { que.Start() }).ToNot(Panic())
				// A no-op Start launches no goroutines, so a subsequent Stop stays a no-op too.
				Expect(func() { que.Stop() }).ToNot(Panic())
			})

			It("is safe to call more than once", func() {
				que.Start()
				Expect(func() { que.Start() }).ToNot(Panic())
				que.Stop()
			})
		})

		Context("Stop", func() {
			It("does not panic when called without a prior call to Start", func() {
				Expect(func() { que.Stop() }).ToNot(Panic())
			})

			It("succeeds if called twice", func() {
				Expect(func() { que.Stop() }).ToNot(Panic())
				Expect(func() { que.Stop() }).ToNot(Panic())
			})
		})
	})

	Context("with a store", func() {
		var lgr *logTest.Logger
		var ctx context.Context
		var str *taskStoreMongo.Store

		BeforeEach(func() {
			lgr = logTest.NewLogger()
			ctx = log.NewContextWithLogger(context.Background(), lgr)

			cfg := storeStructuredMongoTest.NewConfig()
			str = test.Must(taskStoreMongo.NewStore(cfg))
			_ = test.Must(str.GetRepository("tasks").DeleteMany(ctx, bson.M{}))
		})

		AfterEach(func() {
			_ = test.Must(str.GetRepository("tasks").DeleteMany(ctx, bson.M{}))
			Expect(str.Terminate(ctx)).To(Succeed())
		})

		Context("with successful shutdown", func() {
			var cfg *taskQueue.Config
			var que *taskQueue.Queue

			BeforeEach(func() {
				cfg = taskQueue.NewConfig()
				cfg.Workers = 2
				cfg.StartManagerDelay = time.Millisecond
				cfg.DispatchTasksDelay = time.Millisecond
			})

			AfterEach(func() {
				if que != nil {
					que.Stop()
				}
				lgr.AssertDebug("Task queue worker stopped", log.Fields{"error": errors.NewSerializable(context.Canceled)})
				lgr.AssertDebug("Task queue manager stopped", log.Fields{"error": errors.NewSerializable(context.Canceled)})
			})

			It("fails a pending task that does not match any registered runner", func() {
				runner := taskQueueTest.NewStubRunner(taskTest.RandomType())
				que = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, runner))

				unregisteredType := taskTest.RandomType()
				createdTask := test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: unregisteredType}))

				que.Start()

				Eventually(func() string {
					return test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil)).State
				}, "5s", "50ms").To(Equal(task.TaskStateFailed))

				actualTask := test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil))
				Expect(actualTask.State).To(Equal(task.TaskStateFailed))
				Expect(actualTask.Error).ToNot(BeNil())
				Expect(actualTask.Error.Error).To(MatchError("runner not found for task type"))

				lgr.AssertError("Runner not found for task type; task cannot be processed")
				Expect(testutil.ToFloat64(taskQueue.RunnerNotFoundTotal.WithLabelValues(unregisteredType))).To(Equal(float64(1)))
			})

			It("dispatches, runs, and completes a pending task matching a registered runner", func() {
				runner := taskQueueTest.NewCountingRunner(taskTest.RandomType())
				que = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, runner))

				createdTask := test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: runner.GetRunnerType()}))

				que.Start()

				Eventually(func() string {
					return test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil)).State
				}, "5s", "50ms").To(Equal(task.TaskStateCompleted))

				Expect(runner.GetCount()).To(Equal(1))
			})

			It("completes a task that the runner updated while it was running", func() {
				updatedData := metadataTest.RandomMetadataMap()

				runner := taskQueueTest.NewStubRunner(taskTest.RandomType()).
					WithStub(func(runnerContext context.Context, runnerTask *task.Task) {
						*runnerTask = *test.Must(str.NewTaskRepository().UpdateTask(runnerContext, runnerTask.ID, nil, &task.TaskUpdate{Data: &updatedData}))
						runnerTask.SetCompleted()
					})
				que = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, runner))

				createdTask := test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: runner.GetRunnerType()}))

				que.Start()

				Eventually(func() string {
					return test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil)).State
				}, "5s", "50ms").To(Equal(task.TaskStateCompleted))

				actualTask := test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil))
				Expect(actualTask.State).To(Equal(task.TaskStateCompleted))
				Expect(actualTask.Error).To(BeNil())
				Expect(actualTask.Data).To(Equal(updatedData))
			})

			It("logs a warning if the runner update the task while it was running, but did not use the updated task", func() {
				updatedData := metadataTest.RandomMetadataMap()

				runner := taskQueueTest.NewStubRunner(taskTest.RandomType()).
					WithStub(func(runnerContext context.Context, runnerTask *task.Task) {
						test.Must(str.NewTaskRepository().UpdateTask(runnerContext, runnerTask.ID, nil, &task.TaskUpdate{Data: &updatedData}))
						runnerTask.SetCompleted()
					})
				que = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, runner))

				createdTask := test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: runner.GetRunnerType()}))

				que.Start()

				Eventually(func() string {
					return test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil)).State
				}, "5s", "50ms").To(Equal(task.TaskStateCompleted))

				actualTask := test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil))
				Expect(actualTask.State).To(Equal(task.TaskStateCompleted))
				Expect(actualTask.Error).To(BeNil())
				Expect(actualTask.Data).To(BeNil())

				lgr.AssertWarn("Database task revision does not match running task revision; Runner contract broken or concurrent update")
			})

			It("does not complete a task whose claim token changed while it was running", func() {
				runner := taskQueueTest.NewStubRunner(taskTest.RandomType()).
					WithStub(func(runnerContext context.Context, runnerTask *task.Task) {
						// Simulate the task being unstuck and re-claimed elsewhere by changing the claim token out from
						// under this run. The completion must then miss rather than falsely complete another run task.
						test.Must(str.GetCollection("tasks").UpdateOne(ctx, bson.M{"id": runnerTask.ID}, bson.M{"$set": bson.M{"claimToken": ""}}))
						runnerTask.SetCompleted()
					})
				que = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, runner))

				createdTask := test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: runner.GetRunnerType()}))

				que.Start()

				Eventually(func() bool {
					defer func() { _ = recover() }()
					lgr.AssertError("Unable to stop task; no running task matched the id and claim token")
					return true
				}, "5s", "100ms").To(BeTrue())

				actualTask := test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil))
				Expect(actualTask.State).To(Equal(task.TaskStateRunning))
			})

			It("does not complete a task was deleted while it was running", func() {
				runner := taskQueueTest.NewStubRunner(taskTest.RandomType()).
					WithStub(func(runnerContext context.Context, runnerTask *task.Task) {
						test.Must(str.GetCollection("tasks").DeleteOne(ctx, bson.M{"id": runnerTask.ID}))
						runnerTask.SetCompleted()
					})
				que = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, runner))

				createdTask := test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: runner.GetRunnerType()}))

				que.Start()

				Eventually(func() bool {
					defer func() { _ = recover() }()
					lgr.AssertError("Unable to stop task; no running task matched the id and claim token")
					return true
				}, "5s", "100ms").To(BeTrue())

				actualTask := test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil))
				Expect(actualTask).To(BeNil())
			})

			It("cancels a running task when the task is deleted mid-run", func() {
				cfg.MonitorTaskDelay = time.Millisecond
				taskQueue.RunClaimLostTotal.Reset()

				canceled := make(chan error, 1)
				runner := taskQueueTest.NewStubRunner(taskTest.RandomType()).
					WithStub(func(runnerContext context.Context, runnerTask *task.Task) {
						test.Must(str.GetCollection("tasks").DeleteOne(ctx, bson.M{"id": runnerTask.ID}))
						<-runnerContext.Done()
						canceled <- context.Cause(runnerContext)
						runnerTask.SetCompleted()
					})
				que = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, runner))

				createdTask := test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: runner.GetRunnerType()}))

				que.Start()

				Eventually(canceled, "5s").Should(Receive(MatchError("task claim lost")))

				// The outcome is discarded rather than written back as a lost completion.
				Eventually(func() bool {
					defer func() { _ = recover() }()
					lgr.AssertWarn("Task claim lost; task run outcome discarded")
					return true
				}, "5s", "100ms").To(BeTrue())

				Expect(test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil))).To(BeNil())
				Expect(testutil.ToFloat64(taskQueue.RunClaimLostTotal.WithLabelValues(runner.GetRunnerType(), "deleted"))).To(Equal(float64(1)))
			})

			It("cancels a running task deleted after earlier claim checks found the claim intact", func() {
				cfg.MonitorTaskDelay = 50 * time.Millisecond
				taskQueue.RunClaimLostTotal.Reset()

				canceled := make(chan error, 1)
				runner := taskQueueTest.NewStubRunner(taskTest.RandomType()).
					WithStub(func(runnerContext context.Context, runnerTask *task.Task) {
						// Run past several claim checks before losing the claim, so the monitor must keep checking
						// rather than settle after the first.
						time.Sleep(5 * cfg.MonitorTaskDelay)
						test.Must(str.GetCollection("tasks").DeleteOne(ctx, bson.M{"id": runnerTask.ID}))
						<-runnerContext.Done()
						canceled <- context.Cause(runnerContext)
						runnerTask.SetCompleted()
					})
				que = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, runner))

				createdTask := test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: runner.GetRunnerType()}))

				que.Start()

				Eventually(canceled, "5s").Should(Receive(MatchError("task claim lost")))

				Eventually(func() bool {
					defer func() { _ = recover() }()
					lgr.AssertWarn("Task claim lost; task run outcome discarded")
					return true
				}, "5s", "100ms").To(BeTrue())

				Expect(test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil))).To(BeNil())
				Expect(testutil.ToFloat64(taskQueue.RunClaimLostTotal.WithLabelValues(runner.GetRunnerType(), "deleted"))).To(Equal(float64(1)))
			})

			It("cancels a running task when the task is re-claimed mid-run", func() {
				cfg.MonitorTaskDelay = 100 * time.Millisecond
				taskQueue.RunClaimLostTotal.Reset()

				canceled := make(chan error, 1)
				runner := taskQueueTest.NewStubRunner(taskTest.RandomType()).
					WithStub(func(runnerContext context.Context, runnerTask *task.Task) {
						// Simulate the task being unstuck and re-claimed elsewhere by changing the claim token out from
						// under this run.
						test.Must(str.GetCollection("tasks").UpdateOne(ctx, bson.M{"id": runnerTask.ID}, bson.M{"$set": bson.M{"claimToken": "other-claim-token"}}))
						<-runnerContext.Done()
						canceled <- context.Cause(runnerContext)
						runnerTask.SetCompleted()
					})
				que = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, runner))

				createdTask := test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: runner.GetRunnerType()}))

				que.Start()

				Eventually(canceled, "5s").Should(Receive(MatchError("task claim lost")))

				Eventually(func() bool {
					defer func() { _ = recover() }()
					lgr.AssertWarn("Task claim lost; task run outcome discarded")
					return true
				}, "5s", "100ms").To(BeTrue())

				// The foreign claim remains untouched; this run wrote nothing back.
				actualTask := test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil))
				Expect(actualTask.State).To(Equal(task.TaskStateRunning))
				Expect(testutil.ToFloat64(taskQueue.RunClaimLostTotal.WithLabelValues(runner.GetRunnerType(), "reclaimed"))).To(Equal(float64(1)))
			})

			It("does not cancel a running task whose claim is intact", func() {
				cfg.MonitorTaskDelay = time.Millisecond

				runner := taskQueueTest.NewStubRunner(taskTest.RandomType()).
					WithStub(func(runnerContext context.Context, runnerTask *task.Task) {
						// Linger across several claim checks; the claim is intact, so the run must not be canceled.
						select {
						case <-runnerContext.Done():
							runnerTask.AppendError(context.Cause(runnerContext))
							runnerTask.SetFailed()
						case <-time.After(250 * time.Millisecond):
							runnerTask.SetCompleted()
						}
					})
				que = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, runner))

				createdTask := test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: runner.GetRunnerType()}))

				que.Start()

				Eventually(func() string {
					return test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil)).State
				}, "5s", "50ms").To(Equal(task.TaskStateCompleted))

				actualTask := test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil))
				Expect(actualTask.Error).To(BeNil())
			})

			It("cleans up a task that panics during execution", func() {
				runner := taskQueueTest.NewStubRunner(taskTest.RandomType()).
					WithStub(func(runnerContext context.Context, runnerTask *task.Task) { panic("panic test") })
				que = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, runner))

				createdTask := test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: runner.GetRunnerType()}))

				que.Start()

				Eventually(func() string {
					return test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil)).State
				}, "5s", "50ms").To(Equal(task.TaskStateFailed))

				actualTask := test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil))
				Expect(actualTask.State).To(Equal(task.TaskStateFailed))
				Expect(actualTask.Error).ToNot(BeNil())
				Expect(actualTask.Error.Error).To(MatchError("unhandled panic"))

				lgr.AssertError("Unhandled panic while running task")
				Expect(testutil.ToFloat64(taskQueue.RunPanicTotal.WithLabelValues(runner.GetRunnerType()))).To(Equal(float64(1)))
			})

			It("fails a task whose runner leaves it in an unknown state", func() {
				runner := taskQueueTest.NewStubRunner(taskTest.RandomType()).
					WithStub(func(runnerContext context.Context, runnerTask *task.Task) { runnerTask.State = "unknown-state" })
				que = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, runner))

				createdTask := test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: runner.GetRunnerType()}))

				que.Start()

				Eventually(func() string {
					return test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil)).State
				}, "5s", "50ms").To(Equal(task.TaskStateFailed))

				actualTask := test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil))
				Expect(actualTask.State).To(Equal(task.TaskStateFailed))
				Expect(actualTask.Error).ToNot(BeNil())
				Expect(actualTask.Error.Error).To(MatchError("unknown task state"))
			})

			It("warns and sets the available time for a pending task left without one", func() {
				runner := taskQueueTest.NewStubRunner(taskTest.RandomType()).
					WithStub(func(runnerContext context.Context, runnerTask *task.Task) {
						runnerTask.State = task.TaskStatePending
						runnerTask.AvailableTime = nil
					})
				que = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, runner))

				createdTask := test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: runner.GetRunnerType()}))

				que.Start()

				Eventually(func() string {
					defer func() { _ = recover() }()
					lgr.AssertWarn("Available time missing for pending task")
					return test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil)).State
				}, "5s", "50ms").To(Equal(task.TaskStatePending))

				que.Stop()

				actualTask := test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil))
				Expect(actualTask.State).To(Equal(task.TaskStatePending))
			})

			It("warns when a pending task available time is significantly in the past", func() {
				runner := taskQueueTest.NewStubRunner(taskTest.RandomType()).
					WithStub(func(runnerContext context.Context, runnerTask *task.Task) {
						runnerTask.RepeatAvailableAt(time.Now().Add(-2 * time.Minute))
					})
				que = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, runner))

				createdTask := test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: runner.GetRunnerType()}))

				que.Start()

				Eventually(func() string {
					defer func() { _ = recover() }()
					lgr.AssertWarn("Available time significantly before now for pending task")
					return test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil)).State
				}, "5s", "50ms").To(Equal(task.TaskStatePending))

				que.Stop()

				actualTask := test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil))
				Expect(actualTask.State).To(Equal(task.TaskStatePending))
			})

			It("unsticks and logs a task left running past its deadline", func() {
				cfg.UnstickTasksDelay = time.Millisecond
				cfg.UnstickTasksAvailableGracePeriod = 50 * time.Millisecond

				// Keep the unstuck task's availability delay short so it is re-dispatched within the wait below.
				cfg.MonitorTaskDelay = time.Millisecond
				cfg.RunnerWatchdogGracePeriod = 50 * time.Millisecond

				runner := taskQueueTest.NewCountingRunner(taskTest.RandomType())
				que = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, runner))

				createdTask := &task.Task{
					ID:           task.NewID(),
					Type:         runner.GetRunnerType(),
					State:        task.TaskStateRunning,
					CreatedTime:  time.Now(),
					Revision:     1,
					ClaimToken:   pointer.FromString(taskTest.RandomType()),
					DeadlineTime: pointer.FromTime(time.Now().Add(-time.Minute)),
				}
				test.Must(str.GetCollection("tasks").InsertOne(ctx, createdTask))

				que.Start()

				Eventually(func() bool {
					defer func() { _ = recover() }()
					lgr.AssertInfo("Unstuck tasks")
					return true
				}, "5s", "50ms").To(BeTrue())

				// Once unstuck, the task returns to pending and is dispatched, run, and completed.
				Eventually(func() string {
					return test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil)).State
				}, "5s", "50ms").To(Equal(task.TaskStateCompleted))
			})

			It("reverts a task that is still running to pending when the queue is stopped", func() {
				runner := taskQueueTest.NewStubRunner(taskTest.RandomType()).
					WithStub(func(runnerContext context.Context, runnerTask *task.Task) { <-runnerContext.Done() })
				que = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, runner))

				createdTask := test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: runner.GetRunnerType()}))

				que.Start()

				// Wait until the runner has picked up the task and it is running.
				Eventually(func() string {
					return test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil)).State
				}, "5s", "50ms").To(Equal(task.TaskStateRunning))

				// Stopping cancels the worker context; the hanging runner returns leaving the task running, so the
				// queue must revert it to pending for a later retry.
				que.Stop()

				actualTask := test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil))
				Expect(actualTask.State).To(Equal(task.TaskStatePending))
				Expect(actualTask.DeadlineTime).To(BeNil())
			})

			It("cancels the context of a task that exceeds its timeout", func() {
				runner := taskQueueTest.NewSleepRunner(taskTest.RandomType(), time.Minute, 100*time.Millisecond, 50*time.Millisecond, 10*time.Second)
				que = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, runner))

				createdTask := test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: runner.GetRunnerType()}))

				que.Start()

				Eventually(func() string {
					return test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil)).State
				}, "1m", "50ms").To(Equal(task.TaskStateCompleted))

				actualTask := test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil))
				Expect(actualTask.State).To(Equal(task.TaskStateCompleted))
				Expect(actualTask.Error).ToNot(BeNil())
				Expect(actualTask.Error.Error).To(MatchError("task runner timeout exceeded"))
			})

			It("fails a task whose runner returns without setting a terminal state", func() {
				runner := taskQueueTest.NewStubRunner(taskTest.RandomType())
				que = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, runner))

				createdTask := test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: runner.GetRunnerType()}))

				que.Start()

				Eventually(func() string {
					return test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil)).State
				}, "5s", "50ms").To(Equal(task.TaskStateFailed))

				actualTask := test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil))
				Expect(actualTask.State).To(Equal(task.TaskStateFailed))
				Expect(actualTask.Error).ToNot(BeNil())
				Expect(actualTask.Error.Error).To(MatchError("runner failed to set state"))
			})

			It("fails a task that exceeds its timeout without setting its own state", func() {
				runner := taskQueueTest.NewStubRunner(taskTest.RandomType()).
					WithDurationMaximum(100 * time.Millisecond).
					WithStub(func(runnerContext context.Context, runnerTask *task.Task) { <-runnerContext.Done() })
				que = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, runner))

				createdTask := test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: runner.GetRunnerType()}))

				que.Start()

				Eventually(func() string {
					return test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil)).State
				}, "5s", "50ms").To(Equal(task.TaskStateFailed))

				actualTask := test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil))
				Expect(actualTask.State).To(Equal(task.TaskStateFailed))
				Expect(actualTask.Error).ToNot(BeNil())
				Expect(actualTask.Error.Error).To(MatchError("task runner timeout exceeded"))

				lgr.AssertWarn("Task runner exceeded timeout; task will be failed")
				Expect(testutil.ToFloat64(taskQueue.RunnerTimeoutExceededTotal.WithLabelValues(runner.GetRunnerType(), "recovered"))).To(Equal(float64(1)))
			})

			It("logs a warning if a task that exceeds its maximum duration", func() {
				runner := taskQueueTest.NewSleepRunner(taskTest.RandomType(), 2*time.Minute, time.Minute, time.Millisecond, 10*time.Millisecond)
				que = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, runner))

				createdTask := test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: runner.GetRunnerType()}))

				que.Start()

				Eventually(func() string {
					return test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil)).State
				}, "1m", "50ms").To(Equal(task.TaskStateCompleted))

				actualTask := test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil))
				Expect(actualTask.State).To(Equal(task.TaskStateCompleted))
				Expect(actualTask.Error).To(BeNil())

				lgr.AssertWarn("Task duration exceeds maximum")
			})

			It("does not race or panic when Stop is called concurrently", func() {
				runner := taskQueueTest.NewCountingRunner(taskTest.RandomType())
				que = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, runner))

				createdTask := test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: runner.GetRunnerType()}))

				que.Start()

				Eventually(func() string {
					return test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil)).State
				}, "5s", "50ms").To(Equal(task.TaskStateCompleted))

				var waitGroup sync.WaitGroup
				for range 10 {
					waitGroup.Go(func() {
						defer GinkgoRecover()
						que.Stop()
					})
				}
				waitGroup.Wait()
			})
		})

		Context("without successful shutdown", func() {
			var cfg *taskQueue.Config
			var que *taskQueue.Queue

			BeforeEach(func() {
				cfg = taskQueue.NewConfig()
				cfg.Workers = 2
				cfg.StartManagerDelay = time.Millisecond
				cfg.DispatchTasksDelay = time.Millisecond
				cfg.StopWaitTimeout = 250 * time.Millisecond
			})

			AfterEach(func() {
				if que != nil {
					que.Stop()
				}
			})

			It("returns from Stop within the stop timeout when a runner ignores cancellation", func() {
				runner := taskQueueTest.NewStubRunner(taskTest.RandomType()).
					WithDurationMaximum(time.Minute).
					WithStub(func(runnerContext context.Context, runnerTask *task.Task) { select {} })
				que = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, runner))

				createdTask := test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: runner.GetRunnerType()}))

				que.Start()

				// Wait until the blocking runner has picked up the task and it is running.
				Eventually(func() string {
					return test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil)).State
				}, "5s", "50ms").To(Equal(task.TaskStateRunning))

				// Stop must return within roughly the stop timeout even though the runner never returns; it abandons
				// the in-flight task rather than blocking forever.
				stopped := make(chan struct{})
				go func() {
					defer GinkgoRecover()
					que.Stop()
					close(stopped)
				}()
				Eventually(stopped, "5s").Should(BeClosed())

				lgr.AssertError("Task queue workers did not stop within timeout; abandoning in-flight tasks; will be fixed with UnstickTasks later")
			})

			It("logs and counts the run once its timeout elapses", func() {
				cfg.RunnerWatchdogGracePeriod = 50 * time.Millisecond

				// A short duration maximum yields a short runner timeout (2x), and a short grace period keeps the
				// watchdog prompt, so the watchdog fires quickly while the runner is still blocked, without an unstick
				// reclaiming the task.
				runner := taskQueueTest.NewStubRunner(taskTest.RandomType()).
					WithDurationMaximum(20 * time.Millisecond).
					WithStub(func(runnerContext context.Context, runnerTask *task.Task) { select {} })
				que = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, runner))

				runnerType := runner.GetRunnerType()
				taskQueue.RunnerTimeoutExceededTotal.Reset()

				createdTask := test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: runnerType}))

				que.Start()

				// Wait until the runner has picked up the task and it is running.
				Eventually(func() string {
					return test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil)).State
				}, "5s", "50ms").To(Equal(task.TaskStateRunning))

				// The watchdog fires after the runner timeout, recording the stuck run as blocked.
				Eventually(func() float64 {
					return testutil.ToFloat64(taskQueue.RunnerTimeoutExceededTotal.WithLabelValues(runnerType, "blocked"))
				}, "5s", "50ms").Should(BeNumerically(">=", float64(1)))
				lgr.AssertError("Task runner exceeded timeout without returning; worker is blocked until it returns")

				// The runner never returns, so Stop abandons the in-flight worker.
				stopped := make(chan struct{})
				go func() {
					defer GinkgoRecover()
					que.Stop()
					close(stopped)
				}()
				Eventually(stopped, "5s").Should(BeClosed())
			})
		})

		Context("with multiple queues", func() {
			const queueCount = 10
			const workersCount = 10
			const taskCount = 2 * queueCount * workersCount

			var runner *taskQueueTest.StubRunner
			var ques []*taskQueue.Queue

			BeforeEach(func() {
				runner = taskQueueTest.NewStubRunner(taskTest.RandomType()).
					WithStub(func(runnerContext context.Context, runnerTask *task.Task) { runnerTask.State = task.TaskStatePending })

				cfg := taskQueue.NewConfig()
				cfg.Workers = workersCount
				cfg.StartManagerDelay = time.Millisecond
				cfg.DispatchTasksDelay = time.Millisecond
				cfg.UnstickTasksDelay = time.Millisecond
				ques = make([]*taskQueue.Queue, queueCount)
				for index := range len(ques) {
					ques[index] = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, runner))
				}
			})

			AfterEach(func() {
				for _, que := range ques {
					que.Stop()
				}
			})

			It("completes all running tasks when stopped", func() {
				tasks := make(task.Tasks, taskCount)
				for index := range len(tasks) {
					tasks[index] = test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: runner.GetRunnerType()}))
				}

				for _, que := range ques {
					que.Start()
				}

				select {
				case <-ctx.Done():
				case <-time.After(2 * time.Second):
				}

				for _, que := range ques {
					que.Stop()
				}

				for _, tsk := range tasks {
					actualTask := test.Must(str.NewTaskRepository().GetTask(ctx, tsk.ID, nil))
					Expect(actualTask.State).To(Equal(task.TaskStatePending))
				}
			})

			It("eventually a queue attempts to run a task that is already running in another queue", func() {
				createdTask := test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: runner.GetRunnerType()}))

				for _, que := range ques {
					que.Start()
				}

				Eventually(func() bool {
					defer func() { _ = recover() }()
					lgr.AssertInfo("Task no longer available to start")
					return true
				}, "10s", "100ms").To(BeTrue())

				for _, que := range ques {
					que.Stop()
				}

				actualTask := test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil))
				Expect(actualTask.State).To(Equal(task.TaskStatePending))
			})
		})
	})
})
