package base

import (
	"context"
	"math/rand"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/work"
)

type ProcessResultBuilder struct {
	work.ProcessResultPendingBuilder
	work.ProcessResultFailingBuilder
}

func (p *ProcessResultBuilder) Pending(ctx context.Context, wrk *work.Work) *work.ProcessResult {
	if p.ProcessResultPendingBuilder == nil {
		return p.Failed(ctx, wrk, errors.New("pending process result builder is not configured"))
	}
	processingAvailableTime := p.ProcessResultPendingBuilder.ProcessingAvailableTime(ctx, wrk, time.Now())
	return work.NewProcessResultPending(work.PendingUpdate{
		ProcessingAvailableTime: processingAvailableTime,
		ProcessingPriority:      wrk.ProcessingPriority,
		ProcessingTimeout:       wrk.ProcessingTimeout,
		Metadata:                wrk.Metadata,
	})
}

func (p *ProcessResultBuilder) Failing(ctx context.Context, wrk *work.Work, err error) *work.ProcessResult {
	if p.ProcessResultFailingBuilder == nil {
		return p.Failed(ctx, wrk, errors.New("failing process result builder is not configured"))
	}
	failingRetryCount := p.ProcessResultFailingBuilder.FailingRetryCount(ctx, wrk, err)
	failingRetryTime := p.ProcessResultFailingBuilder.FailingRetryTime(ctx, wrk, err, failingRetryCount, time.Now())
	log.LoggerFromContext(ctx).WithFields(log.Fields{"failingRetryCount": failingRetryCount, "failingRetryTime": failingRetryTime}).WithError(err).Error("processor failing")
	return work.NewProcessResultFailing(work.FailingUpdate{
		FailingError:      errors.Serializable{Error: err},
		FailingRetryCount: failingRetryCount,
		FailingRetryTime:  failingRetryTime,
		Metadata:          wrk.Metadata,
	})
}

func (p *ProcessResultBuilder) Failed(ctx context.Context, wrk *work.Work, err error) *work.ProcessResult {
	log.LoggerFromContext(ctx).WithError(err).Error("processor failed")
	return work.NewProcessResultFailed(work.FailedUpdate{
		FailedError: errors.Serializable{Error: err},
		Metadata:    wrk.Metadata,
	})
}

func (p *ProcessResultBuilder) Success(ctx context.Context, wrk *work.Work) *work.ProcessResult {
	return work.NewProcessResultSuccess(work.SuccessUpdate{
		Metadata: wrk.Metadata,
	})
}

func (p *ProcessResultBuilder) Delete(ctx context.Context, wrk *work.Work) *work.ProcessResult {
	return work.NewProcessResultDelete()
}

type ConstantProcessResultPendingBuilder struct {
	Duration time.Duration
}

func (c *ConstantProcessResultPendingBuilder) ProcessingAvailableTime(ctx context.Context, wrk *work.Work, now time.Time) time.Time {
	return now.Add(c.Duration)
}

type LinearProcessResultFailingBuilder struct{}

func (l *LinearProcessResultFailingBuilder) FailingRetryCount(ctx context.Context, wrk *work.Work, err error) int {
	if wrk == nil || wrk.FailingRetryCount == nil {
		return 1
	}
	return *wrk.FailingRetryCount + 1
}

type ConstantProcessResultFailingBuilder struct {
	LinearProcessResultFailingBuilder
	Duration time.Duration
}

func (c *ConstantProcessResultFailingBuilder) FailingRetryTime(ctx context.Context, wrk *work.Work, err error, failingRetryCount int, now time.Time) time.Time {
	return now.Add(c.Duration)
}

type ExponentialProcessResultFailingBuilder struct {
	LinearProcessResultFailingBuilder
	Duration       time.Duration
	DurationJitter time.Duration
}

func (e *ExponentialProcessResultFailingBuilder) FailingRetryTime(ctx context.Context, wrk *work.Work, err error, failingRetryCount int, now time.Time) time.Time {
	if failingRetryCount < 1 {
		return now
	}
	fallbackFactor := time.Duration(1 << (failingRetryCount - 1))
	durationJitter := int64(e.DurationJitter * fallbackFactor)
	duration := e.Duration*fallbackFactor + time.Duration(rand.Int63n(2*durationJitter)-durationJitter)
	return now.Add(duration)
}
