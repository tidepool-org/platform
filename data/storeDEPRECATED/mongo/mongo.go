package mongo

import (
	"context"
	"time"

	mgo "github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/storeDEPRECATED"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func NewStore(cfg *storeStructuredMongo.Config, lgr log.Logger) (*Store, error) {
	baseStore, err := storeStructuredMongo.NewStore(cfg, lgr)
	if err != nil {
		return nil, err
	}

	return &Store{
		Store: baseStore,
	}, nil
}

type Store struct {
	*storeStructuredMongo.Store
}

func (s *Store) EnsureIndexes() error {
	session := s.NewDataSession()
	defer session.Close()
	return session.EnsureIndexes()
}

func (s *Store) NewDataSession() storeDEPRECATED.DataSession {
	return &DataSession{
		Session: s.Store.NewSession("deviceData"),
	}
}

type DataSession struct {
	*storeStructuredMongo.Session
}

func (d *DataSession) EnsureIndexes() error {
	return d.EnsureAllIndexes([]mgo.Index{
    // Additional indexes are also created in `tide-whisperer` and `jellyfish`
		{Key: []string{"_userId", "_active", "type", "-time"}, Background: true, Name: "UserIdTypeWeighted_v2"},
		{Key: []string{"origin.id", "type", "-deletedTime", "_active"}, Background: true, Name: "OriginId"},
		{Key: []string{"uploadId", "type", "-deletedTime", "_active"}, Background: true, Name: "UploadId"},
		{Key: []string{"uploadId"}, Background: true, Unique: true, PartialFilter: bson.M{"type": "upload"}, Name: "UniqueUploadId"},
	})
}

func (d *DataSession) GetDataSetsForUserByID(ctx context.Context, userID string, filter *storeDEPRECATED.Filter, pagination *page.Pagination) ([]*upload.Upload, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	}
	if filter == nil {
		filter = storeDEPRECATED.NewFilter()
	} else if err := structureValidator.New().Validate(filter); err != nil {
		return nil, errors.Wrap(err, "filter is invalid")
	}
	if pagination == nil {
		pagination = page.NewPagination()
	} else if err := structureValidator.New().Validate(pagination); err != nil {
		return nil, errors.Wrap(err, "pagination is invalid")
	}

	if d.IsClosed() {
		return nil, errors.New("session closed")
	}

	now := time.Now()

	var dataSets []*upload.Upload
	selector := bson.M{
		"_active": true,
		"_userId": userID,
		"type":    "upload",
	}
	if !filter.Deleted {
		selector["deletedTime"] = bson.M{"$exists": false}
	}
	err := d.C().Find(selector).Sort("-createdTime").Skip(pagination.Page * pagination.Size).Limit(pagination.Size).All(&dataSets)

	loggerFields := log.Fields{"userId": userID, "dataSetsCount": len(dataSets), "duration": time.Since(now) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("GetDataSetsForUserByID")

	if err != nil {
		return nil, errors.Wrap(err, "unable to get data sets for user by id")
	}

	if dataSets == nil {
		dataSets = []*upload.Upload{}
	}
	return dataSets, nil
}

func (d *DataSession) GetDataSetByID(ctx context.Context, dataSetID string) (*upload.Upload, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if dataSetID == "" {
		return nil, errors.New("data set id is missing")
	}

	if d.IsClosed() {
		return nil, errors.New("session closed")
	}

	now := time.Now()

	dataSets := []*upload.Upload{}
	selector := bson.M{
		"uploadId": dataSetID,
		"type":     "upload",
	}
	err := d.C().Find(selector).Limit(2).All(&dataSets)

	loggerFields := log.Fields{"dataSetId": dataSetID, "duration": time.Since(now) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("GetDataSetByID")

	if err != nil {
		return nil, errors.Wrap(err, "unable to get data set by id")
	}

	if dataSetsCount := len(dataSets); dataSetsCount == 0 {
		return nil, nil
	} else if dataSetsCount > 1 {
		log.LoggerFromContext(ctx).WithField("dataSetId", dataSetID).Warn("Multiple data sets found for data set id")
	}

	return dataSets[0], nil
}

func (d *DataSession) CreateDataSet(ctx context.Context, dataSet *upload.Upload) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if err := validateDataSet(dataSet); err != nil {
		return err
	}

	if d.IsClosed() {
		return errors.New("session closed")
	}

	now := time.Now()
	timestamp := now.Truncate(time.Millisecond).Format(time.RFC3339Nano)

	dataSet.CreatedTime = pointer.FromString(timestamp)

	dataSet.ByUser = dataSet.CreatedUserID

	// This search is protected by the `UniqueUploadId` index, so it's not possible for there to be
	// duplicate `uploadId`s, even if a race condition occurs here.
	selector := bson.M{
		"uploadId": dataSet.UploadID,
		"type":     dataSet.Type,
	}
	count, err := d.C().Find(selector).Count()
	if err == nil {
		if count > 0 {
			err = errors.New("data set already exists")
		} else {
			err = d.C().Insert(dataSet)
		}
	}

	loggerFields := log.Fields{"userId": dataSet.UserID, "dataSetId": dataSet.UploadID, "duration": time.Since(now) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("CreateDataSet")

	if err != nil {
		return errors.Wrap(err, "unable to create data set")
	}
	return nil
}

func (d *DataSession) UpdateDataSet(ctx context.Context, id string, update *data.DataSetUpdate) (*upload.Upload, error) {
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

	if d.IsClosed() {
		return nil, errors.New("session closed")
	}

	now := time.Now()
	timestamp := now.Truncate(time.Millisecond).Format(time.RFC3339Nano)
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
		set["time"] = (*update.Time).Format(data.TimeFormat)
	}
	if update.TimeZoneName != nil {
		set["timezone"] = *update.TimeZoneName
	}
	if update.TimeZoneOffset != nil {
		set["timezoneOffset"] = *update.TimeZoneOffset
	}
	changeInfo, err := d.C().UpdateAll(bson.M{"type": "upload", "uploadId": id}, d.ConstructUpdate(set, unset))
	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("UpdateDataSet")
	if err != nil {
		return nil, errors.Wrap(err, "unable to update data set")
	}

	return d.GetDataSetByID(ctx, id)
}

func (d *DataSession) DeleteDataSet(ctx context.Context, dataSet *upload.Upload) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if err := validateDataSet(dataSet); err != nil {
		return err
	}

	if d.IsClosed() {
		return errors.New("session closed")
	}

	now := time.Now()
	timestamp := now.Truncate(time.Millisecond).Format(time.RFC3339Nano)

	var err error
	var removeInfo *mgo.ChangeInfo
	var updateInfo *mgo.ChangeInfo

	selector := bson.M{
		"_userId":  dataSet.UserID,
		"uploadId": dataSet.UploadID,
		"type":     bson.M{"$ne": "upload"},
	}
	removeInfo, err = d.C().RemoveAll(selector)
	if err == nil {
		selector = bson.M{
			"_userId":       dataSet.UserID,
			"uploadId":      dataSet.UploadID,
			"type":          "upload",
			"deletedTime":   bson.M{"$exists": false},
			"deletedUserId": bson.M{"$exists": false},
		}
		set := bson.M{
			"deletedTime": timestamp,
		}
		unset := bson.M{}
		updateInfo, err = d.C().UpdateAll(selector, d.ConstructUpdate(set, unset))
	}

	loggerFields := log.Fields{"dataSetId": dataSet.UploadID, "removeInfo": removeInfo, "updateInfo": updateInfo, "duration": time.Since(now) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("DeleteDataSet")

	if err != nil {
		return errors.Wrap(err, "unable to delete data set")
	}

	dataSet.SetDeletedTime(&timestamp)
	return nil
}

func (d *DataSession) CreateDataSetData(ctx context.Context, dataSet *upload.Upload, dataSetData []data.Datum) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if err := validateDataSet(dataSet); err != nil {
		return err
	}
	if dataSetData == nil {
		return errors.New("data set data is missing")
	}

	if d.IsClosed() {
		return errors.New("session closed")
	}

	if len(dataSetData) == 0 {
		return nil
	}

	now := time.Now()
	timestamp := now.Truncate(time.Millisecond).Format(time.RFC3339Nano)

	insertData := make([]interface{}, len(dataSetData))
	for index, datum := range dataSetData {
		datum.SetUserID(dataSet.UserID)
		datum.SetDataSetID(dataSet.UploadID)
		datum.SetCreatedTime(&timestamp)
		insertData[index] = datum
	}

	bulk := d.C().Bulk()
	bulk.Unordered()
	bulk.Insert(insertData...)

	_, err := bulk.Run()

	loggerFields := log.Fields{"dataSetId": dataSet.UploadID, "dataCount": len(dataSetData), "duration": time.Since(now) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("CreateDataSetData")

	if err != nil {
		return errors.Wrap(err, "unable to create data set data")
	}
	return nil
}

func (d *DataSession) ActivateDataSetData(ctx context.Context, dataSet *upload.Upload, selectors *data.Selectors) error {
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

	if d.IsClosed() {
		return errors.New("session closed")
	}

	now := time.Now()
	timestamp := now.Truncate(time.Millisecond).Format(time.RFC3339Nano)
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
	changeInfo, err := d.C().UpdateAll(selector, d.ConstructUpdate(set, unset))
	if err != nil {
		logger.WithError(err).Error("Unable to activate data set data")
		return errors.Wrap(err, "unable to activate data set data")
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("ActivateDataSetData")
	return nil
}

func (d *DataSession) ArchiveDataSetData(ctx context.Context, dataSet *upload.Upload, selectors *data.Selectors) error {
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

	if d.IsClosed() {
		return errors.New("session closed")
	}

	now := time.Now()
	timestamp := now.Truncate(time.Millisecond).Format(time.RFC3339Nano)
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
	changeInfo, err := d.C().UpdateAll(selector, d.ConstructUpdate(set, unset))
	if err != nil {
		logger.WithError(err).Error("Unable to archive data set data")
		return errors.Wrap(err, "unable to archive data set data")
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("ArchiveDataSetData")
	return nil
}

func (d *DataSession) DeleteDataSetData(ctx context.Context, dataSet *upload.Upload, selectors *data.Selectors) error {
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

	if d.IsClosed() {
		return errors.New("session closed")
	}

	now := time.Now()
	timestamp := now.Truncate(time.Millisecond).Format(time.RFC3339Nano)
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
	changeInfo, err := d.C().UpdateAll(selector, d.ConstructUpdate(set, unset))
	if err != nil {
		logger.WithError(err).Error("Unable to delete data set data")
		return errors.Wrap(err, "unable to delete data set data")
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("DeleteDataSetData")
	return nil
}

func (d *DataSession) DestroyDeletedDataSetData(ctx context.Context, dataSet *upload.Upload, selectors *data.Selectors) error {
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

	if d.IsClosed() {
		return errors.New("session closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithField("dataSetId", *dataSet.UploadID)

	selector["_userId"] = dataSet.UserID
	selector["uploadId"] = dataSet.UploadID
	selector["type"] = bson.M{"$ne": "upload"}
	selector["deletedTime"] = bson.M{"$exists": true}
	changeInfo, err := d.C().RemoveAll(selector)
	if err != nil {
		logger.WithError(err).Error("Unable to destroy deleted data set data")
		return errors.Wrap(err, "unable to destroy deleted data set data")
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("DestroyDeletedDataSetData")
	return nil
}

func (d *DataSession) DestroyDataSetData(ctx context.Context, dataSet *upload.Upload, selectors *data.Selectors) error {
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

	if d.IsClosed() {
		return errors.New("session closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithField("dataSetId", *dataSet.UploadID)

	selector["_userId"] = dataSet.UserID
	selector["uploadId"] = dataSet.UploadID
	selector["type"] = bson.M{"$ne": "upload"}
	changeInfo, err := d.C().RemoveAll(selector)
	if err != nil {
		logger.WithError(err).Error("Unable to destroy data set data")
		return errors.Wrap(err, "unable to destroy data set data")
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("DestroyDataSetData")
	return nil
}

func (d *DataSession) ArchiveDeviceDataUsingHashesFromDataSet(ctx context.Context, dataSet *upload.Upload) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if err := validateDataSet(dataSet); err != nil {
		return err
	}
	if dataSet.DeviceID == nil || *dataSet.DeviceID == "" {
		return errors.New("data set device id is missing")
	}

	if d.IsClosed() {
		return errors.New("session closed")
	}

	now := time.Now()
	timestamp := now.Truncate(time.Millisecond).Format(time.RFC3339Nano)

	var updateInfo *mgo.ChangeInfo

	var hashes []string
	selector := bson.M{
		"_userId":  dataSet.UserID,
		"uploadId": dataSet.UploadID,
		"type":     bson.M{"$ne": "upload"},
	}
	err := d.C().Find(selector).Distinct("_deduplicator.hash", &hashes)
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
		updateInfo, err = d.C().UpdateAll(selector, d.ConstructUpdate(set, unset))
	}

	loggerFields := log.Fields{"userId": dataSet.UserID, "deviceId": *dataSet.DeviceID, "updateInfo": updateInfo, "duration": time.Since(now) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("ArchiveDeviceDataUsingHashesFromDataSet")

	if err != nil {
		return errors.Wrap(err, "unable to archive device data using hashes from data set")
	}
	return nil
}

func (d *DataSession) UnarchiveDeviceDataUsingHashesFromDataSet(ctx context.Context, dataSet *upload.Upload) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if err := validateDataSet(dataSet); err != nil {
		return err
	}
	if dataSet.DeviceID == nil || *dataSet.DeviceID == "" {
		return errors.New("data set device id is missing")
	}

	if d.IsClosed() {
		return errors.New("session closed")
	}

	now := time.Now()
	timestamp := now.Truncate(time.Millisecond).Format(time.RFC3339Nano)

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
	pipe := d.C().Pipe(pipeline)
	iter := pipe.Iter()

	var overallUpdateInfo mgo.ChangeInfo
	var overallErr error

	result := struct {
		ID struct {
			Active            bool   `bson:"_active"`
			ArchivedDataSetID string `bson:"archivedDatasetId"`
			ArchivedTime      string `bson:"archivedTime"`
		} `bson:"_id"`
		ArchivedHashes []string `bson:"archivedHashes"`
	}{}
	for iter.Next(&result) {
		if result.ID.Active != (result.ID.ArchivedDataSetID == "") || result.ID.Active != (result.ID.ArchivedTime == "") {
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
		updateInfo, err := d.C().UpdateAll(selector, d.ConstructUpdate(set, unset))
		if err != nil {
			loggerFields := log.Fields{"dataSetId": dataSet.UploadID, "result": result}
			log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Error("Unable to update result for UnarchiveDeviceDataUsingHashesFromDataSet")
			if overallErr == nil {
				overallErr = errors.Wrap(err, "unable to transfer device data active")
			}
		} else {
			overallUpdateInfo.Updated += updateInfo.Updated
			overallUpdateInfo.Removed += updateInfo.Removed
		}
	}

	if err := iter.Err(); err != nil {
		if overallErr == nil {
			overallErr = errors.Wrap(err, "unable to iterate to transfer device data active")
		}
	}

	loggerFields := log.Fields{"dataSetId": dataSet.UploadID, "updateInfo": overallUpdateInfo, "duration": time.Since(now) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(overallErr).Debug("UnarchiveDeviceDataUsingHashesFromDataSet")

	return overallErr
}

func (d *DataSession) DeleteOtherDataSetData(ctx context.Context, dataSet *upload.Upload) error {
	if ctx == nil {
		return errors.New("context is missing")
	}

	if err := validateDataSet(dataSet); err != nil {
		return err
	}
	if dataSet.DeviceID == nil || *dataSet.DeviceID == "" {
		return errors.New("data set device id is missing")
	}

	if d.IsClosed() {
		return errors.New("session closed")
	}

	now := time.Now()
	timestamp := now.Truncate(time.Millisecond).Format(time.RFC3339Nano)

	var err error
	var removeInfo *mgo.ChangeInfo
	var updateInfo *mgo.ChangeInfo

	selector := bson.M{
		"_userId":  dataSet.UserID,
		"deviceId": *dataSet.DeviceID,
		"uploadId": bson.M{"$ne": dataSet.UploadID},
		"type":     bson.M{"$ne": "upload"},
	}
	removeInfo, err = d.C().RemoveAll(selector)
	if err == nil {
		selector = bson.M{
			"_userId":       dataSet.UserID,
			"deviceId":      *dataSet.DeviceID,
			"uploadId":      bson.M{"$ne": dataSet.UploadID},
			"type":          "upload",
			"deletedTime":   bson.M{"$exists": false},
			"deletedUserId": bson.M{"$exists": false},
		}
		set := bson.M{
			"deletedTime": timestamp,
		}
		unset := bson.M{}
		updateInfo, err = d.C().UpdateAll(selector, d.ConstructUpdate(set, unset))
	}

	loggerFields := log.Fields{"dataSetId": dataSet.UploadID, "removeInfo": removeInfo, "updateInfo": updateInfo, "duration": time.Since(now) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("DeleteOtherDataSetData")

	if err != nil {
		return errors.Wrap(err, "unable to remove other data set data")
	}
	return nil
}

func (d *DataSession) DestroyDataForUserByID(ctx context.Context, userID string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if userID == "" {
		return errors.New("user id is missing")
	}

	if d.IsClosed() {
		return errors.New("session closed")
	}

	now := time.Now()

	selector := bson.M{
		"_userId": userID,
	}
	removeInfo, err := d.C().RemoveAll(selector)

	loggerFields := log.Fields{"userId": userID, "removeInfo": removeInfo, "duration": time.Since(now) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("DestroyDataForUserByID")

	if err != nil {
		return errors.Wrap(err, "unable to destroy data for user by id")
	}

	return nil
}

func (d *DataSession) ListUserDataSets(ctx context.Context, userID string, filter *data.DataSetFilter, pagination *page.Pagination) (data.DataSets, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	}
	if filter == nil {
		filter = data.NewDataSetFilter()
	} else if err := structureValidator.New().Validate(filter); err != nil {
		return nil, errors.Wrap(err, "filter is invalid")
	}
	if pagination == nil {
		pagination = page.NewPagination()
	} else if err := structureValidator.New().Validate(pagination); err != nil {
		return nil, errors.Wrap(err, "pagination is invalid")
	}

	if d.IsClosed() {
		return nil, errors.New("session closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": userID, "filter": filter, "pagination": pagination})

	dataSets := data.DataSets{}
	selector := bson.M{
		"_active": true,
		"_userId": userID,
		"type":    "upload",
	}
	if filter.ClientName != nil {
		selector["client.name"] = *filter.ClientName
	}
	if filter.Deleted == nil || !*filter.Deleted {
		selector["deletedTime"] = bson.M{"$exists": false}
	}
	if filter.DeviceID != nil {
		selector["deviceId"] = *filter.DeviceID
	}
	err := d.C().Find(selector).Sort("-createdTime").Skip(pagination.Page * pagination.Size).Limit(pagination.Size).All(&dataSets)
	logger.WithFields(log.Fields{"count": len(dataSets), "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("ListUserDataSets")
	if err != nil {
		return nil, errors.Wrap(err, "unable to list user data sets")
	}

	if dataSets == nil {
		dataSets = data.DataSets{}
	}

	return dataSets, nil
}

func (d *DataSession) GetDataSet(ctx context.Context, id string) (*data.DataSet, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	}

	if d.IsClosed() {
		return nil, errors.New("session closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithField("id", id)

	dataSets := data.DataSets{}
	selector := bson.M{
		"uploadId": id,
		"type":     "upload",
	}
	err := d.C().Find(selector).Limit(2).All(&dataSets)
	logger.WithField("duration", time.Since(now)/time.Microsecond).WithError(err).Debug("GetDataSet")
	if err != nil {
		return nil, errors.Wrap(err, "unable to get data set")
	}

	switch count := len(dataSets); count {
	case 0:
		return nil, nil
	case 1:
		return dataSets[0], nil
	default:
		logger.WithField("count", count).Warnf("Multiple data sets found for id %q", id)
		return dataSets[0], nil
	}
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
