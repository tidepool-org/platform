package historic

import (
	"fmt"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/metadata"
	ouraDataWork "github.com/tidepool-org/platform/oura/data/work"
	ouraWork "github.com/tidepool-org/platform/oura/work"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/times"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	Type              = "org.tidepool.oura.data.historic"
	Quantity          = 4
	Frequency         = 5 * time.Second
	ProcessingTimeout = 15 * time.Minute
)

func NewProcessorFactory(dependencies ouraDataWork.Dependencies) (*workBase.ProcessorFactory, error) {
	if err := dependencies.Validate(); err != nil {
		return nil, errors.Wrap(err, "dependencies is invalid")
	}
	processorFactory := func() (work.Processor, error) { return NewProcessor(dependencies) }
	return workBase.NewProcessorFactory(Type, Quantity, Frequency, processorFactory)
}

func NewWorkCreate(providerSessionID string, timeRange *times.TimeRange) (*work.Create, error) {
	if providerSessionID == "" {
		return nil, errors.New("provider session id is missing")
	}
	if timeRange == nil {
		timeRange = &times.TimeRange{}
	}

	hash, err := timeRange.Hash()
	if err != nil {
		return nil, errors.Wrap(err, "unable to hash time range")
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
			TimeRangeMetadata: TimeRangeMetadata{
				TimeRange: timeRange,
			},
		},
	)
}
