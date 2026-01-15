package base

import (
	"context"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/structure"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	"github.com/tidepool-org/platform/work"
)

type Processor struct {
	processResultBuilder work.ProcessResultBuilder
	context              context.Context
	work                 *work.Work
	processingUpdater    work.ProcessingUpdater
}

func NewProcessor(processResultBuilder work.ProcessResultBuilder) (*Processor, error) {
	if processResultBuilder == nil {
		return nil, errors.New("process result builder is missing")
	}
	return &Processor{
		processResultBuilder: processResultBuilder,
	}, nil
}

func (p *Processor) Process(ctx context.Context, wrk *work.Work, processingUpdater work.ProcessingUpdater) *work.ProcessResult {
	if ctx == nil {
		return NewProcessResultFailedFromError(errors.New("context is missing"))
	}
	if wrk == nil {
		return NewProcessResultFailedFromError(errors.New("work is missing"))
	}
	if processingUpdater == nil {
		return NewProcessResultFailedFromError(errors.New("processing updater is missing"))
	}

	p.context = ctx
	p.work = wrk
	p.processingUpdater = processingUpdater

	p.AddFieldToContext("work", p.Work())

	return nil
}

func (p *Processor) ProcessPipelineFunc(ctx context.Context, wrk *work.Work, processingUpdater work.ProcessingUpdater) work.ProcessPipelineFunc {
	return func() *work.ProcessResult {
		return p.Process(ctx, wrk, processingUpdater)
	}
}

func (p *Processor) Context() context.Context {
	return p.context
}

func (p *Processor) AddFieldToContext(key string, value any) {
	p.context = log.ContextWithField(p.Context(), key, value)
}

func (p *Processor) AddFieldsToContext(fields log.Fields) {
	p.context = log.ContextWithFields(p.Context(), fields)
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
	p.Logger().Debug("update work")

	wrk, err := p.processingUpdater.ProcessingUpdate(context.WithoutCancel(p.Context()), work.ProcessingUpdate{Metadata: p.Metadata()})
	if err != nil {
		return p.Failing(errors.New("unable to update work"))
	} else if wrk == nil {
		return p.Failed(errors.New("work is missing"))
	}
	p.work = wrk

	p.AddFieldToContext("work", p.Work())

	return nil
}

func (p *Processor) Pending() *work.ProcessResult {
	return p.processResultBuilder.Pending(p.Context(), p.Work())
}

func (p *Processor) Failing(err error) *work.ProcessResult {
	return p.processResultBuilder.Failing(p.Context(), p.Work(), err)
}

func (p *Processor) Failed(err error) *work.ProcessResult {
	return p.processResultBuilder.Failed(p.Context(), p.Work(), err)
}

func (p *Processor) Success() *work.ProcessResult {
	return p.processResultBuilder.Success(p.Context(), p.Work())
}

func (p *Processor) Delete() *work.ProcessResult {
	return p.processResultBuilder.Delete(p.Context(), p.Work())
}

func NewProcessResultFailedFromError(err error) *work.ProcessResult {
	return work.NewProcessResultFailed(work.FailedUpdate{
		FailedError: errors.Serializable{Error: err},
	})
}
