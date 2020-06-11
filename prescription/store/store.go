package store

import (
	"context"

	"github.com/tidepool-org/platform/prescription"
	"github.com/tidepool-org/platform/status"
)

type Store interface {
	status.StoreStatusReporter
	GetPrescriptionRepository() PrescriptionRepository
	CreateIndexes(ctx context.Context) error
}

type PrescriptionRepository interface {
	prescription.Accessor
	CreateIndexes(ctx context.Context) error
}
