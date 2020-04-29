package mongo

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/user"

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
		{Key: []string{"patientId"}, Background: true},
		{Key: []string{"prescriberId"}, Background: true},
		{Key: []string{"createdUserId"}, Background: true},
		{Key: []string{"accessCode"}, Unique: true, Sparse: true, Background: true},
		{Key: []string{"latestRevision.attributes.email"}, Background: true, Name: "latest_patient_email"},
		{Key: []string{"_id", "revisionHistory.revisionId"}, Background: true, Unique: true, Name: "unique_revision_id"},
	})
}

func (p *PrescriptionSession) CreatePrescription(ctx context.Context, userID string, create *prescription.RevisionCreate) (*prescription.Prescription, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if p.IsClosed() {
		return nil, errors.New("session closed")
	}
	if userID == "" {
		return nil, errors.New("userID is missing")
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
		return nil, errors.Wrap(err, "unable to create prescription")
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
		return nil, errors.New("filter is missing")
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

	selector := newMongoSelectorFromFilter(filter)
	selector["deletedTime"] = nil

	prescriptions := prescription.Prescriptions{}
	err := p.C().Find(selector).Skip(pagination.Page * pagination.Size).Limit(pagination.Size).All(&prescriptions)
	logger.WithFields(log.Fields{"duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("ListPrescriptions")
	if err != nil {
		return nil, errors.Wrap(err, "unable to list prescriptions")
	}

	return prescriptions, nil
}

func (p *PrescriptionSession) DeletePrescription(ctx context.Context, clinicianID string, id string) (bool, error) {
	if ctx == nil {
		return false, errors.New("context is missing")
	}
	if p.IsClosed() {
		return false, errors.New("session closed")
	}
	if clinicianID == "" {
		return false, errors.New("clinician id is missing")
	}
	if id == "" {
		return false, errors.New("prescription id is missing")
	} else if !bson.IsObjectIdHex(id) {
		return false, nil
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"clinicianId": clinicianID, "id": id})

	selector := bson.M{
		"_id": bson.ObjectIdHex(id),
		"$or": []bson.M{
			{"prescriberId": clinicianID},
			{"createdUserId": clinicianID},
		},
		"state": bson.M{
			"$in": []string{prescription.StateDraft, prescription.StatePending},
		},
		"deletedTime": nil,
	}
	update := bson.M{
		"$set": bson.M{
			"deletedUserId": clinicianID,
			"deletedTime":   now,
		},
	}

	err := p.C().Update(selector, update)
	logger.WithFields(log.Fields{"duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("DeletePrescription")
	if err == mgo.ErrNotFound {
		return false, nil
	} else if err != nil {
		return false, errors.Wrap(err, "unable to delete prescription")
	} else {
		return true, nil
	}
}

func (p *PrescriptionSession) AddRevision(ctx context.Context, usr *user.User, id string, create *prescription.RevisionCreate) (*prescription.Prescription, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if p.IsClosed() {
		return nil, errors.New("session closed")
	}
	if usr == nil {
		return nil, errors.New("user is missing")
	}
	if id == "" {
		return nil, errors.New("prescription id is missing")
	}

	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": usr.UserID, "id": id, "create": create})

	selector := bson.M{
		"_id": bson.ObjectIdHex(id),
		"$or": []bson.M{
			{"prescriberId": *usr.UserID},
			{"createdUserId": *usr.UserID},
		},
	}

	prescr := &prescription.Prescription{}
	err := p.C().Find(selector).One(prescr)
	if err == mgo.ErrNotFound {
		return nil, nil
	}

	prescriptionUpdate := prescription.NewPrescriptionAddRevisionUpdate(usr, prescr, create)
	if err := structureValidator.New().Validate(prescriptionUpdate); err != nil {
		return nil, errors.Wrap(err, "the prescription update is invalid")
	}

	// Concurrent updates are safe, because updates are atomic at the document level and
	// because there's a unique index on the revision id.
	updateSelector := bson.M{
		"_id":                       prescr.ID,
		"latestRevision.revisionId": prescr.LatestRevision.RevisionID,
	}

	update := newMongoUpdateFromPrescriptionUpdate(prescriptionUpdate)

	now := time.Now()
	err = p.C().Update(updateSelector, update)
	logger.WithFields(log.Fields{"id": id, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("UpdatePrescription")
	if err != nil {
		return nil, errors.Wrap(err, "unable to update prescription")
	}

	err = p.C().FindId(prescr.ID).One(prescr)
	if err != nil {
		return nil, errors.Wrap(err, "unable to find updated prescription")
	}

	return prescr, nil
}

func (p *PrescriptionSession) ClaimPrescription(ctx context.Context, usr *user.User, claim *prescription.Claim) (*prescription.Prescription, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if p.IsClosed() {
		return nil, errors.New("session closed")
	}
	if usr == nil {
		return nil, errors.New("user is missing")
	}

	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": usr.UserID, "claim": claim})

	selector := bson.M{
		"accessCode": claim.AccessCode,
		"patientId":  nil,
		"state":      prescription.StateSubmitted,
	}

	prescr := &prescription.Prescription{}
	err := p.C().Find(selector).One(prescr)
	if err == mgo.ErrNotFound {
		return nil, nil
	}

	prescriptionUpdate := prescription.NewPrescriptionClaimUpdate(usr, prescr)
	if err := structureValidator.New().Validate(prescriptionUpdate); err != nil {
		return nil, errors.Wrap(err, "the prescription update is invalid")
	}

	update := newMongoUpdateFromPrescriptionUpdate(prescriptionUpdate)

	now := time.Now()
	err = p.C().Update(selector, update)
	logger.WithFields(log.Fields{"id": prescr.ID, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("UpdatePrescription")
	if err != nil {
		return nil, errors.Wrap(err, "unable to update prescription")
	}

	err = p.C().FindId(prescr.ID).One(prescr)
	if err != nil {
		return nil, errors.Wrap(err, "unable to find updated prescription")
	}

	return prescr, nil
}

func (p *PrescriptionSession) UpdatePrescriptionState(ctx context.Context, usr *user.User, id string, update *prescription.StateUpdate) (*prescription.Prescription, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if p.IsClosed() {
		return nil, errors.New("session closed")
	}
	if usr == nil {
		return nil, errors.New("user is missing")
	}
	if id == "" {
		return nil, errors.New("prescription id is missing")
	}

	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": usr.UserID, "id": id, "update": update})

	selector := bson.M{
		"_id":       bson.ObjectIdHex(id),
		"patientId": *usr.UserID,
	}

	prescr := &prescription.Prescription{}
	err := p.C().Find(selector).One(prescr)
	if err == mgo.ErrNotFound {
		return nil, nil
	}

	prescriptionUpdate := prescription.NewPrescriptionStateUpdate(usr, prescr, update)
	if err := structureValidator.New().Validate(prescriptionUpdate); err != nil {
		return nil, errors.Wrap(err, "the prescription update is invalid")
	}
	mongoUpdate := newMongoUpdateFromPrescriptionUpdate(prescriptionUpdate)

	if err = p.deactiveActivePrescriptions(ctx, usr); err != nil {
		return nil, err
	}

	now := time.Now()
	err = p.C().Update(selector, mongoUpdate)
	logger.WithFields(log.Fields{"id": prescr.ID, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("UpdatePrescription")
	if err != nil {
		return nil, errors.Wrap(err, "unable to update prescription")
	}

	err = p.C().FindId(prescr.ID).One(prescr)
	if err != nil {
		return nil, errors.Wrap(err, "unable to find updated prescription")
	}

	return prescr, nil
}

func (p *PrescriptionSession) deactiveActivePrescriptions(ctx context.Context, usr *user.User) error {
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": usr.UserID})

	selector := bson.M{
		"patientId": usr.UserID,
		"state":     prescription.StateActive,
	}
	update := bson.M{
		"$set": bson.M{
			"state": prescription.StateInactive,
		},
	}

	now := time.Now()
	_, err := p.C().UpdateAll(selector, update)
	logger.WithFields(log.Fields{"duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("DeactivatePrescriptions")
	if err != nil {
		return errors.Wrap(err, "unable to deactivate prescriptions for user")
	}

	return err
}

func newMongoSelectorFromFilter(filter *prescription.Filter) bson.M {
	selector := bson.M{}
	if filter.ClinicianID != "" {
		selector["$or"] = []bson.M{
			{"prescriberId": filter.ClinicianID},
			{"createdUserId": filter.ClinicianID},
		}
	}
	if filter.PatientID != "" {
		selector["patientId"] = filter.PatientID
	}
	if filter.PatientEmail != "" {
		selector["latestRevision.attributes.email"] = filter.PatientEmail
	}
	if filter.ID != "" {
		if bson.IsObjectIdHex(filter.ID) {
			selector["_id"] = bson.ObjectIdHex(filter.ID)
		} else {
			selector["_id"] = nil
		}
	}
	if filter.State != "" {
		selector["state"] = filter.State
	}
	if filter.CreatedTimeStart != nil {
		selector["createdTime"] = bson.M{"$gt": filter.CreatedTimeStart}
	}
	if filter.CreatedTimeEnd != nil {
		selector["createdTime"] = bson.M{"$lt": filter.CreatedTimeEnd}
	}
	if filter.ModifiedTimeStart != nil {
		selector["latestRevision.attributes.modifiedTime"] = bson.M{"$gt": filter.ModifiedTimeStart}
	}
	if filter.ModifiedTimeEnd != nil {
		selector["latestRevision.attributes.modifiedTime"] = bson.M{"$lt": filter.ModifiedTimeEnd}
	}

	return selector
}

func newMongoUpdateFromPrescriptionUpdate(prescrUpdate *prescription.Update) bson.M {
	set := bson.M{}
	update := bson.M{
		"$set": &set,
	}

	set["state"] = prescrUpdate.State
	set["expirationTime"] = prescrUpdate.ExpirationTime

	if prescrUpdate.Revision != nil {
		set["latestRevision"] = prescrUpdate.Revision
		update["$push"] = bson.M{
			"revisionHistory": prescrUpdate.Revision,
		}
	}

	if prescrUpdate.GetUpdatedAccessCode() != nil {
		code := *prescrUpdate.GetUpdatedAccessCode()
		if code != "" {
			set["accessCode"] = code
		} else {
			set["accessCode"] = nil
		}
	}

	if prescrUpdate.PrescriberUserID != "" {
		set["prescriberId"] = prescrUpdate.PrescriberUserID
	}

	if prescrUpdate.PatientID != "" {
		set["patientId"] = prescrUpdate.PatientID
	}

	return update
}
