package work

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/oura/jotform"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	"github.com/tidepool-org/platform/work"
)

const (
	processorType     = "org.tidepool.processors.oura.jotform.reconcile"
	quantity          = 1
	frequency         = 60 * time.Minute
	processingTimeout = 3 * time.Minute

	JotformReconcileMetadataFormIDKey                 = "formId"
	JotformReconcileMetadataLastProcessedSubmissionID = "lastProcessedSubmissionID"
)

type Metadata struct {
	FormID                    string
	LastProcessedSubmissionID string
}

func (m *Metadata) Parse(parser structure.ObjectParser) {
	if formId := parser.String(JotformReconcileMetadataFormIDKey); formId != nil {
		m.FormID = *formId
	}
	parser.String(JotformReconcileMetadataLastProcessedSubmissionID)
}

func (m *Metadata) Validate(validator structure.Validator) {
	validator.String(JotformReconcileMetadataFormIDKey, &m.FormID).NotEmpty()
	validator.String(JotformReconcileMetadataLastProcessedSubmissionID, &m.LastProcessedSubmissionID).NotEmpty()
}

func CreateReconcilerWorkItemIfNotExists(ctx context.Context, client work.Client, formID string) error {
	create := &work.Create{
		Type:                    processorType,
		DeduplicationID:         pointer.FromAny(formID),
		ProcessingTimeout:       int(processingTimeout.Seconds()),
		ProcessingAvailableTime: time.Now(),
		Metadata: map[string]any{
			JotformReconcileMetadataFormIDKey:                 formID,
			JotformReconcileMetadataLastProcessedSubmissionID: "0",
		},
	}
	if _, err := client.Create(ctx, create); err != nil {
		return err
	}
	return nil
}

type Processor struct {
	submissionProcessor *jotform.SubmissionProcessor
	logger              log.Logger
}

func NewProcessor(submissionProcessor *jotform.SubmissionProcessor) *Processor {
	return &Processor{
		submissionProcessor: submissionProcessor,
	}
}

func (p *Processor) Type() string {
	return processorType
}

func (p *Processor) Quantity() int {
	return quantity
}

func (p *Processor) Frequency() time.Duration {
	return frequency
}

func (p *Processor) Process(ctx context.Context, wrk *work.Work, updater work.ProcessingUpdater) work.ProcessResult {
	parser := structureParser.NewObject(p.logger, &wrk.Metadata)
	metadata := &Metadata{}
	metadata.Parse(parser)

	if metadata.FormID == "" {
		return *work.NewProcessResultFailed(work.FailedUpdate{
			FailedError: *errors.NewSerializable(errors.New("form id is missing")),
		})
	}

	id, err := p.submissionProcessor.Reconcile(ctx, metadata.FormID, metadata.LastProcessedSubmissionID)
	if err != nil {
		return *work.NewProcessResultFailing(work.FailingUpdate{
			FailingError: *errors.NewSerializable(err),
		})
	}

	return *work.NewProcessResultSuccess(work.SuccessUpdate{
		Metadata: map[string]any{
			JotformReconcileMetadataFormIDKey:                 metadata.FormID,
			JotformReconcileMetadataLastProcessedSubmissionID: id,
		},
	})
}
