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

	It("DelayDefault is expected", func() {
		Expect(taskQueue.DelayDefault).To(Equal(time.Minute))
	})

	It("DelayInitialDefault is expected", func() {
		Expect(taskQueue.DelayInitialDefault).To(Equal(time.Minute))
	})

	It("DelayUnstickDefault is expected", func() {
		Expect(taskQueue.DelayUnstickDefault).To(Equal(5 * time.Minute))
	})

	It("StopWaitTimeoutDefault is expected", func() {
		Expect(taskQueue.StopWaitTimeoutDefault).To(Equal(10 * time.Second))
	})

	It("DurationJitterFactor is expected", func() {
		Expect(taskQueue.DurationJitterFactor).To(Equal(0.2))
	})

	It("TaskDeadlineDefault is expected", func() {
		Expect(taskQueue.TaskDeadlineDefault).To(Equal(time.Minute))
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
				Expect(cfg.Delay).To(Equal(taskQueue.DelayDefault))
				Expect(cfg.DelayInitial).To(Equal(taskQueue.DelayInitialDefault))
				Expect(cfg.DelayUnstick).To(Equal(taskQueue.DelayUnstickDefault))
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

				It("returns an error when delay is not parsable", func() {
					configReporter.Config["delay"] = test.RandomStringFromCharset(test.CharsetAlpha)
					Expect(cfg.Load(configReporter)).To(MatchError("delay is invalid"))
				})

				It("returns an error when delay initial is not parsable", func() {
					configReporter.Config["delay_initial"] = test.RandomStringFromCharset(test.CharsetAlpha)
					Expect(cfg.Load(configReporter)).To(MatchError("delay initial is invalid"))
				})

				It("returns an error when delay unstick is not parsable", func() {
					configReporter.Config["delay_unstick"] = test.RandomStringFromCharset(test.CharsetAlpha)
					Expect(cfg.Load(configReporter)).To(MatchError("delay unstick is invalid"))
				})

				It("returns an error when stop wait timeout is not parsable", func() {
					configReporter.Config["stop_wait_timeout"] = test.RandomStringFromCharset(test.CharsetAlpha)
					Expect(cfg.Load(configReporter)).To(MatchError("stop wait timeout is invalid"))
				})

				It("uses existing workers if not set", func() {
					Expect(cfg.Load(configReporter)).To(Succeed())
					Expect(cfg.Workers).To(Equal(taskQueue.WorkersDefault))
				})

				It("uses existing delay if not set", func() {
					Expect(cfg.Load(configReporter)).To(Succeed())
					Expect(cfg.Delay).To(Equal(taskQueue.DelayDefault))
				})

				It("uses existing delay initial if not set", func() {
					Expect(cfg.Load(configReporter)).To(Succeed())
					Expect(cfg.DelayInitial).To(Equal(taskQueue.DelayInitialDefault))
				})

				It("uses existing delay unstick if not set", func() {
					Expect(cfg.Load(configReporter)).To(Succeed())
					Expect(cfg.DelayUnstick).To(Equal(taskQueue.DelayUnstickDefault))
				})

				It("uses existing stop wait timeout if not set", func() {
					Expect(cfg.Load(configReporter)).To(Succeed())
					Expect(cfg.StopWaitTimeout).To(Equal(taskQueue.StopWaitTimeoutDefault))
				})

				It("returns successfully and uses values from the config reporter", func() {
					configReporter.Config["workers"] = "5"
					configReporter.Config["delay"] = "30"
					configReporter.Config["delay_initial"] = "45"
					configReporter.Config["delay_unstick"] = "60"
					configReporter.Config["stop_wait_timeout"] = "15"
					Expect(cfg.Load(configReporter)).To(Succeed())
					Expect(cfg.Workers).To(Equal(5))
					Expect(cfg.Delay).To(Equal(30 * time.Second))
					Expect(cfg.DelayInitial).To(Equal(45 * time.Second))
					Expect(cfg.DelayUnstick).To(Equal(60 * time.Second))
					Expect(cfg.StopWaitTimeout).To(Equal(15 * time.Second))
				})
			})

			Context("Validate", func() {
				It("returns an error when workers is less than 1", func() {
					cfg.Workers = 0
					Expect(cfg.Validate()).To(MatchError("workers is invalid"))
				})

				It("returns an error when delay is invalid", func() {
					cfg.Delay = 0
					Expect(cfg.Validate()).To(MatchError("delay is invalid"))
				})

				It("returns an error when delay initial is invalid", func() {
					cfg.DelayInitial = 0
					Expect(cfg.Validate()).To(MatchError("delay initial is invalid"))
				})

				It("returns an error when delay unstick is invalid", func() {
					cfg.DelayUnstick = 0
					Expect(cfg.Validate()).To(MatchError("delay unstick is invalid"))
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
			runner := taskQueueTest.NewSleepRunner(taskTest.RandomType(), 3*time.Minute, 2*time.Minute, 0, 0)
			que, err := taskQueue.New(taskTest.RandomType(), cfg, lgr, str, runner)
			Expect(err).To(MatchError("runner duration maximum is invalid"))
			Expect(que).To(BeNil())
		})

		It("returns an error when a runner timeout does not exceed its duration maximum", func() {
			runner := taskQueueTest.NewSleepRunner(taskTest.RandomType(), 3*time.Minute, time.Minute, time.Minute, 0)
			que, err := taskQueue.New(taskTest.RandomType(), cfg, lgr, str, runner)
			Expect(err).To(MatchError("runner timeout is invalid"))
			Expect(que).To(BeNil())
		})

		It("returns an error when a runner deadline does not exceed its timeout", func() {
			runner := taskQueueTest.NewSleepRunner(taskTest.RandomType(), 2*time.Minute, 2*time.Minute, time.Minute, 0)
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
		var que taskQueue.Queue

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
			if str != nil {
				Expect(str.Terminate(ctx)).To(Succeed())
			}
		})

		Context("with a single queue", func() {
			var countingRunner *taskQueueTest.CountingRunner
			var panicRunner *taskQueueTest.PanicRunner
			var cfg *taskQueue.Config
			var que taskQueue.Queue

			BeforeEach(func() {
				countingRunner = taskQueueTest.NewCountingRunner(taskTest.RandomType())
				panicRunner = taskQueueTest.NewPanicRunner(taskTest.RandomType())

				cfg = &taskQueue.Config{Workers: 2, Delay: time.Millisecond, DelayInitial: time.Millisecond, DelayUnstick: taskQueue.DelayUnstickDefault, StopWaitTimeout: taskQueue.StopWaitTimeoutDefault}
				que = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, countingRunner, panicRunner))
			})

			AfterEach(func() {
				que.Stop()
				lgr.AssertDebug("Task queue worker stopped", log.Fields{"error": errors.NewSerializable(context.Canceled)})
				lgr.AssertDebug("Task queue manager stopped", log.Fields{"error": errors.NewSerializable(context.Canceled)})
			})

			It("dispatches, runs, and completes a pending task matching a registered runner", func() {
				createdTask := test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: countingRunner.GetRunnerType()}))

				que.Start()

				Eventually(func() string {
					return test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil)).State
				}, "5s", "50ms").To(Equal(task.TaskStateCompleted))

				Expect(countingRunner.GetCount()).To(Equal(1))
			})

			It("completes a task that the runner updated while it was running", func() {
				updatingRunner := taskQueueTest.NewCallbackRunner(taskTest.RandomType(), func(ctx context.Context, tsk *task.Task) {
					// A runner may update its own task mid-run. The update bumps the revision but
					// leaves the state lock intact, so the queue's completion must still match.
					data := map[string]any{"key": "value"}
					updated, err := str.NewTaskRepository().UpdateTask(ctx, tsk.ID, nil, &task.TaskUpdate{Data: &data})
					if err != nil || updated == nil {
						tsk.AppendError(errors.New("unable to update task during run"))
						return
					}
					*tsk = *updated // replace the in-memory task with the updated one, per the runner contract
					tsk.SetCompleted()
				})
				que = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, updatingRunner))

				createdTask := test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: updatingRunner.GetRunnerType()}))

				que.Start()

				Eventually(func() string {
					return test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil)).State
				}, "5s", "50ms").To(Equal(task.TaskStateCompleted))

				actualTask := test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil))
				Expect(actualTask.State).To(Equal(task.TaskStateCompleted))
				Expect(actualTask.Error).To(BeNil())
				Expect(actualTask.Data).To(HaveKeyWithValue("key", "value"))
			})

			It("does not complete a task whose state lock changed while it was running", func() {
				stealingRunner := taskQueueTest.NewCallbackRunner(taskTest.RandomType(), func(ctx context.Context, tsk *task.Task) {
					// Simulate the task being unstuck and re-claimed elsewhere by changing the
					// state lock out from under this run. The completion must then miss rather
					// than falsely complete another run's task.
					_, err := str.GetCollection("tasks").UpdateOne(ctx, bson.M{"id": tsk.ID}, bson.M{"$set": bson.M{"stateLock": "ffffffffffffffffffffffffffffffff"}})
					if err != nil {
						tsk.AppendError(err)
						return
					}
					tsk.SetCompleted()
				})
				que = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, stealingRunner))

				createdTask := test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: stealingRunner.GetRunnerType()}))

				que.Start()

				// The completion's compare-and-swap misses, which is logged rather than silently swallowed.
				Eventually(func() bool {
					defer func() { _ = recover() }()
					lgr.AssertWarn("Unable to stop task; no running task matched the expected condition")
					return true
				}, "5s", "100ms").To(BeTrue())

				// The task was not falsely completed; it remains running (recovered later by unstick).
				actualTask := test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil))
				Expect(actualTask.State).To(Equal(task.TaskStateRunning))
			})

			It("cleans up a task that panics during execution", func() {
				createdTask := test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: panicRunner.GetRunnerType()}))

				que.Start()

				Eventually(func() string {
					return test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil)).State
				}, "5s", "50ms").To(Equal(task.TaskStateFailed))

				Expect(countingRunner.GetCount()).To(Equal(0))

				actualTask := test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil))
				Expect(actualTask.State).To(Equal(task.TaskStateFailed))
				Expect(actualTask.Error).ToNot(BeNil())
				Expect(actualTask.Error.Error).To(MatchError("unhandled panic"))

				lgr.AssertError("Unhandled panic while running task")
				Expect(testutil.ToFloat64(taskQueue.RunPanicTotal.WithLabelValues(panicRunner.GetRunnerType()))).To(Equal(float64(1)))
			})

			It("fails a pending task that does not match any registered runner", func() {
				createdTask := test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: taskTest.RandomType()}))

				que.Start()

				Eventually(func() string {
					return test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil)).State
				}, "5s", "50ms").To(Equal(task.TaskStateFailed))

				Expect(countingRunner.GetCount()).To(Equal(0))

				actualTask := test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil))
				Expect(actualTask.State).To(Equal(task.TaskStateFailed))
				Expect(actualTask.Error).ToNot(BeNil())
				Expect(actualTask.Error.Error).To(MatchError("runner not found for task"))
			})

			It("fails a task whose runner leaves it in an unknown state", func() {
				unknownStateRunner := taskQueueTest.NewCallbackRunner(taskTest.RandomType(), func(ctx context.Context, tsk *task.Task) {
					tsk.State = "unknown-state"
				})
				que = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, unknownStateRunner))

				createdTask := test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: unknownStateRunner.GetRunnerType()}))

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
				missingAvailableTimeRunner := taskQueueTest.NewCallbackRunner(taskTest.RandomType(), func(ctx context.Context, tsk *task.Task) {
					tsk.State = task.TaskStatePending
					tsk.AvailableTime = nil
				})
				que = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, missingAvailableTimeRunner))

				test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: missingAvailableTimeRunner.GetRunnerType()}))

				que.Start()

				Eventually(func() bool {
					defer func() { _ = recover() }()
					lgr.AssertWarn("Available time missing for pending task")
					return true
				}, "5s", "50ms").To(BeTrue())
			})

			It("warns when a pending task's available time is significantly in the past", func() {
				staleAvailableTimeRunner := taskQueueTest.NewCallbackRunner(taskTest.RandomType(), func(ctx context.Context, tsk *task.Task) {
					tsk.RepeatAvailableAt(time.Now().Add(-2 * time.Minute))
				})
				que = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, staleAvailableTimeRunner))

				test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: staleAvailableTimeRunner.GetRunnerType()}))

				que.Start()

				Eventually(func() bool {
					defer func() { _ = recover() }()
					lgr.AssertWarn("Available time significantly before now for pending task")
					return true
				}, "5s", "50ms").To(BeTrue())
			})

			It("unsticks and logs a task left running past its deadline", func() {
				stuckTask := &task.Task{
					ID:           task.NewID(),
					Type:         countingRunner.GetRunnerType(),
					State:        task.TaskStateRunning,
					DeadlineTime: pointer.FromTime(time.Now().Add(-time.Minute)),
					StateLock:    pointer.FromString(taskTest.RandomType()),
					CreatedTime:  time.Now(),
					Revision:     1,
				}
				_, err := str.GetCollection("tasks").InsertOne(ctx, stuckTask)
				Expect(err).ToNot(HaveOccurred())

				unstickConfig := &taskQueue.Config{Workers: 2, Delay: time.Millisecond, DelayInitial: time.Millisecond, DelayUnstick: time.Millisecond, StopWaitTimeout: taskQueue.StopWaitTimeoutDefault}
				que = test.Must(taskQueue.New(taskTest.RandomType(), unstickConfig, lgr, str, countingRunner))

				que.Start()

				Eventually(func() bool {
					defer func() { _ = recover() }()
					lgr.AssertInfo("Unstuck tasks")
					return true
				}, "5s", "50ms").To(BeTrue())

				// Once unstuck, the task returns to pending and is dispatched, run, and completed.
				Eventually(func() string {
					return test.Must(str.NewTaskRepository().GetTask(ctx, stuckTask.ID, nil)).State
				}, "5s", "50ms").To(Equal(task.TaskStateCompleted))
			})

			It("reverts a task that is still running to pending when the queue is stopped", func() {
				hangingRunner := taskQueueTest.NewHangingRunner(taskTest.RandomType(), time.Minute)
				que = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, hangingRunner))

				createdTask := test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: hangingRunner.GetRunnerType()}))

				que.Start()

				// Wait until the runner has picked up the task and it is running.
				Eventually(func() string {
					return test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil)).State
				}, "5s", "50ms").To(Equal(task.TaskStateRunning))

				// Stopping cancels the worker context; the hanging runner returns leaving the
				// task running, so the queue must revert it to pending for a later retry.
				que.Stop()

				actualTask := test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil))
				Expect(actualTask.State).To(Equal(task.TaskStatePending))
				Expect(actualTask.DeadlineTime).To(BeNil())
			})

			It("cancels the context of a task that exceeds its timeout", func() {
				sleepRunner := taskQueueTest.NewSleepRunner(taskTest.RandomType(), time.Minute, 100*time.Millisecond, 50*time.Millisecond, 10*time.Second)
				que = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, sleepRunner))

				createdTask := test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: sleepRunner.GetRunnerType()}))

				que.Start()

				Eventually(func() string {
					return test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil)).State
				}, "1m", "50ms").To(Equal(task.TaskStateCompleted))

				Expect(countingRunner.GetCount()).To(Equal(0))

				actualTask := test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil))
				Expect(actualTask.State).To(Equal(task.TaskStateCompleted))
				Expect(actualTask.Error).ToNot(BeNil())
				Expect(actualTask.Error.Error).To(MatchError("task runner timeout exceeded"))
			})

			It("fails a task whose runner returns without setting a terminal state", func() {
				noopRunner := taskQueueTest.NewCallbackRunner(taskTest.RandomType(), nil)
				que = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, noopRunner))

				createdTask := test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: noopRunner.GetRunnerType()}))

				que.Start()

				Eventually(func() string {
					return test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil)).State
				}, "5s", "50ms").To(Equal(task.TaskStateFailed))

				actualTask := test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil))
				Expect(actualTask.State).To(Equal(task.TaskStateFailed))
				Expect(actualTask.Error).ToNot(BeNil())
				Expect(actualTask.Error.Error).To(MatchError("runner failed to set terminal task state"))
			})

			It("fails a task that exceeds its timeout without setting its own state", func() {
				hangingRunner := taskQueueTest.NewHangingRunner(taskTest.RandomType(), 200*time.Millisecond)
				que = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, hangingRunner))

				createdTask := test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: hangingRunner.GetRunnerType()}))

				que.Start()

				Eventually(func() string {
					return test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil)).State
				}, "5s", "50ms").To(Equal(task.TaskStateFailed))

				actualTask := test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil))
				Expect(actualTask.State).To(Equal(task.TaskStateFailed))
				Expect(actualTask.Error).ToNot(BeNil())
				Expect(actualTask.Error.Error).To(MatchError("task runner timeout exceeded"))
			})

			It("logs a warning if a task that exceeds its maximum duration", func() {
				sleepRunner := taskQueueTest.NewSleepRunner(taskTest.RandomType(), 2*time.Minute, time.Minute, time.Millisecond, 10*time.Millisecond)
				que = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, sleepRunner))

				createdTask := test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: sleepRunner.GetRunnerType()}))

				que.Start()

				Eventually(func() string {
					return test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil)).State
				}, "1m", "50ms").To(Equal(task.TaskStateCompleted))

				Expect(countingRunner.GetCount()).To(Equal(0))

				actualTask := test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil))
				Expect(actualTask.State).To(Equal(task.TaskStateCompleted))
				Expect(actualTask.Error).To(BeNil())

				lgr.AssertWarn("Task duration exceeds maximum")
			})

			It("does not race or panic when Stop is called concurrently", func() {
				createdTask := test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: countingRunner.GetRunnerType()}))

				que.Start()

				Eventually(func() string {
					return test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil)).State
				}, "5s", "50ms").To(Equal(task.TaskStateCompleted))

				var waitGroup sync.WaitGroup
				for range 3 {
					waitGroup.Go(func() {
						defer GinkgoRecover()
						que.Stop()
					})
				}
				waitGroup.Wait()
			})
		})

		Context("with a blocking runner", func() {
			var blockingRunner *taskQueueTest.BlockingRunner
			var que taskQueue.Queue

			BeforeEach(func() {
				blockingRunner = taskQueueTest.NewBlockingRunner(taskTest.RandomType(), time.Minute)

				cfg := &taskQueue.Config{Workers: 1, Delay: time.Millisecond, DelayInitial: time.Millisecond, DelayUnstick: taskQueue.DelayUnstickDefault, StopWaitTimeout: 250 * time.Millisecond}
				que = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, blockingRunner))
			})

			It("returns from Stop within the stop timeout when a runner ignores cancellation", func() {
				createdTask := test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: blockingRunner.GetRunnerType()}))

				que.Start()

				// Wait until the blocking runner has picked up the task and it is running.
				Eventually(func() string {
					return test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil)).State
				}, "5s", "50ms").To(Equal(task.TaskStateRunning))

				// Stop must return within roughly the stop timeout even though the runner never
				// returns; it abandons the in-flight task rather than blocking forever.
				stopped := make(chan struct{})
				go func() {
					defer GinkgoRecover()
					que.Stop()
					close(stopped)
				}()
				Eventually(stopped, "5s").Should(BeClosed())

				lgr.AssertError("Task queue workers did not stop within timeout; abandoning in-flight tasks")
			})
		})

		Context("with a runner that exceeds its timeout without returning", func() {
			var blockingRunner *taskQueueTest.BlockingRunner
			var que taskQueue.Queue

			BeforeEach(func() {
				// A short duration maximum yields a short runner timeout (3x), so the watchdog fires
				// quickly while the runner is still blocked, without an unstick reclaiming the task.
				blockingRunner = taskQueueTest.NewBlockingRunner(taskTest.RandomType(), 20*time.Millisecond)

				cfg := &taskQueue.Config{Workers: 1, Delay: time.Millisecond, DelayInitial: time.Millisecond, DelayUnstick: taskQueue.DelayUnstickDefault, StopWaitTimeout: 250 * time.Millisecond}
				que = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, blockingRunner))
			})

			It("logs and counts the run once its timeout elapses", func() {
				runnerType := blockingRunner.GetRunnerType()
				taskQueue.RunnerTimeoutExceededTotal.Reset()

				createdTask := test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: runnerType}))

				que.Start()

				// Wait until the runner has picked up the task and it is running.
				Eventually(func() string {
					return test.Must(str.NewTaskRepository().GetTask(ctx, createdTask.ID, nil)).State
				}, "5s", "50ms").To(Equal(task.TaskStateRunning))

				// The watchdog fires after the runner timeout, recording the stuck run.
				Eventually(func() float64 {
					return testutil.ToFloat64(taskQueue.RunnerTimeoutExceededTotal.WithLabelValues(runnerType))
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

			var runner *taskQueueTest.RepeatRunner
			var ques []taskQueue.Queue

			BeforeEach(func() {
				runner = taskQueueTest.NewRepeatRunner(taskTest.RandomType())

				cfg := &taskQueue.Config{Workers: workersCount, Delay: time.Millisecond, DelayInitial: time.Millisecond, DelayUnstick: time.Millisecond, StopWaitTimeout: taskQueue.StopWaitTimeoutDefault}
				ques = make([]taskQueue.Queue, queueCount)
				for index := range len(ques) {
					ques[index] = test.Must(taskQueue.New(taskTest.RandomType(), cfg, lgr, str, runner))
				}
			})

			AfterEach(func() {
				for _, que := range ques {
					que.Stop()
				}
			})

			It("completes all runnings tasks when stopped", func() {
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
				tsk := test.Must(str.NewTaskRepository().CreateTask(ctx, &task.TaskCreate{Type: runner.GetRunnerType()}))

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

				actualTask := test.Must(str.NewTaskRepository().GetTask(ctx, tsk.ID, nil))
				Expect(actualTask.State).To(Equal(task.TaskStatePending))
			})
		})
	})
})
