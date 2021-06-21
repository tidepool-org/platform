package test

import (
	"context"

	"github.com/tidepool-org/platform/page"

	"github.com/tidepool-org/platform/prescription"

	"github.com/onsi/gomega"
)

type CreatePrescriptionInput struct {
	Context        context.Context
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

type DeletePrescriptionInput struct {
	Ctx         context.Context
	ClinicID    string
	ID          string
	ClinicianID string
}

type DeletePrescriptionOutput struct {
	Success bool
	Err     error
}

type AddRevisionInput struct {
	Ctx    context.Context
	ID     string
	Create *prescription.RevisionCreate
}

type AddRevisionOutput struct {
	Prescr *prescription.Prescription
	Err    error
}

type ClaimPrescriptionInput struct {
	Ctx   context.Context
	Claim *prescription.Claim
}

type ClaimPrescriptionOutput struct {
	Prescr *prescription.Prescription
	Err    error
}

type UpdatePrescriptionStateInput struct {
	Ctx    context.Context
	ID     string
	Update *prescription.StateUpdate
}

type UpdatePrescriptionStateOutput struct {
	Prescr *prescription.Prescription
	Err    error
}

type PrescriptionAccessor struct {
	CreatePrescriptionInvocations      int
	CreatePrescriptionInputs           []CreatePrescriptionInput
	CreatePrescriptionOutputs          []CreatePrescriptionOutput
	ListPrescriptionsInvocations       int
	ListPrescriptionsInputs            []ListPrescriptionsInput
	ListPrescriptionOutputs            []ListPrescriptionsOutput
	DeletePrescriptionInvocations      int
	DeletePrescriptionInputs           []DeletePrescriptionInput
	DeletePrescriptionOutputs          []DeletePrescriptionOutput
	AddRevisionInvocations             int
	AddRevisionInputs                  []AddRevisionInput
	AddRevisionOutputs                 []AddRevisionOutput
	ClaimPrescriptionInvocations       int
	ClaimPrescriptionInputs            []ClaimPrescriptionInput
	ClaimPrescriptionOutputs           []ClaimPrescriptionOutput
	UpdatePrescriptionStateInvocations int
	UpdatePrescriptionStateInputs      []UpdatePrescriptionStateInput
	UpdatePrescriptionStateOutputs     []UpdatePrescriptionStateOutput
}

func NewPrescriptionAccessor() *PrescriptionAccessor {
	return &PrescriptionAccessor{}
}

func (p *PrescriptionAccessor) CreatePrescription(ctx context.Context, create *prescription.RevisionCreate) (*prescription.Prescription, error) {
	p.CreatePrescriptionInvocations++

	p.CreatePrescriptionInputs = append(p.CreatePrescriptionInputs, CreatePrescriptionInput{Context: ctx, RevisionCreate: create})

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

func (p *PrescriptionAccessor) DeletePrescription(ctx context.Context, clinicID, prescriptionID, clinicianID string) (bool, error) {
	p.DeletePrescriptionInvocations++

	p.DeletePrescriptionInputs = append(p.DeletePrescriptionInputs, DeletePrescriptionInput{Ctx: ctx, ClinicID: clinicID, ClinicianID: clinicianID, ID: prescriptionID})

	gomega.Expect(p.DeletePrescriptionOutputs).ToNot(gomega.BeEmpty())

	output := p.DeletePrescriptionOutputs[0]
	p.DeletePrescriptionOutputs = p.DeletePrescriptionOutputs[1:]
	return output.Success, output.Err
}

func (p *PrescriptionAccessor) AddRevision(ctx context.Context, id string, create *prescription.RevisionCreate) (*prescription.Prescription, error) {
	p.AddRevisionInvocations++

	p.AddRevisionInputs = append(p.AddRevisionInputs, AddRevisionInput{Ctx: ctx, ID: id, Create: create})

	gomega.Expect(p.AddRevisionOutputs).ToNot(gomega.BeEmpty())

	output := p.AddRevisionOutputs[0]
	p.AddRevisionOutputs = p.AddRevisionOutputs[1:]
	return output.Prescr, output.Err
}

func (p *PrescriptionAccessor) ClaimPrescription(ctx context.Context, claim *prescription.Claim) (*prescription.Prescription, error) {
	p.ClaimPrescriptionInvocations++

	p.ClaimPrescriptionInputs = append(p.ClaimPrescriptionInputs, ClaimPrescriptionInput{Ctx: ctx, Claim: claim})

	gomega.Expect(p.ClaimPrescriptionOutputs).ToNot(gomega.BeEmpty())

	output := p.ClaimPrescriptionOutputs[0]
	p.ClaimPrescriptionOutputs = p.ClaimPrescriptionOutputs[1:]
	return output.Prescr, output.Err
}

func (p *PrescriptionAccessor) UpdatePrescriptionState(ctx context.Context, id string, update *prescription.StateUpdate) (*prescription.Prescription, error) {
	p.UpdatePrescriptionStateInvocations++

	p.UpdatePrescriptionStateInputs = append(p.UpdatePrescriptionStateInputs, UpdatePrescriptionStateInput{Ctx: ctx, ID: id, Update: update})

	gomega.Expect(p.UpdatePrescriptionStateOutputs).ToNot(gomega.BeEmpty())

	output := p.UpdatePrescriptionStateOutputs[0]
	p.UpdatePrescriptionStateOutputs = p.UpdatePrescriptionStateOutputs[1:]
	return output.Prescr, output.Err
}

func (p *PrescriptionAccessor) Expectations() {
	gomega.Expect(p.CreatePrescriptionOutputs).To(gomega.BeEmpty())
	gomega.Expect(p.ListPrescriptionOutputs).To(gomega.BeEmpty())
	gomega.Expect(p.DeletePrescriptionOutputs).To(gomega.BeEmpty())
	gomega.Expect(p.AddRevisionOutputs).To(gomega.BeEmpty())
	gomega.Expect(p.ClaimPrescriptionOutputs).To(gomega.BeEmpty())
	gomega.Expect(p.UpdatePrescriptionStateOutputs).To(gomega.BeEmpty())
}
