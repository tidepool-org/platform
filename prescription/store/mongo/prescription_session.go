package mongo

import (
	"context"
	"time"

	"github.com/globalsign/mgo/bson"

	"github.com/tidepool-org/platform/page"

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

func (p *PrescriptionSession) EnsureIndexes() error {
	return p.EnsureAllIndexes([]mgo.Index{
		{Key: []string{"id"}, Unique: true, Background: true},
		{Key: []string{"patientId"}, Background: true},
		{Key: []string{"prescriberId"}, Background: true},
		{Key: []string{"createdUserId"}, Background: true},
		{Key: []string{"accessCode"}, Unique: true, Background: true},
	})
}

func (p *PrescriptionSession) CreatePrescription(ctx context.Context, userID string, create *prescription.RevisionCreate) (*prescription.Prescription, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("userID is missing")
	}
	if p.IsClosed() {
		return nil, errors.New("session closed")
	}

	model := prescription.NewPrescription(userID, create)
	if err := structureValidator.New().Validate(model); err != nil {
		return nil, errors.Wrap(err, "prescription is invalid")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": userID, "create": create})

	err := p.C().Insert(model)
	logger.WithFields(log.Fields{"id": model.ID, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("CreatePrescription")
	if err != nil {
		return nil, errors.Wrap(err, "unable to create user restricted token")
	}

	return model, nil
}

func (p *PrescriptionSession) ListPrescriptions(ctx context.Context, filter *prescription.Filter, pagination *page.Pagination) (prescription.Prescriptions, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if p.IsClosed() {
		return nil, errors.New("session closed")
	}

	if filter == nil {
		filter = prescription.NewFilter()
	} else if err := structureValidator.New().Validate(filter); err != nil {
		return nil, errors.Wrap(err, "filter is invalid")
	}
	if pagination == nil {
		pagination = page.NewPagination()
	} else if err := structureValidator.New().Validate(pagination); err != nil {
		return nil, errors.Wrap(err, "pagination is invalid")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"filter": filter})

	selector := bson.M{}
	if filter.PatientID != "" {
		selector["patientId"] = filter.PatientID
	}
	if filter.ClinicianID != "" {
		selector["$or"] = []bson.M{
			{"prescriberId": filter.ClinicianID},
			{"createdUserId": filter.ClinicianID},
		}
	}
	if filter.State != "" {
		selector["state"] = filter.State
	}

	prescriptions := prescription.Prescriptions{}
	err := p.C().Find(selector).Skip(pagination.Page * pagination.Size).Limit(pagination.Size).All(&prescriptions)
	logger.WithFields(log.Fields{"duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("ListPrescriptions")
	if err != nil {
		return nil, errors.Wrap(err, "unable to list prescriptions")
	}

	return prescriptions, nil
}

func (p *PrescriptionSession) GetUnclaimedPrescription(ctx context.Context, accessCode string) (*prescription.Prescription, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if p.IsClosed() {
		return nil, errors.New("session closed")
	}
	if accessCode == "" {
		return nil, errors.New("access code is missing")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"accessCode": accessCode})

	selector := bson.M{
		"accessCode": accessCode,
		"patientId":  nil,
		"state":      prescription.StateSubmitted,
	}

	prescr := &prescription.Prescription{}
	err := p.C().Find(selector).One(prescr)
	logger.WithFields(log.Fields{"duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("GetUnclaimedPrescription")
	if err == mgo.ErrNotFound {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "unable to find unclaimed prescription")
	}

	return prescr, nil
}
