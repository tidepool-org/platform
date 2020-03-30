package test

import (
	"context"

	"github.com/tidepool-org/platform/page"

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

type ListPrescriptionsInput struct {
	Ctx        context.Context
	Filter     *prescription.Filter
	Pagination *page.Pagination
}

type ListPrescriptionsOutput struct {
	Prescriptions prescription.Prescriptions
	Err           error
}

type GetUnclaimedPrescriptionInput struct {
	Ctx        context.Context
	AccessCode string
}

type GetUnclaimedPrescriptionOutput struct {
	Prescr *prescription.Prescription
	Err    error
}

type PrescriptionAccessor struct {
	CreatePrescriptionInvocations       int
	CreatePrescriptionInputs            []CreatePrescriptionInput
	CreatePrescriptionOutputs           []CreatePrescriptionOutput
	ListPrescriptionsInvocations        int
	ListPrescriptionsInputs             []ListPrescriptionsInput
	ListPrescriptionOutputs             []ListPrescriptionsOutput
	GetUnclaimedPrescriptionInvocations int
	GetUnclaimedPrescriptionInputs      []GetUnclaimedPrescriptionInput
	GetUnclaimedPrescriptionOutputs     []GetUnclaimedPrescriptionOutput
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

func (p *PrescriptionAccessor) ListPrescriptions(ctx context.Context, filter *prescription.Filter, pagination *page.Pagination) (prescription.Prescriptions, error) {
	p.ListPrescriptionsInvocations++

	p.ListPrescriptionsInputs = append(p.ListPrescriptionsInputs, ListPrescriptionsInput{Ctx: ctx, Filter: filter, Pagination: pagination})

	gomega.Expect(p.ListPrescriptionOutputs).ToNot(gomega.BeEmpty())

	output := p.ListPrescriptionOutputs[0]
	p.ListPrescriptionOutputs = p.ListPrescriptionOutputs[1:]
	return output.Prescriptions, output.Err
}

func (p *PrescriptionAccessor) GetUnclaimedPrescription(ctx context.Context, accessCode string) (*prescription.Prescription, error) {
	p.GetUnclaimedPrescriptionInvocations++

	p.GetUnclaimedPrescriptionInputs = append(p.GetUnclaimedPrescriptionInputs, GetUnclaimedPrescriptionInput{Ctx: ctx, AccessCode: accessCode})

	gomega.Expect(p.GetUnclaimedPrescriptionOutputs).ToNot(gomega.BeEmpty())

	output := p.GetUnclaimedPrescriptionOutputs[0]
	p.GetUnclaimedPrescriptionOutputs = p.GetUnclaimedPrescriptionOutputs[1:]
	return output.Prescr, output.Err
}

func (p *PrescriptionAccessor) Expectations() {
	gomega.Expect(p.CreatePrescriptionOutputs).To(gomega.BeEmpty())
}
