package service

import (
	"context"

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
