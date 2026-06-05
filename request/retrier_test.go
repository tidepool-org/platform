package request_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Retrier", func() {
	var lgr *logTest.Logger
	var ctx context.Context
	var invocations int

	BeforeEach(func() {
		lgr = logTest.NewLogger()
		ctx = log.NewContextWithLogger(context.Background(), lgr)
		invocations = 0
	})

	Context("RetryNone", func() {
		It("does not retry on failure", func() {
			expectedErr := errorsTest.RandomError()
			success, err := request.RetryNone.Retry(ctx, func(ctx context.Context) (bool, error) {
				invocations++
				return false, expectedErr
			})
			errorsTest.ExpectEqual(err, expectedErr)
			Expect(success).To(BeFalse())
			Expect(invocations).To(Equal(1))
		})

		It("returns success on the first attempt", func() {
			success, err := request.RetryNone.Retry(ctx, func(ctx context.Context) (bool, error) {
				invocations++
				return true, nil
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeTrue())
			Expect(invocations).To(Equal(1))
		})
	})

	Context("RetryFailure", func() {
		It("returns nil when the handler succeeds immediately", func() {
			success, err := request.RetryFailure(ctx, request.RetryNone, func(ctx context.Context) (bool, error) {
				invocations++
				return true, nil
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeTrue())
			Expect(invocations).To(Equal(1))
		})

		It("returns nil after retrying when the handler eventually succeeds", func() {
			retrier := request.NewRetrier(3, 10*time.Millisecond, 0.1)
			success, err := request.RetryFailure(ctx, retrier, func(ctx context.Context) (bool, error) {
				invocations++
				if invocations < 2 {
					return false, errorsTest.RandomError()
				}
				return true, nil
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeTrue())
			Expect(invocations).To(Equal(2))
		})

		It("returns the accumulated errors when the handler always fails", func() {
			expectedErr := errorsTest.RandomError()
			success, err := request.RetryFailure(ctx, request.RetryNone, func(ctx context.Context) (bool, error) {
				invocations++
				return false, expectedErr
			})
			errorsTest.ExpectEqual(err, expectedErr)
			Expect(success).To(BeFalse())
			Expect(invocations).To(Equal(1))
		})
	})

	Context("RetryError", func() {
		It("returns nil when the handler succeeds immediately", func() {
			err := request.RetryError(ctx, request.RetryNone, func(ctx context.Context) error {
				invocations++
				return nil
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(invocations).To(Equal(1))
		})

		It("returns nil after retrying when the handler eventually succeeds", func() {
			retrier := request.NewRetrier(3, 10*time.Millisecond, 0.1)
			err := request.RetryError(ctx, retrier, func(ctx context.Context) error {
				invocations++
				if invocations < 2 {
					return errorsTest.RandomError()
				}
				return nil
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(invocations).To(Equal(2))
		})

		It("returns the accumulated errors when the handler always fails", func() {
			expectedErr := errorsTest.RandomError()
			err := request.RetryError(ctx, request.RetryNone, func(ctx context.Context) error {
				invocations++
				return expectedErr
			})
			errorsTest.ExpectEqual(err, expectedErr)
			Expect(invocations).To(Equal(1))
		})
	})

	Context("RetryMissing", func() {
		It("returns the result when the handler succeeds immediately", func() {
			expectedResult := pointer.From(test.RandomString())
			result, err := request.RetryMissing(ctx, request.RetryNone, func(ctx context.Context) (*string, error) {
				invocations++
				return expectedResult, nil
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(expectedResult))
			Expect(invocations).To(Equal(1))
		})

		It("retries when the handler returns a nil result without an error", func() {
			retrier := request.NewRetrier(3, 10*time.Millisecond, 0.1)
			expectedResult := pointer.From(test.RandomString())
			result, err := request.RetryMissing(ctx, retrier, func(ctx context.Context) (*string, error) {
				invocations++
				if invocations < 2 {
					return nil, nil
				}
				return expectedResult, nil
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(expectedResult))
			Expect(invocations).To(Equal(2))
		})

		It("returns the accumulated errors and a nil result when the handler always fails", func() {
			expectedErr := errorsTest.RandomError()
			result, err := request.RetryMissing(ctx, request.RetryNone, func(ctx context.Context) (*string, error) {
				invocations++
				return nil, expectedErr
			})
			errorsTest.ExpectEqual(err, expectedErr)
			Expect(result).To(BeNil())
			Expect(invocations).To(Equal(1))
		})
	})

	Context("NewRetrier", func() {
		It("returns successfully", func() {
			Expect(request.NewRetrier(3, time.Second, 0.1)).ToNot(BeNil())
		})

		It("clamps negative retries to zero", func() {
			retrier := request.NewRetrier(-1, time.Second, 0.1)
			expectedErr := errorsTest.RandomError()
			success, err := retrier.Retry(ctx, func(ctx context.Context) (bool, error) {
				invocations++
				return false, expectedErr
			})
			errorsTest.ExpectEqual(err, expectedErr)
			Expect(success).To(BeFalse())
			Expect(invocations).To(Equal(1))
		})

		It("clamps negative delay to zero", func() {
			retrier := request.NewRetrier(2, -time.Second, 0.1)
			start := time.Now()
			success, err := retrier.Retry(ctx, func(ctx context.Context) (bool, error) {
				invocations++
				return invocations >= 3, nil
			})
			elapsed := time.Since(start)
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeTrue())
			Expect(invocations).To(Equal(3))
			Expect(elapsed).To(BeNumerically("<", time.Second))
		})

		It("uses the absolute value of a negative jitter", func() {
			retrier := request.NewRetrier(1, 10*time.Millisecond, -0.5)
			success, err := retrier.Retry(ctx, func(ctx context.Context) (bool, error) {
				invocations++
				return invocations >= 2, nil
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeTrue())
			Expect(invocations).To(Equal(2))
		})
	})

	Context("Retry", func() {
		It("passes the context to the handler", func() {
			retrier := request.NewRetrier(0, time.Second, 0.1)
			type key struct{}
			expectedContext := context.WithValue(ctx, key{}, "value")
			success, err := retrier.Retry(expectedContext, func(ctx context.Context) (bool, error) {
				Expect(ctx).To(Equal(expectedContext))
				return true, nil
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeTrue())
		})

		It("returns success without retrying when the handler succeeds immediately", func() {
			retrier := request.NewRetrier(3, 10*time.Millisecond, 0.1)
			success, err := retrier.Retry(ctx, func(ctx context.Context) (bool, error) {
				invocations++
				return true, nil
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeTrue())
			Expect(invocations).To(Equal(1))
		})

		It("returns success after retrying when the handler eventually succeeds", func() {
			retrier := request.NewRetrier(3, 10*time.Millisecond, 0.1)
			success, err := retrier.Retry(ctx, func(ctx context.Context) (bool, error) {
				invocations++
				return invocations >= 3, nil
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeTrue())
			Expect(invocations).To(Equal(3))
		})

		It("clears a previous error once the handler succeeds", func() {
			retrier := request.NewRetrier(3, 10*time.Millisecond, 0.1)
			success, err := retrier.Retry(ctx, func(ctx context.Context) (bool, error) {
				invocations++
				if invocations < 2 {
					return false, errorsTest.RandomError()
				}
				return true, nil
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeTrue())
			Expect(invocations).To(Equal(2))
		})

		It("returns failure and accumulated errors after exhausting all retries", func() {
			retrier := request.NewRetrier(2, 10*time.Millisecond, 0.1)
			expectedErrs := []error{}
			success, err := retrier.Retry(ctx, func(ctx context.Context) (bool, error) {
				invocations++
				expectedErr := errorsTest.RandomError()
				expectedErrs = append(expectedErrs, expectedErr)
				return false, expectedErr
			})
			errorsTest.ExpectEqual(err, expectedErrs...)
			Expect(success).To(BeFalse())
			Expect(invocations).To(Equal(3))
		})

		It("returns failure without error when the handler never succeeds and never errors", func() {
			retrier := request.NewRetrier(2, 10*time.Millisecond, 0.1)
			success, err := retrier.Retry(ctx, func(ctx context.Context) (bool, error) {
				invocations++
				return false, nil
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeFalse())
			Expect(invocations).To(Equal(3))
		})

		It("stops retrying and returns the context error when the context is cancelled during the delay", func() {
			retrier := request.NewRetrier(5, time.Hour, 0.1)
			cancelCtx, cancel := context.WithCancel(ctx)
			success, err := retrier.Retry(cancelCtx, func(ctx context.Context) (bool, error) {
				invocations++
				if invocations == 1 {
					cancel()
				}
				return false, errorsTest.RandomError()
			})
			Expect(err).To(Equal(context.Canceled))
			Expect(success).To(BeFalse())
			Expect(invocations).To(Equal(1))
		})

		It("stops retrying and returns the context error when the context is already done", func() {
			retrier := request.NewRetrier(5, time.Hour, 0.1)
			cancelCtx, cancel := context.WithCancel(ctx)
			cancel()
			success, err := retrier.Retry(cancelCtx, func(ctx context.Context) (bool, error) {
				invocations++
				return false, errorsTest.RandomError()
			})
			Expect(err).To(Equal(context.Canceled))
			Expect(success).To(BeFalse())
			Expect(invocations).To(Equal(1))
		})

		It("delays with exponential backoff between retries", func() {
			retrier := request.NewRetrier(3, 20*time.Millisecond, 0)
			start := time.Now()
			success, err := retrier.Retry(ctx, func(ctx context.Context) (bool, error) {
				invocations++
				return invocations >= 4, nil
			})
			elapsed := time.Since(start)
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeTrue())
			Expect(invocations).To(Equal(4))
			Expect(elapsed).To(BeNumerically(">=", 140*time.Millisecond)) // Delays are approximately 20ms, 40ms, and 80ms (1x, 2x, 4x the base delay).
		})

		It("applies jitter to the delay between retries", func() {
			retrier := request.NewRetrier(1, 100*time.Millisecond, 0.5)
			start := time.Now()
			success, err := retrier.Retry(ctx, func(ctx context.Context) (bool, error) {
				invocations++
				return invocations >= 2, nil
			})
			elapsed := time.Since(start)
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeTrue())
			Expect(invocations).To(Equal(2))
			Expect(elapsed).To(BeNumerically(">=", 50*time.Millisecond)) // With 50% jitter, the single delay is in the range [50ms, 150ms].
		})

		It("logs info for every retry", func() {
			retrier := request.NewRetrier(2, 10*time.Millisecond, 0.1)
			success, err := retrier.Retry(ctx, func(ctx context.Context) (bool, error) {
				invocations++
				return false, nil
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeFalse())
			Expect(invocations).To(Equal(3))
			Expect(lgr.SerializedFields).To(HaveLen(2))
			lgr.AssertInfo("Delaying before retrying")
		})
	})
})
