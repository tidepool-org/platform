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
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type DatumRepository struct {
	*storeStructuredMongo.Repository
}

const (
	ModifiedTimeIndexRaw = "2023-04-01T00:00:00Z"
)

func (d *DatumRepository) EnsureIndexes() error {
	modifiedTime, err := time.Parse(time.RFC3339, ModifiedTimeIndexRaw)
	if err != nil {
		return err
	}
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

func (d *DatumRepository) CreateDataSetData(ctx context.Context, dataSet *upload.Upload, dataSetData []data.Datum) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if err := validateDataSet(dataSet); err != nil {
		return err
	}
	if dataSetData == nil {
		return errors.New("data set data is missing")
	}

	if len(dataSetData) == 0 {
		return nil
	}

	now := time.Now().UTC()
	timestamp := now.Truncate(time.Millisecond)

	var insertData []mongo.WriteModel

	for _, datum := range dataSetData {
		datum.SetUserID(dataSet.UserID)
		datum.SetDataSetID(dataSet.UploadID)
		datum.SetCreatedTime(&timestamp)
		datum.SetModifiedTime(&timestamp)
		datum.SetModifiedTime(&timestamp)
		insertData = append(insertData, mongo.NewInsertOneModel().SetDocument(datum))
	}

	opts := options.BulkWrite().SetOrdered(false)

	_, err := d.BulkWrite(ctx, insertData, opts)

	loggerFields := log.Fields{"dataSetId": dataSet.UploadID, "dataCount": len(dataSetData), "duration": time.Since(now) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("CreateDataSetData")

	if err != nil {
		return errors.Wrap(err, "unable to create data set data")
	}
	return nil
}

func (d *DatumRepository) ActivateDataSetData(ctx context.Context, dataSet *upload.Upload, selectors *data.Selectors) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if err := validateDataSet(dataSet); err != nil {
		return err
	}
	selector, err := validateAndTranslateSelectors(selectors)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	timestamp := now.Truncate(time.Millisecond)
	logger := log.LoggerFromContext(ctx).WithField("dataSetId", *dataSet.UploadID)

	selector["_userId"] = dataSet.UserID
	selector["uploadId"] = dataSet.UploadID
	selector["type"] = bson.M{"$ne": "upload"}
	selector["_active"] = false
	selector["deletedTime"] = bson.M{"$exists": false}
	set := bson.M{
		"_active":      true,
		"modifiedTime": timestamp,
	}
	unset := bson.M{
		"archivedDatasetId": 1,
		"archivedTime":      1,
		"modifiedUserId":    1,
	}
	changeInfo, err := d.UpdateMany(ctx, selector, d.ConstructUpdate(set, unset))
	if err != nil {
		logger.WithError(err).Error("Unable to activate data set data")
		return errors.Wrap(err, "unable to activate data set data")
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("ActivateDataSetData")
	return nil
}

func (d *DatumRepository) ArchiveDataSetData(ctx context.Context, dataSet *upload.Upload, selectors *data.Selectors) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if err := validateDataSet(dataSet); err != nil {
		return err
	}
	selector, err := validateAndTranslateSelectors(selectors)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	timestamp := now.Truncate(time.Millisecond)
	logger := log.LoggerFromContext(ctx).WithField("dataSetId", *dataSet.UploadID)

	selector["_userId"] = dataSet.UserID
	selector["uploadId"] = dataSet.UploadID
	selector["type"] = bson.M{"$ne": "upload"}
	selector["_active"] = true
	selector["deletedTime"] = bson.M{"$exists": false}
	set := bson.M{
		"_active":      false,
		"archivedTime": timestamp,
		"modifiedTime": timestamp,
	}
	unset := bson.M{
		"archivedDatasetId": 1,
		"modifiedUserId":    1,
	}
	changeInfo, err := d.UpdateMany(ctx, selector, d.ConstructUpdate(set, unset))
	if err != nil {
		logger.WithError(err).Error("Unable to archive data set data")
		return errors.Wrap(err, "unable to archive data set data")
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("ArchiveDataSetData")
	return nil
}

func (d *DatumRepository) DeleteDataSetData(ctx context.Context, dataSet *upload.Upload, selectors *data.Selectors) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if err := validateDataSet(dataSet); err != nil {
		return err
	}
	selector, err := validateAndTranslateSelectors(selectors)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	timestamp := now.Truncate(time.Millisecond)
	logger := log.LoggerFromContext(ctx).WithField("dataSetId", *dataSet.UploadID)

	selector["_userId"] = dataSet.UserID
	selector["uploadId"] = dataSet.UploadID
	selector["type"] = bson.M{"$ne": "upload"}
	selector["deletedTime"] = bson.M{"$exists": false}
	set := bson.M{
		"_active":      false,
		"archivedTime": timestamp,
		"deletedTime":  timestamp,
		"modifiedTime": timestamp,
	}
	unset := bson.M{
		"archivedDatasetId": 1,
		"deletedUserId":     1,
		"modifiedUserId":    1,
	}
	changeInfo, err := d.UpdateMany(ctx, selector, d.ConstructUpdate(set, unset))
	if err != nil {
		logger.WithError(err).Error("Unable to delete data set data")
		return errors.Wrap(err, "unable to delete data set data")
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("DeleteDataSetData")
	return nil
}

func (d *DatumRepository) DestroyDeletedDataSetData(ctx context.Context, dataSet *upload.Upload, selectors *data.Selectors) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if err := validateDataSet(dataSet); err != nil {
		return err
	}
	selector, err := validateAndTranslateSelectors(selectors)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	logger := log.LoggerFromContext(ctx).WithField("dataSetId", *dataSet.UploadID)

	selector["_userId"] = dataSet.UserID
	selector["uploadId"] = dataSet.UploadID
	selector["type"] = bson.M{"$ne": "upload"}
	selector["deletedTime"] = bson.M{"$exists": true}
	changeInfo, err := d.DeleteMany(ctx, selector)
	if err != nil {
		logger.WithError(err).Error("Unable to destroy deleted data set data")
		return errors.Wrap(err, "unable to destroy deleted data set data")
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("DestroyDeletedDataSetData")
	return nil
}

func (d *DatumRepository) DestroyDataSetData(ctx context.Context, dataSet *upload.Upload, selectors *data.Selectors) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if err := validateDataSet(dataSet); err != nil {
		return err
	}
	selector, err := validateAndTranslateSelectors(selectors)
	if err != nil {
		return err
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithField("dataSetId", *dataSet.UploadID)

	selector["_userId"] = dataSet.UserID
	selector["uploadId"] = dataSet.UploadID
	selector["type"] = bson.M{"$ne": "upload"}
	changeInfo, err := d.DeleteMany(ctx, selector)
	if err != nil {
		logger.WithError(err).Error("Unable to destroy data set data")
		return errors.Wrap(err, "unable to destroy data set data")
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("DestroyDataSetData")
	return nil
}

func (d *DatumRepository) ArchiveDeviceDataUsingHashesFromDataSet(ctx context.Context, dataSet *upload.Upload) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if err := validateDataSet(dataSet); err != nil {
		return err
	}
	if dataSet.DeviceID == nil || *dataSet.DeviceID == "" {
		return errors.New("data set device id is missing")
	}

	now := time.Now().UTC()
	timestamp := now.Truncate(time.Millisecond)

	var updateInfo *mongo.UpdateResult

	selector := bson.M{
		"_userId":  dataSet.UserID,
		"uploadId": dataSet.UploadID,
		"type":     bson.M{"$ne": "upload"},
	}

	hashes, err := d.Distinct(ctx, "_deduplicator.hash", selector)
	if err == nil && len(hashes) > 0 {
		selector = bson.M{
			"_userId":            dataSet.UserID,
			"deviceId":           *dataSet.DeviceID,
			"type":               bson.M{"$ne": "upload"},
			"_active":            true,
			"_deduplicator.hash": bson.M{"$in": hashes},
		}
		set := bson.M{
			"_active":           false,
			"archivedDatasetId": dataSet.UploadID,
			"archivedTime":      timestamp,
			"modifiedTime":      timestamp,
		}
		unset := bson.M{}
		updateInfo, err = d.UpdateMany(ctx, selector, d.ConstructUpdate(set, unset))
	}

	loggerFields := log.Fields{"userId": dataSet.UserID, "deviceId": *dataSet.DeviceID, "updateInfo": updateInfo, "duration": time.Since(now) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("ArchiveDeviceDataUsingHashesFromDataSet")

	if err != nil {
		return errors.Wrap(err, "unable to archive device data using hashes from data set")
	}
	return nil
}

func (d *DatumRepository) UnarchiveDeviceDataUsingHashesFromDataSet(ctx context.Context, dataSet *upload.Upload) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if err := validateDataSet(dataSet); err != nil {
		return err
	}
	if dataSet.DeviceID == nil || *dataSet.DeviceID == "" {
		return errors.New("data set device id is missing")
	}

	now := time.Now().UTC()
	timestamp := now.Truncate(time.Millisecond)

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"uploadId": dataSet.UploadID,
				"type":     bson.M{"$ne": "upload"},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"_active":           "$_active",
					"archivedDatasetId": "$archivedDatasetId",
					"archivedTime":      "$archivedTime",
				},
				"archivedHashes": bson.M{"$push": "$_deduplicator.hash"},
			},
		},
	}
	cursor, _ := d.Aggregate(ctx, pipeline)

	var overallUpdateInfo mongo.UpdateResult
	var overallErr error

	result := struct {
		ID struct {
			Active            bool      `bson:"_active"`
			ArchivedDataSetID string    `bson:"archivedDatasetId"`
			ArchivedTime      time.Time `bson:"archivedTime"`
		} `bson:"_id"`
		ArchivedHashes []string `bson:"archivedHashes"`
	}{}
	for cursor.Next(ctx) {
		err := cursor.Decode(&result)
		if err != nil {
			loggerFields := log.Fields{"dataSetId": dataSet.UploadID, "result": result}
			log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Error("Unable to decode result for UnarchiveDeviceDataUsingHashesFromDataSet")
			if overallErr == nil {
				overallErr = errors.Wrap(err, "unable to decode device data results")
			}
		}
		if result.ID.Active != (result.ID.ArchivedDataSetID == "") || result.ID.Active != (result.ID.ArchivedTime.IsZero()) {
			loggerFields := log.Fields{"dataSetId": dataSet.UploadID, "result": result}
			log.LoggerFromContext(ctx).WithFields(loggerFields).Error("Unexpected pipe result for UnarchiveDeviceDataUsingHashesFromDataSet")
			continue
		}

		selector := bson.M{
			"_userId":            dataSet.UserID,
			"deviceId":           dataSet.DeviceID,
			"archivedDatasetId":  dataSet.UploadID,
			"_deduplicator.hash": bson.M{"$in": result.ArchivedHashes},
		}
		set := bson.M{
			"_active":      result.ID.Active,
			"modifiedTime": timestamp,
		}
		unset := bson.M{}
		if result.ID.Active {
			unset["archivedDatasetId"] = true
			unset["archivedTime"] = true
		} else {
			set["archivedDatasetId"] = result.ID.ArchivedDataSetID
			set["archivedTime"] = result.ID.ArchivedTime
		}
		updateInfo, err := d.UpdateMany(ctx, selector, d.ConstructUpdate(set, unset))
		if err != nil {
			loggerFields := log.Fields{"dataSetId": dataSet.UploadID, "result": result}
			log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Error("Unable to update result for UnarchiveDeviceDataUsingHashesFromDataSet")
			if overallErr == nil {
				overallErr = errors.Wrap(err, "unable to transfer device data active")
			}
		} else {
			overallUpdateInfo.ModifiedCount += updateInfo.ModifiedCount
		}
	}

	if err := cursor.Err(); err != nil {
		if overallErr == nil {
			overallErr = errors.Wrap(err, "unable to iterate to transfer device data active")
		}
	}

	loggerFields := log.Fields{"dataSetId": dataSet.UploadID, "updateInfo": overallUpdateInfo, "duration": time.Since(now) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(overallErr).Debug("UnarchiveDeviceDataUsingHashesFromDataSet")

	return overallErr
}

