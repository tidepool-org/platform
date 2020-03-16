package service

import (
	"context"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/prescription"
	prescriptionStoreMongo "github.com/tidepool-org/platform/prescription/store/mongo"
)

type Client struct {
	prescriptionStore *prescriptionStoreMongo.Store
}

func NewClient(logger log.Logger, prescriptionStore *prescriptionStoreMongo.Store) (*Client, error) {
	if logger == nil {
		return nil, errors.New("logger is missing")
	}
	if prescriptionStore == nil {
		return nil, errors.New("prescription store is missing")
	}

	return &Client{
		prescriptionStore: prescriptionStore,
	}, nil
}

func (c *Client) CreatePrescription(ctx context.Context, userID string, create *prescription.RevisionCreate) (*prescription.Prescription, error) {
	ssn := c.prescriptionStore.NewPrescriptionSession()
	defer ssn.Close()

	return ssn.CreatePrescription(ctx, userID, create)
}
