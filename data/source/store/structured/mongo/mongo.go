package mongo

import (
	"context"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	dataSource "github.com/tidepool-org/platform/data/source"
	dataSourceStoreStructured "github.com/tidepool-org/platform/data/source/store/structured"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/user"
)

type Store struct {
	*storeStructuredMongo.Store
}

func NewStore(config *storeStructuredMongo.Config, logger log.Logger) (*Store, error) {
	store, err := storeStructuredMongo.NewStore(config, logger)
	if err != nil {
		return nil, err
	}

	return &Store{
		Store: store,
	}, nil
}

func (s *Store) EnsureIndexes() error {
	session := s.newSession()
	defer session.Close()
	return session.EnsureIndexes()
}

func (s *Store) NewSession() dataSourceStoreStructured.Session {
	return s.newSession()
}

func (s *Store) newSession() *Session {
	return &Session{
		Session: s.Store.NewSession("data_sources"),
	}
}

type Session struct {
	*storeStructuredMongo.Session
}

func (s *Session) EnsureIndexes() error {
	return s.EnsureAllIndexes([]mgo.Index{
		{Key: []string{"id"}, Background: true, Unique: true},
		{Key: []string{"userId"}, Background: true},
	})
}

func (s *Session) List(ctx context.Context, userID string, filter *dataSource.Filter, pagination *page.Pagination) (dataSource.SourceArray, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	} else if !user.IsValidID(userID) {
		return nil, errors.New("user id is invalid")
	}
	if filter == nil {
		filter = dataSource.NewFilter()
	} else if err := structureValidator.New().Validate(filter); err != nil {
		return nil, errors.Wrap(err, "filter is invalid")
	}
	if pagination == nil {
		pagination = page.NewPagination()
	} else if err := structureValidator.New().Validate(pagination); err != nil {
		return nil, errors.Wrap(err, "pagination is invalid")
	}

	if s.IsClosed() {
		return nil, errors.New("session closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": userID, "filter": filter, "pagination": pagination})

	result := dataSource.SourceArray{}
	query := bson.M{
		"userId": userID,
	}
	if filter.ProviderType != nil {
		query["providerType"] = bson.M{
			"$in": *filter.ProviderType,
		}
	}
	if filter.ProviderName != nil {
		query["providerName"] = bson.M{
			"$in": *filter.ProviderName,
		}
	}
	if filter.ProviderSessionID != nil {
		query["providerSessionId"] = bson.M{
			"$in": *filter.ProviderSessionID,
		}
	}
	if filter.State != nil {
		query["state"] = bson.M{
			"$in": *filter.State,
		}
	}
	err := s.C().Find(query).Sort("-createdTime").Skip(pagination.Page * pagination.Size).Limit(pagination.Size).All(&result)
	if err != nil {
		logger.WithError(err).Error("Unable to list data sources")
		return nil, errors.Wrap(err, "unable to list data sources")
	}

	logger.WithFields(log.Fields{"count": len(result), "duration": time.Since(now) / time.Microsecond}).Debug("List")
	return result, nil
}

func (s *Session) Create(ctx context.Context, userID string, create *dataSource.Create) (*dataSource.Source, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	} else if !user.IsValidID(userID) {
		return nil, errors.New("user id is invalid")
	}
	if create == nil {
		return nil, errors.New("create is missing")
	} else if err := structureValidator.New().Validate(create); err != nil {
		return nil, errors.Wrap(err, "create is invalid")
	}

	if s.IsClosed() {
		return nil, errors.New("session closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": userID, "create": create})

	doc := &dataSource.Source{
		UserID:            pointer.FromString(userID),
		ProviderType:      create.ProviderType,
		ProviderName:      create.ProviderName,
		ProviderSessionID: create.ProviderSessionID,
		State:             create.State,
		CreatedTime:       pointer.FromTime(now),
		Revision:          pointer.FromInt(0),
	}

	var id string
	var err error
	for retry := 0; retry < 3; retry++ {
		id = dataSource.NewID()
		logger = logger.WithField("id", id)

		doc.ID = pointer.FromString(id)
		if err = s.C().Insert(doc); mgo.IsDup(err) {
			logger.WithError(err).Error("Duplicate data source id")
		} else {
			break
		}
	}
	if err != nil {
		logger.WithError(err).Error("Unable to create data source")
		return nil, errors.Wrap(err, "unable to create data source")
	}

	result, err := s.get(logger, id, nil)
	if err != nil {
		return nil, err
	}

	logger.WithField("duration", time.Since(now)/time.Microsecond).Debug("Create")
	return result, nil
}

