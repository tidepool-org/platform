package event

import (
	"context"
	"fmt"
	"time"

	"github.com/tidepool-org/platform/customerio"
	dataSource "github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
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

	ProcessingTimeout = 1 * time.Minute
)

const (
	MetadataKeyUserID                 = "userId"
	MetadataKeyEventType              = "eventType"
	MetadataKeyEventData              = "eventData"
	MetadataKeyEventDeduplicationTime = "eventDeduplicationTime"
	MetadataKeyEventDeduplicationID   = "eventDeduplicationId"
)

type Metadata struct {
	UserID                 *string    `json:"userId,omitempty"`
	EventType              *string    `json:"eventType,omitempty"`
	EventData              any        `json:"eventData,omitempty"`
	EventDeduplicationTime *time.Time `json:"eventDeduplicationTime,omitempty"`
	EventDeduplicationID   *string    `json:"eventDeduplicationId,omitempty"`
}

func (m *Metadata) Parse(parser structure.ObjectParser) {
	m.UserID = parser.String(MetadataKeyUserID)
	m.EventType = parser.String(MetadataKeyEventType)
	m.EventData = parser.Object(MetadataKeyEventData)
	m.EventDeduplicationTime = parser.Time(MetadataKeyEventDeduplicationTime, time.RFC3339Nano)
	m.EventDeduplicationID = parser.String(MetadataKeyEventDeduplicationID)
}

func (m *Metadata) Validate(validator structure.Validator) {
	validator.String(MetadataKeyUserID, m.UserID).Exists().NotEmpty()
	validator.String(MetadataKeyEventType, m.EventType).Exists().NotEmpty()
	validator.Time(MetadataKeyEventDeduplicationTime, m.EventDeduplicationTime).Exists().NotZero()
	validator.String(MetadataKeyEventDeduplicationID, m.EventDeduplicationID).Exists().NotEmpty()
}

type Dependencies struct {
	workBase.Dependencies
	CustomerIOClient *customerio.Client
}

func (d Dependencies) Validate() error {
	if err := d.Dependencies.Validate(); err != nil {
		return err
	}
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
	*workBase.Processor[Metadata]
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

	processor, err := workBase.NewProcessor[Metadata](dependencies.Dependencies, processResultBuilder)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create processor")
	}

	return &Processor{
		Processor: processor,
		Client:    dependencies.CustomerIOClient,
	}, nil
}

func (p *Processor) Process(ctx context.Context, wrk *work.Work, processingUpdater work.ProcessingUpdater) *work.ProcessResult {
	return append(p.ProcessPipeline(ctx, wrk, processingUpdater),
		p.sendEvent,
	).Process(p.Delete)
}

func (p *Processor) sendEvent() *work.ProcessResult {
	metadata := p.Metadata()
	if metadata == nil {
		return p.Failed(errors.New("metadata is missing"))
	} else if err := structureValidator.New(log.LoggerFromContext(p.Context())).Validate(metadata); err != nil {
		return p.Failed(errors.Wrap(err, "metadata is invalid"))
	}

	event, err := p.eventFromMetadata(metadata)
	if err != nil {
		return p.Failed(errors.Wrap(err, "unable to create event from metadata"))
	}
	userID, err := p.userIDFromMetadata(metadata)
	if err != nil {
		return p.Failed(errors.Wrap(err, "unable to get user id from metadata"))
	}

	if err = p.Client.SendEvent(p.Context(), userID, event); err != nil {
		return p.Failing(errors.Wrap(err, "unable to send event"))
	}

	return nil
}

func (p *Processor) eventFromMetadata(metadata *Metadata) (*customerio.Event, error) {
	event := &customerio.Event{
		Name: *metadata.EventType,
		Data: metadata.EventData,
	}
	deduplicationTime := *metadata.EventDeduplicationTime
	deduplicationID := *metadata.EventDeduplicationID

	if err := event.SetDeduplicationID(deduplicationTime, deduplicationID); err != nil {
		return nil, errors.Wrap(err, "unable to set event deduplication id")
	}
	if err := structureValidator.New(log.LoggerFromContext(p.Context())).Validate(event); err != nil {
		return nil, errors.Wrap(err, "event is invalid")
	}

	return event, nil
}

func (p *Processor) userIDFromMetadata(metadata *Metadata) (string, error) {
	userID := *metadata.UserID
	if userID == "" {
		return "", errors.New("userID is missing")
	}

	return userID, nil
}

func NewDataSourceStateChangedEventWorkCreate(dataSrc *dataSource.Source) (*work.Create, error) {
	if dataSrc == nil {
		return nil, errors.New("data source is missing")
	}

	workMetadata := &Metadata{}
	workMetadata.UserID = pointer.FromString(dataSrc.UserID)
	workMetadata.EventType = pointer.FromString(customerio.DataSourceStateChangedEventType)
	workMetadata.EventData = &customerio.DataSourceStateChangedEvent{
		ProviderName: dataSrc.ProviderName,
		State:        dataSrc.State,
	}
	workMetadata.EventDeduplicationTime = pointer.FromTime(DeduplicationTimeFromDataSource(*dataSrc))
	workMetadata.EventDeduplicationID = pointer.FromString(EventDeduplicationIDFromDataSource(customerio.DataSourceStateChangedEventType, *dataSrc))

	encodedWorkMetadata, err := metadata.Encode(workMetadata)
	if err != nil {
		return nil, errors.Wrap(err, "unable to encode work metadata")
	}

	return &work.Create{
		Type:              Type,
		DeduplicationID:   pointer.FromString(WorkDeduplicationIDFromDataSource(customerio.DataSourceStateChangedEventType, *dataSrc)),
		ProcessingTimeout: int(ProcessingTimeout.Seconds()),
		Metadata:          encodedWorkMetadata,
	}, nil
}

func WorkDeduplicationIDFromDataSource(eventType string, dataSrc dataSource.Source) string {
	return fmt.Sprintf("%s:%s:%s:%s", eventType, dataSrc.UserID, dataSrc.ID, DeduplicationTimeFromDataSource(dataSrc).Format(time.RFC3339))
}

func DeduplicationTimeFromDataSource(dataSrc dataSource.Source) time.Time {
	return pointer.Default(dataSrc.ModifiedTime, dataSrc.CreatedTime)
}

func EventDeduplicationIDFromDataSource(eventType string, dataSrc dataSource.Source) string {
	return fmt.Sprintf("%s:%s:%s", eventType, dataSrc.UserID, dataSrc.ID)
}
