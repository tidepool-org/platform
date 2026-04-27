package base

import (
	"context"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/metadata"
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

type ProcessorWithoutMetadata = Processor[map[string]any]

func NewProcessorWithoutMetadata(dependencies Dependencies, processResultBuilder work.ProcessResultBuilder) (*ProcessorWithoutMetadata, error) {
	return NewProcessor[map[string]any](dependencies, processResultBuilder)
}

type Processor[W any] struct {
	processResultBuilder work.ProcessResultBuilder
	workClient           work.Client
	context              context.Context
	work                 *work.Work
	processingUpdater    work.ProcessingUpdater
	metadata             *W
}

func NewProcessor[W any](dependencies Dependencies, processResultBuilder work.ProcessResultBuilder) (*Processor[W], error) {
	if err := dependencies.Validate(); err != nil {
		return nil, errors.Wrap(err, "dependencies is invalid")
	}
	if processResultBuilder == nil {
		return nil, errors.New("process result builder is missing")
	}
	return &Processor[W]{
		processResultBuilder: processResultBuilder,
		workClient:           dependencies.WorkClient,
		metadata:             new(W),
	}, nil
}

func (p *Processor[W]) process(ctx context.Context, wrk *work.Work, processingUpdater work.ProcessingUpdater) *work.ProcessResult {
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

	return p.decodeMetadata()
}

func (p *Processor[W]) ProcessPipeline(ctx context.Context, wrk *work.Work, processingUpdater work.ProcessingUpdater) work.ProcessPipeline {
	return work.ProcessPipeline{func() *work.ProcessResult { return p.process(ctx, wrk, processingUpdater) }}
}

func (p *Processor[W]) WorkClient() work.Client {
	return p.workClient
}

func (p *Processor[W]) Context() context.Context {
	return p.context
}

func (p *Processor[W]) AddFieldToContext(key string, value any) {
	p.context = log.ContextWithField(p.context, key, value)
}

func (p *Processor[W]) AddFieldsToContext(fields log.Fields) {
	p.context = log.ContextWithFields(p.context, fields)
}

func (p *Processor[W]) ProcessingUpdate() *work.ProcessResult {
	log.LoggerFromContext(p.context).Debug("update work")

	if result := p.encodeMetadata(); result != nil {
		return result
	}

	wrk, err := p.processingUpdater.ProcessingUpdate(context.WithoutCancel(p.context), work.ProcessingUpdate{Metadata: p.work.Metadata})
	if err != nil {
		return p.Failing(errors.Wrap(err, "unable to update work"))
	} else if wrk == nil {
		return p.Failed(errors.New("work is missing"))
	}
	p.work = wrk

	p.AddFieldToContext("work", p.work)

	return p.decodeMetadata()
}

func (p *Processor[W]) Metadata() *W {
	return p.metadata
}

func (p *Processor[W]) Pending() *work.ProcessResult {
	if result := p.encodeMetadata(); result != nil {
		return result
	}
	return p.processResultBuilder.Pending(p.context, p.work)
}

func (p *Processor[W]) Failing(err error) *work.ProcessResult {
	if result := p.encodeMetadata(); result != nil {
		return result
	}
	return p.processResultBuilder.Failing(p.context, p.work, err)
}

func (p *Processor[W]) Failed(err error) *work.ProcessResult {
	if result := p.encodeMetadata(); result != nil {
		return result
	}
	return p.processResultBuilder.Failed(p.context, p.work, err)
}

func (p *Processor[W]) Success() *work.ProcessResult {
	if result := p.encodeMetadata(); result != nil {
		return result
	}
	return p.processResultBuilder.Success(p.context, p.work)
}

func (p *Processor[W]) Delete() *work.ProcessResult {
	if result := p.encodeMetadata(); result != nil {
		return result
	}
	return p.processResultBuilder.Delete(p.context, p.work)
}

func (p *Processor[W]) decodeMetadata() *work.ProcessResult {
	if workMetadata, err := metadata.Decode[W](p.context, p.work.Metadata); err != nil {
		return p.processResultBuilder.Failed(p.context, p.work, err) // Do not encode metadata if decoding fails (otherwise we potentially corrupt metadata)
	} else if workMetadata != nil {
		*p.metadata = *workMetadata
	}
	return nil
}

func (p *Processor[W]) encodeMetadata() *work.ProcessResult {
	if workMetadata, err := metadata.Encode(p.metadata); err != nil {
		return p.processResultBuilder.Failed(p.context, p.work, err)
	} else if workMetadata != nil {
		p.work.Metadata = workMetadata
	}
	return nil
}
