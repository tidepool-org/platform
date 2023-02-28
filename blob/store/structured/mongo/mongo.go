package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/blob"
	blobStoreStructured "github.com/tidepool-org/platform/blob/store/structured"
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

func (s *Store) NewBlobRepository() blobStoreStructured.BlobRepository {
	return s.newRepository()
}

func (s *Store) newRepository() *BlobRepository {
	return &BlobRepository{
		s.Store.GetRepository("blobs"),
	}
}

func (s *Store) NewDeviceLogsRepository() blobStoreStructured.DeviceLogsRepository {
	return &DeviceLogsRepository{
		s.Store.GetRepository("deviceLogs"),
	}
}

type DeviceLogsRepository struct {
	*storeStructuredMongo.Repository
}

func (d *DeviceLogsRepository) Create(ctx context.Context, userID string, create *blobStoreStructured.Create) (*blob.DeviceLogsBlob, error) {
	ctx, logger := log.ContextAndLoggerWithFields(ctx, log.Fields{"userId": userID, "create": create})

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

	now := time.Now()

	doc := &blob.DeviceLogsBlob{
		UserID:      pointer.FromString(userID),
		MediaType:   create.MediaType,
		CreatedTime: pointer.FromTime(now.Truncate(time.Millisecond)),
		Revision:    pointer.FromInt(0),
	}

	var id string
	var err error
	for retry := 0; retry < 3; retry++ {
		id = blob.NewID()
		logger = logger.WithField("id", id)

		doc.ID = pointer.FromString(id)
		if _, err = d.InsertOne(ctx, doc); storeStructuredMongo.IsDup(err) {
			logger.WithError(err).Error("Duplicate blob id")
		} else {
			break
		}
	}
	if err != nil {
		logger.WithError(err).Error("Unable to create blob")
		return nil, errors.Wrap(err, "unable to create blob")
	}

	result, err := d.get(ctx, logger, id, nil)
	if err != nil {
		return nil, err
	}

	logger.WithField("duration", time.Since(now)/time.Microsecond).Debug("Create")
	return result, nil
}

func (d *DeviceLogsRepository) Update(ctx context.Context, id string, condition *request.Condition, update *blobStoreStructured.DeviceLogsUpdate) (*blob.DeviceLogsBlob, error) {
	ctx, logger := log.ContextAndLoggerWithFields(ctx, log.Fields{"id": id, "condition": condition, "update": update})

	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	} else if !blob.IsValidID(id) {
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

	now := time.Now()

	if !update.IsEmpty() {
		query := bson.M{
			"id": id,
			"deletedTime": bson.M{
				"$exists": false,
			},
		}
		if condition.Revision != nil {
			query["revision"] = *condition.Revision
		}
		set := bson.M{
			"modifiedTime": now.Truncate(time.Millisecond),
		}
		unset := bson.M{}
		if update.MediaType != nil {
			set["mediaType"] = *update.MediaType
		}
		if update.DigestMD5 != nil {
			set["digestMD5"] = *update.DigestMD5
		}
		if update.Size != nil {
			set["size"] = *update.Size
		}
		if update.StartAt != nil {
			set["startAtTime"] = *update.StartAt
		}
		if update.EndAt != nil {
			set["endAtTime"] = *update.EndAt
		}
		changeInfo, err := d.UpdateMany(ctx, query, d.ConstructUpdate(set, unset))
		if err != nil {
			logger.WithError(err).Error("Unable to update blob")
			return nil, errors.Wrap(err, "unable to update blob")
		} else if changeInfo.MatchedCount > 0 {
			condition = nil
		} else {
			update = nil
		}

		logger = logger.WithField("changeInfo", changeInfo)
	}

	var result *blob.DeviceLogsBlob
	if update != nil {
		var err error
		if result, err = d.get(ctx, logger, id, condition); err != nil {
			return nil, err
		}
	}

	logger.WithField("duration", time.Since(now)/time.Microsecond).Debug("Update")
	return result, nil
}

