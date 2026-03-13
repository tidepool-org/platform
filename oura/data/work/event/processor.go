package event

import (
	"context"
	"time"

	dataSource "github.com/tidepool-org/platform/data/source"
	dataSourceWork "github.com/tidepool-org/platform/data/source/work"
	dataWork "github.com/tidepool-org/platform/data/work"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/oura"
	ouraDataWork "github.com/tidepool-org/platform/oura/data/work"
	ouraWebhook "github.com/tidepool-org/platform/oura/webhook"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	Type      = "org.tidepool.oura.work.data.event"
	Quantity  = 4
	Frequency = 5 * time.Second

	FailingRetryDuration       = 1 * time.Minute
	FailingRetryDurationJitter = 5 * time.Second

	ProcessingTimeout = 1 * 60 // Seconds

	MetadataKeyEvent = "event"
)

type Dependencies struct {
	workBase.Dependencies
	DataDependencies dataWork.Dependencies
	Client           oura.Client
}

func (d Dependencies) Validate() error {
	if err := d.Dependencies.Validate(); err != nil {
		return errors.Wrap(err, "dependencies is invalid")
	}
	if err := d.DataDependencies.Validate(); err != nil {
		return err
	}
	if d.Client == nil {
		return errors.New("client is missing")
	}
	return nil
}

func NewProcessorFactory(dependencies Dependencies) (*workBase.ProcessorFactory, error) {
	if err := dependencies.Validate(); err != nil {
		return nil, errors.Wrap(err, "dependencies is invalid")
	}
	processorFactory := func() (work.Processor, error) { return NewProcessor(dependencies) }
	return workBase.NewProcessorFactory(Type, Quantity, Frequency, processorFactory)
}

type Processor struct {
	*workBase.Processor
	*dataWork.Mixin
	Client oura.Client
}

func NewProcessor(dependencies Dependencies) (*Processor, error) {
	if err := dependencies.Validate(); err != nil {
		return nil, errors.Wrap(err, "dependencies is invalid")
	}

	processResultBuilder := &workBase.ProcessResultBuilder{
		ProcessResultFailingBuilder: &workBase.ExponentialProcessResultFailingBuilder{
			Duration:       FailingRetryDuration,
			DurationJitter: FailingRetryDurationJitter,
		},
	}

	processor, err := workBase.NewProcessor(dependencies.Dependencies, processResultBuilder)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create processor")
	}

	mixin, err := dataWork.NewMixin(processor, dependencies.DataDependencies)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create mixin")
	}

	return &Processor{
		Processor: processor,
		Mixin:     mixin,
	}, nil
}

func (p *Processor) Process(ctx context.Context, wrk *work.Work, processingUpdater work.ProcessingUpdater) *work.ProcessResult {
	return append(p.ProcessPipeline(ctx, wrk, processingUpdater),
		func() *work.ProcessResult { return nil }, // TODO: Implement
	).Process(p.Delete)
}

func (o *Processor) EventFromMetadata() (*ouraWebhook.Event, error) {
	parser := o.MetadataParser()
	event := ouraWebhook.ParseEvent(parser.WithReferenceObjectParser(MetadataKeyEvent))
	if err := parser.Error(); err != nil {
		return nil, errors.Wrap(err, "unable to parse event from metadata")
	}
	return event, nil
}

func NewWorkCreate(dataSrc *dataSource.Source, event *ouraWebhook.Event) (*work.Create, error) {
	if dataSrc == nil {
		return nil, errors.New("data source is missing")
	}
	if event == nil {
		return nil, errors.New("event is missing")
	}
	dataSrcID := dataSrc.ID
	serialID := pointer.FromString(ouraDataWork.SerialIDFromDataSourceID(dataSrcID))
	return &work.Create{
		Type:              Type,
		GroupID:           serialID,
		DeduplicationID:   pointer.FromString(dataSrcID),
		SerialID:          serialID,
		ProcessingTimeout: ProcessingTimeout,
		Metadata: map[string]any{
			dataSourceWork.MetadataKeyID: dataSrcID,
			MetadataKeyEvent:             event,
		},
	}, nil
}
