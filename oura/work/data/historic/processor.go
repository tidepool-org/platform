package historic

import (
	"context"
	"time"

	dataSource "github.com/tidepool-org/platform/data/source"
	dataSourceWork "github.com/tidepool-org/platform/data/source/work"
	dataWork "github.com/tidepool-org/platform/data/work"
	"github.com/tidepool-org/platform/errors"
	ouraWork "github.com/tidepool-org/platform/oura/work"
	ouraWorkData "github.com/tidepool-org/platform/oura/work/data"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	Type      = "org.tidepool.oura.work.data.historic"
	Quantity  = 4
	Frequency = 5 * time.Second

	FailingRetryDuration       = 1 * time.Minute
	FailingRetryDurationJitter = 5 * time.Second

	ProcessingTimeout = 15 * 60 // Seconds
)

type Dependencies struct {
	DataDependencies dataWork.Dependencies
	Client           ouraWork.Client
}

func (d Dependencies) Validate() error {
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

type DataMixin = dataWork.Mixin

type Processor struct {
	*workBase.Processor
	*DataMixin
	Client    ouraWork.Client
	timeRange *dataWork.TimeRange
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

	processor, err := workBase.NewProcessor(processResultBuilder)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create processor")
	}

	dataMixin, err := dataWork.NewMixin(processor, dependencies.DataDependencies)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create data mixin")
	}

	return &Processor{
		Processor: processor,
		DataMixin: dataMixin,
	}, nil
}

func (p *Processor) Process(ctx context.Context, wrk *work.Work, processingUpdater work.ProcessingUpdater) *work.ProcessResult {
	return work.ProcessPipeline{
		p.ProcessPipelineFunc(ctx, wrk, processingUpdater),
		p.prepareTimeRange,
		// TODO: Implement
		p.Delete,
	}.Process()
}

func (p *Processor) prepareTimeRange() *work.ProcessResult {
	timeRange, err := p.TimeRangeFromMetadata()
	if err != nil {
		return p.Failed(errors.Wrap(err, "unable to parse time range from metadata"))
	} else if timeRange == nil {
		timeRange = &dataWork.TimeRange{}
	}

	p.AddFieldToContext("timeRangeInitial", *timeRange)

	to := time.Now()
	if timeRange.To == nil || timeRange.To.After(to) {
		timeRange.To = pointer.FromTime(to)
	}

	*timeRange = timeRange.Truncate(time.Millisecond)

	p.AddFieldToContext("timeRangeFinal", *timeRange)

	p.timeRange = timeRange
	return nil
}

func NewWorkCreate(dataSrc *dataSource.Source, timeRange dataWork.TimeRange) (*work.Create, error) {
	if dataSrc == nil {
		return nil, errors.New("data source is missing")
	}

	dataSrcID := *dataSrc.ID
	return &work.Create{
		Type:              Type,
		GroupID:           pointer.FromString(ouraWorkData.GroupIDFromDataSourceID(dataSrcID)),
		DeduplicationID:   pointer.FromString(dataSrcID),
		SerialID:          pointer.FromString(ouraWorkData.SerialIDFromDataSourceID(dataSrcID)),
		ProcessingTimeout: ProcessingTimeout,
		Metadata: map[string]any{
			dataSourceWork.MetadataKeyID:  dataSrcID,
			dataWork.MetadataKeyTimeRange: timeRange.Truncate(time.Millisecond),
		},
	}, nil
}
