package base

import (
	"context"
	"math"
	"time"

	"github.com/tidepool-org/platform/crypto"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/work"
)

type ProcessResultBuilder struct {
	work.ProcessResultPendingBuilder
	work.ProcessResultFailingBuilder
}

func (p *ProcessResultBuilder) Pending(ctx context.Context, wrk *work.Work, tm time.Time) *work.ProcessResult {
	if p.ProcessResultPendingBuilder == nil {
		return p.Failed(ctx, wrk, errors.New("pending process result builder is not configured"), tm)
	}
	processingAvailableTime := p.ProcessingAvailableTime(ctx, wrk, tm)
	return work.NewProcessResultPending(work.PendingUpdate{
		ProcessingAvailableTime: processingAvailableTime,
		ProcessingPriority:      wrk.ProcessingPriority,
		ProcessingTimeout:       wrk.ProcessingTimeout,
		Metadata:                wrk.Metadata,
	})
}

func (p *ProcessResultBuilder) Failing(ctx context.Context, wrk *work.Work, err error, tm time.Time) *work.ProcessResult {
	if p.ProcessResultFailingBuilder == nil {
		return p.Failed(ctx, wrk, errors.New("failing process result builder is not configured"), tm)
	}
	failingRetryCount := p.FailingRetryCount(ctx, wrk, err)
	failingRetryTime := p.FailingRetryTime(ctx, wrk, err, failingRetryCount, tm)
	log.LoggerFromContext(ctx).WithFields(log.Fields{"failingRetryCount": failingRetryCount, "failingRetryTime": failingRetryTime}).WithError(err).Error("processor failing")
	return work.NewProcessResultFailing(work.FailingUpdate{
		FailingError:      errors.Serializable{Error: err},
		FailingRetryCount: failingRetryCount,
		FailingRetryTime:  failingRetryTime,
		Metadata:          wrk.Metadata,
	})
}

func (p *ProcessResultBuilder) Failed(ctx context.Context, wrk *work.Work, err error, tm time.Time) *work.ProcessResult {
	log.LoggerFromContext(ctx).WithError(err).Error("processor failed")
	return work.NewProcessResultFailed(work.FailedUpdate{
		FailedError: errors.Serializable{Error: err},
		Metadata:    wrk.Metadata,
	})
}

func (p *ProcessResultBuilder) Success(ctx context.Context, wrk *work.Work, tm time.Time) *work.ProcessResult {
	return work.NewProcessResultSuccess(work.SuccessUpdate{
		Metadata: wrk.Metadata,
	})
}

func (p *ProcessResultBuilder) Delete(ctx context.Context, wrk *work.Work, tm time.Time) *work.ProcessResult {
	return work.NewProcessResultDelete()
}

type ConstantProcessResultPendingBuilder struct {
	Duration time.Duration
}

func (c *ConstantProcessResultPendingBuilder) ProcessingAvailableTime(ctx context.Context, wrk *work.Work, tm time.Time) time.Time {
	return tm.Add(c.Duration)
}

type LinearProcessResultFailingBuilder struct{}

func (l *LinearProcessResultFailingBuilder) FailingRetryCount(ctx context.Context, wrk *work.Work, err error) int {
	if wrk == nil || wrk.FailingRetryCount == nil {
		return 1
	} else if *wrk.FailingRetryCount < math.MaxInt {
		return *wrk.FailingRetryCount + 1
	} else {
		return math.MaxInt
	}
}

type ConstantProcessResultFailingBuilder struct {
	LinearProcessResultFailingBuilder
	Duration time.Duration
}

func (c *ConstantProcessResultFailingBuilder) FailingRetryTime(ctx context.Context, wrk *work.Work, err error, failingRetryCount int, tm time.Time) time.Time {
	return tm.Add(c.Duration)
}

type ExponentialProcessResultFailingBuilder struct {
	LinearProcessResultFailingBuilder
	Duration        time.Duration
	DurationJitter  time.Duration
	DurationMaximum *time.Duration
}

func (e *ExponentialProcessResultFailingBuilder) FailingRetryTime(ctx context.Context, wrk *work.Work, err error, failingRetryCount int, tm time.Time) time.Time {
	duration := durationWithJitterExponential(e.Duration, e.DurationJitter, failingRetryCount-1)
	if e.DurationMaximum != nil && duration > *e.DurationMaximum {
		duration = *e.DurationMaximum
	}
	return tm.Add(duration)
}

func durationWithJitterExponential(duration time.Duration, durationJitter time.Duration, exponent int) time.Duration {
	if duration < 0 || exponent < 0 {
		return 0
	}
	duration = durationExponential(duration, exponent)
	durationJitter = time.Duration(crypto.RandomInt64N(int64(durationExponential(durationAbsolute(durationJitter), exponent))))
	if crypto.RandomBool() {
		durationJitter = min(durationJitter, durationMaximum-duration)
	} else {
		durationJitter = -min(durationJitter, duration)
	}
	return duration + durationJitter
}

func durationExponential(duration time.Duration, exponent int) time.Duration {
	if duration <= 0 || exponent < 0 {
		return 0
	} else if exponent == 0 {
		return duration
	} else if exponent > int(math.Log2(float64(durationMaximum)/float64(duration))) {
		return durationMaximum
	} else {
		return time.Duration(int64(duration) * (1 << exponent))
	}
}

func durationAbsolute(duration time.Duration) time.Duration {
	if duration < 0 {
		return -duration
	} else {
		return duration
	}
}

const durationMaximum = time.Duration(math.MaxInt64)
