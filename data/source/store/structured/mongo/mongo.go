package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

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

func NewStore(config *storeStructuredMongo.Config) (*Store, error) {
	store, err := storeStructuredMongo.NewStore(config)
	if err != nil {
		return nil, err
	}

	return &Store{
		Store: store,
	}, nil
}

func (s *Store) EnsureIndexes() error {
	repository := s.newRepository()
	return repository.EnsureIndexes()
}

func (s *Store) NewDataSourcesRepository() dataSourceStoreStructured.DataSourcesRepository {
	return s.newRepository()
}

func (s *Store) newRepository() *DataSourcesRepository {
	return &DataSourcesRepository{
		s.Store.GetRepository("data_sources"),
	}
}

type DataSourcesRepository struct {
	*storeStructuredMongo.Repository
}

func (c *DataSourcesRepository) EnsureIndexes() error {
	return c.CreateAllIndexes(context.Background(), []mongo.IndexModel{
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
			Keys: bson.D{
				{Key: "providerName", Value: 1},
				{Key: "providerExternalId", Value: 1},
			},
		},
	})
}

func (c *DataSourcesRepository) List(ctx context.Context, userID string, filter *dataSource.Filter, pagination *page.Pagination) (dataSource.SourceArray, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if filter == nil {
		filter = dataSource.NewFilter()
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

	result := dataSource.SourceArray{}
	query := bson.M{}

	if userID != "" {
		if !user.IsValidID(userID) {
			return nil, errors.New("user id is invalid")
		}
		query["userId"] = userID
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
	if filter.ProviderExternalID != nil {
		query["providerExternalId"] = bson.M{
			"$in": *filter.ProviderExternalID,
		}
	}
	if filter.State != nil {
		query["state"] = bson.M{
			"$in": *filter.State,
		}
	}
	opts := storeStructuredMongo.FindWithPagination(pagination).
		SetSort(bson.M{"createdTime": -1})
	cursor, err := c.Find(ctx, query, opts)

	if err != nil {
		logger.WithError(err).Error("Unable to list data sources")
		return nil, errors.Wrap(err, "unable to list data sources")
	}

	if err = cursor.All(ctx, &result); err != nil {
		return nil, errors.Wrap(err, "unable to decode data sources")
	}

	logger.WithFields(log.Fields{"count": len(result), "duration": time.Since(now) / time.Microsecond}).Debug("List")
	return result, nil
}

func (c *DataSourcesRepository) Create(ctx context.Context, userID string, create *dataSource.Create) (*dataSource.Source, error) {
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
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(create); err != nil {
		return nil, errors.Wrap(err, "create is invalid")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": userID, "create": create})

	doc := &dataSource.Source{
		UserID:             pointer.FromString(userID),
		ProviderType:       create.ProviderType,
		ProviderName:       create.ProviderName,
		ProviderSessionID:  create.ProviderSessionID,
		ProviderExternalID: create.ProviderExternalID,
		State:              pointer.FromString(dataSource.StateDisconnected),
		Metadata:           create.Metadata,
		CreatedTime:        pointer.FromTime(now),
		Revision:           pointer.FromInt(0),
	}

	var id string
	var err error
	for retry := 0; retry < 3; retry++ {
		id = dataSource.NewID()
		logger = logger.WithField("id", id)

		doc.ID = pointer.FromString(id)
		if _, err = c.InsertOne(ctx, doc); storeStructuredMongo.IsDup(err) {
			logger.WithError(err).Error("Duplicate data source id")
		} else {
			break
		}
	}
	if err != nil {
		logger.WithError(err).Error("Unable to create data source")
		return nil, errors.Wrap(err, "unable to create data source")
	}

	result, err := c.get(ctx, logger, id, nil)
	if err != nil {
		return nil, err
	}

	logger.WithField("duration", time.Since(now)/time.Microsecond).Debug("Create")
	return result, nil
}

func (c *DataSourcesRepository) DestroyAll(ctx context.Context, userID string) (bool, error) {
	if ctx == nil {
		return false, errors.New("context is missing")
	}
	if userID == "" {
		return false, errors.New("user id is missing")
	} else if !user.IsValidID(userID) {
		return false, errors.New("user id is invalid")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithField("userId", userID)

	query := bson.M{
		"userId": userID,
	}
	changeInfo, err := c.DeleteMany(ctx, query)
	if err != nil {
		logger.WithError(err).Error("Unable to destroy all data sources")
		return false, errors.Wrap(err, "unable to destroy all data sources")
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("DestroyAll")
	return changeInfo.DeletedCount > 0, nil
}

func (c *DataSourcesRepository) Get(ctx context.Context, id string) (*dataSource.Source, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	} else if !dataSource.IsValidID(id) {
		return nil, errors.New("id is invalid")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithField("id", id)

	result, err := c.get(ctx, logger, id, nil)
	if err != nil {
		return nil, err
	}

	logger.WithField("duration", time.Since(now)/time.Microsecond).Debug("Get")
	return result, nil
}

func (c *DataSourcesRepository) Update(ctx context.Context, id string, condition *request.Condition, update *dataSource.Update) (*dataSource.Source, error) {
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
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(condition); err != nil {
		return nil, errors.Wrap(err, "condition is invalid")
	}
	if update == nil {
		return nil, errors.New("update is missing")
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(update); err != nil {
		return nil, errors.Wrap(err, "update is invalid")
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
		if update.State != nil {
			set["state"] = *update.State
			switch *update.State {
			case dataSource.StateDisconnected:
				unset["providerSessionId"] = true
				unset["error"] = true
			case dataSource.StateConnected:
				unset["error"] = true
			}
		}
		if update.Metadata != nil {
			set["metadata"] = update.Metadata
		}
		if update.ProviderSessionID != nil {
			delete(unset, "providerSessionId")
			set["providerSessionId"] = *update.ProviderSessionID
		}
		if update.ProviderExternalID != nil {
			set["providerExternalId"] = *update.ProviderExternalID
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
		changeInfo, err := c.UpdateMany(ctx, query, c.ConstructUpdate(set, unset))
		if err != nil {
			logger.WithError(err).Error("Unable to update data source")
			return nil, errors.Wrap(err, "unable to update data source")
		} else if changeInfo.MatchedCount > 0 {
			condition = nil
		} else {
			update = nil
		}

		logger = logger.WithField("changeInfo", changeInfo)
	}

	var result *dataSource.Source
	if update != nil {
		var err error
		if result, err = c.get(ctx, logger, id, condition); err != nil {
			return nil, err
		}
	}

	logger.WithField("duration", time.Since(now)/time.Microsecond).Debug("Update")
	return result, nil
}

func (c *DataSourcesRepository) Destroy(ctx context.Context, id string, condition *request.Condition) (bool, error) {
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
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(condition); err != nil {
		return false, errors.Wrap(err, "condition is invalid")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"id": id, "condition": condition})

	query := bson.M{
		"id": id,
	}
	if condition.Revision != nil {
		query["revision"] = *condition.Revision
	}
	changeInfo, err := c.DeleteMany(ctx, query)
	if err != nil {
		logger.WithError(err).Error("Unable to destroy data source")
		return false, errors.Wrap(err, "unable to destroy data source")
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("Destroy")
	return changeInfo.DeletedCount > 0, nil
}

func (c *DataSourcesRepository) get(ctx context.Context, logger log.Logger, id string, condition *request.Condition) (*dataSource.Source, error) {
	var result *dataSource.Source
	query := bson.M{
		"id": id,
	}
	if condition != nil && condition.Revision != nil {
		query["revision"] = *condition.Revision
	}
	err := c.FindOne(ctx, query).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		logger.WithError(err).Error("Unable to get data source")
		return nil, errors.Wrap(err, "unable to decode data source")
	}

	if result.Revision == nil {
		result.Revision = pointer.FromInt(0)
	}

	return result, nil
}
