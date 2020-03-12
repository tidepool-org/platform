package test

import (
	"context"

	"github.com/tidepool-org/platform/prescription"

	"github.com/onsi/gomega"
)

type CreatePrescriptionInput struct {
	Context        context.Context
	UserID         string
	RevisionCreate *prescription.RevisionCreate
}

type CreatePrescriptionOutput struct {
	Prescription *prescription.Prescription
	Error        error
}

type PrescriptionAccessor struct {
	CreatePrescriptionInvocations int
	CreatePrescriptionInputs      []CreatePrescriptionInput
	CreatePrescriptionOutputs     []CreatePrescriptionOutput
}

func NewPrescriptionAccessor() *PrescriptionAccessor {
	return &PrescriptionAccessor{}
}

func (p *PrescriptionAccessor) CreatePrescription(ctx context.Context, userID string, create *prescription.RevisionCreate) (*prescription.Prescription, error) {
	p.CreatePrescriptionInvocations++

	p.CreatePrescriptionInputs = append(p.CreatePrescriptionInputs, CreatePrescriptionInput{Context: ctx, UserID: userID, RevisionCreate: create})

	gomega.Expect(p.CreatePrescriptionOutputs).ToNot(gomega.BeEmpty())

	output := p.CreatePrescriptionOutputs[0]
	p.CreatePrescriptionOutputs = p.CreatePrescriptionOutputs[1:]
	return output.Prescription, output.Error
}

func (p *PrescriptionAccessor) Expectations() {
	gomega.Expect(p.CreatePrescriptionOutputs).To(gomega.BeEmpty())
}
