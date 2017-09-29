package mongo

import (
	"context"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/storeDEPRECATED"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/store/mongo"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func New(cfg *mongo.Config, lgr log.Logger) (*Store, error) {
	baseStore, err := mongo.New(cfg, lgr)
	if err != nil {
		return nil, err
	}

	return &Store{
		Store: baseStore,
	}, nil
}

type Store struct {
	*mongo.Store
}

func (s *Store) NewDataSession() storeDEPRECATED.DataSession {
	return &DataSession{
		Session: s.Store.NewSession("deviceData"),
	}
}

type DataSession struct {
	*mongo.Session
}

func (d *DataSession) GetDatasetsForUserByID(ctx context.Context, userID string, filter *storeDEPRECATED.Filter, pagination *page.Pagination) ([]*upload.Upload, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	}
	if filter == nil {
		filter = storeDEPRECATED.NewFilter()
	} else if err := filter.Validate(); err != nil {
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

	startTime := time.Now()

	var datasets []*upload.Upload
	selector := bson.M{
		"_userId": userID,
		"type":    "upload",
	}
	if !filter.Deleted {
		selector["deletedTime"] = bson.M{"$exists": false}
	}
	err := d.C().Find(selector).Sort("-createdTime").Skip(pagination.Page * pagination.Size).Limit(pagination.Size).All(&datasets)

	loggerFields := log.Fields{"userId": userID, "datasetsCount": len(datasets), "duration": time.Since(startTime) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("GetDatasetsForUserByID")

	if err != nil {
		return nil, errors.Wrap(err, "unable to get datasets for user by id")
	}

	if datasets == nil {
		datasets = []*upload.Upload{}
	}
	return datasets, nil
}

func (d *DataSession) GetDatasetByID(ctx context.Context, datasetID string) (*upload.Upload, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if datasetID == "" {
		return nil, errors.New("dataset id is missing")
	}

	if d.IsClosed() {
		return nil, errors.New("session closed")
	}

	startTime := time.Now()

	datasets := []*upload.Upload{}
	selector := bson.M{
		"uploadId": datasetID,
		"type":     "upload",
	}
	err := d.C().Find(selector).Limit(2).All(&datasets)

	loggerFields := log.Fields{"dataSetId": datasetID, "duration": time.Since(startTime) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("GetDatasetByID")

	if err != nil {
		return nil, errors.Wrap(err, "unable to get dataset by id")
	}

	if datasetsCount := len(datasets); datasetsCount == 0 {
		return nil, nil
	} else if datasetsCount > 1 {
		log.LoggerFromContext(ctx).WithField("dataSetId", datasetID).Warn("Multiple datasets found for dataset id")
	}

	return datasets[0], nil
}

func (d *DataSession) CreateDataset(ctx context.Context, dataset *upload.Upload) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if err := d.validateDataset(dataset); err != nil {
		return err
	}

	if d.IsClosed() {
		return errors.New("session closed")
	}

	startTime := time.Now()

	dataset.CreatedTime = d.Timestamp()

	dataset.ByUser = dataset.CreatedUserID

	// TODO: Consider upsert instead to prevent multiples being created?

	selector := bson.M{
		"_userId":  dataset.UserID,
		"uploadId": dataset.UploadID,
		"type":     dataset.Type,
	}
	count, err := d.C().Find(selector).Count()
	if err == nil {
		if count > 0 {
			err = errors.New("dataset already exists")
		} else {
			err = d.C().Insert(dataset)
		}
	}

	loggerFields := log.Fields{"userId": dataset.UserID, "dataSetId": dataset.UploadID, "duration": time.Since(startTime) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("CreateDataset")

	if err != nil {
		return errors.Wrap(err, "unable to create dataset")
	}
	return nil
}

func (d *DataSession) UpdateDataSet(ctx context.Context, id string, update *data.DataSetUpdate) (*upload.Upload, error) {
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

	if d.IsClosed() {
		return nil, errors.New("session closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"id": id, "update": update})

	set := bson.M{
		"modifiedTime": now.Truncate(time.Second),
	}
	unset := bson.M{}
	if update.Active != nil {
		set["_active"] = *update.Active
	}
	if update.Deduplicator != nil {
		set["_deduplicator"] = update.Deduplicator
	}
	if update.State != nil {
		set["state"] = *update.State
	}
	if update.Time != nil {
		set["time"] = (*update.Time).Format(data.TimeFormat)
	}
	if update.Timezone != nil {
		set["timezone"] = *update.Timezone
	}
	if update.TimezoneOffset != nil {
		set["timezoneOffset"] = *update.TimezoneOffset
	}
	changeInfo, err := d.C().UpdateAll(bson.M{"type": "upload", "uploadId": id}, d.ConstructUpdate(set, unset))
	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("UpdateDataSet")
	if err != nil {
		return nil, errors.Wrap(err, "unable to update data set")
	}

	return d.GetDatasetByID(ctx, id)
}

func (d *DataSession) DeleteDataset(ctx context.Context, dataset *upload.Upload) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if err := d.validateDataset(dataset); err != nil {
		return err
	}

	if d.IsClosed() {
		return errors.New("session closed")
	}

	startTime := time.Now()

	timestamp := d.Timestamp()

	var err error
	var removeInfo *mgo.ChangeInfo
	var updateInfo *mgo.ChangeInfo

	selector := bson.M{
		"_userId":  dataset.UserID,
		"uploadId": dataset.UploadID,
		"type":     bson.M{"$ne": "upload"},
	}
	removeInfo, err = d.C().RemoveAll(selector)
	if err == nil {
		selector = bson.M{
			"_userId":       dataset.UserID,
			"uploadId":      dataset.UploadID,
			"type":          "upload",
			"deletedTime":   bson.M{"$exists": false},
			"deletedUserId": bson.M{"$exists": false},
		}
		set := bson.M{
			"deletedTime": timestamp,
		}
		unset := bson.M{}
		updateInfo, err = d.C().UpdateAll(selector, d.constructUpdate(set, unset))
	}

	loggerFields := log.Fields{"dataSetId": dataset.UploadID, "removeInfo": removeInfo, "updateInfo": updateInfo, "duration": time.Since(startTime) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("DeleteDataset")

	if err != nil {
		return errors.Wrap(err, "unable to delete dataset")
	}

	dataset.SetDeletedTime(timestamp)
	return nil
}

func (d *DataSession) CreateDatasetData(ctx context.Context, dataset *upload.Upload, datasetData []data.Datum) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if err := d.validateDataset(dataset); err != nil {
		return err
	}
	if datasetData == nil {
		return errors.New("dataset data is missing")
	}

	if d.IsClosed() {
		return errors.New("session closed")
	}

	startTime := time.Now()

	timestamp := d.Timestamp()

	insertData := make([]interface{}, len(datasetData))
	for index, datum := range datasetData {
		datum.SetUserID(dataset.UserID)
		datum.SetDatasetID(dataset.UploadID)
		datum.SetCreatedTime(timestamp)
		insertData[index] = datum
	}

	bulk := d.C().Bulk()
	bulk.Unordered()
	bulk.Insert(insertData...)

	_, err := bulk.Run()

	loggerFields := log.Fields{"dataSetId": dataset.UploadID, "dataCount": len(datasetData), "duration": time.Since(startTime) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("CreateDatasetData")

	if err != nil {
		return errors.Wrap(err, "unable to create dataset data")
	}
	return nil
}

func (d *DataSession) ActivateDatasetData(ctx context.Context, dataset *upload.Upload) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if err := d.validateDataset(dataset); err != nil {
		return err
	}

	if d.IsClosed() {
		return errors.New("session closed")
	}

	startTime := time.Now()

	timestamp := d.Timestamp()

	selector := bson.M{
		"_userId":  dataset.UserID,
		"uploadId": dataset.UploadID,
	}
	set := bson.M{
		"_active":      true,
		"modifiedTime": timestamp,
	}
	unset := bson.M{
		"archivedDatasetId": 1,
		"archivedTime":      1,
	}
	updateInfo, err := d.C().UpdateAll(selector, d.constructUpdate(set, unset))

	loggerFields := log.Fields{"dataSetId": dataset.UploadID, "updateInfo": updateInfo, "duration": time.Since(startTime) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("ActivateDatasetData")

	if err != nil {
		return errors.Wrap(err, "unable to activate dataset data")
	}

	dataset.SetActive(true)
	dataset.SetModifiedTime(timestamp)
	return nil
}

func (d *DataSession) ArchiveDeviceDataUsingHashesFromDataset(ctx context.Context, dataset *upload.Upload) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if err := d.validateDataset(dataset); err != nil {
		return err
	}

	if d.IsClosed() {
		return errors.New("session closed")
	}

	startTime := time.Now()

	timestamp := d.Timestamp()

	var updateInfo *mgo.ChangeInfo

	var hashes []string
	selector := bson.M{
		"uploadId": dataset.UploadID,
		"type":     bson.M{"$ne": "upload"},
	}
	err := d.C().Find(selector).Distinct("_deduplicator.hash", &hashes)
	if err == nil && len(hashes) > 0 {
		selector = bson.M{
			"_userId":            dataset.UserID,
			"deviceId":           *dataset.DeviceID,
			"type":               bson.M{"$ne": "upload"},
			"_active":            true,
			"_deduplicator.hash": bson.M{"$in": hashes},
		}
		set := bson.M{
			"_active":           false,
			"archivedDatasetId": dataset.UploadID,
			"archivedTime":      timestamp,
			"modifiedTime":      timestamp,
		}
		unset := bson.M{}
		updateInfo, err = d.C().UpdateAll(selector, d.constructUpdate(set, unset))
	}

	loggerFields := log.Fields{"userId": dataset.UserID, "deviceId": *dataset.DeviceID, "updateInfo": updateInfo, "duration": time.Since(startTime) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("ArchiveDeviceDataUsingHashesFromDataset")

	if err != nil {
		return errors.Wrap(err, "unable to archive device data using hashes from dataset")
	}
	return nil
}

func (d *DataSession) UnarchiveDeviceDataUsingHashesFromDataset(ctx context.Context, dataset *upload.Upload) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if err := d.validateDataset(dataset); err != nil {
		return err
	}

	if d.IsClosed() {
		return errors.New("session closed")
	}

	startTime := time.Now()

	timestamp := d.Timestamp()

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"uploadId": dataset.UploadID,
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
			ArchivedDatasetID string `bson:"archivedDatasetId"`
			ArchivedTime      string `bson:"archivedTime"`
		} `bson:"_id"`
		ArchivedHashes []string `bson:"archivedHashes"`
	}{}
	for iter.Next(&result) {
		if result.ID.Active != (result.ID.ArchivedDatasetID == "") || result.ID.Active != (result.ID.ArchivedTime == "") {
			loggerFields := log.Fields{"dataSetId": dataset.UploadID, "result": result}
			log.LoggerFromContext(ctx).WithFields(loggerFields).Error("Unexpected pipe result for UnarchiveDeviceDataUsingHashesFromDataset")
			continue
		}

		selector := bson.M{
			"_userId":            dataset.UserID,
			"deviceId":           dataset.DeviceID,
			"archivedDatasetId":  dataset.UploadID,
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
			set["archivedDatasetId"] = result.ID.ArchivedDatasetID
			set["archivedTime"] = result.ID.ArchivedTime
		}
		updateInfo, err := d.C().UpdateAll(selector, d.constructUpdate(set, unset))
		if err != nil {
			loggerFields := log.Fields{"dataSetId": dataset.UploadID, "result": result}
			log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Error("Unable to update result for UnarchiveDeviceDataUsingHashesFromDataset")
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

	loggerFields := log.Fields{"dataSetId": dataset.UploadID, "updateInfo": overallUpdateInfo, "duration": time.Since(startTime) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(overallErr).Debug("UnarchiveDeviceDataUsingHashesFromDataset")

	return overallErr
}

func (d *DataSession) DeleteOtherDatasetData(ctx context.Context, dataset *upload.Upload) error {
	if ctx == nil {
		return errors.New("context is missing")
	}

	if err := d.validateDataset(dataset); err != nil {
		return err
	}

	if d.IsClosed() {
		return errors.New("session closed")
	}

	startTime := time.Now()

	timestamp := d.Timestamp()

	var err error
	var removeInfo *mgo.ChangeInfo
	var updateInfo *mgo.ChangeInfo

	selector := bson.M{
		"_userId":  dataset.UserID,
		"deviceId": *dataset.DeviceID,
		"uploadId": bson.M{"$ne": dataset.UploadID},
		"type":     bson.M{"$ne": "upload"},
	}
	removeInfo, err = d.C().RemoveAll(selector)
	if err == nil {
		selector = bson.M{
			"_userId":       dataset.UserID,
			"deviceId":      *dataset.DeviceID,
			"uploadId":      bson.M{"$ne": dataset.UploadID},
			"type":          "upload",
			"deletedTime":   bson.M{"$exists": false},
			"deletedUserId": bson.M{"$exists": false},
		}
		set := bson.M{
			"deletedTime": timestamp,
		}
		unset := bson.M{}
		updateInfo, err = d.C().UpdateAll(selector, d.constructUpdate(set, unset))
	}

	loggerFields := log.Fields{"dataSetId": dataset.UploadID, "removeInfo": removeInfo, "updateInfo": updateInfo, "duration": time.Since(startTime) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("DeleteOtherDatasetData")

	if err != nil {
		return errors.Wrap(err, "unable to remove other dataset data")
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

	startTime := time.Now()

	selector := bson.M{
		"_userId": userID,
	}
	removeInfo, err := d.C().RemoveAll(selector)

	loggerFields := log.Fields{"userId": userID, "removeInfo": removeInfo, "duration": time.Since(startTime) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("DestroyDataForUserByID")

	if err != nil {
		return errors.Wrap(err, "unable to destroy data for user by id")
	}

	return nil
}

func (d *DataSession) validateDataset(dataset *upload.Upload) error {
	if dataset == nil {
		return errors.New("dataset is missing")
	}
	if dataset.UserID == "" {
		return errors.New("dataset user id is missing")
	}
	if dataset.UploadID == "" {
		return errors.New("dataset upload id is missing")
	}
	if dataset.DeviceID == nil || *dataset.DeviceID == "" {
		return errors.New("dataset device id is missing")
	}

	return nil
}

func (d *DataSession) constructUpdate(set bson.M, unset bson.M) bson.M {
	update := bson.M{}
	if len(set) > 0 {
		update["$set"] = set
	}
	if len(unset) > 0 {
		update["$unset"] = unset
	}
	return update
}