func (d *DeviceLogsRepository) Destroy(ctx context.Context, id string, condition *request.Condition) (bool, error) {
	ctx, logger := log.ContextAndLoggerWithFields(ctx, log.Fields{"id": id, "condition": condition})

	if ctx == nil {
		return false, errors.New("context is missing")
	}
	if id == "" {
		return false, errors.New("id is missing")
	} else if !blob.IsValidID(id) {
		return false, errors.New("id is invalid")
	}
	if condition == nil {
		condition = request.NewCondition()
	} else if err := structureValidator.New().Validate(condition); err != nil {
		return false, errors.Wrap(err, "condition is invalid")
	}

	now := time.Now()

	query := bson.M{
		"id": id,
	}
	if condition.Revision != nil {
		query["revision"] = *condition.Revision
	}
	changeInfo, err := d.DeleteMany(ctx, query)
	if err != nil {
		logger.WithError(err).Error("Unable to destroy blob")
		return false, errors.Wrap(err, "unable to destroy blob")
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("Destroy")
	return changeInfo.DeletedCount > 0, nil
}

func (d *DeviceLogsRepository) get(ctx context.Context, logger log.Logger, id string, condition *request.Condition, queryModifiers ...storeStructuredMongo.QueryModifier) (*blob.DeviceLogsBlob, error) {
	logger = logger.WithFields(log.Fields{"id": id, "condition": condition})

	var result *blob.DeviceLogsBlob
	query := bson.M{
		"id": id,
	}
	query = storeStructuredMongo.ModifyQuery(query, queryModifiers...)
	err := d.FindOne(ctx, query).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		logger.WithError(err).Error("Unable to get device logs blob")
		return nil, errors.Wrap(err, "unable to get device logs blob")
	}
	return result, nil
}

type BlobRepository struct {
	*storeStructuredMongo.Repository
}

func (b *BlobRepository) EnsureIndexes() error {
	return b.CreateAllIndexes(context.Background(),
		[]mongo.IndexModel{
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
				Keys: bson.D{{Key: "mediaType", Value: 1}},
				Options: options.Index().
					SetBackground(true),
			},
			{
				Keys: bson.D{{Key: "status", Value: 1}},
				Options: options.Index().
					SetBackground(true),
			},
		})
}

func (b *BlobRepository) List(ctx context.Context, userID string, filter *blob.Filter, pagination *page.Pagination) (blob.BlobArray, error) {
	ctx, logger := log.ContextAndLoggerWithFields(ctx, log.Fields{"userId": userID, "filter": filter, "pagination": pagination})

	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	} else if !user.IsValidID(userID) {
		return nil, errors.New("user id is invalid")
	}
	if filter == nil {
		filter = blob.NewFilter()
	} else if err := structureValidator.New().Validate(filter); err != nil {
		return nil, errors.Wrap(err, "filter is invalid")
	}
	if pagination == nil {
		pagination = page.NewPagination()
	} else if err := structureValidator.New().Validate(pagination); err != nil {
		return nil, errors.Wrap(err, "pagination is invalid")
	}

	now := time.Now()

	var status []string
	if filter.Status != nil {
		status = *filter.Status
	} else {
		status = []string{blob.StatusAvailable}
	}

	result := blob.BlobArray{}
	query := bson.M{
		"userId": userID,
		"status": bson.M{
			"$in": status,
		},
		"deletedTime": bson.M{
			"$exists": false,
		},
	}
	if filter.MediaType != nil {
		query["mediaType"] = bson.M{
			"$in": *filter.MediaType,
		}
	}
	opts := storeStructuredMongo.FindWithPagination(pagination).
		SetSort(bson.M{"createdTime": -1})
	cursor, err := b.Find(ctx, query, opts)
	if err != nil {
		logger.WithError(err).Error("Unable to list blobs")
		return nil, errors.Wrap(err, "unable to list blobs")
	}

	if err = cursor.All(ctx, &result); err != nil {
		return nil, errors.Wrap(err, "unable to decode blobs")
	}

	logger.WithFields(log.Fields{"count": len(result), "duration": time.Since(now) / time.Microsecond}).Debug("List")
	return result, nil
}

