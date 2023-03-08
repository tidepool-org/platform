package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/pointer"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type DataSetRepository struct {
	*storeStructuredMongo.Repository
}

func (d *DataSetRepository) EnsureIndexes() error {
	modifiedTime, err := time.Parse(time.RFC3339, ModifiedTimeIndexRaw)
	if err != nil {
		return err
	}
	// Note "type" field isn't really needed because datasets/uploads are
	// always type == "upload" but this is just to keep the original queries
	// untouched.
	return d.CreateAllIndexes(context.Background(), []mongo.IndexModel{
		// Additional indexes are also created in `tide-whisperer` and `jellyfish`
		{
			Keys: bson.D{
				{Key: "_userId", Value: 1},
				{Key: "_active", Value: 1},
				{Key: "type", Value: 1},
				{Key: "time", Value: -1},
			},
			Options: options.Index().
				SetBackground(true).
				SetName("UserIdTypeWeighted_v2"),
		},
		{
			Keys: bson.D{
				{Key: "_userId", Value: 1},
				{Key: "_active", Value: 1},
				{Key: "type", Value: 1},
				{Key: "modifiedTime", Value: 1},
			},
			Options: options.Index().
				SetBackground(true).
				SetPartialFilterExpression(bson.D{
					{
						Key: "modifiedTime",
						Value: bson.D{
							{Key: "$gt", Value: modifiedTime},
						},
					},
				}).
				SetName("UserIdTypeModifiedTime"),
		},
		{
			Keys: bson.D{
				{Key: "origin.id", Value: 1},
				{Key: "type", Value: 1},
				{Key: "deletedTime", Value: -1},
				{Key: "_active", Value: 1},
			},
			Options: options.Index().
				SetBackground(true).
				SetName("OriginId"),
		},
		{
			Keys: bson.D{
				{Key: "uploadId", Value: 1},
			},
			Options: options.Index().
				SetUnique(true).
				SetPartialFilterExpression(bson.D{{Key: "type", Value: "upload"}}).
				SetName("UniqueUploadId"),
		},
		{
			Keys: bson.D{
				{Key: "uploadId", Value: 1},
				{Key: "type", Value: 1},
				{Key: "deletedTime", Value: -1},
				{Key: "_active", Value: 1},
			},
			Options: options.Index().
				SetBackground(true).
				SetName("UploadId"),
		},
		{
			Keys: bson.D{
				{Key: "_userId", Value: 1},
				{Key: "deviceId", Value: 1},
				{Key: "type", Value: 1},
				{Key: "_active", Value: 1},
				{Key: "_deduplicator.hash", Value: 1},
			},
			Options: options.Index().
				SetBackground(true).
				SetPartialFilterExpression(bson.D{
					{Key: "_active", Value: true},
					{Key: "_deduplicator.hash", Value: bson.D{{Key: "$exists", Value: true}}},
					{Key: "deviceId", Value: bson.D{{Key: "$exists", Value: true}}},
				}).
				SetName("DeduplicatorHash"),
		},
	})
}

func (d *DataSetRepository) GetDataSetByID(ctx context.Context, dataSetID string) (*upload.Upload, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if dataSetID == "" {
		return nil, errors.New("data set id is missing")
	}

	now := time.Now().UTC()

	var dataSet *upload.Upload
	selector := bson.M{
		"uploadId": dataSetID,
	}
	err := d.FindOne(ctx, selector).Decode(&dataSet)

	loggerFields := log.Fields{"dataSetId": dataSetID, "duration": time.Since(now) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("DataSet.GetDataSetByID")

	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "unable to get data set by id")
	}

	return dataSet, nil
}

