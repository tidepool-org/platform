package work

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	ResultPending = "pending"
	ResultFailing = "failing"
	ResultFailed  = "failed"
	ResultSuccess = "success"
	ResultDelete  = "delete"
)

func Results() []string {
	return []string{
		ResultPending,
		ResultFailing,
		ResultFailed,
		ResultSuccess,
		ResultDelete,
	}
}

// The result of processing the work.
type ProcessResult struct {
	Result        string         `json:"result,omitempty" bson:"result,omitempty"`
	PendingUpdate *PendingUpdate `json:"pendingUpdate,omitempty" bson:"pendingUpdate,omitempty"`
	FailingUpdate *FailingUpdate `json:"failingUpdate,omitempty" bson:"failingUpdate,omitempty"`
	FailedUpdate  *FailedUpdate  `json:"failedUpdate,omitempty" bson:"failedUpdate,omitempty"`
	SuccessUpdate *SuccessUpdate `json:"successUpdate,omitempty" bson:"successUpdate,omitempty"`
}

func NewProcessResultPending(pendingUpdate PendingUpdate) *ProcessResult {
	return &ProcessResult{Result: ResultPending, PendingUpdate: &pendingUpdate}
}

func NewProcessResultFailing(failingUpdate FailingUpdate) *ProcessResult {
	return &ProcessResult{Result: ResultFailing, FailingUpdate: &failingUpdate}
}

func NewProcessResultFailed(failedUpdate FailedUpdate) *ProcessResult {
	return &ProcessResult{Result: ResultFailed, FailedUpdate: &failedUpdate}
}

func NewProcessResultSuccess(successUpdate SuccessUpdate) *ProcessResult {
	return &ProcessResult{Result: ResultSuccess, SuccessUpdate: &successUpdate}
}

func NewProcessResultDelete() *ProcessResult {
	return &ProcessResult{Result: ResultDelete}
}

func (p *ProcessResult) Parse(parser structure.ObjectParser) {
	if ptr := parser.String("result"); ptr != nil {
		p.Result = *ptr
	}
	p.PendingUpdate = ParsePendingUpdate(parser.WithReferenceObjectParser("pendingUpdate"))
	p.FailingUpdate = ParseFailingUpdate(parser.WithReferenceObjectParser("failingUpdate"))
	p.FailedUpdate = ParseFailedUpdate(parser.WithReferenceObjectParser("failedUpdate"))
	p.SuccessUpdate = ParseSuccessUpdate(parser.WithReferenceObjectParser("successUpdate"))
}

func (p *ProcessResult) Validate(validator structure.Validator) {
	validator.String("result", &p.Result).OneOf(Results()...)
	if pendingUpdateValidator := validator.WithReference("pendingUpdate"); p.PendingUpdate != nil {
		if p.Result == ResultPending {
			p.PendingUpdate.Validate(pendingUpdateValidator)
		} else {
			pendingUpdateValidator.ReportError(structureValidator.ErrorValueExists())
		}
	} else if p.Result == ResultPending {
		pendingUpdateValidator.ReportError(structureValidator.ErrorValueNotExists())
	}
	if failingUpdateValidator := validator.WithReference("failingUpdate"); p.FailingUpdate != nil {
		if p.Result == ResultFailing {
			p.FailingUpdate.Validate(failingUpdateValidator)
		} else {
			failingUpdateValidator.ReportError(structureValidator.ErrorValueExists())
		}
	} else if p.Result == ResultFailing {
		failingUpdateValidator.ReportError(structureValidator.ErrorValueNotExists())
	}
	if failedUpdateValidator := validator.WithReference("failedUpdate"); p.FailedUpdate != nil {
		if p.Result == ResultFailed {
			p.FailedUpdate.Validate(failedUpdateValidator)
		} else {
			failedUpdateValidator.ReportError(structureValidator.ErrorValueExists())
		}
	} else if p.Result == ResultFailed {
		failedUpdateValidator.ReportError(structureValidator.ErrorValueNotExists())
	}
	if successUpdateValidator := validator.WithReference("successUpdate"); p.SuccessUpdate != nil {
		if p.Result == ResultSuccess {
			p.SuccessUpdate.Validate(successUpdateValidator)
		} else {
			successUpdateValidator.ReportError(structureValidator.ErrorValueExists())
		}
	} else if p.Result == ResultSuccess {
		successUpdateValidator.ReportError(structureValidator.ErrorValueNotExists())
	}
}

func (p *ProcessResult) Metadata() map[string]any {
	switch p.Result {
	case ResultPending:
		if p.PendingUpdate != nil {
			return p.PendingUpdate.Metadata
		}
	case ResultFailing:
		if p.FailingUpdate != nil {
			return p.FailingUpdate.Metadata
		}
	case ResultFailed:
		if p.FailedUpdate != nil {
			return p.FailedUpdate.Metadata
		}
	case ResultSuccess:
		if p.SuccessUpdate != nil {
			return p.SuccessUpdate.Metadata
		}
	}
	return nil
}

func (p *ProcessResult) Error() error {
	switch p.Result {
	case ResultFailing:
		if p.FailingUpdate != nil {
			return p.FailingUpdate.FailingError.Error
		}
	case ResultFailed:
		if p.FailedUpdate != nil {
			return p.FailedUpdate.FailedError.Error
		}
	}
	return nil
}

type ProcessResultBuilder interface {
	Pending(ctx context.Context, wrk *Work) *ProcessResult
	Failing(ctx context.Context, wrk *Work, err error) *ProcessResult
	Failed(ctx context.Context, wrk *Work, err error) *ProcessResult
	Success(ctx context.Context, wrk *Work) *ProcessResult
	Delete(ctx context.Context, wrk *Work) *ProcessResult
}

type ProcessResultPendingBuilder interface {
	ProcessingAvailableDuration(ctx context.Context, wrk *Work) time.Duration
}

type ProcessResultFailingBuilder interface {
	FailingRetryCount(ctx context.Context, wrk *Work, err error) int
	FailingRetryDuration(ctx context.Context, wrk *Work, err error, retryCount int) time.Duration
}

type ProcessPipelineFunc func() *ProcessResult

type ProcessPipeline []ProcessPipelineFunc

func (p ProcessPipeline) Process() *ProcessResult {
	for _, fn := range p {
		if result := fn(); result != nil {
			return result
		}
	}
	return nil
}
