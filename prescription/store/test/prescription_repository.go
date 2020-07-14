package test

import (
	"context"

	prescriptionTest "github.com/tidepool-org/platform/prescription/test"
)

type PrescriptionRepository struct {
	*prescriptionTest.PrescriptionAccessor
}

func NewPrescriptionSession() *PrescriptionRepository {
	return &PrescriptionRepository{
		PrescriptionAccessor: prescriptionTest.NewPrescriptionAccessor(),
	}
}

func (p *PrescriptionRepository) CreateIndexes(ctx context.Context) error {
	return nil
}

func (p *PrescriptionRepository) Expectations() {
	p.PrescriptionAccessor.Expectations()
}
