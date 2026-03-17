package subscribe

import (
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/oura"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	Type              = "org.tidepool.oura.webhook.subscribe"
	Quantity          = 1
	Frequency         = time.Minute
	ProcessingTimeout = 5 * time.Minute
)

type OuraClient = oura.Client

type Dependencies struct {
	workBase.Dependencies
	OuraClient
}

func (d Dependencies) Validate() error {
	if err := d.Dependencies.Validate(); err != nil {
		return err
	}
	if d.OuraClient == nil {
		return errors.New("oura client is missing")
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

func NewWorkCreate() (*work.Create, error) {
	return &work.Create{
		Type:              Type,
		GroupID:           pointer.FromString(Type),
		DeduplicationID:   pointer.FromString(Type),
		ProcessingTimeout: int(ProcessingTimeout.Seconds()),
	}, nil
}
