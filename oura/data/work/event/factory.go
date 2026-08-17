package event

import (
	"fmt"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/oura"
	ouraDataWork "github.com/tidepool-org/platform/oura/data/work"
	ouraWork "github.com/tidepool-org/platform/oura/work"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	Type              = "org.tidepool.oura.data.event"
	Quantity          = 4
	Frequency         = 5 * time.Second
	ProcessingTimeout = 3 * time.Minute
)

func NewProcessorFactory(dependencies ouraDataWork.Dependencies) (*workBase.ProcessorFactory, error) {
	if err := dependencies.Validate(); err != nil {
		return nil, errors.Wrap(err, "dependencies is invalid")
	}
	processorFactory := func() (work.Processor, error) { return NewProcessor(dependencies) }
	return workBase.NewProcessorFactory(Type, Quantity, Frequency, processorFactory)
}

func NewWorkCreate(providerSessionID string, event *oura.Event) (*work.Create, error) {
	if providerSessionID == "" {
		return nil, errors.New("provider session id is missing")
	}
	if event == nil {
		return nil, errors.New("event is missing")
	}

	hash, err := event.Hash()
	if err != nil {
		return nil, errors.Wrap(err, "unable to hash event")
	}

	return metadata.WithMetadata(
		&work.Create{
			Type:              Type,
			GroupID:           pointer.From(ouraWork.GroupIDFromProviderSessionID(providerSessionID)),
			DeduplicationID:   pointer.From(fmt.Sprintf("%s:%s", providerSessionID, hash)),
			SerialID:          pointer.From(ouraWork.SerialIDFromProviderSessionID(providerSessionID)),
			ProcessingTimeout: int(ProcessingTimeout.Seconds()),
		},
		&Metadata{
			ProviderSessionMetadata: ProviderSessionMetadata{
				ProviderSessionID: pointer.From(providerSessionID),
			},
			EventMetadata: EventMetadata{
				Event: event,
			},
		},
	)
}
