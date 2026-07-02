package request

import (
	"context"
	"math"
	"time"

	"github.com/tidepool-org/platform/crypto"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/log"
)

var RetryNone = NewRetrier(0, 0, 0)

func RetryFailure(ctx context.Context, retrier Retrier, handler func(ctx context.Context) (bool, error)) (bool, error) {
	return retrier.Retry(ctx, handler)
}

func RetryError(ctx context.Context, retrier Retrier, handler func(ctx context.Context) error) error {
	_, err := retrier.Retry(ctx, func(ctx context.Context) (bool, error) {
		err := handler(ctx)
		return err == nil, err
	})
	return err
}

func RetryMissing[T any](ctx context.Context, retrier Retrier, handler func(ctx context.Context) (*T, error)) (*T, error) {
	var result *T
	_, err := retrier.Retry(ctx, func(ctx context.Context) (bool, error) {
		var err error
		result, err = handler(ctx)
		return result != nil, err
	})
	return result, err
}

type Retrier interface {
	Retry(ctx context.Context, handler func(ctx context.Context) (bool, error)) (bool, error)
}

func NewRetrier(retries int, delay time.Duration, jitter float64) Retrier {
	return &retrier{
		retries: max(0, retries),
		delay:   max(0, delay),
		jitter:  math.Abs(jitter),
	}
}

type retrier struct {
	retries int
	delay   time.Duration
	jitter  float64
}

func (r *retrier) Retry(ctx context.Context, handler func(ctx context.Context) (bool, error)) (bool, error) {
	var err error

	lgr := log.LoggerFromContext(ctx)

	retry := retry{ID: id.Must(id.New(16))}
	for retry.Retry = range r.retries + 1 {
		if retry.Retry > 0 {
			retry.Delay = r.delayWithExponentialBackoff(retry.Retry)
			lgr.WithField("retry", retry).Info("Delaying before retrying")
			select {
			case <-ctx.Done():
				return false, ctx.Err()
			case <-time.After(retry.Delay):
			}
		}

		retrySuccess, retryErr := handler(ctx)
		if retryErr != nil {
			err = errors.Append(err, retryErr)
		} else {
			err = nil
		}

		if retrySuccess {
			return true, err
		}

		retry.Error.Error = retryErr
	}

	return false, err
}

func (r *retrier) delayWithExponentialBackoff(retry int) time.Duration {
	if retry <= 0 || r.delay <= 0 {
		return 0
	}
	backoff := time.Duration(1 << (retry - 1))
	delay := float64(r.delay) * float64(backoff)
	jitter := int64(r.jitter * delay)
	if jitter > 0 {
		jitter = crypto.RandomInt64N(2*jitter) - jitter
	}
	return time.Duration(int64(delay) + jitter)
}

type retry struct {
	ID    string              `json:"id"`
	Retry int                 `json:"retry"`
	Error errors.Serializable `json:"error,omitzero"`
	Delay time.Duration       `json:"delay"`
}