func (d *DataSetRepository) createDataSet(ctx context.Context, dataSet *upload.Upload, now time.Time) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if err := validateDataSet(dataSet); err != nil {
		return err
	}

	now = now.UTC()
	timestamp := now.Truncate(time.Millisecond)

	dataSet.CreatedTime = pointer.FromTime(timestamp)
	dataSet.ModifiedTime = pointer.FromTime(timestamp)

	dataSet.ByUser = dataSet.CreatedUserID

	var err error
	if _, err = d.InsertOne(ctx, dataSet); storeStructuredMongo.IsDup(err) {
		err = errors.New("data set already exists")
	}

	loggerFields := log.Fields{"userId": dataSet.UserID, "dataSetId": dataSet.UploadID, "duration": time.Since(now) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("DataSet.CreateDataSet")

	if err != nil {
		return errors.Wrap(err, "unable to create data set")
	}
	return nil
}

// upsertDataSet upserts a DataSet. It is not exported as the caller of this
// code has to first check if the update succeeded in the original, old,
// deviceData collection. If it did, then an upsert in this collection is
// allowed. When migration is complete and all device upload data is moved to
// the new collection, the upsert option can be removed.
func (d *DataSetRepository) upsertDataSet(ctx context.Context, id string, update *data.DataSetUpdate, now time.Time) (*upload.Upload, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	} else if !data.IsValidSetID(id) {
		return nil, errors.New("id is invalid")
	}
	if update == nil {
		return nil, errors.New("update is missing")
	} else if err := structureValidator.New().Validate(update); err != nil {
		return nil, errors.Wrap(err, "update is invalid")
	}

	now = now.UTC()
	timestamp := now.Truncate(time.Millisecond)
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"id": id, "update": update})

	set := bson.M{
		"modifiedTime": timestamp,
	}
	unset := bson.M{}
	if update.Active != nil {
		set["_active"] = *update.Active
	}
	if update.DeviceID != nil {
		set["deviceId"] = *update.DeviceID
	}
	if update.DeviceModel != nil {
		set["deviceModel"] = *update.DeviceModel
	}
	if update.DeviceSerialNumber != nil {
		set["deviceSerialNumber"] = *update.DeviceSerialNumber
	}
	if update.Deduplicator != nil {
		set["_deduplicator"] = update.Deduplicator
	}
	if update.State != nil {
		set["_state"] = *update.State
	}
	if update.Time != nil {
		set["time"] = *update.Time
	}
	if update.TimeZoneName != nil {
		set["timezone"] = *update.TimeZoneName
	}
	if update.TimeZoneOffset != nil {
		set["timezoneOffset"] = *update.TimeZoneOffset
	}
	// Remove this upsert option when migration of device upload data from deviceData to deviceDataSets is complete
	opts := options.Update().SetUpsert(true)
	changeInfo, err := d.UpdateMany(ctx, bson.M{"uploadId": id}, d.ConstructUpdate(set, unset), opts)
	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("DataSetRepository.upsertDataSet")
	if err != nil {
		return nil, errors.Wrap(err, "unable to upsert data set")
	}

	return d.GetDataSetByID(ctx, id)
}

func (d *DataSetRepository) GetDataSet(ctx context.Context, id string) (*data.DataSet, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithField("id", id)

	var dataSet *data.DataSet
	selector := bson.M{
		"uploadId": id,
	}

	err := d.FindOne(ctx, selector).Decode(&dataSet)
	logger.WithField("duration", time.Since(now)/time.Microsecond).WithError(err).Debug("DataSet.GetDataSet")
	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "unable to get data set")
	}

	return dataSet, nil
}

func validateDataSet(dataSet *upload.Upload) error {
	if dataSet == nil {
		return errors.New("data set is missing")
	}
	if dataSet.UserID == nil {
		return errors.New("data set user id is missing")
	}
	if *dataSet.UserID == "" {
		return errors.New("data set user id is empty")
	}
	if dataSet.UploadID == nil {
		return errors.New("data set upload id is missing")
	}
	if *dataSet.UploadID == "" {
		return errors.New("data set upload id is empty")
	}
	return nil
}
