package mongo

import (
	"context"

	mgo "github.com/globalsign/mgo"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/prescription"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

type PrescriptionSession struct {
	*storeStructuredMongo.Session
}

func (r *PrescriptionSession) EnsureIndexes() error {
	return r.EnsureAllIndexes([]mgo.Index{
		{Key: []string{"id"}, Unique: true, Background: true},
	})
}

func (r *PrescriptionSession) CreatePrescription(ctx context.Context, userID string, create *prescription.RevisionCreate) (*prescription.Prescription, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}

	return nil, nil
}
