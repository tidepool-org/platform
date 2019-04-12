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
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type ProviderSessionSession struct {
	*storeStructuredMongo.Session
}

func (p *ProviderSessionSession) EnsureIndexes() error {
	return p.EnsureAllIndexes([]mgo.Index{
		{Key: []string{"id"}, Unique: true, Background: true},
		{Key: []string{"userId"}, Background: true},
		{Key: []string{"userId", "type", "name"}, Unique: true, Background: true},
	})
}

func (p *ProviderSessionSession) ListUserProviderSessions(ctx context.Context, userID string, filter *auth.ProviderSessionFilter, pagination *page.Pagination) (auth.ProviderSessions, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	}
	if filter == nil {
		filter = auth.NewProviderSessionFilter()
	} else if err := structureValidator.New().Validate(filter); err != nil {
		return nil, errors.Wrap(err, "filter is invalid")
	}
	if pagination == nil {
		pagination = page.NewPagination()
	} else if err := structureValidator.New().Validate(pagination); err != nil {
		return nil, errors.Wrap(err, "pagination is invalid")
	}

	if p.IsClosed() {
		return nil, errors.New("session closed")
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
	err := p.C().Find(selector).Sort("-createdTime").Skip(pagination.Page * pagination.Size).Limit(pagination.Size).All(&providerSessions)
	logger.WithFields(log.Fields{"count": len(providerSessions), "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("ListUserProviderSessions")
	if err != nil {
		return nil, errors.Wrap(err, "unable to list user provider sessions")
	}

	if providerSessions == nil {
		providerSessions = auth.ProviderSessions{}
	}

	return providerSessions, nil
}

func (p *ProviderSessionSession) CreateUserProviderSession(ctx context.Context, userID string, create *auth.ProviderSessionCreate) (*auth.ProviderSession, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}

	providerSession, err := auth.NewProviderSession(userID, create)
	if err != nil {
		return nil, err
	} else if err = structureValidator.New().Validate(providerSession); err != nil {
		return nil, errors.Wrap(err, "provider session is invalid")
	}

	if p.IsClosed() {
		return nil, errors.New("session closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": userID, "create": create})

	err = p.C().Insert(providerSession)
	logger.WithFields(log.Fields{"id": providerSession.ID, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("CreateUserProviderSession")
	if err != nil {
		return nil, errors.Wrap(err, "unable to create user provider session")
	}

	return providerSession, nil
}

func (p *ProviderSessionSession) DeleteAllProviderSessions(ctx context.Context, userID string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if userID == "" {
		return errors.New("user id is missing")
	}

	if p.IsClosed() {
		return errors.New("session closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithField("userId", userID)

	changeInfo, err := p.C().RemoveAll(bson.M{"userId": userID})
	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("DeleteAllProviderSessions")
	if err != nil {
		return errors.Wrap(err, "unable to delete all provider sessions")
	}

	return nil
}

func (p *ProviderSessionSession) GetProviderSession(ctx context.Context, id string) (*auth.ProviderSession, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	}

	if p.IsClosed() {
		return nil, errors.New("session closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithField("id", id)

	providerSessions := auth.ProviderSessions{}
	err := p.C().Find(bson.M{"id": id}).Limit(2).All(&providerSessions)
	logger.WithField("duration", time.Since(now)/time.Microsecond).WithError(err).Debug("GetProviderSession")
	if err != nil {
		return nil, errors.Wrap(err, "unable to get provider session")
	}

	switch count := len(providerSessions); count {
	case 0:
		return nil, nil
	case 1:
		return providerSessions[0], nil
	default:
		logger.WithField("count", count).Warnf("Multiple provider sessions found for id %q", id)
		return providerSessions[0], nil
	}
}

func (p *ProviderSessionSession) UpdateProviderSession(ctx context.Context, id string, update *auth.ProviderSessionUpdate) (*auth.ProviderSession, error) {
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

	if p.IsClosed() {
		return nil, errors.New("session closed")
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
	changeInfo, err := p.C().UpdateAll(bson.M{"id": id}, p.ConstructUpdate(set, unset))
	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("UpdateProviderSession")
	if err != nil {
		return nil, errors.Wrap(err, "unable to update provider session")
	}

	return p.GetProviderSession(ctx, id)
}

func (p *ProviderSessionSession) DeleteProviderSession(ctx context.Context, id string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if id == "" {
		return errors.New("id is missing")
	}

	if p.IsClosed() {
		return errors.New("session closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithField("id", id)

	changeInfo, err := p.C().RemoveAll(bson.M{"id": id})
	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("DeleteProviderSession")
	if err != nil {
		return errors.Wrap(err, "unable to delete provider session")
	}

	return nil
}
