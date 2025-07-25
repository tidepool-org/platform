package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tidepool-org/platform/summary/types"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/store"
	baseDatum "github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/data/types/dosingdecision"
	platerrors "github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type DatumRepository struct {
	*storeStructuredMongo.Repository
}

const (
	LowerTimeIndexRaw = "2022-01-01T00:00:00Z"
)

var ErrSelectorsInvalid = errors.New("selectors is invalid")

func (d *DatumRepository) EnsureIndexes() error {
	lowerTimeBound, err := time.Parse(time.RFC3339, LowerTimeIndexRaw)
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
				SetName("UserIdTypeWeighted_v2"),
		},
		{
			Keys: bson.D{
				{Key: "_userId", Value: 1},
				{Key: "type", Value: 1},
				{Key: "time", Value: 1},
				{Key: "_active", Value: 1},
				{Key: "modifiedTime", Value: 1},
			},
			Options: options.Index().
				SetName("ShardKeyIndex"),
		},
		{
			Keys: bson.D{
				{Key: "_userId", Value: 1},
				{Key: "_active", Value: 1},
				{Key: "type", Value: 1},
				{Key: "time", Value: 1},
				{Key: "modifiedTime", Value: 1},
			},
			Options: options.Index().
				SetName("UserIdActiveTypeTimeModifiedTime").
				SetPartialFilterExpression(bson.D{
					{
						Key: "time",
						Value: bson.D{
							{Key: "$gt", Value: lowerTimeBound},
						},
					},
				}),
		},
		{
			Keys: bson.D{
				{Key: "_userId", Value: 1},
				{Key: "_active", Value: 1},
				{Key: "type", Value: 1},
				{Key: "modifiedTime", Value: 1},
				{Key: "time", Value: 1},
			},
			Options: options.Index().
				SetName("UserIdActiveTypeModifiedTimeTime").
				SetPartialFilterExpression(bson.D{
					{
						Key: "time",
						Value: bson.D{
							{Key: "$gt", Value: lowerTimeBound},
						},
					},
				}),
		},
		{
			Keys: bson.D{
				{Key: "_userId", Value: 1},
				{Key: "origin.id", Value: 1},
				{Key: "deletedTime", Value: -1},
				{Key: "_active", Value: 1},
			},
			Options: options.Index().
				SetPartialFilterExpression(bson.D{
					{Key: "origin.id", Value: bson.D{{Key: "$exists", Value: true}}},
				}).
				SetName("UserIdOriginId"),
		},
		// Future optimization after release.
		// Rebuild index to to move _active
		// before type for better compression and more
		// closely follow ESR
		{
			Keys: bson.D{
				{Key: "uploadId", Value: 1},
				{Key: "type", Value: 1},
				{Key: "deletedTime", Value: -1},
				{Key: "_active", Value: 1},
			},
			Options: options.Index().
				SetName("UploadId"),
		},

		// Future optimization - remove the PFE on deviceId as the Base datum
		// already makes sure it exists and prod DB has already been checked to
		// ensure there are no datums w/ no deviceId. Other possible
		// optimization remove the _active in the PFE to use this in the
		// ArchiveDeviceDataUsingHashesFromDataSet > Distinct quiery. Can also
		// remove "type" field w/ corresponding removal of "$ne": "upload" in
		// queries where appropriate.
		{
			Keys: bson.D{
				{Key: "_userId", Value: 1},
				{Key: "deviceId", Value: 1},
				{Key: "type", Value: 1},
				{Key: "_active", Value: 1},
				{Key: "_deduplicator.hash", Value: 1},
			},
			Options: options.Index().
				SetPartialFilterExpression(bson.D{
					{Key: "_active", Value: true},
					{Key: "_deduplicator.hash", Value: bson.D{{Key: "$exists", Value: true}}},
					{Key: "deviceId", Value: bson.D{{Key: "$exists", Value: true}}},
				}).
				SetName("DeduplicatorHash"),
		},
	})
}

