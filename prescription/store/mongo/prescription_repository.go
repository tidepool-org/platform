package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/page"

	"github.com/tidepool-org/platform/log"
	structureValidator "github.com/tidepool-org/platform/structure/validator"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/prescription"
	structuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

type PrescriptionRepository struct {
	*structuredMongo.Repository
}

func (p *PrescriptionRepository) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "patientUserId", Value: 1}},
			Options: options.Index().
				SetName("GetByPatientId").
				SetBackground(true),
		},
		{
			Keys: bson.D{{Key: "prescriberUserId", Value: 1}},
			Options: options.Index().
				SetName("GetByPrescriberId").
				SetBackground(true),
		},
		{
			Keys: bson.D{{Key: "createdUserId", Value: 1}},
			Options: options.Index().
				SetName("GetByCreatedUserId").
				SetBackground(true),
		},
		{
			Keys: bson.D{{Key: "accessCode", Value: 1}},
			Options: options.Index().
				SetName("GetByUniqueAccessCode").
				SetUnique(true).
				SetSparse(true).
				SetBackground(true),
		},
		{
			Keys: bson.D{{Key: "latestRevision.attributes.email", Value: 1}},
			Options: options.Index().
				SetName("GetByPatientEmail").
				SetBackground(true),
		},
		{
			Keys: bson.D{{Key: "_id", Value: 1}, {Key: "revisionHistory.revisionId", Value: 1}},
			Options: options.Index().
				SetName("UniqueRevisionId").
				SetUnique(true).
				SetBackground(true),
		},
	}

	return p.CreateAllIndexes(ctx, indexes)
}

