package service

import (
	"context"
	"fmt"

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

	// Fetch prescription using the claim
	prescr, err := p.GetClaimablePrescription(ctx, claim)
	if err != nil || prescr == nil {
		return nil, err
	}

	// Verify the prescription integrity
	if err = prescription.VerifyRevisionIntegrityHash(*prescr.LatestRevision); err != nil {
		return nil, fmt.Errorf("integrity check for prescription %v failed: %w", prescr.ID, err)
	}

	// Claim the prescription atomically
	claim.RevisionHash = prescr.LatestRevision.IntegrityHash.Hash
	prescr, err = repo.ClaimPrescription(ctx, claim)
	if err != nil {
		return nil, err
	}

	// Share patient account with the clinic that created the prescription
	_, err = p.clinicsClient.SharePatientAccount(ctx, prescr.ClinicID, prescr.PatientUserID)
	if err != nil {
		return nil, err
	}
	return prescr, nil
}

func (p *PrescriptionService) GetClaimablePrescription(ctx context.Context, claim *prescription.Claim) (*prescription.Prescription, error) {
	repo := p.prescriptionStore.GetPrescriptionRepository()
	return repo.GetClaimablePrescription(ctx, claim)
}

func (p *PrescriptionService) UpdatePrescriptionState(ctx context.Context, prescriptionID string, update *prescription.StateUpdate) (*prescription.Prescription, error) {
	repository := p.prescriptionStore.GetPrescriptionRepository()
	return repository.UpdatePrescriptionState(ctx, prescriptionID, update)
}