func (b *BlobRepository) Create(ctx context.Context, userID string, create *blobStoreStructured.Create) (*blob.Blob, error) {
	ctx, logger := log.ContextAndLoggerWithFields(ctx, log.Fields{"userId": userID, "create": create})

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

	now := time.Now()

	doc := &blob.Blob{
		UserID:      pointer.FromString(userID),
		MediaType:   create.MediaType,
		Status:      pointer.FromString(blob.StatusCreated),
		CreatedTime: pointer.FromTime(now.Truncate(time.Millisecond)),
		Revision:    pointer.FromInt(0),
	}

	var id string
	var err error
	for retry := 0; retry < 3; retry++ {
		id = blob.NewID()
		logger = logger.WithField("id", id)

		doc.ID = pointer.FromString(id)
		if _, err = b.InsertOne(ctx, doc); storeStructuredMongo.IsDup(err) {
			logger.WithError(err).Error("Duplicate blob id")
		} else {
			break
		}
	}
	if err != nil {
		logger.WithError(err).Error("Unable to create blob")
		return nil, errors.Wrap(err, "unable to create blob")
	}

	result, err := b.get(ctx, logger, id, nil)
	if err != nil {
		return nil, err
	}

	logger.WithField("duration", time.Since(now)/time.Microsecond).Debug("Create")
	return result, nil
}

