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
		return p.Failed(ctx, wrk, errors.New("pending process resulter is not configured"))
	}
	processingAvailableDuration := p.ProcessResultPendingBuilder.ProcessingAvailableDuration(ctx, wrk)
	return work.NewProcessResultPending(work.PendingUpdate{
		ProcessingAvailableTime: time.Now().Add(processingAvailableDuration),
		Metadata:                wrk.Metadata,
	})
}

func (p *ProcessResultBuilder) Failing(ctx context.Context, wrk *work.Work, err error) *work.ProcessResult {
	if p.ProcessResultFailingBuilder == nil {
		return p.Failed(ctx, wrk, errors.New("failing process resulter is not configured"))
	}
	failingRetryCount := p.ProcessResultFailingBuilder.FailingRetryCount(ctx, wrk, err)
	failingRetryDuration := p.ProcessResultFailingBuilder.FailingRetryDuration(ctx, wrk, err, failingRetryCount)
	log.LoggerFromContext(ctx).WithFields(log.Fields{"failingRetryCount": failingRetryCount, "failingRetryDuration": failingRetryDuration}).WithError(err).Error("processor failing")
	return work.NewProcessResultFailing(work.FailingUpdate{
		FailingError:      errors.Serializable{Error: err},
		FailingRetryCount: failingRetryCount,
		FailingRetryTime:  time.Now().Add(failingRetryDuration),
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

// Constant duration pending process resulter
type ConstantProcessResultPendingBuilder struct {
	Duration time.Duration
}

func (c *ConstantProcessResultPendingBuilder) ProcessingAvailableDuration(ctx context.Context, wrk *work.Work) time.Duration {
	return c.Duration
}

// Linear failing process resulter
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

func (c *ConstantProcessResultFailingBuilder) FailingRetryDuration(ctx context.Context, wrk *work.Work, err error, retryCount int) time.Duration {
	return c.Duration
}

type ExponentialProcessResultFailingBuilder struct {
	LinearProcessResultFailingBuilder
	Duration       time.Duration
	JitterDuration time.Duration
}

func (e *ExponentialProcessResultFailingBuilder) FailingRetryDuration(ctx context.Context, wrk *work.Work, err error, retryCount int) time.Duration {
	if retryCount < 1 {
		return 0
	}
	fallbackFactor := time.Duration(1 << (retryCount - 1))
	jitterDuration := int64(e.JitterDuration * fallbackFactor)
	return e.Duration*fallbackFactor + time.Duration(rand.Int63n(2*jitterDuration)-jitterDuration)
}