func (d *DatumRepository) GetDataSet(ctx context.Context, id string) (*data.DataSet, error) {
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
		"type":     "upload",
	}

	err := d.FindOne(ctx, selector).Decode(&dataSet)
	logger.WithField("duration", time.Since(now)/time.Microsecond).WithError(err).Debug("DatumRepository.GetDataSet")
	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "unable to get data set")
	}

	return dataSet, nil
}

func validateAndTranslateSelectors(selectors *data.Selectors) (bson.M, error) {
	if selectors == nil {
		return bson.M{}, nil
	} else if err := structureValidator.New().Validate(selectors); err != nil {
		return nil, errors.Wrap(err, "selectors is invalid")
	}

	var selectorIDs []string
	var selectorOriginIDs []string
	for _, selector := range *selectors {
		if selector != nil {
			if selector.ID != nil {
				selectorIDs = append(selectorIDs, *selector.ID)
			} else if selector.Origin != nil && selector.Origin.ID != nil {
				selectorOriginIDs = append(selectorOriginIDs, *selector.Origin.ID)
			}
		}
	}

	selector := bson.M{}
	if len(selectorIDs) > 0 && len(selectorOriginIDs) > 0 {
		selector["$or"] = []bson.M{
			{"id": bson.M{"$in": selectorIDs}},
			{"origin.id": bson.M{"$in": selectorOriginIDs}},
		}
	} else if len(selectorIDs) > 0 {
		selector["id"] = bson.M{"$in": selectorIDs}
	} else if len(selectorOriginIDs) > 0 {
		selector["origin.id"] = bson.M{"$in": selectorOriginIDs}
	}

	if len(selector) == 0 {
		return nil, errors.New("selectors is invalid")
	}

	return selector, nil
}
