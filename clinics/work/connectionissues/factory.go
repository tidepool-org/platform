// Package connectionissues provides a singleton work unit that periodically asks the
// clinic service to recompute the connection issues of its patients.
package connectionissues

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	Type              = "org.tidepool.clinic.patient.connection.issues.update"
	Quantity          = 1
	Frequency         = 1 * time.Minute
	ProcessingTimeout = 5 * time.Minute
)

// ClinicClient is the part of the clinic client used by the processor.
type ClinicClient interface {
	UpdateConnectionIssues(ctx context.Context) error
}

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

// NewWorkCreate describes the singleton work item. Creating it again is a no-op, so it
// can be created on every service start.
func NewWorkCreate() (*work.Create, error) {
	return &work.Create{
		Type:              Type,
		DeduplicationID:   pointer.From(work.DeduplicationIDSingleton),
		ProcessingTimeout: int(ProcessingTimeout.Seconds()),
	}, nil
}
