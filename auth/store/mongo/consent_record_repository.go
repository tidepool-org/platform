package mongo

import (
	"context"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/summary/store"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type ConsentRecordRepository struct {
	*storeStructuredMongo.Repository
	consentRepository *ConsentRepository
}

func (p *ConsentRecordRepository) EnsureIndexes() error {
	return p.CreateAllIndexes(context.Background(), []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "userId", Value: 1},
				{Key: "type", Value: 1},
				{Key: "status", Value: 1},
			},
			Options: options.Index().
				SetUnique(true).
				SetName("UniqueActiveConsentRecordPerUser").
				SetPartialFilterExpression(bson.M{
					"status": "active",
				}),
		},
		{
			Keys: bson.D{
				{Key: "userId", Value: 1},
				{Key: "type", Value: 1},
				{Key: "createdTime", Value: -1},
			},
			Options: options.Index().
				SetName("MostRecentConsentRecordForUserAndType"),
		},
		{
			Keys: bson.D{
				{Key: "id", Value: 1},
			},
			Options: options.Index().
				SetUnique(true).
				SetName("UniqueConsentRecordId"),
		},
	})
}

func (p *ConsentRecordRepository) Get(ctx context.Context, userID string, id string) (*auth.ConsentRecord, error) {
	selector := bson.M{
		"userId": userID,
		"id":     id,
	}

	consentRecord := &auth.ConsentRecord{}
	err := p.FindOne(ctx, selector).Decode(consentRecord)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return consentRecord, nil
}

