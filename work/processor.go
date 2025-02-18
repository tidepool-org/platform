package work

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/metadata"
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

// Allows processor to update the work in the database. Returns the resulting
// updated work or an error.
type ProcessingUpdater interface {
	ProcessingUpdate(ctx context.Context, processingUpdate ProcessingUpdate) (*Work, error)
}

// Required interface for a processor of work.
type Processor interface {

	// The type of work supported by this processor. Must be in the form of a reverse DNS.
	Type() string

	// The quantity of work supported by this processor. Must be greater than zero.
	Quantity() int

	// The minimum frequency to check for new work for this processor. Must be greater than zero.
	Frequency() time.Duration

	// Process the specified work within the specified context providing intermediate updates
	// with the specified updater. The specified context will be forcefully canceled after
	// the processing timeout specified by the work.
	Process(ctx context.Context, wrk *Work, updater ProcessingUpdater) ProcessResult
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

func (p *ProcessResult) Metadata() *metadata.Metadata {
	if p.PendingUpdate != nil && p.PendingUpdate.Metadata != nil {
		return p.PendingUpdate.Metadata
	}
	if p.FailingUpdate != nil && p.FailingUpdate.Metadata != nil {
		return p.FailingUpdate.Metadata
	}
	if p.FailedUpdate != nil && p.FailedUpdate.Metadata != nil {
		return p.FailedUpdate.Metadata
	}
	if p.SuccessUpdate != nil && p.SuccessUpdate.Metadata != nil {
		return p.SuccessUpdate.Metadata
	}
	return nil
}
