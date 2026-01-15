package work

import (
	"context"
	"time"
)

// Allows a processor to update the work in the database while processing that work. Returns the resulting
// updated work or an error.
//
//go:generate mockgen -source=processor.go -destination=test/processor_mocks.go -package=test ProcessingUpdater
type ProcessingUpdater interface {

	// Update the work in the database while processing. Returns the resulting
	// updated work or an error.
	ProcessingUpdate(ctx context.Context, processingUpdate ProcessingUpdate) (*Work, error)
}

// Required interface for a processor of work.
//
//go:generate mockgen -source=processor.go -destination=test/processor_mocks.go -package=test Processor
type Processor interface {

	// Process the specified work within the specified context providing intermediate updates
	// with the specified processing updater. The specified context will be forcefully canceled after
	// the processing timeout specified by the work.
	Process(ctx context.Context, wrk *Work, processingUpdater ProcessingUpdater) *ProcessResult
}

// Required interface for a processor factory.
//
//go:generate mockgen -source=processor.go -destination=test/processor_mocks.go -package=test ProcessorFactory
type ProcessorFactory interface {

	// The type of work supported by the processor this factory creates. Must be in the form of a reverse DNS.
	Type() string

	// The quantity of work supported by the processor this factory creates. Must be greater than zero.
	Quantity() int

	// The minimum frequency to check for new work for the processor this factory creates. Must be greater than zero.
	Frequency() time.Duration

	// Create a new processor to handle a work.
	New() (Processor, error)
}
