package mongo

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/log"
	structureValidator "github.com/tidepool-org/platform/structure/validator"

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
	if userID == "" {
		return nil, errors.New("userID is missing")
	}

	model, err := prescription.NewPrescription(ctx, userID, create)
	if err != nil {
		return nil, err
	} else if err = structureValidator.New().Validate(model); err != nil {
		return nil, errors.Wrap(err, "prescription is invalid")
	}

	if r.IsClosed() {
		return nil, errors.New("session closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": userID, "create": create})

	err = r.C().Insert(model)
	logger.WithFields(log.Fields{"id": model.ID, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("CreatePrescription")
	if err != nil {
		return nil, errors.Wrap(err, "unable to create user restricted token")
	}

	return model, nil
}