func (d *DatumRepository) CreateDataSetData(ctx context.Context, dataSet *data.DataSet, dataSetData []data.Datum) error {
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

	insertData := make([]mongo.WriteModel, 0, len(dataSetData))

	for _, datum := range dataSetData {
		datum.SetUserID(dataSet.UserID)
		datum.SetDataSetID(dataSet.UploadID)
		datum.SetCreatedTime(&timestamp)
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

func (d *DatumRepository) ExistingDataSetData(ctx context.Context, dataSet *data.DataSet, selectors *data.Selectors) (*data.Selectors, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if err := validateDataSet(dataSet); err != nil {
		return nil, err
	}
	selector, _, err := validateAndTranslateSelectors(ctx, selectors)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	logger := log.LoggerFromContext(ctx).WithField("dataSetId", *dataSet.UploadID)

	selector["_userId"] = dataSet.UserID
	selector["uploadId"] = dataSet.UploadID
	selector["_active"] = true
	selector["deletedTime"] = bson.M{"$exists": false}

	findOptions := options.Find()
	findOptions.SetProjection(bson.M{"_id": 0, "id": 1, "time": 1, "_deduplicator.hash": 1, "origin.id": 1, "origin.time": 1})

	cursor, err := d.Find(ctx, selector, findOptions)
	if err != nil {
		logger.WithError(err).Error("Unable to get newer data set data selectors")
		return nil, fmt.Errorf("unable to get newer data set data selectors: %w", err)
	}

	existingSelectors := data.NewSelectors()
	if err = cursor.All(ctx, existingSelectors); err != nil {
		logger.WithError(err).Error("Unable to decode newer data set data selectors")
		return nil, fmt.Errorf("unable to decode newer data set data selectors: %w", err)
	}

	logger.WithFields(log.Fields{"existingSelector": existingSelectors, "duration": time.Since(now) / time.Microsecond}).Debug("ExistingDataSetData")
	return existingSelectors, nil
}

func (d *DatumRepository) ActivateDataSetData(ctx context.Context, dataSet *data.DataSet, selectors *data.Selectors) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if err := validateDataSet(dataSet); err != nil {
		return err
	}
	selector, _, err := validateAndTranslateSelectors(ctx, selectors)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	timestamp := now.Truncate(time.Millisecond)
	logger := log.LoggerFromContext(ctx).WithField("dataSetId", *dataSet.UploadID)

	selector["_userId"] = dataSet.UserID
	selector["uploadId"] = dataSet.UploadID
	selector["type"] = bson.M{"$ne": "upload"} // Note we WILL keep the "type" field in the UploadId index as that's a query need in tide-whisperer
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

func (d *DatumRepository) ArchiveDataSetData(ctx context.Context, dataSet *data.DataSet, selectors *data.Selectors) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if err := validateDataSet(dataSet); err != nil {
		return err
	}
	selector, hasOriginID, err := validateAndTranslateSelectors(ctx, selectors)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	timestamp := now.Truncate(time.Millisecond)
	logger := log.LoggerFromContext(ctx).WithField("dataSetId", *dataSet.UploadID)

	selector["_userId"] = dataSet.UserID
	selector["uploadId"] = dataSet.UploadID
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
	opts := options.Update()
	if hasOriginID {
		opts.SetHint("UserIdOriginId")
	}
	changeInfo, err := d.UpdateMany(ctx, selector, d.ConstructUpdate(set, unset), opts)
	if err != nil {
		logger.WithError(err).Error("Unable to archive data set data")
		return fmt.Errorf("unable to archive data set data: %w", err)
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("ArchiveDataSetData")
	return nil
}

func (d *DatumRepository) DeleteDataSetData(ctx context.Context, dataSet *data.DataSet, selectors *data.Selectors) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if err := validateDataSet(dataSet); err != nil {
		return err
	}
	selector, hasOriginID, err := validateAndTranslateSelectors(ctx, selectors)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	timestamp := now.Truncate(time.Millisecond)
	logger := log.LoggerFromContext(ctx).WithField("dataSetId", *dataSet.UploadID)

	selector["_userId"] = dataSet.UserID
	selector["uploadId"] = dataSet.UploadID
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
	opts := options.Update()
	if hasOriginID {
		opts.SetHint("UserIdOriginId")
	}
	changeInfo, err := d.UpdateMany(ctx, selector, d.ConstructUpdate(set, unset), opts)
	if err != nil {
		logger.WithError(err).Error("Unable to delete data set data")
		return fmt.Errorf("unable to delete data set data: %w", err)
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("DeleteDataSetData")
	return nil
}

func (d *DatumRepository) DestroyDeletedDataSetData(ctx context.Context, dataSet *data.DataSet, selectors *data.Selectors) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if err := validateDataSet(dataSet); err != nil {
		return err
	}
	selector, hasOriginID, err := validateAndTranslateSelectors(ctx, selectors)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	logger := log.LoggerFromContext(ctx).WithField("dataSetId", *dataSet.UploadID)

	selector["_userId"] = dataSet.UserID
	selector["uploadId"] = dataSet.UploadID
	selector["deletedTime"] = bson.M{"$exists": true}
	opts := options.Delete()
	if hasOriginID {
		opts.SetHint("UserIdOriginId")
	}
	changeInfo, err := d.DeleteMany(ctx, selector, opts)
	if err != nil {
		logger.WithError(err).Error("Unable to destroy deleted data set data")
		return fmt.Errorf("unable to destroy deleted data set data: %w", err)
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("DestroyDeletedDataSetData")
	return nil
}

func (d *DatumRepository) DestroyDataSetData(ctx context.Context, dataSet *data.DataSet, selectors *data.Selectors) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if err := validateDataSet(dataSet); err != nil {
		return err
	}
	selector, _, err := validateAndTranslateSelectors(ctx, selectors)
	if err != nil {
		return err
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithField("dataSetId", *dataSet.UploadID)

	selector["_userId"] = dataSet.UserID
	selector["uploadId"] = dataSet.UploadID
	changeInfo, err := d.DeleteMany(ctx, selector)
	if err != nil {
		logger.WithError(err).Error("Unable to destroy data set data")
		return fmt.Errorf("unable to destroy data set data: %w", err)
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("DestroyDataSetData")
	return nil
}

func (d *DatumRepository) ArchiveDeviceDataUsingHashesFromDataSet(ctx context.Context, dataSet *data.DataSet) error {
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

	// Note that the "DeduplicatorHash" index is NOT used here as the fields in the query don't match the the index definition. On average an upload only has one device anyways (P90 ~ 1). However the "DeduplicatorHash" index is still useful for the UpdateMany operation that follows.
	selector := bson.M{
		"_userId":            dataSet.UserID,
		"uploadId":           dataSet.UploadID,
		"type":               bson.M{"$ne": "upload"},
		"_deduplicator.hash": bson.M{"$ne": nil},
	}

	hashes, err := d.Distinct(ctx, "_deduplicator.hash", selector)
	if err == nil && len(hashes) > 0 {
		selector = bson.M{
			"_userId":            dataSet.UserID,
			"deviceId":           *dataSet.DeviceID,
			"type":               bson.M{"$ne": "upload"}, // Until we update the indexes to NOT have type, the planner will sometimes not use the correct index w/o the type range so we are leaving $ne upload in some cases. The actual performance and size gains are minor (~5%) TODO: for a future update, create a version of the index WITHOUT the type
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

func (d *DatumRepository) UnarchiveDeviceDataUsingHashesFromDataSet(ctx context.Context, dataSet *data.DataSet) error {
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

// FUTURE: Currently does not translate time or origin.time fields. Since origin.time is currently persisted as a string
// (and not time.Time) we cannot reliably query for it due to potentially variable timezone offsets. Eventually, migrate
// origin.time to time.Time and add additional qualifiers to the database selector that document origin.time must be
// greater than or equal to the incoming selector origin.time. For now, we can only query on id and origin.id.
// See: https://tidepool.atlassian.net/browse/BACK-3548
func validateAndTranslateSelectors(ctx context.Context, selectors *data.Selectors) (filter bson.M, hasOriginID bool, err error) {
	if selectors == nil {
		return bson.M{}, false, nil
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(selectors); err != nil {
		return nil, false, errors.Join(ErrSelectorsInvalid, err)
	}

	var selectorIDs []string
	var selectorDeduplicatorHashes []string
	var selectorOriginIDs []string
	for _, selector := range *selectors {
		if selector != nil {
			if selector.ID != nil {
				selectorIDs = append(selectorIDs, *selector.ID)
			} else if selector.Deduplicator != nil && selector.Deduplicator.Hash != nil {
				selectorDeduplicatorHashes = append(selectorDeduplicatorHashes, *selector.Deduplicator.Hash)
			} else if selector.Origin != nil && selector.Origin.ID != nil {
				selectorOriginIDs = append(selectorOriginIDs, *selector.Origin.ID)
			}
		}
	}

	var filters []bson.M
	if len(selectorIDs) > 0 {
		filters = append(filters, bson.M{"id": bson.M{"$in": selectorIDs}})
	}
	if len(selectorDeduplicatorHashes) > 0 {
		filters = append(filters, bson.M{"_deduplicator.hash": bson.M{"$in": selectorDeduplicatorHashes}})
	}
	if len(selectorOriginIDs) > 0 {
		filters = append(filters, bson.M{"origin.id": bson.M{"$in": selectorOriginIDs}})
	}

	switch len(filters) {
	case 0:
		return nil, false, errors.New("selectors is empty")
	case 1:
		return filters[0], len(selectorOriginIDs) > 0, nil
	default:
		return nil, false, errors.New("selectors is invalid, only one type of selector allowed")
	}
}

func (d *DatumRepository) GetDataRange(ctx context.Context, userId string, typ []string, status *data.UserDataStatus) (*mongo.Cursor, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}

	if userId == "" {
		return nil, errors.New("userId is empty")
	}

	if len(typ) == 0 {
		return nil, errors.New("typ is empty")
	}

	// quit early if range is 0
	if status.FirstData.Equal(status.LastData) {
		return nil, fmt.Errorf("FirstData (%s) equals LastData (%s) for user %s", status.FirstData, status.LastData, userId)
	}

	// return error if ranges are inverted, as this can produce unexpected results
	if status.FirstData.After(status.LastData) {
		return nil, fmt.Errorf("FirstData (%s) after LastData (%s) for user %s", status.FirstData, status.LastData, userId)
	}

	// quit early if range is 0
	if status.LastUpdated.Equal(status.NextLastUpdated) {
		return nil, fmt.Errorf("LastUpdated (%s) equals NextLastUpdated (%s) for user %s", status.LastUpdated, status.NextLastUpdated, userId)
	}

	// return error if ranges are inverted, as this can produce unexpected results
	if status.LastUpdated.After(status.NextLastUpdated) {
		return nil, fmt.Errorf("LastUpdated (%s) after NextLastUpdated (%s) for user %s", status.LastUpdated, status.NextLastUpdated, userId)
	}

	selector := bson.M{
		"_active": true,
		"_userId": userId,
		"time": bson.M{
			"$gt":  status.FirstData,
			"$lte": status.LastData,
		},
	}

	if len(typ) > 1 {
		selector["type"] = bson.M{"$in": typ}
	} else {
		selector["type"] = typ[0]
	}

	// we have everything we need to pull only modified records, but other areas are not ready for this yet
	//selector["modifiedTime"]= bson.M{
	//	"$gt":  status.LastUpdated,
	//	"$lte": status.NextLastUpdated,
	//}

	opts := options.Find()
	opts.SetSort(bson.D{{Key: "time", Value: 1}})
	opts.SetBatchSize(300)

	cursor, err := d.Find(ctx, selector, opts)
	if err != nil {
		return nil, fmt.Errorf("unable to get %s data in date range for user: %w", typ, err)
	}

	return cursor, nil
}

func (d *DatumRepository) GetAlertableData(ctx context.Context,
	params store.AlertableParams) (*store.AlertableResponse, error) {

	if params.End.IsZero() {
		params.End = time.Now()
	}

	cursor, err := d.getAlertableData(ctx, params, dosingdecision.Type)
	if err != nil {
		return nil, err
	}
	dosingDecisions := []*dosingdecision.DosingDecision{}
	if err := cursor.All(ctx, &dosingDecisions); err != nil {
		return nil, platerrors.Wrap(err, "Unable to load alertable dosing documents")
	}
	cursor, err = d.getAlertableData(ctx, params, continuous.Type)
	if err != nil {
		return nil, err
	}
	glucoseData := []*glucose.Glucose{}
	if err := cursor.All(ctx, &glucoseData); err != nil {
		return nil, platerrors.Wrap(err, "Unable to load alertable glucose documents")
	}
	response := &store.AlertableResponse{
		DosingDecisions: dosingDecisions,
		Glucose:         glucoseData,
	}

	return response, nil
}

func (d *DatumRepository) getAlertableData(ctx context.Context,
	params store.AlertableParams, typ string) (*mongo.Cursor, error) {

	selector := bson.M{
		"_active":  true,
		"uploadId": params.UploadID,
		"type":     typ,
		"_userId":  params.UserID,
		"time":     bson.M{"$gte": params.Start, "$lte": params.End},
	}
	findOptions := options.Find().SetSort(bson.D{{Key: "time", Value: -1}})
	cursor, err := d.Find(ctx, selector, findOptions)
	if err != nil {
		format := "Unable to find alertable %s data in dataset %s"
		return nil, platerrors.Wrapf(err, format, typ, params.UploadID)
	}
	return cursor, nil
}

func (d *DatumRepository) getTimeRange(ctx context.Context, userId string, typ []string, status *data.UserDataStatus) (err error) {
	timestamp := time.Now().UTC()
	futureCutoff := timestamp.AddDate(0, 0, 1)
	pastCutoff := timestamp.AddDate(-2, 0, 0)

	// get latest active record
	selector := bson.M{
		"_active": true,
		"_userId": userId,
		"time": bson.M{
			"$gte": pastCutoff,
			"$lte": futureCutoff,
		},
	}

	if len(typ) == 1 {
		selector["type"] = typ[0]
	} else {
		selector["type"] = bson.M{"$in": typ}
	}

	findOptions := options.Find()
	findOptions.SetProjection(bson.M{"_id": 0, "time": 1})
	findOptions.SetSort(bson.D{{Key: "time", Value: -1}})
	findOptions.SetLimit(1)

	var cursor *mongo.Cursor
	cursor, err = d.Find(ctx, selector, findOptions)
	if err != nil {
		return fmt.Errorf("unable to get last %s time: %w", typ, err)
	}

	var dataSet []*baseDatum.Base
	if err = cursor.All(ctx, &dataSet); err != nil {
		return fmt.Errorf("unable to decode last %s time: %w", typ, err)
	}

	// if we have a record
	if len(dataSet) > 0 {
		status.LastData = dataSet[0].Time.UTC()
		status.FirstData = status.LastData.AddDate(0, 0, -types.HoursAgoToKeep/24)
	}

	return nil
}

func (d *DatumRepository) populateLastUpload(ctx context.Context, userId string, typ []string, status *data.UserDataStatus) (err error) {
	// get latest modified record
	selector := bson.M{
		"_userId": userId,
		"_active": bson.M{"$in": bson.A{true, false}},
		"time": bson.M{
			"$gte": status.FirstData,
			"$lte": status.LastData,
		},
	}

	if len(typ) == 1 {
		selector["type"] = typ[0]
	} else {
		selector["type"] = bson.M{"$in": typ}
	}

	findOptions := options.Find().SetProjection(bson.M{"_id": 0, "modifiedTime": 1, "createdTime": 1})
	if lowerTimeBound, err := time.Parse(time.RFC3339, LowerTimeIndexRaw); err == nil && status.FirstData.After(lowerTimeBound) {
		findOptions.SetHint("UserIdActiveTypeModifiedTimeTime")
	}
	findOptions.SetLimit(1)
	findOptions.SetSort(bson.D{{Key: "modifiedTime", Value: -1}})

	var cursor *mongo.Cursor
	cursor, err = d.Find(ctx, selector, findOptions)
	if err != nil {
		return fmt.Errorf("unable to get last %s  modifiedTime: %w", typ, err)
	}

	var dataSet []*baseDatum.Base
	if err = cursor.All(ctx, &dataSet); err != nil {
		return fmt.Errorf("unable to decode last %s modifiedTime: %w", typ, err)
	}

	// if we have a record
	if len(dataSet) > 0 {
		// handle data without modifiedTime, as older data may not have it
		// this will only be triggered on fresh summaries of old data
		if dataSet[0].ModifiedTime != nil {
			status.LastUpload = dataSet[0].ModifiedTime.UTC()
		} else {
			status.LastUpload = dataSet[0].CreatedTime.UTC()
		}
	}

	return nil
}

func (d *DatumRepository) populateEarliestModified(ctx context.Context, userId string, typ []string, status *data.UserDataStatus) (err error) {
	// get earliest modified record which is newer than LastUpdated
	selector := bson.M{
		"_userId": userId,
		"_active": bson.M{"$in": bson.A{true, false}},
		"time": bson.M{
			"$gte": status.FirstData,
			"$lte": status.LastData,
		},
	}

	if len(typ) == 1 {
		selector["type"] = typ[0]
	} else {
		selector["type"] = bson.M{"$in": typ}
	}

	findOptions := options.Find()
	findOptions.SetLimit(1)
	findOptions.SetSort(bson.D{{Key: "time", Value: 1}}).
		SetProjection(bson.M{"_id": 0, "time": 1})

	// this skips using modifiedTime on fresh calculations as it may cause trouble with initial calculation of summaries
	// for users with only data old enough to not have a modifiedTime, which would be excluded by this.
	// this is not a concern for subsequent updates, as they would be triggered by new data, which would have modifiedTime
	if !status.LastUpdated.IsZero() {
		selector["modifiedTime"] = bson.M{
			"$gte": status.LastUpdated,
		}
		if lowerTimeBound, err := time.Parse(time.RFC3339, LowerTimeIndexRaw); err == nil && status.FirstData.After(lowerTimeBound) {
			// has blocking sort, but more selective so usually performs better.
			findOptions.SetHint("UserIdActiveTypeModifiedTimeTime")
		}
	}

	var cursor *mongo.Cursor
	cursor, err = d.Find(ctx, selector, findOptions)
	if err != nil {
		return fmt.Errorf("unable to get earliest %s recently modified time: %w", typ, err)
	}

	var dataSet []*baseDatum.Base
	if err = cursor.All(ctx, &dataSet); err != nil {
		return fmt.Errorf("unable to decode earliest %s recently modified time: %w", typ, err)
	}

	// if we have a record
	if len(dataSet) > 0 {
		status.EarliestModified = dataSet[0].Time.UTC()
	}

	return nil
}

func (d *DatumRepository) GetLastUpdatedForUser(ctx context.Context, userId string, typ []string, lastUpdated time.Time) (*data.UserDataStatus, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}

	if userId == "" {
		return nil, errors.New("userId is empty")
	}

	if len(typ) == 0 {
		return nil, errors.New("typ is empty")
	}

	status := &data.UserDataStatus{
		LastUpdated:     lastUpdated,
		NextLastUpdated: time.Now().UTC().Truncate(time.Millisecond),
	}

	err := d.getTimeRange(ctx, userId, typ, status)
	if err != nil {
		return nil, err
	}

	// the user has no eligible data, quit out early
	if status.LastData.IsZero() {
		return nil, nil
	}

	err = d.populateLastUpload(ctx, userId, typ, status)
	if err != nil {
		return nil, err
	}

	err = d.populateEarliestModified(ctx, userId, typ, status)
	if err != nil {
		return nil, err
	}

	return status, nil
}
