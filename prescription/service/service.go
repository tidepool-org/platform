package service

import (
	"context"

	"github.com/tidepool-org/platform/clinics"

	"github.com/tidepool-org/platform/page"

	prescriptionStore "github.com/tidepool-org/platform/prescription/store"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/prescription"
)

type PrescriptionService struct {
	prescriptionStore prescriptionStore.Store
	clinicsClient     clinics.Client
}

func NewService(store prescriptionStore.Store, clinicsClient clinics.Client) (prescription.Service, error) {
	if store == nil {
		return nil, errors.New("prescription store is missing")
	}
	if clinicsClient == nil {
		return nil, errors.New("clinics client is missing")
	}

	return &PrescriptionService{
		clinicsClient:     clinicsClient,
		prescriptionStore: store,
	}, nil
}

func (p *PrescriptionService) CreatePrescription(ctx context.Context, create *prescription.RevisionCreate) (*prescription.Prescription, error) {
	repo := p.prescriptionStore.GetPrescriptionRepository()
	return repo.CreatePrescription(ctx, create)
}

func (p *PrescriptionService) ListPrescriptions(ctx context.Context, filter *prescription.Filter, pagination *page.Pagination) (prescription.Prescriptions, error) {
	repo := p.prescriptionStore.GetPrescriptionRepository()
	return repo.ListPrescriptions(ctx, filter, pagination)
}

func (p *PrescriptionService) DeletePrescription(ctx context.Context, clinicID, prescriptionID, clinicianID string) (bool, error) {
	repo := p.prescriptionStore.GetPrescriptionRepository()
	return repo.DeletePrescription(ctx, clinicID, prescriptionID, clinicianID)
}

func (p *PrescriptionService) AddRevision(ctx context.Context, prescriptionID string, create *prescription.RevisionCreate) (*prescription.Prescription, error) {
	repo := p.prescriptionStore.GetPrescriptionRepository()
	return repo.AddRevision(ctx, prescriptionID, create)
}

func (p *PrescriptionService) ClaimPrescription(ctx context.Context, claim *prescription.Claim) (*prescription.Prescription, error) {
	repo := p.prescriptionStore.GetPrescriptionRepository()
	prescr, err := repo.ClaimPrescription(ctx, claim)
	if err != nil {
		return nil, err
	}
	_, err = p.clinicsClient.SharePatientAccount(ctx, prescr.ClinicID, prescr.PatientUserID)
	if err != nil {
		return nil, err
	}
	return prescr, nil
}

func (p *PrescriptionService) UpdatePrescriptionState(ctx context.Context, prescriptionID string, update *prescription.StateUpdate) (*prescription.Prescription, error) {
	repository := p.prescriptionStore.GetPrescriptionRepository()
	return repository.UpdatePrescriptionState(ctx, prescriptionID, update)
}