func (p *ConsentRecordRepository) List(ctx context.Context, userID string, filter *auth.ConsentRecordFilter, pagination *page.Pagination) (auth.ConsentRecords, error) {
	if filter == nil {
		filter = auth.NewConsentRecordFilter()
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(filter); err != nil {
		return nil, errors.Wrap(err, "filter is invalid")
	}
	if pagination == nil {
		pagination = page.NewPagination()
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(pagination); err != nil {
		return nil, errors.Wrap(err, "pagination is invalid")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"filter": filter, "pagination": pagination})

	selector := bson.M{
		"userId": userID,
	}
	if filter.ID != nil {
		selector["id"] = *filter.ID
	}
	if filter.Type != nil {
		selector["type"] = *filter.Type
	}
	if filter.Status != nil {
		selector["status"] = *filter.Status
	}
	if filter.Version != nil {
		selector["version"] = *filter.Version
	}

	sort := bson.M{
		"type":        1,
		"createdTime": -1,
	}

	consentRecords := auth.ConsentRecords{}
	var err error

	if *filter.Latest {
		consentRecords, err = p.listLatest(ctx, selector, sort, pagination)
	} else {
		consentRecords, err = p.listAll(ctx, selector, sort, pagination)
	}

	if consentRecords == nil {
		consentRecords = auth.ConsentRecords{}
	}

	logger.WithFields(log.Fields{"count": len(consentRecords), "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("ListConsentRecords")

	return consentRecords, nil
}

func (p *ConsentRecordRepository) listLatest(ctx context.Context, selector bson.M, sort bson.M, pagination *page.Pagination) (auth.ConsentRecords, error) {
	pipeline := bson.A{
		bson.M{
			"$match": selector,
		},
		bson.M{
			"$sort": sort,
		},
		bson.M{
			"$group": bson.M{
				"_id":        "$type",
				"mostRecent": bson.M{"$first": "$$ROOT"},
			},
		},
		bson.M{
			"$replaceRoot": bson.M{"$newRoot": "$mostRecent"},
		},
		bson.M{
			"$skip": pagination.Page * pagination.Size,
		},
		bson.M{
			"$limit": pagination.Size,
		},
	}

	cursor, err := p.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list consent records")
	}

	consentRecords := auth.ConsentRecords{}
	if err = cursor.All(ctx, &consentRecords); err != nil {
		return nil, errors.Wrap(err, "unable to decode consent records")
	}

	return consentRecords, nil
}

func (p *ConsentRecordRepository) listAll(ctx context.Context, selector bson.M, sort bson.M, pagination *page.Pagination) (auth.ConsentRecords, error) {
	opts := storeStructuredMongo.FindWithPagination(pagination).SetSort(sort)
	cursor, err := p.Find(ctx, selector, opts)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list consent records")
	}

	consentRecords := auth.ConsentRecords{}
	if err = cursor.All(ctx, &consentRecords); err != nil {
		return nil, errors.Wrap(err, "unable to decode consent records")
	}

	return consentRecords, nil
}

func (p *ConsentRecordRepository) Create(ctx context.Context, userID string, create *auth.ConsentRecordCreate) (*auth.ConsentRecord, error) {
	consentRecord, err := auth.NewConsentRecord(ctx, userID, create)
	if err != nil {
		return nil, err
	} else if err = structureValidator.New(log.LoggerFromContext(ctx)).Validate(consentRecord); err != nil {
		return nil, errors.Wrap(err, "consent record is invalid")
	}

	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": userID, "create": create})

	result, err := p.List(ctx, userID, &auth.ConsentRecordFilter{
		Type:   pointer.FromAny(create.Type),
		Latest: pointer.FromAny(true),
		Status: pointer.FromAny(auth.ConsentRecordStatusActive),
	}, &page.Pagination{Page: 0, Size: 1})
	if err != nil {
		return nil, errors.Wrap(err, "unable to list existing active consent records for type")
	}
	if len(result) > 0 {
		return nil, errors.Newf("active consent record with type %s already exists", consentRecord.Type)
	}

	_, err = store.WithTransaction(ctx, nil, func(sCtx mongo.SessionContext) (any, error) {
		// Revoke existing records
		selector := bson.M{
			"status": auth.ConsentRecordStatusActive,
			"type":   consentRecord.Type,
			"userId": userID,
		}
		update := bson.M{
			"$currentDate": bson.M{
				"modifiedTime": true,
			},
			"$set": bson.M{
				// Make sure we have a non-interrupted stream of data in case the user re-consents to a new version
				"revocationTime": consentRecord.CreatedTime,
				"status":         auth.ConsentRecordStatusRevoked,
			},
		}

		result, err := p.UpdateOne(ctx, selector, update)
		if err != nil {
			return nil, errors.Wrap(err, "unable to revoke existing consent records")
		}
		if result.ModifiedCount > 0 {
			logger.Debugf("revoked %d existing consent records", result.ModifiedCount)
		}

		_, err = p.InsertOne(ctx, consentRecord)
		logger.WithFields(log.Fields{"id": consentRecord.ID}).WithError(err).Debug("CreateConsentRecord")

		return nil, err
	})

	if err != nil {
		return nil, errors.Wrap(err, "unable to create consent record")
	}

	return consentRecord, nil
}

func (p *ConsentRecordRepository) Revoke(ctx context.Context, userID string, id string) error {
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": userID, "id": id})

	selector := bson.M{
		"id":     id,
		"status": auth.ConsentRecordStatusActive,
		"userId": userID,
	}

	update := bson.M{
		"$set": bson.M{
			"status": auth.ConsentRecordStatusRevoked,
		},
		"$currentDate": bson.M{
			"revocationTime": true,
			"modifiedTime":   true,
		},
	}

	result, err := p.UpdateOne(ctx, selector, update)
	if err != nil {
		return errors.Wrap(err, "unable to revoke existing consent records")
	}

	if result.ModifiedCount == 0 {
		logger.Debugf("could not find active consent record")
	}

	return nil
}

func (p *ConsentRecordRepository) Update(ctx context.Context, consentRecord *auth.ConsentRecord) (*auth.ConsentRecord, error) {
	if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(consentRecord); err != nil {
		return nil, errors.Wrap(err, "consent record is invalid")
	}

	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": consentRecord.UserID, "id": consentRecord.ID})

	selector := bson.M{
		"id": consentRecord.ID,
	}
	update := bson.M{
		"$set": consentRecord,
		"$currentDate": bson.M{
			"modifiedTime": true,
		},
	}

	result, err := p.UpdateOne(ctx, selector, update)
	if err != nil {
		return nil, errors.Wrap(err, "unable to update existing consent record")
	}
	if result.ModifiedCount == 0 {
		logger.Debugf("could not find active consent record")
	}

	return consentRecord, nil
}
