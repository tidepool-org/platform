package service

import (
	"context"
	"github.com/tidepool-org/platform/device"
	"go.uber.org/fx"

	"github.com/tidepool-org/platform/user"

	"github.com/tidepool-org/platform/page"

	prescriptionStore "github.com/tidepool-org/platform/prescription/store"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/prescription"
)

type PrescriptionService struct {
	prescriptionStore prescriptionStore.Store
	settingsValidator device.SettingsValidator
}

type Params struct {
	fx.In

	store             prescriptionStore.Store
	settingsValidator device.SettingsValidator
}

func NewService(params Params) (prescription.Service, error) {
	if params.store == nil {
		return nil, errors.New("prescription store is missing")
	}
	if params.settingsValidator == nil {
		return nil, errors.New("settings validator is missing")
	}

	return &PrescriptionService{
		prescriptionStore: params.store,
		settingsValidator: params.settingsValidator,
	}, nil
}

func (p *PrescriptionService) CreatePrescription(ctx context.Context, userID string, create *prescription.RevisionCreate) (*prescription.Prescription, error) {
	repo := p.prescriptionStore.GetPrescriptionRepository()
	return repo.CreatePrescription(ctx, userID, create)
}

func (p *PrescriptionService) ListPrescriptions(ctx context.Context, filter *prescription.Filter, pagination *page.Pagination) (prescription.Prescriptions, error) {
	repo := p.prescriptionStore.GetPrescriptionRepository()
	return repo.ListPrescriptions(ctx, filter, pagination)
}

func (p *PrescriptionService) DeletePrescription(ctx context.Context, clinicianID string, id string) (bool, error) {
	repo := p.prescriptionStore.GetPrescriptionRepository()
	return repo.DeletePrescription(ctx, clinicianID, id)
}

func (p *PrescriptionService) AddRevision(ctx context.Context, usr *user.User, id string, create *prescription.RevisionCreate) (*prescription.Prescription, error) {
	repo := p.prescriptionStore.GetPrescriptionRepository()
	return repo.AddRevision(ctx, usr, id, create)
}

func (p *PrescriptionService) ClaimPrescription(ctx context.Context, usr *user.User, claim *prescription.Claim) (*prescription.Prescription, error) {
	repo := p.prescriptionStore.GetPrescriptionRepository()
	return repo.ClaimPrescription(ctx, usr, claim)
}

func (p *PrescriptionService) UpdatePrescriptionState(ctx context.Context, usr *user.User, id string, update *prescription.StateUpdate) (*prescription.Prescription, error) {
	ssn := p.prescriptionStore.GetPrescriptionRepository()
	return ssn.UpdatePrescriptionState(ctx, usr, id, update)
}
