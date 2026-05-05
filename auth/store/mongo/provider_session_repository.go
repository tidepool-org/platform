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
				SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "userId", Value: 1}, {Key: "type", Value: 1}, {Key: "name", Value: 1}, {Key: "externalId", Value: 1}},
			Options: options.Index().
				SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "type", Value: 1}, {Key: "name", Value: 1}, {Key: "externalId", Value: 1}},
		},
	})
}

func (p *ProviderSessionRepository) CreateProviderSession(ctx context.Context, create *auth.ProviderSessionCreate) (*auth.ProviderSession, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if create == nil {
		return nil, errors.New("create is missing")
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(create); err != nil {
		return nil, errors.Wrap(err, "create is invalid")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithField("create", create)

	providerSession := &auth.ProviderSession{
		ID:          auth.NewProviderSessionID(),
		UserID:      create.UserID,
		Type:        create.Type,
		Name:        create.Name,
		OAuthToken:  create.OAuthToken,
		ExternalID:  create.ExternalID,
		CreatedTime: now,
	}
	_, err := p.InsertOne(ctx, providerSession)
	logger.WithFields(log.Fields{"id": providerSession.ID, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("CreateProviderSession")
	if err != nil {
		return nil, errors.Wrap(err, "unable to create provider session")
	}

	return providerSession, nil
}

func (p *ProviderSessionRepository) ListProviderSessions(ctx context.Context, filter *auth.ProviderSessionFilter, pagination *page.Pagination) (auth.ProviderSessions, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if filter == nil {
		return nil, errors.New("filter is missing")
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
	if filter.UserID != nil {
		selector["userId"] = *filter.UserID
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
	if err != nil {
		return nil, errors.Wrap(err, "unable to list provider sessions")
	}

	providerSessions := auth.ProviderSessions{}
	if err = cursor.All(ctx, &providerSessions); err != nil {
		return nil, errors.Wrap(err, "unable to decode provider sessions")
	}

	logger.WithFields(log.Fields{"count": len(providerSessions), "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("ListProviderSessions")

	return providerSessions, nil
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