func (p *PrescriptionRepository) CreatePrescription(ctx context.Context, create *prescription.RevisionCreate) (*prescription.Prescription, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}

	model := prescription.NewPrescription(create)
	if err := structureValidator.New().Validate(model); err != nil {
		return nil, errors.Wrap(err, "prescription is invalid")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": create.ClinicianId, "create": create})

	_, err := p.InsertOne(ctx, model)
	logger.WithFields(log.Fields{"id": model.ID, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("CreatePrescription")
	if err != nil {
		return nil, errors.Wrap(err, "unable to create prescription")
	}

	return model, nil
}

func (p *PrescriptionRepository) ListPrescriptions(ctx context.Context, filter *prescription.Filter, pagination *page.Pagination) (prescription.Prescriptions, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
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

	opts := structuredMongo.FindWithPagination(pagination)
	cursor, err := p.Find(ctx, selector, opts)

	logger.WithFields(log.Fields{"duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("ListPrescriptions")
	if err != nil {
		return nil, errors.Wrap(err, "unable to list prescriptions")
	}

	prescriptions := prescription.Prescriptions{}
	defer cursor.Close(ctx)
	if err = cursor.All(ctx, &prescriptions); err != nil {
		return nil, errors.Wrap(err, "unable to decode prescriptions")
	}

	return prescriptions, nil
}

func (p *PrescriptionRepository) DeletePrescription(ctx context.Context, clinicId, prescriptionId, clinicianId string) (bool, error) {
	if ctx == nil {
		return false, errors.New("context is missing")
	}
	if clinicianId == "" {
		return false, errors.New("clinician id is missing")
	}
	if clinicId == "" {
		return false, errors.New("clinic id is missing")
	}
	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"clinic": clinicId, "id": prescriptionId})

	prescriptionID, err := primitive.ObjectIDFromHex(prescriptionId)
	if err == primitive.ErrInvalidHex {
		return false, nil
	} else if err != nil {
		return false, err
	}

	selector := bson.M{
		"_id": prescriptionID,
		"clinicId": clinicId,
		"state": bson.M{
			"$in": []string{prescription.StateDraft, prescription.StatePending},
		},
		"deletedTime": nil,
	}
	update := bson.M{
		"$set": bson.M{
			"deletedUserId": clinicianId,
			"deletedTime":   now,
		},
	}

	res, err := p.UpdateOne(ctx, selector, update)
	logger.WithFields(log.Fields{"duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("DeletePrescription")
	if err != nil {
		return false, errors.Wrap(err, "unable to delete prescription")
	}

	success := res.ModifiedCount == 1
	return success, nil
}

func (p *PrescriptionRepository) AddRevision(ctx context.Context, prescriptionId string, create *prescription.RevisionCreate) (*prescription.Prescription, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}

	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": create.ClinicianId, "prescriptionId": prescriptionId, "create": create})

	prescriptionID, err := primitive.ObjectIDFromHex(prescriptionId)
	if err == primitive.ErrInvalidHex {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	selector := bson.M{
		"_id": prescriptionID,
		"clinicId": create.ClinicId,
	}

	prescr := &prescription.Prescription{}
	err = p.FindOne(ctx, selector).Decode(prescr)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "could not get prescription to add revision to")
	}

	prescriptionUpdate := prescription.NewPrescriptionAddRevisionUpdate(prescr, create)
	if err := structureValidator.New().Validate(prescriptionUpdate); err != nil {
		return nil, errors.Wrap(err, "the prescription update is invalid")
	}

	// Concurrent updates are safe, because updates are atomic at the document level and
	// because revision ids are guaranteed to be unique in a document.
	updateSelector := bson.M{
		"_id":                       prescr.ID,
		"latestRevision.revisionId": prescr.LatestRevision.RevisionID,
	}

	update := newMongoUpdateFromPrescriptionUpdate(prescriptionUpdate)

	now := time.Now()
	res, err := p.UpdateOne(ctx, updateSelector, update)
	logger.WithFields(log.Fields{"id": prescriptionId, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("UpdatePrescription")
	if err != nil {
		return nil, errors.Wrap(err, "unable to update prescription")
	} else if res.ModifiedCount == 0 {
		return nil, errors.New("unable to find prescription to update")
	}

	err = p.FindOneByID(ctx, prescr.ID, prescr)
	if err != nil {
		return nil, errors.Wrap(err, "unable to find updated prescription")
	}

	return prescr, nil
}

func (p *PrescriptionRepository) ClaimPrescription(ctx context.Context, claim *prescription.Claim) (*prescription.Prescription, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}

	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": claim.PatientId, "claim": claim})

	selector := bson.M{
		"accessCode":                         claim.AccessCode,
		"latestRevision.attributes.birthday": claim.Birthday,
		"patientUserId":                      nil,
		"state":                              prescription.StateSubmitted,
	}

	prescr := &prescription.Prescription{}
	err := p.FindOne(ctx, selector).Decode(prescr)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "could not get prescription to add revision to")
	}

	id := prescr.ID
	prescriptionUpdate := prescription.NewPrescriptionClaimUpdate(claim.PatientId, prescr)
	if err := structureValidator.New().Validate(prescriptionUpdate); err != nil {
		return nil, errors.Wrap(err, "the prescription update is invalid")
	}

	update := newMongoUpdateFromPrescriptionUpdate(prescriptionUpdate)

	now := time.Now()
	res, err := p.UpdateOne(ctx, selector, update)
	logger.WithFields(log.Fields{"id": id, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("UpdatePrescription")
	if err != nil {
		return nil, errors.Wrap(err, "unable to update prescription")
	} else if res.ModifiedCount == 0 {
		return nil, errors.New("unable to find prescription to update")
	}

	prescr = &prescription.Prescription{}
	err = p.FindOneByID(ctx, id, prescr)
	if err != nil {
		return nil, errors.Wrap(err, "unable to find updated prescription")
	}

	return prescr, nil
}

func (p *PrescriptionRepository) UpdatePrescriptionState(ctx context.Context, prescriptionId string, update *prescription.StateUpdate) (*prescription.Prescription, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}

	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": update.PatientId, "id": prescriptionId, "update": update})

	prescriptionID, err := primitive.ObjectIDFromHex(prescriptionId)
	if err == primitive.ErrInvalidHex {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	selector := bson.M{
		"_id":           prescriptionID,
		"patientUserId": update.PatientId,
	}

	prescr := &prescription.Prescription{}
	err = p.FindOne(ctx, selector).Decode(prescr)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}

	prescriptionUpdate := prescription.NewPrescriptionStateUpdate(prescr, update)
	if err := structureValidator.New().Validate(prescriptionUpdate); err != nil {
		return nil, errors.Wrap(err, "the prescription update is invalid")
	}
	mongoUpdate := newMongoUpdateFromPrescriptionUpdate(prescriptionUpdate)

	if err = p.deactiveActivePrescriptions(ctx, update.PatientId); err != nil {
		return nil, err
	}

	now := time.Now()
	res, err := p.UpdateOne(ctx, selector, mongoUpdate)
	logger.WithFields(log.Fields{"id": prescr.ID, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("UpdatePrescription")
	if err != nil {
		return nil, errors.Wrap(err, "unable to update prescription")
	} else if res.ModifiedCount == 0 {
		return nil, errors.New("unable to find prescription to update")
	}

	err = p.FindOneByID(ctx, prescr.ID, prescr)
	if err != nil {
		return nil, errors.Wrap(err, "unable to find updated prescription")
	}

	return prescr, nil
}

func (p *PrescriptionRepository) deactiveActivePrescriptions(ctx context.Context, patientUserId string) error {
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": patientUserId})

	selector := bson.M{
		"patientUserId": patientUserId,
		"state":         prescription.StateActive,
	}
	update := bson.M{
		"$set": bson.M{
			"state": prescription.StateInactive,
		},
	}

	now := time.Now()
	_, err := p.UpdateMany(ctx, selector, update)
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
			{"prescriberUserId": filter.ClinicianID},
			{"createdUserId": filter.ClinicianID},
		}
	}
	if filter.PatientUserID != "" {
		selector["patientUserId"] = filter.PatientUserID
	}
	if filter.PatientEmail != "" {
		selector["latestRevision.attributes.email"] = filter.PatientEmail
	}
	if filter.ID != "" {
		objID, err := primitive.ObjectIDFromHex(filter.ID)
		if err != nil {
			selector["_id"] = nil
		} else {
			selector["_id"] = objID
		}
	}
	if filter.State != "" {
		selector["state"] = filter.State
	}
	if filter.CreatedAfter != nil {
		selector["createdTime"] = bson.M{"$gte": filter.CreatedAfter}
	}
	if filter.CreatedBefore != nil {
		selector["createdTime"] = bson.M{"$lt": filter.CreatedBefore}
	}
	if filter.ModifiedAfter != nil {
		selector["modifiedTime"] = bson.M{"$gte": filter.ModifiedAfter}
	}
	if filter.ModifiedBefore != nil {
		selector["modifiedTime"] = bson.M{"$lt": filter.ModifiedBefore}
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
			update["$unset"] = bson.M{
				"accessCode": "",
			}
		}
	}

	if prescrUpdate.PrescriberUserID != "" {
		set["prescriberUserId"] = prescrUpdate.PrescriberUserID
	}

	if prescrUpdate.PatientUserID != "" {
		set["patientUserId"] = prescrUpdate.PatientUserID
	}

	if prescrUpdate.SubmittedTime != nil {
		set["submittedTime"] = prescrUpdate.SubmittedTime
	}

	return update
}
