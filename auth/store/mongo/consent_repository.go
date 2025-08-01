package mongo

import (
	"context"
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

type ConsentRepository struct {
	*storeStructuredMongo.Repository
}

func (p *ConsentRepository) EnsureIndexes() error {
	return p.CreateAllIndexes(context.Background(), []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "type", Value: 1}, {Key: "version", Value: -1}},
			Options: options.Index().
				SetUnique(true).
				SetName("UniqueConsentVersion"),
		},
	})
}

func (p *ConsentRepository) List(ctx context.Context, filter *auth.ConsentFilter, pagination *page.Pagination) (auth.Consents, error) {
	if filter == nil {
		filter = auth.NewConsentFilter()
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

	selector := bson.M{}
	if filter.Type != nil {
		selector["type"] = *filter.Type
	}
	if filter.Version != nil {
		selector["version"] = *filter.Version
	}

	opts := storeStructuredMongo.FindWithPagination(pagination).
		SetSort(bson.M{"version": -1})
	cursor, err := p.Find(ctx, selector, opts)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list consents")
	}

	consents := auth.Consents{}
	if err = cursor.All(ctx, &consents); err != nil {
		return nil, errors.Wrap(err, "unable to decode consents")
	}

	if consents == nil {
		consents = auth.Consents{}
	}
	logger.WithFields(log.Fields{"count": len(consents), "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("ListConsents")

	return consents, nil
}

func (p *ConsentRepository) EnsureConsent(ctx context.Context, consent *auth.Consent) error {
	if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(consent); err != nil {
		return errors.Wrap(err, "filter is invalid")
	}

	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"type": consent.Type, "version": consent.Version})

	selector := bson.M{
		"type":    consent.Type,
		"version": consent.Version,
	}
	update := bson.M{
		"$setOnInsert": consent,
	}

	opts := options.Update().SetUpsert(true)

	result, err := p.UpdateOne(ctx, selector, update, opts)
	if err != nil {
		return errors.Wrap(err, "unable to ensure consent")
	}

	logger.WithFields(log.Fields{"result": *result}).Info("ensured consent")

	return nil
}
