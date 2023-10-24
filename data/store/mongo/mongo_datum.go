package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/tidepool-org/platform/data/types/basal"
	"github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/data/types/blood/glucose/selfmonitored"
	"github.com/tidepool-org/platform/data/types/bolus"
	"github.com/tidepool-org/platform/data/types/insulin"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"errors"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/summary/types"
	baseDatum "github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/upload"
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

var ErrSelectorsInvalid = errors.New("selectors is invalid")

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
		return fmt.Errorf("unable to create data set data: %w", err)
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
		return fmt.Errorf("unable to activate data set data: %w", err)
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
		return fmt.Errorf("unable to archive data set data: %w", err)
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
		return fmt.Errorf("unable to delete data set data: %w", err)
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
		return fmt.Errorf("unable to destroy deleted data set data: %w", err)
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
		return fmt.Errorf("unable to destroy data set data: %w", err)
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
		return fmt.Errorf("unable to archive device data using hashes from data set: %w", err)
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
				overallErr = fmt.Errorf("unable to decode device data results: %w", err)
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
				overallErr = fmt.Errorf("unable to transfer device data active: %w", err)
			}
		} else {
			overallUpdateInfo.ModifiedCount += updateInfo.ModifiedCount
		}
	}

	if err := cursor.Err(); err != nil {
		if overallErr == nil {
			overallErr = fmt.Errorf("unable to iterate to transfer device data active: %w", err)
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
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to get data set: %w", err)
	}

	return dataSet, nil
}

func validateAndTranslateSelectors(selectors *data.Selectors) (bson.M, error) {
	if selectors == nil {
		return bson.M{}, nil
	} else if err := structureValidator.New().Validate(selectors); err != nil {
		return nil, errors.Join(ErrSelectorsInvalid, err)
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

func (d *DatumRepository) CheckDataSetContainsType(ctx context.Context, dataSetID string, typ string) (bool, error) {
	twoYearsPast := time.Now().UTC().AddDate(0, -24, 0)
	oneDayFuture := time.Now().UTC().AddDate(0, 0, 1)

	selector := bson.M{
		"_active":  true,
		"uploadId": dataSetID,
		"type":     typ,
		"time": bson.M{
			"$gt":  twoYearsPast,
			"$lte": oneDayFuture,
		},
	}

	var result bson.M
	if err := d.FindOne(ctx, selector).Decode(result); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, fmt.Errorf("unable to check for type %s in dataset %s: %w", typ, dataSetID, err)
	}

	return true, nil
}

func (d *DatumRepository) GetDataRange(ctx context.Context, dataRecords interface{}, userId string, typ string, startTime time.Time, endTime time.Time) error {
	if ctx == nil {
		return errors.New("context is missing")
	}

	if userId == "" {
		return errors.New("userId is empty")
	}

	if typ == "" {
		return errors.New("typ is empty")
	}

	// This is never expected to be an upload.
	if isTypeUpload(typ) {
		return fmt.Errorf("unexpected type: %v", upload.Type)
	}

	switch v := dataRecords.(type) {
	case *[]*glucose.Glucose:
		if typ != continuous.Type && typ != selfmonitored.Type {
			return fmt.Errorf("invalid type and destination pointer pair, %s cannot be decoded into glucose slice", typ)
		}
	case *[]*insulin.Insulin:
		if typ != bolus.Type && typ != basal.Type {
			return fmt.Errorf("invalid type and destination pointer pair, %s cannot be decoded into insulin slice", typ)
		}
	case *[]interface{}:
		// we cant check the type match, but at least the structure should work
	default:
		return fmt.Errorf("provided dataRecords type %T cannot be decoded into", v)
	}

	// quit early if range is 0
	if startTime.Equal(endTime) {
		return nil
	}

	// return error if ranges are inverted, as this can produce unexpected results
	if startTime.After(endTime) {
		return fmt.Errorf("startTime (%s) after endTime (%s) for user %s", startTime, endTime, userId)
	}

	selector := bson.M{
		"_active": true,
		"_userId": userId,
		"type":    typ,
		"time": bson.M{
			"$gt":  startTime,
			"$lte": endTime,
		},
	}

	opts := options.Find()
	opts.SetSort(bson.D{{Key: "time", Value: 1}})

	cursor, err := d.Find(ctx, selector, opts)
	if err != nil {
		return fmt.Errorf("unable to get cgm data in date range for user: %w", err)
	}

	if err = cursor.All(ctx, dataRecords); err != nil {
		return fmt.Errorf("unable to decode data sets, %w", err)
	}

	return nil
}

func (d *DatumRepository) GetLastUpdatedForUser(ctx context.Context, id string, typ string) (*types.UserLastUpdated, error) {
	var err error
	var cursor *mongo.Cursor
	var status = &types.UserLastUpdated{}
	var dataSet []*baseDatum.Base

	if ctx == nil {
		return nil, errors.New("context is missing")
	}

	if id == "" {
		return nil, errors.New("id is missing")
	}

	// This is never expected to by an upload.
	if isTypeUpload(typ) {
		return nil, fmt.Errorf("unexpected type: %v", upload.Type)
	}

	futureCutoff := time.Now().AddDate(0, 0, 1).UTC()
	pastCutoff := time.Now().AddDate(-2, 0, 0).UTC()

	selector := bson.M{
		"_active": true,
		"_userId": id,
		"type":    typ,
		"time": bson.M{
			"$lte": futureCutoff,
			"$gte": pastCutoff,
		},
	}

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "time", Value: -1}})
	findOptions.SetLimit(1)

	cursor, err = d.Find(ctx, selector, findOptions)
	if err != nil {
		return nil, fmt.Errorf("unable to get last %s date: %w", typ, err)
	}

	if err = cursor.All(ctx, &dataSet); err != nil {
		return nil, fmt.Errorf("unable to decode last %s date: %w", typ, err)
	}

	// if we have no record
	if len(dataSet) < 1 {
		return status, nil
	}

	status.LastUpload = *dataSet[0].CreatedTime
	status.LastUpload = status.LastUpload.UTC()

	status.LastData = *dataSet[0].Time
	status.LastData = status.LastData.UTC()

	return status, nil
}

func (d *DatumRepository) DistinctUserIDs(ctx context.Context, typ string) ([]string, error) {
	var distinctUserIDMap = make(map[string]struct{})
	var empty struct{}

	if ctx == nil {
		return nil, errors.New("context is missing")
	}

	// This is never expected to by an upload.
	if isTypeUpload(typ) {
		return nil, fmt.Errorf("unexpected type: %v", upload.Type)
	}

	// allow for a small margin on the pastCutoff to allow for calculation delay
	pastCutoff := time.Now().AddDate(0, -23, -20).UTC()
	futureCutoff := time.Now().AddDate(0, 0, 1).UTC()

	selector := bson.M{
		"_userId": bson.M{"$ne": -1111},
		"_active": true,
		"type":    typ,
		"time":    bson.M{"$gte": pastCutoff, "$lte": futureCutoff},
	}

	result, err := d.Distinct(ctx, "_userId", selector)
	if err != nil {
		return nil, fmt.Errorf("error fetching distinct userIDs: %w", err)
	}

	for _, v := range result {
		distinctUserIDMap[v.(string)] = empty
	}

	userIDs := make([]string, 0, len(distinctUserIDMap))
	for k := range distinctUserIDMap {
		userIDs = append(userIDs, k)
	}

	return userIDs, nil
}
