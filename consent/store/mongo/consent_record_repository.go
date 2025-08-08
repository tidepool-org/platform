package mongo

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/consent"
	"github.com/tidepool-org/platform/pointer"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type ConsentRecordRepository struct {
	*storeStructuredMongo.Repository
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

func (p *ConsentRecordRepository) GetConsentRecord(ctx context.Context, userID string, id string) (*consent.Record, error) {
	selector := bson.M{
		"userId": userID,
		"id":     id,
	}

	consentRecord := &consent.Record{}
	err := p.FindOne(ctx, selector).Decode(consentRecord)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return consentRecord, nil
}

func (p *ConsentRecordRepository) ListConsentRecords(ctx context.Context, userID string, filter *consent.RecordFilter, pagination *page.Pagination) (*storeStructuredMongo.ListResult[consent.Record], error) {
	if filter == nil {
		filter = consent.NewConsentRecordFilter()
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

	var pipeline []bson.M
	if *filter.Latest {
		pipeline = listLatestConsentRecordsPipeline(selector, sort, *pagination)
	} else {
		pipeline = listAllConsentRecordsPipeline(selector, sort, *pagination)
	}

	cursor, err := p.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list consent records")
	}

	result := storeStructuredMongo.ListResult[consent.Record]{}
	if cursor.Next(ctx) {
		if err = cursor.Decode(&result); err != nil {
			return nil, errors.Wrap(err, "unable to decode consent records")
		}
	}

	logger.WithFields(log.Fields{"count": len(result.Data), "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("ListConsentRecords")

	return &result, nil
}

func (p *ConsentRecordRepository) CreateConsentRecord(ctx context.Context, userID string, create *consent.RecordCreate) (*consent.Record, error) {
	consentRecord, err := consent.NewConsentRecord(ctx, userID, create)
	if err != nil {
		return nil, err
	} else if err = structureValidator.New(log.LoggerFromContext(ctx)).Validate(consentRecord); err != nil {
		return nil, errors.Wrap(err, "consent record is invalid")
	}

	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": userID, "create": create})

	result, err := p.ListConsentRecords(ctx, userID, &consent.RecordFilter{
		Type:   pointer.FromAny(create.Type),
		Latest: pointer.FromAny(true),
		Status: pointer.FromAny(consent.RecordStatusActive),
	}, &page.Pagination{Page: 0, Size: 1})
	if err != nil {
		return nil, errors.Wrap(err, "unable to list existing active consent records for type")
	}
	if len(result.Data) > 0 {
		return nil, errors.Newf("active consent record with type %s already exists", consentRecord.Type)
	}

	_, err = p.InsertOne(ctx, consentRecord)
	logger.WithFields(log.Fields{"id": consentRecord.ID}).WithError(err).Debug("CreateConsentRecord")

	return consentRecord, err
}

func (p *ConsentRecordRepository) RevokeConsentRecord(ctx context.Context, userID string, revoke *consent.RecordRevoke) error {
	if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(revoke); err != nil {
		return errors.Wrap(err, "revoke is invalid")
	}

	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": userID, "id": revoke.ID})

	selector := bson.M{
		"id":     revoke.ID,
		"status": consent.RecordStatusActive,
		"userId": userID,
	}

	update := bson.M{
		"$set": bson.M{
			"status":         consent.RecordStatusRevoked,
			"revocationTime": revoke.RevocationTime,
		},
		"$currentDate": bson.M{
			"modifiedTime": true,
		},
	}

	result, err := p.UpdateOne(ctx, selector, update)
	if err != nil {
		return errors.Wrap(err, "unable to revoke existing consent record")
	}

	if result.ModifiedCount == 0 {
		logger.Debugf("could not find active consent record")
	}

	return nil
}

func (p *ConsentRecordRepository) UpdateConsentRecord(ctx context.Context, consentRecord *consent.Record) (*consent.Record, error) {
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

func listLatestConsentRecordsPipeline(selector bson.M, sort bson.M, pagination page.Pagination) []bson.M {
	pipeline := []bson.M{
		{
			"$match": selector,
		},
		{
			"$sort": sort,
		},
		{
			"$group": bson.M{
				"_id":        "$type",
				"mostRecent": bson.M{"$first": "$$ROOT"},
			},
		},
		{
			"$replaceRoot": bson.M{"newRoot": "$mostRecent"},
		},
	}

	pipeline = append(pipeline, storeStructuredMongo.PaginationFacetPipelineStages(pagination)...)
	return pipeline
}

func listAllConsentRecordsPipeline(selector bson.M, sort bson.M, pagination page.Pagination) []bson.M {
	return storeStructuredMongo.ListResultQueryPipeline(selector, sort, pagination)
}
