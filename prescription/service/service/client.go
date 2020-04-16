package service

import (
	"context"
	"github.com/tidepool-org/platform/user"

	"github.com/tidepool-org/platform/page"

	prescriptionStore "github.com/tidepool-org/platform/prescription/store"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/prescription"
)

type Client struct {
	prescriptionStore prescriptionStore.Store
}

func NewClient(logger log.Logger, store prescriptionStore.Store) (*Client, error) {
	if logger == nil {
		return nil, errors.New("logger is missing")
	}
	if store == nil {
		return nil, errors.New("prescription store is missing")
	}

	return &Client{
		prescriptionStore: store,
	}, nil
}

func (c *Client) CreatePrescription(ctx context.Context, userID string, create *prescription.RevisionCreate) (*prescription.Prescription, error) {
	ssn := c.prescriptionStore.NewPrescriptionSession()
	defer ssn.Close()

	return ssn.CreatePrescription(ctx, userID, create)
}

func (c *Client) ListPrescriptions(ctx context.Context, filter *prescription.Filter, pagination *page.Pagination) (prescription.Prescriptions, error) {
	ssn := c.prescriptionStore.NewPrescriptionSession()
	defer ssn.Close()

	return ssn.ListPrescriptions(ctx, filter, pagination)
}

func (c *Client) DeletePrescription(ctx context.Context, clinicianID string, id string) (bool, error) {
	ssn := c.prescriptionStore.NewPrescriptionSession()
	defer ssn.Close()

	return ssn.DeletePrescription(ctx, clinicianID, id)
}

func (c *Client) AddRevision(ctx context.Context, usr *user.User, id string, create *prescription.RevisionCreate) (*prescription.Prescription, error) {
	ssn := c.prescriptionStore.NewPrescriptionSession()
	defer ssn.Close()

	return ssn.AddRevision(ctx, usr, id, create)
}

func (c *Client) ClaimPrescription(ctx context.Context, usr *user.User, claim *prescription.Claim) (*prescription.Prescription, error) {
	ssn := c.prescriptionStore.NewPrescriptionSession()
	defer ssn.Close()

	return ssn.ClaimPrescription(ctx, usr, claim)
}

func (c *Client) UpdatePrescriptionState(ctx context.Context, usr *user.User, id string, update *prescription.StateUpdate) (*prescription.Prescription, error) {
	ssn := c.prescriptionStore.NewPrescriptionSession()
	defer ssn.Close()

	return ssn.UpdatePrescriptionState(ctx, usr, id, update)
}
