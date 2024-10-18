package mongo

import (
	"context"
	stdErrs "errors"
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

type DeviceLogsRepository struct {
	*storeStructuredMongo.Repository
}

func (d *DeviceLogsRepository) EnsureIndexes() error {
	return d.CreateAllIndexes(context.Background(),
		[]mongo.IndexModel{
			{
				Keys: bson.D{{Key: "id", Value: 1}},
				Options: options.Index().
					SetUnique(true),
			},
			{
				Keys: bson.D{
					{Key: "userId", Value: 1},
					{Key: "startAtTime", Value: 1},
				},
			},
			{
				Keys: bson.D{
					{Key: "userId", Value: 1},
					{Key: "endAtTime", Value: 1},
				},
			},
			{
				Keys: bson.D{{Key: "startAtTime", Value: 1}},
			},
			{
				Keys: bson.D{{Key: "endAtTime", Value: 1}},
			},
		})
}

func (d *DeviceLogsRepository) List(ctx context.Context, userID string, filter *blob.DeviceLogsFilter, pagination *page.Pagination) (blob.DeviceLogsBlobArray, error) {
	ctx, logger := log.ContextAndLoggerWithFields(ctx, log.Fields{"userId": userID, "filter": filter, "pagination": pagination})

	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errorUserIDMissing
	}
	if filter == nil {
		filter = blob.NewDeviceLogsFilter()
	} else if err := structureValidator.New().Validate(filter); err != nil {
		return nil, errors.Wrap(err, "filter is invalid")
	}
	if pagination == nil {
		pagination = page.NewPagination()
	} else if err := structureValidator.New().Validate(pagination); err != nil {
		return nil, errors.Wrap(err, "pagination is invalid")
	}

	now := time.Now()

	var result blob.DeviceLogsBlobArray
	query := bson.M{
		"userId": userID,
		"deletedTime": bson.M{
			"$exists": false,
		},
	}
	if filter.StartAtTime != nil {
		query["startAtTime"] = bson.M{
			"$gte": *filter.StartAtTime,
		}
	}
	if filter.EndAtTime != nil {
		query["endAtTime"] = bson.M{
			"$lt": *filter.EndAtTime,
		}
	}
	opts := storeStructuredMongo.FindWithPagination(pagination).
		SetSort(bson.M{"createdTime": -1})
	cursor, err := d.Find(ctx, query, opts)
	if err != nil {
		logger.WithError(err).Error("Unable to list device logs")
		return nil, errors.Wrap(err, "unable to list device logs")
	}

	if err = cursor.All(ctx, &result); err != nil {
		return nil, errors.Wrap(err, "unable to decode device logs")
	}

	logger.WithFields(log.Fields{"count": len(result), "duration": time.Since(now) / time.Microsecond}).Debug("List")
	return result, nil
}

func (d *DeviceLogsRepository) Get(ctx context.Context, deviceLogID string) (*blob.DeviceLogsBlob, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if deviceLogID == "" {
		return nil, errors.New("deviceLogID is missing")
	}

	var result blob.DeviceLogsBlob
	query := bson.M{
		"id": deviceLogID,
	}
	err := d.Repository.FindOne(ctx, query).Decode(&result)
	if stdErrs.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (d *DeviceLogsRepository) Create(ctx context.Context, userID string, create *blobStoreStructured.Create) (*blob.DeviceLogsBlob, error) {
	ctx, logger := log.ContextAndLoggerWithFields(ctx, log.Fields{"userId": userID, "create": create})

	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if !user.IsValidID(userID) {
		if userID == "" {
			return nil, errorUserIDMissing
		}
		return nil, errorUserIDNotValid
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
	if !blob.IsValidID(id) {
		if id == "" {
			return nil, errorBlobIDMissing
		}
		return nil, errorBlobIDNotValid
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
		if update.StartAt != nil && !update.StartAt.IsZero() {
			set["startAtTime"] = *update.StartAt
		}
		if update.EndAt != nil && !update.EndAt.IsZero() {
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
	if !blob.IsValidID(id) {
		if id == "" {
			return false, errorBlobIDMissing
		}
		return false, errorBlobIDNotValid
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
