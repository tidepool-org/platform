package work

import (
	"context"
	"math/rand"
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
	processorType = "org.tidepool.processors.oura.jotform.reconcile"
	quantity      = 1
	frequency     = 30 * time.Minute

	processingTimeout = 3 * time.Minute
	reconcilerWorkID  = "reconciler"

	retryDurationJitter = 10 * time.Second
	retryDuration       = processingTimeout * 2

	initialSubmissionID                               = "0"
	JotformReconcileMetadataLastProcessedSubmissionID = "lastProcessedSubmissionId"
)

type Metadata struct {
	LastProcessedSubmissionID string
}

func (m *Metadata) Parse(parser structure.ObjectParser) {
	m.LastProcessedSubmissionID = pointer.Default(parser.String(JotformReconcileMetadataLastProcessedSubmissionID), initialSubmissionID)
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
			JotformReconcileMetadataLastProcessedSubmissionID: initialSubmissionID,
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
	return time.Minute
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

		failingRetryCount := pointer.Default(wrk.FailingRetryCount, 0) + 1

		return *work.NewProcessResultFailing(work.FailingUpdate{
			FailingError:      *errors.NewSerializable(err),
			FailingRetryCount: failingRetryCount,
			FailingRetryTime:  time.Now().Add(exponentialBackoff(failingRetryCount)),
			Metadata:          updatedMetadata,
		})
	}

	logger.Info("reconciled submissions")

	return *work.NewProcessResultPending(work.PendingUpdate{
		Metadata:                updatedMetadata,
		ProcessingAvailableTime: time.Now().Add(frequency),
		ProcessingTimeout:       int(processingTimeout.Seconds()),
	})
}

func exponentialBackoff(retryCount int) time.Duration {
	fallbackFactor := time.Duration(1 << (retryCount - 1))
	retryDurationJitter := int64(retryDurationJitter * fallbackFactor)
	return retryDuration*fallbackFactor + time.Duration(rand.Int63n(2*retryDurationJitter)-retryDurationJitter)
}
