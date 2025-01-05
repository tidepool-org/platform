package work

import (
	"context"
	"time"
)

// Allows processor to update the work in the database. Returns the resulting
// updated work or an error.
type ProcessingUpdater interface {
	ProcessingUpdate(ctx context.Context, update *ProcessingUpdate) (*Work, error)
}

// Required interface for a processor of work.
type Processor interface {

	// The tyoe of work supported by this processor. Must be in the form of a reverse DNS.
	Type() string

	// The quantity of work supported by this processor. Must be greater than zero.
	Quantity() int

	// The minimum frequency to check for new work for this processor. Must be greater than zero.
	Frequency() time.Duration

	// Process the specified work within the specified context providing intermediate updates
	// with the specified updater. The specified context will be forcefully canceled after
	// the processing timeout specified by the work.
	Process(ctx context.Context, wrk *Work, updater ProcessingUpdater) *PendingUpdate
}
