package event

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tidepool-org/platform/customerio"
	dataSource "github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/pointer"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	Type      = "io.customer.event"
	Quantity  = 1
	Frequency = 5 * time.Second

	FailingRetryDuration       = 1 * time.Minute
	FailingRetryDurationJitter = 5 * time.Second

	ProcessingTimeout = 60 // Seconds
)

const (
	MetadataKeyUserID                 = "userId"
	MetadataKeyEventType              = "eventType"
	MetadataKeyEventData              = "eventData"
	MetadataKeyEventDeduplicationTime = "eventDeduplicationTime"
	MetadataKeyEventDeduplicationID   = "eventDeduplicationId"
)

type Dependencies struct {
	CustomerIOClient *customerio.Client
}

func (d Dependencies) Validate() error {
	if d.CustomerIOClient == nil {
		return errors.New("customer io client is missing")
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
	Client *customerio.Client
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

	return &Processor{
		Processor: processor,
		Client:    dependencies.CustomerIOClient,
	}, nil
}

func (p *Processor) Process(ctx context.Context, wrk *work.Work, processingUpdater work.ProcessingUpdater) *work.ProcessResult {
	return work.ProcessPipeline{
		p.ProcessPipelineFunc(ctx, wrk, processingUpdater),
		p.sendEvent,
		p.Delete,
	}.Process()
}

func (p *Processor) sendEvent() *work.ProcessResult {
	event, err := p.eventFromMetadata()
	if err != nil {
		return p.Failed(errors.Wrap(err, "unable to create event from metadata"))
	}
	userID, err := p.userIDFromMetadata()
	if err != nil {
		return p.Failed(errors.Wrap(err, "unable to get user id from metadata"))
	}

	if err = p.Client.SendEvent(p.Context(), userID, event); err != nil {
		return p.Failing(errors.Wrap(err, "unable to send event"))
	}

	return nil
}

func (p *Processor) eventFromMetadata() (*customerio.Event, error) {
	parser := p.MetadataParser()
	event := &customerio.Event{
		Name: pointer.Default(parser.String(MetadataKeyEventType), ""),
		Data: parser.Object(MetadataKeyEventData),
	}
	deduplicationTime := pointer.Default(parser.Time(MetadataKeyEventDeduplicationTime, ""), time.Time{})
	deduplicationID := pointer.Default(parser.String(MetadataKeyEventDeduplicationID), "")

	if err := parser.Error(); err != nil {
		return nil, errors.Wrap(parser.Error(), "unable to parse metadata")
	}
	if err := event.SetDeduplicationID(deduplicationTime, deduplicationID); err != nil {
		return nil, errors.Wrap(err, "unable to set event deduplication id")
	}
	if err := structureValidator.New(p.Logger()).Validate(event); err != nil {
		return nil, errors.Wrap(err, "event is invalid")
	}

	return event, nil
}

func (p *Processor) userIDFromMetadata() (string, error) {
	parser := p.MetadataParser()
	userID := strings.TrimSpace(pointer.Default(parser.String(MetadataKeyUserID), ""))

	if err := parser.Error(); err != nil {
		return "", errors.Wrap(parser.Error(), "unable to parse metadata")
	}
	if userID == "" {
		return "", errors.New("userID is missing")
	}

	return userID, nil
}

func NewDataSourceStateChangedEventWorkCreate(dataSrc *dataSource.Source) (*work.Create, error) {
	if dataSrc == nil {
		return nil, errors.New("data source is missing")
	}
	return &work.Create{
		Type:              Type,
		DeduplicationID:   pointer.FromString(WorkDeduplicationIDFromDataSource(customerio.DataSourceStateChangedEventType, *dataSrc)),
		GroupID:           pointer.FromString(GroupIDFromDataSource(customerio.DataSourceStateChangedEventType, *dataSrc)),
		SerialID:          pointer.FromString(SerialIDFromDataSource(customerio.DataSourceStateChangedEventType, *dataSrc)),
		ProcessingTimeout: ProcessingTimeout,
		Metadata: map[string]any{
			MetadataKeyUserID:    *dataSrc.UserID,
			MetadataKeyEventType: customerio.DataSourceStateChangedEventType,
			MetadataKeyEventData: customerio.DataSourceStateChangedEvent{
				ProviderName: *dataSrc.ProviderName,
				State:        *dataSrc.State,
			},
			MetadataKeyEventDeduplicationTime: DeduplicationTimeFromDataSource(*dataSrc),
			MetadataKeyEventDeduplicationID:   EventDeduplicationIDFromDataSource(customerio.DataSourceStateChangedEventType, *dataSrc),
		},
	}, nil
}

func GroupIDFromDataSource(eventType string, dataSrc dataSource.Source) string {
	return fmt.Sprintf("%s:%s:%s", Type, eventType, *dataSrc.ID)
}

func SerialIDFromDataSource(eventType string, dataSrc dataSource.Source) string {
	return fmt.Sprintf("%s:%s:%s", Type, eventType, *dataSrc.ID)
}

func WorkDeduplicationIDFromDataSource(eventType string, dataSrc dataSource.Source) string {
	return fmt.Sprintf("%s:%s:%s:%s", eventType, *dataSrc.UserID, *dataSrc.ID, DeduplicationTimeFromDataSource(dataSrc).Format(time.RFC3339))
}

func DeduplicationTimeFromDataSource(dataSrc dataSource.Source) time.Time {
	return pointer.Default(dataSrc.ModifiedTime, *dataSrc.CreatedTime)
}

func EventDeduplicationIDFromDataSource(eventType string, dataSrc dataSource.Source) string {
	return fmt.Sprintf("%s:%s:%s", eventType, *dataSrc.UserID, *dataSrc.ID)
}
