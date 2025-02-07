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

type ProviderSessionRepository struct {
	*storeStructuredMongo.Repository
}

func (p *ProviderSessionRepository) EnsureIndexes() error {
	return p.CreateAllIndexes(context.Background(), []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "id", Value: 1}},
			Options: options.Index().
				SetUnique(true).
				SetBackground(true),
		},
		{
			Keys: bson.D{{Key: "userId", Value: 1}},
			Options: options.Index().
				SetBackground(true),
		},
		{
			Keys: bson.D{{Key: "userId", Value: 1}, {Key: "type", Value: 1}, {Key: "name", Value: 1}},
			Options: options.Index().
				SetUnique(true).
				SetBackground(true),
		},
	})
}

func (p *ProviderSessionRepository) ListUserProviderSessions(ctx context.Context, userID string, filter *auth.ProviderSessionFilter, pagination *page.Pagination) (auth.ProviderSessions, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	}
	if filter == nil {
		filter = auth.NewProviderSessionFilter()
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(filter); err != nil {
		return nil, errors.Wrap(err, "filter is invalid")
	}
	if pagination == nil {
		pagination = page.NewPagination()
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(pagination); err != nil {
		return nil, errors.Wrap(err, "pagination is invalid")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": userID, "filter": filter, "pagination": pagination})

	providerSessions := auth.ProviderSessions{}
	selector := bson.M{
		"userId": userID,
	}
	if filter.Type != nil {
		selector["type"] = *filter.Type
	}
	if filter.Name != nil {
		selector["name"] = *filter.Name
	}
	if filter.ExternalID != nil {
		selector["externalId"] = *filter.ExternalID
	}
	opts := storeStructuredMongo.FindWithPagination(pagination).
		SetSort(bson.M{"createdTime": -1})
	cursor, err := p.Find(ctx, selector, opts)
	logger.WithFields(log.Fields{"count": len(providerSessions), "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("ListUserProviderSessions")
	if err != nil {
		return nil, errors.Wrap(err, "unable to list user provider sessions")
	}
	if err = cursor.All(ctx, &providerSessions); err != nil {
		return nil, errors.Wrap(err, "unable to decode user provider sessions")
	}

	if providerSessions == nil {
		providerSessions = auth.ProviderSessions{}
	}

	return providerSessions, nil
}

func (p *ProviderSessionRepository) CreateUserProviderSession(ctx context.Context, userID string, create *auth.ProviderSessionCreate) (*auth.ProviderSession, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}

	providerSession, err := auth.NewProviderSession(ctx, userID, create)
	if err != nil {
		return nil, err
	} else if err = structureValidator.New(log.LoggerFromContext(ctx)).Validate(providerSession); err != nil {
		return nil, errors.Wrap(err, "provider session is invalid")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": userID, "create": create})

	_, err = p.InsertOne(ctx, providerSession)
	logger.WithFields(log.Fields{"id": providerSession.ID, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("CreateUserProviderSession")
	if err != nil {
		return nil, errors.Wrap(err, "unable to create user provider session")
	}

	return providerSession, nil
}

func (p *ProviderSessionRepository) DeleteAllProviderSessions(ctx context.Context, userID string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if userID == "" {
		return errors.New("user id is missing")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithField("userId", userID)

	changeInfo, err := p.DeleteMany(ctx, bson.M{"userId": userID})
	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("DeleteAllProviderSessions")
	if err != nil {
		return errors.Wrap(err, "unable to delete all provider sessions")
	}

	return nil
}

func (p *ProviderSessionRepository) GetProviderSession(ctx context.Context, id string) (*auth.ProviderSession, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithField("id", id)

	var providerSession *auth.ProviderSession
	err := p.FindOne(ctx, bson.M{"id": id}).Decode(&providerSession)
	logger.WithField("duration", time.Since(now)/time.Microsecond).WithError(err).Debug("GetProviderSession")
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}

	return providerSession, err
}

func (p *ProviderSessionRepository) UpdateProviderSession(ctx context.Context, id string, update *auth.ProviderSessionUpdate) (*auth.ProviderSession, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	}
	if update == nil {
		return nil, errors.New("update is missing")
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(update); err != nil {
		return nil, errors.Wrap(err, "update is invalid")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"id": id, "update": update})

	set := bson.M{
		"modifiedTime": now,
	}
	unset := bson.M{}
	if update.OAuthToken != nil {
		set["oauthToken"] = update.OAuthToken
	} else {
		unset["oauthToken"] = true
	}
	if update.ExternalID != nil {
		set["externalId"] = update.ExternalID
	} else {
		unset["externalId"] = true
	}
	changeInfo, err := p.UpdateMany(ctx, bson.M{"id": id}, p.ConstructUpdate(set, unset))
	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("UpdateProviderSession")
	if err != nil {
		return nil, errors.Wrap(err, "unable to update provider session")
	}

	return p.GetProviderSession(ctx, id)
}

func (p *ProviderSessionRepository) DeleteProviderSession(ctx context.Context, id string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if id == "" {
		return errors.New("id is missing")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithField("id", id)

	changeInfo, err := p.DeleteMany(ctx, bson.M{"id": id})
	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("DeleteProviderSession")
	if err != nil {
		return errors.Wrap(err, "unable to delete provider session")
	}

	return nil
}
