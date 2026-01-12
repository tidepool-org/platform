package work

import (
	"context"
	"time"
)

// Allows processor to update the work in the database. Returns the resulting
// updated work or an error.
//
//go:generate mockgen -source=processor.go -destination=test/processor_mocks.go -package=test ProcessingUpdater
type ProcessingUpdater interface {

	// Update the work in the database while processing. Returns the resulting
	// updated work or an error.
	ProcessingUpdate(ctx context.Context, processingUpdate ProcessingUpdate) (*Work, error)
}

// Required interface for a processing work.
//
//go:generate mockgen -source=processor.go -destination=test/processor_mocks.go -package=test Processor
type Processing interface {

	// Process the specified work within the specified context providing intermediate updates
	// with the specified updater. The specified context will be forcefully canceled after
	// the processing timeout specified by the work.
	Process(ctx context.Context, wrk *Work, updater ProcessingUpdater) *ProcessResult
}

// Required interface for a processor of work.
//
//go:generate mockgen -source=processor.go -destination=test/processor_mocks.go -package=test Processor
type Processor interface {
	Processing

	// The type of work supported by this processor. Must be in the form of a reverse DNS.
	Type() string

	// The quantity of work supported by this processor. Must be greater than zero.
	Quantity() int

	// The minimum frequency to check for new work for this processor. Must be greater than zero.
	Frequency() time.Duration
}
