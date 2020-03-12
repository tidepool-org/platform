package test

import (
	prescriptionTest "github.com/tidepool-org/platform/prescription/test"
	"github.com/tidepool-org/platform/test"
)

type PrescriptionSession struct {
	*test.Closer
	*prescriptionTest.PrescriptionAccessor
}

func NewPrescriptionSession() *PrescriptionSession {
	return &PrescriptionSession{
		Closer:               test.NewCloser(),
		PrescriptionAccessor: prescriptionTest.NewPrescriptionAccessor(),
	}
}

func (p *PrescriptionSession) Expectations() {
	p.Closer.AssertOutputsEmpty()
	p.PrescriptionAccessor.Expectations()
}
