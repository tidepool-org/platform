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

type RestrictedTokenRepository struct {
	*storeStructuredMongo.Repository
}

func (r *RestrictedTokenRepository) EnsureIndexes() error {
	return r.CreateAllIndexes(context.Background(), []mongo.IndexModel{
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
	})
}

func (r *RestrictedTokenRepository) ListUserRestrictedTokens(ctx context.Context, userID string, filter *auth.RestrictedTokenFilter, pagination *page.Pagination) (auth.RestrictedTokens, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	}
	if filter == nil {
		filter = auth.NewRestrictedTokenFilter()
	} else if err := structureValidator.New().Validate(filter); err != nil {
		return nil, errors.Wrap(err, "filter is invalid")
	}
	if pagination == nil {
		pagination = page.NewPagination()
	} else if err := structureValidator.New().Validate(pagination); err != nil {
		return nil, errors.Wrap(err, "pagination is invalid")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": userID, "filter": filter, "pagination": pagination})

	restrictedTokens := auth.RestrictedTokens{}
	selector := bson.M{
		"userId": userID,
	}
	opts := storeStructuredMongo.FindWithPagination(pagination).
		SetSort(bson.M{"createdTime": -1})
	cursor, err := r.Find(ctx, selector, opts)
	logger.WithFields(log.Fields{"count": len(restrictedTokens), "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("ListUserRestrictedTokens")
	if err != nil {
		return nil, errors.Wrap(err, "unable to list user restricted tokens")
	}

	if err = cursor.All(ctx, &restrictedTokens); err != nil {
		return nil, errors.Wrap(err, "unable to decode user restricted tokens")
	}

	if restrictedTokens == nil {
		restrictedTokens = auth.RestrictedTokens{}
	}

	return restrictedTokens, nil
}

func (r *RestrictedTokenRepository) CreateUserRestrictedToken(ctx context.Context, userID string, create *auth.RestrictedTokenCreate) (*auth.RestrictedToken, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}

	restrictedToken, err := auth.NewRestrictedToken(userID, create)
	if err != nil {
		return nil, err
	} else if err = structureValidator.New().Validate(restrictedToken); err != nil {
		return nil, errors.Wrap(err, "restricted token is invalid")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": userID, "create": create})

	_, err = r.InsertOne(ctx, restrictedToken)
	logger.WithFields(log.Fields{"id": restrictedToken.ID, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("CreateUserRestrictedToken")
	if err != nil {
		return nil, errors.Wrap(err, "unable to create user restricted token")
	}

	return restrictedToken, nil
}

func (r *RestrictedTokenRepository) DeleteAllRestrictedTokens(ctx context.Context, userID string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if userID == "" {
		return errors.New("user id is missing")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithField("userId", userID)

	changeInfo, err := r.DeleteMany(ctx, bson.M{"userId": userID})
	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("DeleteAllRestrictedTokens")
	if err != nil {
		return errors.Wrap(err, "unable to delete all restricted tokens")
	}

	return nil
}

func (r *RestrictedTokenRepository) GetRestrictedToken(ctx context.Context, id string) (*auth.RestrictedToken, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithField("id", id)

	restrictedTokens := auth.RestrictedTokens{}
	opts := options.Find().SetLimit(2)
	cursor, err := r.Find(ctx, bson.M{"id": id}, opts)
	logger.WithField("duration", time.Since(now)/time.Microsecond).WithError(err).Debug("GetRestrictedToken")
	if err != nil {
		return nil, errors.Wrap(err, "unable to get restricted token")
	}

	if err = cursor.All(ctx, &restrictedTokens); err != nil {
		return nil, errors.Wrap(err, "unable to decode restricted tokens")
	}

	switch count := len(restrictedTokens); count {
	case 0:
		return nil, nil
	case 1:
		return restrictedTokens[0], nil
	default:
		logger.WithField("count", count).Warnf("Multiple restricted tokens found for id %q", id)
		return restrictedTokens[0], nil
	}
}

func (r *RestrictedTokenRepository) UpdateRestrictedToken(ctx context.Context, id string, update *auth.RestrictedTokenUpdate) (*auth.RestrictedToken, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	}
	if update == nil {
		return nil, errors.New("update is missing")
	} else if err := structureValidator.New().Validate(update); err != nil {
		return nil, errors.Wrap(err, "update is invalid")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"id": id, "update": update})

	set := bson.M{
		"modifiedTime": now,
	}
	unset := bson.M{}
	if update.Paths != nil {
		set["path"] = *update.Paths
	}
	if update.ExpirationTime != nil {
		set["expirationTime"] = *update.ExpirationTime
	}
	changeInfo, err := r.UpdateMany(ctx, bson.M{"id": id}, r.ConstructUpdate(set, unset))
	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("UpdateRestrictedToken")
	if err != nil {
		return nil, errors.Wrap(err, "unable to update restricted token")
	}

	return r.GetRestrictedToken(ctx, id)
}

func (r *RestrictedTokenRepository) DeleteRestrictedToken(ctx context.Context, id string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if id == "" {
		return errors.New("id is missing")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithField("id", id)

	changeInfo, err := r.DeleteMany(ctx, bson.M{"id": id})
	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("DeleteRestrictedToken")
	if err != nil {
		return errors.Wrap(err, "unable to delete restricted token")
	}

	return nil
}
