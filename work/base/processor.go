package base

import (
	"context"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/structure"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	"github.com/tidepool-org/platform/work"
)

type Dependencies struct {
	WorkClient work.Client
}

func (d Dependencies) Validate() error {
	if d.WorkClient == nil {
		return errors.New("work client is missing")
	}
	return nil
}

type Processor struct {
	processResultBuilder work.ProcessResultBuilder
	workClient           work.Client
	context              context.Context
	work                 *work.Work
	processingUpdater    work.ProcessingUpdater
}

func NewProcessor(dependencies Dependencies, processResultBuilder work.ProcessResultBuilder) (*Processor, error) {
	if err := dependencies.Validate(); err != nil {
		return nil, errors.Wrap(err, "dependencies is invalid")
	}
	if processResultBuilder == nil {
		return nil, errors.New("process result builder is missing")
	}
	return &Processor{
		processResultBuilder: processResultBuilder,
		workClient:           dependencies.WorkClient,
	}, nil
}

func (p *Processor) process(ctx context.Context, wrk *work.Work, processingUpdater work.ProcessingUpdater) *work.ProcessResult {
	if ctx == nil {
		return work.ProcessResultFailedFromError(errors.New("context is missing"))
	}
	if wrk == nil {
		return work.ProcessResultFailedFromError(errors.New("work is missing"))
	}
	if processingUpdater == nil {
		return work.ProcessResultFailedFromError(errors.New("processing updater is missing"))
	}

	p.context = ctx
	p.work = wrk
	p.processingUpdater = processingUpdater

	p.AddFieldToContext("work", p.work)

	return nil
}

func (p *Processor) ProcessPipeline(ctx context.Context, wrk *work.Work, processingUpdater work.ProcessingUpdater) work.ProcessPipeline {
	return work.ProcessPipeline{func() *work.ProcessResult { return p.process(ctx, wrk, processingUpdater) }}
}

func (p *Processor) WorkClient() work.Client {
	return p.workClient
}

func (p *Processor) Context() context.Context {
	return p.context
}

func (p *Processor) AddFieldToContext(key string, value any) {
	p.context = log.ContextWithField(p.context, key, value)
}

func (p *Processor) AddFieldsToContext(fields log.Fields) {
	p.context = log.ContextWithFields(p.context, fields)
}

func (p *Processor) Logger() log.Logger {
	return log.LoggerFromContext(p.Context())
}

func (p *Processor) Work() *work.Work {
	return p.work
}

func (p *Processor) Metadata() map[string]any {
	return p.Work().Metadata
}

func (p *Processor) MetadataParser() structure.ObjectParser {
	var parsableMetadata *map[string]any
	if metadata := p.Metadata(); metadata != nil {
		parsableMetadata = &metadata
	}
	return structureParser.NewObject(p.Logger(), parsableMetadata)
}

func (p *Processor) ProcessingUpdate() *work.ProcessResult {
	log.LoggerFromContext(p.context).Debug("update work")

	wrk, err := p.processingUpdater.ProcessingUpdate(context.WithoutCancel(p.context), work.ProcessingUpdate{Metadata: p.work.Metadata})
	if err != nil {
		return p.Failing(errors.Wrap(err, "unable to update work"))
	} else if wrk == nil {
		return p.Failed(errors.New("work is missing"))
	}
	p.work = wrk

	p.AddFieldToContext("work", p.work)

	return nil
}

func (p *Processor) Pending() *work.ProcessResult {
	return p.processResultBuilder.Pending(p.context, p.work)
}

func (p *Processor) Failing(err error) *work.ProcessResult {
	return p.processResultBuilder.Failing(p.context, p.work, err)
}

func (p *Processor) Failed(err error) *work.ProcessResult {
	return p.processResultBuilder.Failed(p.context, p.work, err)
}

func (p *Processor) Success() *work.ProcessResult {
	return p.processResultBuilder.Success(p.context, p.work)
}

func (p *Processor) Delete() *work.ProcessResult {
	return p.processResultBuilder.Delete(p.context, p.work)
}