func (b *BlobRepository) DeleteAll(ctx context.Context, userID string) (bool, error) {
	ctx, logger := log.ContextAndLoggerWithField(ctx, "userId", userID)

	if ctx == nil {
		return false, errors.New("context is missing")
	}
	if userID == "" {
		return false, errors.New("user id is missing")
	} else if !user.IsValidID(userID) {
		return false, errors.New("user id is invalid")
	}

	now := time.Now()

	query := bson.M{
		"userId": userID,
	}
	set := bson.M{
		"modifiedTime": now.Truncate(time.Millisecond),
		"deletedTime":  now.Truncate(time.Millisecond),
	}
	unset := bson.M{}
	changeInfo, err := b.UpdateMany(ctx, query, b.ConstructUpdate(set, unset))
	if err != nil {
		logger.WithError(err).Error("Unable to delete all blobs")
		return false, errors.Wrap(err, "unable to delete all blobs")
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("DeleteAll")
	return changeInfo.ModifiedCount > 0, nil
}

func (b *BlobRepository) DestroyAll(ctx context.Context, userID string) (bool, error) {
	ctx, logger := log.ContextAndLoggerWithField(ctx, "userId", userID)

	if ctx == nil {
		return false, errors.New("context is missing")
	}
	if userID == "" {
		return false, errors.New("user id is missing")
	} else if !user.IsValidID(userID) {
		return false, errors.New("user id is invalid")
	}

	now := time.Now()

	query := bson.M{
		"userId": userID,
	}
	changeInfo, err := b.DeleteMany(ctx, query)
	if err != nil {
		logger.WithError(err).Error("Unable to destroy all blobs")
		return false, errors.Wrap(err, "unable to destroy all blobs")
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("DestroyAll")
	return changeInfo.DeletedCount > 0, nil
}

func (b *BlobRepository) Get(ctx context.Context, id string, condition *request.Condition) (*blob.Blob, error) {
	ctx, logger := log.ContextAndLoggerWithFields(ctx, log.Fields{"id": id, "condition": condition})

	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	} else if !blob.IsValidID(id) {
		return nil, errors.New("id is invalid")
	}
	if condition == nil {
		condition = request.NewCondition()
	} else if err := structureValidator.New().Validate(condition); err != nil {
		return nil, errors.Wrap(err, "condition is invalid")
	}

	now := time.Now()

	result, err := b.get(ctx, logger, id, condition, storeStructuredMongo.NotDeleted)
	if err != nil {
		return nil, err
	}

	logger.WithField("duration", time.Since(now)/time.Microsecond).Debug("Get")
	return result, nil
}

func (b *BlobRepository) Update(ctx context.Context, id string, condition *request.Condition, update *blobStoreStructured.Update) (*blob.Blob, error) {
	ctx, logger := log.ContextAndLoggerWithFields(ctx, log.Fields{"id": id, "condition": condition, "update": update})

	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	} else if !blob.IsValidID(id) {
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

	now := time.Now()

	if !update.IsEmpty() {
		query := bson.M{
			"id": id,
			"deletedTime": bson.M{
				"$exists": false,
			},
		}
		if condition.Revision != nil {
			query["revision"] = *condition.Revision
		}
		set := bson.M{
			"modifiedTime": now.Truncate(time.Millisecond),
		}
		unset := bson.M{}
		if update.MediaType != nil {
			set["mediaType"] = *update.MediaType
		}
		if update.DigestMD5 != nil {
			set["digestMD5"] = *update.DigestMD5
		}
		if update.Size != nil {
			set["size"] = *update.Size
		}
		if update.Status != nil {
			set["status"] = *update.Status
		}
		changeInfo, err := b.UpdateMany(ctx, query, b.ConstructUpdate(set, unset))
		if err != nil {
			logger.WithError(err).Error("Unable to update blob")
			return nil, errors.Wrap(err, "unable to update blob")
		} else if changeInfo.MatchedCount > 0 {
			condition = nil
		} else {
			update = nil
		}

		logger = logger.WithField("changeInfo", changeInfo)
	}

	var result *blob.Blob
	if update != nil {
		var err error
		if result, err = b.get(ctx, logger, id, condition); err != nil {
			return nil, err
		}
	}

	logger.WithField("duration", time.Since(now)/time.Microsecond).Debug("Update")
	return result, nil
}

func (b *BlobRepository) Delete(ctx context.Context, id string, condition *request.Condition) (bool, error) {
	ctx, logger := log.ContextAndLoggerWithFields(ctx, log.Fields{"id": id, "condition": condition})

	if ctx == nil {
		return false, errors.New("context is missing")
	}
	if id == "" {
		return false, errors.New("id is missing")
	} else if !blob.IsValidID(id) {
		return false, errors.New("id is invalid")
	}
	if condition == nil {
		condition = request.NewCondition()
	} else if err := structureValidator.New().Validate(condition); err != nil {
		return false, errors.Wrap(err, "condition is invalid")
	}

	now := time.Now()

	query := bson.M{
		"id": id,
	}
	if condition.Revision != nil {
		query["revision"] = *condition.Revision
	}
	set := bson.M{
		"modifiedTime": now.Truncate(time.Millisecond),
		"deletedTime":  now.Truncate(time.Millisecond),
	}
	unset := bson.M{}
	changeInfo, err := b.UpdateMany(ctx, query, b.ConstructUpdate(set, unset))
	if err != nil {
		logger.WithError(err).Error("Unable to delete blob")
		return false, errors.Wrap(err, "unable to delete blob")
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("Delete")
	return changeInfo.ModifiedCount > 0, nil
}

func (b *BlobRepository) Destroy(ctx context.Context, id string, condition *request.Condition) (bool, error) {
	ctx, logger := log.ContextAndLoggerWithFields(ctx, log.Fields{"id": id, "condition": condition})

	if ctx == nil {
		return false, errors.New("context is missing")
	}
	if id == "" {
		return false, errors.New("id is missing")
	} else if !blob.IsValidID(id) {
		return false, errors.New("id is invalid")
	}
	if condition == nil {
		condition = request.NewCondition()
	} else if err := structureValidator.New().Validate(condition); err != nil {
		return false, errors.Wrap(err, "condition is invalid")
	}

	now := time.Now()

	query := bson.M{
		"id": id,
	}
	if condition.Revision != nil {
		query["revision"] = *condition.Revision
	}
	changeInfo, err := b.DeleteMany(ctx, query)
	if err != nil {
		logger.WithError(err).Error("Unable to destroy blob")
		return false, errors.Wrap(err, "unable to destroy blob")
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("Destroy")
	return changeInfo.DeletedCount > 0, nil
}

func (c *BlobRepository) get(ctx context.Context, logger log.Logger, id string, condition *request.Condition, queryModifiers ...storeStructuredMongo.QueryModifier) (*blob.Blob, error) {
	logger = logger.WithFields(log.Fields{"id": id, "condition": condition})

	var result *blob.Blob
	query := bson.M{
		"id": id,
	}
	if condition != nil && condition.Revision != nil {
		query["revision"] = *condition.Revision
	}
	query = storeStructuredMongo.ModifyQuery(query, queryModifiers...)
	err := c.FindOne(ctx, query).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		logger.WithError(err).Error("Unable to get blob")
		return nil, errors.Wrap(err, "unable to get blob")
	}

	if result.Revision == nil {
		result.Revision = pointer.FromInt(0)
	}

	return result, nil
}
