package mongo

import (
	"context"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/store/mongo"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type RestrictedTokenSession struct {
	*mongo.Session
}

func (r *RestrictedTokenSession) EnsureIndexes() error {
	return r.EnsureAllIndexes([]mgo.Index{
		{Key: []string{"id"}, Unique: true, Background: true},
		{Key: []string{"userId"}, Background: true},
	})
}

func (r *RestrictedTokenSession) ListUserRestrictedTokens(ctx context.Context, userID string, filter *auth.RestrictedTokenFilter, pagination *page.Pagination) (auth.RestrictedTokens, error) {
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

	if r.IsClosed() {
		return nil, errors.New("session closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": userID, "filter": filter, "pagination": pagination})

	restrictedTokens := auth.RestrictedTokens{}
	selector := bson.M{
		"userId": userID,
	}
	err := r.C().Find(selector).Sort("-createdTime").Skip(pagination.Page * pagination.Size).Limit(pagination.Size).All(&restrictedTokens)
	logger.WithFields(log.Fields{"count": len(restrictedTokens), "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("ListUserRestrictedTokens")
	if err != nil {
		return nil, errors.Wrap(err, "unable to list user restricted tokens")
	}

	if restrictedTokens == nil {
		restrictedTokens = auth.RestrictedTokens{}
	}

	return restrictedTokens, nil
}

func (r *RestrictedTokenSession) CreateUserRestrictedToken(ctx context.Context, userID string, create *auth.RestrictedTokenCreate) (*auth.RestrictedToken, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}

	restrictedToken, err := auth.NewRestrictedToken(userID, create)
	if err != nil {
		return nil, err
	} else if err = structureValidator.New().Validate(restrictedToken); err != nil {
		return nil, errors.Wrap(err, "restricted token is invalid")
	}

	if r.IsClosed() {
		return nil, errors.New("session closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": userID, "create": create})

	err = r.C().Insert(restrictedToken)
	logger.WithFields(log.Fields{"id": restrictedToken.ID, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("CreateUserRestrictedToken")
	if err != nil {
		return nil, errors.Wrap(err, "unable to create user restricted token")
	}

	return restrictedToken, nil
}

func (r *RestrictedTokenSession) GetRestrictedToken(ctx context.Context, id string) (*auth.RestrictedToken, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	}

	if r.IsClosed() {
		return nil, errors.New("session closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithField("id", id)

	restrictedTokens := auth.RestrictedTokens{}
	err := r.C().Find(bson.M{"id": id}).Limit(2).All(&restrictedTokens)
	logger.WithField("duration", time.Since(now)/time.Microsecond).WithError(err).Debug("GetRestrictedToken")
	if err != nil {
		return nil, errors.Wrap(err, "unable to get restricted token")
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

func (r *RestrictedTokenSession) UpdateRestrictedToken(ctx context.Context, id string, update *auth.RestrictedTokenUpdate) (*auth.RestrictedToken, error) {
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

	if r.IsClosed() {
		return nil, errors.New("session closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"id": id, "update": update})

	set := bson.M{
		"modifiedTime": now.Truncate(time.Second),
	}
	unset := bson.M{}
	if update.Paths != nil {
		set["path"] = *update.Paths
	}
	if update.ExpirationTime != nil {
		set["expirationTime"] = *update.ExpirationTime
	}
	changeInfo, err := r.C().UpdateAll(bson.M{"id": id}, r.ConstructUpdate(set, unset))
	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("UpdateRestrictedToken")
	if err != nil {
		return nil, errors.Wrap(err, "unable to update restricted token")
	}

	return r.GetRestrictedToken(ctx, id)
}

func (r *RestrictedTokenSession) DeleteRestrictedToken(ctx context.Context, id string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if id == "" {
		return errors.New("id is missing")
	}

	if r.IsClosed() {
		return errors.New("session closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithField("id", id)

	changeInfo, err := r.C().RemoveAll(bson.M{"id": id})
	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("DeleteRestrictedToken")
	if err != nil {
		return errors.Wrap(err, "unable to delete restricted token")
	}

	return nil
}
