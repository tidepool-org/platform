package user

import (
	"time"

	dataWorkSummary "github.com/tidepool-org/platform/data/work/summary"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	// Quantity is the number of users whose summaries each instance of the service
	// recalculates at once.
	Quantity  = 10
	Frequency = 5 * time.Second

	FailingRetryDuration        = 1 * time.Minute
	FailingRetryDurationJitter  = 0.1
	FailingRetryDurationMaximum = 1 * time.Hour
)

type ClinicsClient = dataWorkSummary.ClinicsClient

type Dependencies struct {
	workBase.Dependencies
	Summarizers
	ClinicsClient
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
	return nil
}

func NewProcessorFactory(dependencies Dependencies) (*workBase.ProcessorFactory, error) {
	if err := dependencies.Validate(); err != nil {
		return nil, errors.Wrap(err, "dependencies is invalid")
	}
	processorFactory := func() (work.Processor, error) { return NewProcessor(dependencies) }
	return workBase.NewProcessorFactory(Type, Quantity, Frequency, processorFactory)
}
