package postprocess

import (
	"time"

	"github.com/tidepool-org/platform/clinics"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/user"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	Quantity  = 15
	Frequency = 30 * time.Second

	FailingRetryDuration        = 1 * time.Minute
	FailingRetryDurationJitter  = 5 * time.Second
	FailingRetryDurationMaximum = 1 * time.Hour
)

type (
	ClinicsClient = clinics.Client
	UserClient    = user.Client
)

type Dependencies struct {
	workBase.Dependencies
	Summarizers
	ClinicsClient
	UserClient
}

func (d Dependencies) Validate() error {
	if err := d.Dependencies.Validate(); err != nil {
		return err
	}
	if d.Summarizers == nil {
		return errors.New("summarizers is missing")
	}
	if d.ClinicsClient == nil {
		return errors.New("clinics client is missing")
	}
	if d.UserClient == nil {
		return errors.New("user client is missing")
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
