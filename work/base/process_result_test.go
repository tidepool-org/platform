package base_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	gomock "go.uber.org/mock/gomock"

	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("process_result", func() {
	var ctx context.Context
	var mockController *gomock.Controller
	var wrk *work.Work
	var tm time.Time

	BeforeEach(func() {
		ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
		mockController, ctx = gomock.WithContext(ctx, GinkgoT())
		wrk = workTest.RandomWork()
		tm = test.RandomTime()
	})

	Context("ProcessResultBuilder", func() {
		var builder *workBase.ProcessResultBuilder

		BeforeEach(func() {
			builder = &workBase.ProcessResultBuilder{}
		})

		Context("Pending", func() {
			It("returns Failed ProcessResult when PendingBuilder is nil", func() {
				Expect(builder.Pending(ctx, wrk, tm)).To(workTest.MatchFailedProcessResult(MatchAllFields(Fields{
					"FailedError": errorsTest.MatchSerialized(MatchError("pending process result builder is not configured")),
					"Metadata":    Equal(wrk.Metadata),
				})))
			})

			It("returns Pending ProcessResult when PendingBuilder is configured", func() {
				processingAvailableTime := test.RandomTimeAfterNow()

				mockProcessResultPendingBuilder := workTest.NewMockProcessResultPendingBuilder(mockController)
				mockProcessResultPendingBuilder.EXPECT().
					ProcessingAvailableTime(ctx, wrk, gomock.AssignableToTypeOf(time.Time{})).
					Return(processingAvailableTime)

				builder.ProcessResultPendingBuilder = mockProcessResultPendingBuilder
				Expect(builder.Pending(ctx, wrk, tm)).To(workTest.MatchPendingProcessResult(Equal(work.PendingUpdate{
					ProcessingAvailableTime: processingAvailableTime,
					ProcessingPriority:      wrk.ProcessingPriority,
					ProcessingTimeout:       wrk.ProcessingTimeout,
					Metadata:                wrk.Metadata,
				})))
			})
		})

		Context("Failing", func() {
			var err error

			BeforeEach(func() {
				err = errorsTest.RandomError()
			})

			It("returns Failed ProcessResult when FailingBuilder is nil", func() {
				Expect(builder.Failing(ctx, wrk, err, tm)).To(workTest.MatchFailedProcessResult(MatchAllFields(Fields{
					"FailedError": errorsTest.MatchSerialized(MatchError("failing process result builder is not configured")),
					"Metadata":    Equal(wrk.Metadata),
				})))
			})

			It("returns Failing ProcessResult when FailingBuilder is configured", func() {
				failingRetryCount := test.RandomIntFromRange(0, test.RandomIntMaximum())
				failingRetryTime := test.RandomTimeAfterNow()

				mockProcessResultFailingBuilder := workTest.NewMockProcessResultFailingBuilder(mockController)
				mockProcessResultFailingBuilder.EXPECT().
					FailingRetryCount(ctx, wrk, err).
					Return(failingRetryCount)
				mockProcessResultFailingBuilder.EXPECT().
					FailingRetryTime(ctx, wrk, err, failingRetryCount, gomock.AssignableToTypeOf(time.Time{})).
					Return(failingRetryTime)

				builder.ProcessResultFailingBuilder = mockProcessResultFailingBuilder
				Expect(builder.Failing(ctx, wrk, err, tm)).To(workTest.MatchFailingProcessResult(Equal(work.FailingUpdate{
					FailingError:      errors.Serializable{Error: err},
					FailingRetryCount: failingRetryCount,
					FailingRetryTime:  failingRetryTime,
					Metadata:          wrk.Metadata,
				})))
			})
		})

		Context("Failed", func() {
			It("returns Failed ProcessResult", func() {
				err := errorsTest.RandomError()
				Expect(builder.Failed(ctx, wrk, err, tm)).To(workTest.MatchFailedProcessResult(Equal(work.FailedUpdate{
					FailedError: errors.Serializable{Error: err},
					Metadata:    wrk.Metadata,
				})))
			})
		})

		Context("Success", func() {
			It("returns Success ProcessResult", func() {
				Expect(builder.Success(ctx, wrk, tm)).To(workTest.MatchSuccessProcessResult(Equal(work.SuccessUpdate{
					Metadata: wrk.Metadata,
				})))
			})
		})

		Context("Delete", func() {
			It("returns Delete ProcessResult", func() {
				Expect(builder.Delete(ctx, wrk, tm)).To(workTest.MatchDeleteProcessResult())
			})
		})
	})

	Context("ConstantProcessResultPendingBuilder", func() {
		It("returns duration after now", func() {
			now := test.RandomTime()
			duration := test.RandomDuration()
			builder := &workBase.ConstantProcessResultPendingBuilder{Duration: duration}
			Expect(builder.ProcessingAvailableTime(ctx, wrk, now)).To(Equal(now.Add(duration)))
		})
	})

	Context("LinearProcessResultFailingBuilder", func() {
		var err error
		var builder *workBase.LinearProcessResultFailingBuilder

		BeforeEach(func() {
			err = errorsTest.RandomError()
			builder = &workBase.LinearProcessResultFailingBuilder{}
		})

		It("returns 1 when work is nil", func() {
			Expect(builder.FailingRetryCount(ctx, nil, err)).To(Equal(1))
		})

		It("returns 1 when failing retry count is nil", func() {
			wrk.FailingRetryCount = nil
			Expect(builder.FailingRetryCount(ctx, wrk, err)).To(Equal(1))
		})

		It("increments failing retry count", func() {
			failingRetryCount := test.RandomIntFromRange(0, test.RandomIntMaximum()-1)
			wrk.FailingRetryCount = &failingRetryCount
			Expect(builder.FailingRetryCount(ctx, wrk, err)).To(Equal(failingRetryCount + 1))
		})
	})

	Context("ConstantProcessResultFailingBuilder", func() {
		var err error
		var failingRetryCount int
		var tm time.Time

		BeforeEach(func() {
			err = errorsTest.RandomError()
			failingRetryCount = test.RandomIntFromRange(0, 10)
			tm = test.RandomTime()
		})

		It("returns duration after now", func() {
			duration := test.RandomDuration()
			builder := &workBase.ConstantProcessResultFailingBuilder{Duration: duration}
			Expect(builder.FailingRetryTime(ctx, wrk, err, failingRetryCount, tm)).To(Equal(tm.Add(duration)))
		})
	})

	Context("ExponentialProcessResultFailingBuilder", func() {
		var err error
		var tm time.Time

		BeforeEach(func() {
			err = errorsTest.RandomError()
			tm = test.RandomTime()
		})

		It("returns duration after now if less than absolute maximum duration", func() {
			builder := &workBase.ExponentialProcessResultFailingBuilder{
				Duration: tm.AddDate(10, 0, 0).Sub(tm) - 1,
			}
			Expect(builder.FailingRetryTime(ctx, nil, err, 1, tm)).To(BeTemporally("<", tm.AddDate(10, 0, 0)))
		})

		It("returns absolute maximum duration after now", func() {
			builder := &workBase.ExponentialProcessResultFailingBuilder{
				Duration: tm.AddDate(10, 0, 0).Sub(tm) + 1,
			}
			Expect(builder.FailingRetryTime(ctx, nil, err, 1, tm)).To(Equal(tm.AddDate(10, 0, 0)))
		})

		It("returns duration after now if failing retry count within limits", func() {
			builder := &workBase.ExponentialProcessResultFailingBuilder{
				Duration: time.Second,
			}
			Expect(builder.FailingRetryTime(ctx, nil, err, 29, tm)).To(BeTemporally("<", tm.AddDate(10, 0, 0)))
		})

		It("returns absolute maximum duration after now if failing retry count exceeds limits", func() {
			builder := &workBase.ExponentialProcessResultFailingBuilder{
				Duration: time.Second,
			}
			Expect(builder.FailingRetryTime(ctx, nil, err, 30, tm)).To(Equal(tm.AddDate(10, 0, 0)))
		})

		It("returns duration after now if failing retry count within limits", func() {
			durationMaximum := 24 * time.Hour
			builder := &workBase.ExponentialProcessResultFailingBuilder{
				Duration:        time.Second,
				DurationMaximum: pointer.FromDuration(durationMaximum),
			}
			Expect(builder.FailingRetryTime(ctx, nil, err, 17, tm)).To(BeTemporally("<", tm.Add(durationMaximum)))
		})

		It("returns specified maximum duration if failing retry count exceeds limits", func() {
			durationMaximum := 24 * time.Hour
			builder := &workBase.ExponentialProcessResultFailingBuilder{
				Duration:        time.Second,
				DurationMaximum: pointer.FromDuration(durationMaximum),
			}
			Expect(builder.FailingRetryTime(ctx, nil, err, 18, tm)).To(Equal(tm.Add(durationMaximum)))
		})

		Context("with random duration and duration jitter", func() {
			var builder *workBase.ExponentialProcessResultFailingBuilder

			BeforeEach(func() {
				builder = &workBase.ExponentialProcessResultFailingBuilder{
					Duration:       test.RandomDurationFromRange(0, time.Hour),
					DurationJitter: test.RandomDurationFromRange(0, time.Minute),
				}
			})

			It("returns now for failing retry count less than 1", func() {
				Expect(builder.FailingRetryTime(ctx, nil, err, 0, tm)).To(Equal(tm))
			})

			It("returns duration after now for failing retry count of 1", func() {
				builder.DurationJitter = 0
				Expect(builder.FailingRetryTime(ctx, nil, err, 1, tm)).To(BeTemporally("==", tm.Add(builder.Duration)))
			})

			It("returns duration within duration jitter after now for failing retry count of 1", func() {
				Expect(builder.FailingRetryTime(ctx, nil, err, 1, tm)).To(And(BeTemporally(">=", tm.Add(builder.Duration-builder.DurationJitter)), BeTemporally("<=", tm.Add(builder.Duration+builder.DurationJitter))))
			})

			It("calculates exponential duration", func() {
				for index := range 17 {
					duration := time.Duration(int64(builder.Duration) * (int64(1) << index))
					durationJitter := time.Duration(int64(builder.DurationJitter) * (int64(1) << index))
					Expect(builder.FailingRetryTime(ctx, nil, err, index+1, tm)).To(And(BeTemporally(">=", tm.Add(duration-durationJitter)), BeTemporally("<=", tm.Add(duration+durationJitter))))
				}
			})
		})
	})
})
