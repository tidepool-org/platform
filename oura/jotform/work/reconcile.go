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
	reconcilerWorkID  = "reconciler"

	JotformReconcileMetadataLastProcessedSubmissionID = "lastProcessedSubmissionID"
)

type Metadata struct {
	LastProcessedSubmissionID string
}

func (m *Metadata) Parse(parser structure.ObjectParser) {
	parser.String(JotformReconcileMetadataLastProcessedSubmissionID)
}

func (m *Metadata) Validate(validator structure.Validator) {
	validator.String(JotformReconcileMetadataLastProcessedSubmissionID, &m.LastProcessedSubmissionID).NotEmpty()
}

func EnsureReconcilerWorkItemExists(ctx context.Context, client work.Client) error {
	create := &work.Create{
		Type:                    processorType,
		DeduplicationID:         pointer.FromString(reconcilerWorkID),
		ProcessingTimeout:       int(processingTimeout.Seconds()),
		ProcessingAvailableTime: time.Now(),
		Metadata: map[string]any{
			JotformReconcileMetadataLastProcessedSubmissionID: "0",
		},
	}
	if _, err := client.Create(ctx, create); err != nil {
		return err
	}
	return nil
}

type Processor struct {
	logger              log.Logger
	submissionProcessor *jotform.SubmissionProcessor
}

func NewProcessor(submissionProcessor *jotform.SubmissionProcessor, logger log.Logger) *Processor {
	return &Processor{
		logger:              logger,
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

	result, err := p.submissionProcessor.Reconcile(ctx, metadata.LastProcessedSubmissionID)
	logger := p.logger.WithFields(log.Fields{
		"processed": result.TotalProcessed,
		"errors":    result.TotalErrors,
	})

	updatedMetadata := map[string]any{
		JotformReconcileMetadataLastProcessedSubmissionID: metadata.LastProcessedSubmissionID,
	}

	if result.LastProcessedID != "" {
		updatedMetadata[JotformReconcileMetadataLastProcessedSubmissionID] = result.LastProcessedID
	}

	if err != nil {
		logger.WithError(err).Error("unable to reconcile submissions")

		return *work.NewProcessResultFailing(work.FailingUpdate{
			FailingError: *errors.NewSerializable(err),
			Metadata:     updatedMetadata,
		})
	}

	logger.Info("reconciled submissions")

	return *work.NewProcessResultSuccess(work.SuccessUpdate{
		Metadata: updatedMetadata,
	})
}