func (s *Session) DestroyAll(ctx context.Context, userID string) (bool, error) {
	if ctx == nil {
		return false, errors.New("context is missing")
	}
	if userID == "" {
		return false, errors.New("user id is missing")
	} else if !user.IsValidID(userID) {
		return false, errors.New("user id is invalid")
	}

	if s.IsClosed() {
		return false, errors.New("session closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithField("userId", userID)

	query := bson.M{
		"userId": userID,
	}
	changeInfo, err := s.C().RemoveAll(query)
	if err != nil {
		logger.WithError(err).Error("Unable to destroy all data sources")
		return false, errors.Wrap(err, "unable to destroy all data sources")
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("DestroyAll")
	return changeInfo.Removed > 0, nil
}

func (s *Session) Get(ctx context.Context, id string) (*dataSource.Source, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	} else if !dataSource.IsValidID(id) {
		return nil, errors.New("id is invalid")
	}

	if s.IsClosed() {
		return nil, errors.New("session closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithField("id", id)

	result, err := s.get(logger, id, nil)
	if err != nil {
		return nil, err
	}

	logger.WithField("duration", time.Since(now)/time.Microsecond).Debug("Get")
	return result, nil
}

func (s *Session) Update(ctx context.Context, id string, condition *request.Condition, update *dataSource.Update) (*dataSource.Source, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	} else if !dataSource.IsValidID(id) {
		return nil, errors.New("id is invalid")
	}
	if condition == nil {
		condition = request.NewCondition()
	} else if err := structureValidator.New().Validate(condition); err != nil {
		return nil, errors.Wrap(err, "condition is invalid")
	}
	if update == nil {
		return nil, errors.New("update is missing")
	} else if err := structureValidator.New().Validate(update); err != nil {
		return nil, errors.Wrap(err, "update is invalid")
	}

	if s.IsClosed() {
		return nil, errors.New("session closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"id": id, "condition": condition, "update": update})

	if !update.IsEmpty() {
		query := bson.M{
			"id": id,
		}
		if condition.Revision != nil {
			query["revision"] = *condition.Revision
		}
		set := bson.M{
			"modifiedTime": now,
		}
		unset := bson.M{}
		if update.ProviderSessionID != nil {
			set["providerSessionId"] = *update.ProviderSessionID
		}
		if update.State != nil {
			set["state"] = *update.State
			switch *update.State {
			case dataSource.StateDisconnected:
				delete(set, "providerSessionId")
				unset["providerSessionId"] = true
				unset["error"] = true
			case dataSource.StateConnected:
				unset["error"] = true
			}
		}
		if update.Error != nil {
			delete(unset, "error")
			set["error"] = *update.Error
		}
		if update.DataSetIDs != nil {
			set["dataSetIds"] = *update.DataSetIDs
		}
		if update.EarliestDataTime != nil {
			set["earliestDataTime"] = *update.EarliestDataTime
		}
		if update.LatestDataTime != nil {
			set["latestDataTime"] = *update.LatestDataTime
		}
		if update.LastImportTime != nil {
			set["lastImportTime"] = *update.LastImportTime
		}
		changeInfo, err := s.C().UpdateAll(query, s.ConstructUpdate(set, unset))
		if err != nil {
			logger.WithError(err).Error("Unable to update data source")
			return nil, errors.Wrap(err, "unable to update data source")
		} else if changeInfo.Matched > 0 {
			condition = nil
		} else {
			update = nil
		}

		logger = logger.WithField("changeInfo", changeInfo)
	}

	var result *dataSource.Source
	if update != nil {
		var err error
		if result, err = s.get(logger, id, condition); err != nil {
			return nil, err
		}
	}

	logger.WithField("duration", time.Since(now)/time.Microsecond).Debug("Update")
	return result, nil
}

func (s *Session) Destroy(ctx context.Context, id string, condition *request.Condition) (bool, error) {
	if ctx == nil {
		return false, errors.New("context is missing")
	}
	if id == "" {
		return false, errors.New("id is missing")
	} else if !dataSource.IsValidID(id) {
		return false, errors.New("id is invalid")
	}
	if condition == nil {
		condition = request.NewCondition()
	} else if err := structureValidator.New().Validate(condition); err != nil {
		return false, errors.Wrap(err, "condition is invalid")
	}

	if s.IsClosed() {
		return false, errors.New("session closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"id": id, "condition": condition})

	query := bson.M{
		"id": id,
	}
	if condition.Revision != nil {
		query["revision"] = *condition.Revision
	}
	changeInfo, err := s.C().RemoveAll(query)
	if err != nil {
		logger.WithError(err).Error("Unable to destroy data source")
		return false, errors.Wrap(err, "unable to destroy data source")
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("Destroy")
	return changeInfo.Removed > 0, nil
}

func (s *Session) get(logger log.Logger, id string, condition *request.Condition) (*dataSource.Source, error) {
	results := dataSource.SourceArray{}
	query := bson.M{
		"id": id,
	}
	if condition != nil && condition.Revision != nil {
		query["revision"] = *condition.Revision
	}
	err := s.C().Find(query).Limit(2).All(&results)
	if err != nil {
		logger.WithError(err).Error("Unable to get data source")
		return nil, errors.Wrap(err, "unable to get data source")
	}

	var result *dataSource.Source
	switch len(results) {
	case 0:
		return nil, nil
	case 1:
		result = results[0]
	default:
		logger.Error("Multiple data sources found")
		result = results[0]
	}

	if result.Revision == nil {
		result.Revision = pointer.FromInt(0)
	}

	return result, nil
}
