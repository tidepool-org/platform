package base

import (
	"context"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/structure"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	"github.com/tidepool-org/platform/work"
)

type Processing struct {
	processResultBuilder work.ProcessResultBuilder
	context              context.Context
	work                 *work.Work
	processingUpdater    work.ProcessingUpdater
}

func NewProcessing(processResultBuilder work.ProcessResultBuilder) *Processing {
	return &Processing{
		processResultBuilder: processResultBuilder,
	}
}

func (p *Processing) Process(ctx context.Context, wrk *work.Work, processingUpdater work.ProcessingUpdater) func() *work.ProcessResult {
	return func() *work.ProcessResult {
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

		p.ContextWithField("work", p.Work())

		return nil
	}
}

func (p *Processing) Context() context.Context {
	return p.context
}

func (p *Processing) ContextWithField(key string, value any) {
	p.context = log.ContextWithField(p.Context(), key, value)
}

func (p *Processing) ContextWithFields(fields log.Fields) {
	p.context = log.ContextWithFields(p.Context(), fields)
}

func (p *Processing) Logger() log.Logger {
	return log.LoggerFromContext(p.Context())
}

func (p *Processing) Work() *work.Work {
	return p.work
}

func (p *Processing) Metadata() map[string]any {
	return p.Work().Metadata
}

func (p *Processing) MetadataParser() structure.ObjectParser {
	var parsableMetadata *map[string]any
	if metadata := p.Metadata(); metadata != nil {
		parsableMetadata = &metadata
	}
	return structureParser.NewObject(p.Logger(), parsableMetadata)
}

func (p *Processing) ProcessingUpdate() *work.ProcessResult {
	p.Logger().Debug("update work")

	wrk, err := p.processingUpdater.ProcessingUpdate(context.WithoutCancel(p.Context()), work.ProcessingUpdate{Metadata: p.Metadata()})
	if err != nil {
		return p.Failing(errors.New("unable to update work"))
	} else if wrk == nil {
		return p.Failed(errors.New("work is missing"))
	}
	p.work = wrk

	p.ContextWithField("work", p.Work())

	return nil
}

func (p *Processing) Pending() *work.ProcessResult {
	return p.processResultBuilder.Pending(p.Context(), p.Work())
}

func (p *Processing) Failing(err error) *work.ProcessResult {
	return p.processResultBuilder.Failing(p.Context(), p.Work(), err)
}

func (p *Processing) Failed(err error) *work.ProcessResult {
	return p.processResultBuilder.Failed(p.Context(), p.Work(), err)
}

func (p *Processing) Success() *work.ProcessResult {
	return p.processResultBuilder.Success(p.Context(), p.Work())
}

func (p *Processing) Delete() *work.ProcessResult {
	return p.processResultBuilder.Delete(p.Context(), p.Work())
}

func NewProcessResultFailedFromError(err error) *work.ProcessResult {
	return work.NewProcessResultFailed(work.FailedUpdate{
		FailedError: errors.Serializable{Error: err},
	})
}
