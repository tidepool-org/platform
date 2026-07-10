package deviceissues

import (
	"time"

	"github.com/tidepool-org/platform/clinics"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	Type      = "org.tidepool.clinic.device.issues.update"
	Quantity  = 1
	Frequency = time.Minute // coordinator poll cadence, not the run interval

	PendingRetryDuration       = 15 * time.Minute // the every-15-minute run interval
	FailingRetryDuration       = time.Minute
	FailingRetryDurationJitter = 10 * time.Second
	ProcessingTimeout          = 5 * time.Minute
)

type ClinicClient = clinics.Client

type Dependencies struct {
	workBase.Dependencies
	ClinicClient
}

func (d Dependencies) Validate() error {
	if err := d.Dependencies.Validate(); err != nil {
		return err
	}
	if d.ClinicClient == nil {
		return errors.New("clinic client is missing")
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
		DeduplicationID:   pointer.FromString(work.DeduplicationIDSingleton),
		ProcessingTimeout: int(ProcessingTimeout.Seconds()),
	}, nil
}
