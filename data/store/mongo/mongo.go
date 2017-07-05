package mongo

import (
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/store/mongo"
)

func New(logger log.Logger, config *mongo.Config) (*Store, error) {
	baseStore, err := mongo.New(logger, config)
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

func (s *Store) NewSession(logger log.Logger) store.Session {
	return &Session{
		Session: s.Store.NewSession(logger),
	}
}

type Session struct {
	*mongo.Session
}

func (s *Session) GetDatasetsForUserByID(userID string, filter *store.Filter, pagination *store.Pagination) ([]*upload.Upload, error) {
	if userID == "" {
		return nil, errors.New("mongo", "user id is missing")
	}
	if filter == nil {
		filter = store.NewFilter()
	} else if err := filter.Validate(); err != nil {
		return nil, errors.Wrap(err, "mongo", "filter is invalid")
	}
	if pagination == nil {
		pagination = store.NewPagination()
	} else if err := pagination.Validate(); err != nil {
		return nil, errors.Wrap(err, "mongo", "pagination is invalid")
	}

	if s.IsClosed() {
		return nil, errors.New("mongo", "session closed")
	}

	startTime := time.Now()

	var datasets []*upload.Upload
	query := bson.M{
		"_userId": userID,
		"type":    "upload",
	}
	if !filter.Deleted {
		query["deletedTime"] = bson.M{"$exists": false}
	}
	err := s.C().Find(query).Sort("-createdTime").Skip(pagination.Page * pagination.Size).Limit(pagination.Size).All(&datasets)

	loggerFields := log.Fields{"userId": userID, "datasetsCount": len(datasets), "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("GetDatasetsForUserByID")

	if err != nil {
		return nil, errors.Wrap(err, "mongo", "unable to get datasets for user by id")
	}

	if datasets == nil {
		datasets = []*upload.Upload{}
	}
	return datasets, nil
}

func (s *Session) GetDatasetByID(datasetID string) (*upload.Upload, error) {
	if datasetID == "" {
		return nil, errors.New("mongo", "dataset id is missing")
	}

	if s.IsClosed() {
		return nil, errors.New("mongo", "session closed")
	}

	startTime := time.Now()

	datasets := []*upload.Upload{}
	query := bson.M{
		"uploadId": datasetID,
		"type":     "upload",
	}
	err := s.C().Find(query).Limit(2).All(&datasets)

	loggerFields := log.Fields{"datasetId": datasetID, "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("GetDatasetByID")

	if err != nil {
		return nil, errors.Wrap(err, "mongo", "unable to get dataset by id")
	}

	if datasetsCount := len(datasets); datasetsCount == 0 {
		return nil, nil
	} else if datasetsCount > 1 {
		s.Logger().WithField("datasetId", datasetID).Warn("Multiple datasets found for dataset id")
	}

	return datasets[0], nil
}

func (s *Session) CreateDataset(dataset *upload.Upload) error {
	if err := s.validateDataset(dataset); err != nil {
		return err
	}

	if s.IsClosed() {
		return errors.New("mongo", "session closed")
	}

	startTime := time.Now()

	dataset.CreatedTime = s.Timestamp()
	dataset.CreatedUserID = s.AgentUserID()

	dataset.ByUser = dataset.CreatedUserID

	// TODO: Consider upsert instead to prevent multiples being created?

	query := bson.M{
		"_userId":  dataset.UserID,
		"_groupId": dataset.GroupID,
		"uploadId": dataset.UploadID,
		"type":     dataset.Type,
	}
	count, err := s.C().Find(query).Count()
	if err == nil {
		if count > 0 {
			err = errors.New("mongo", "dataset already exists")
		} else {
			err = s.C().Insert(dataset)
		}
	}

	loggerFields := log.Fields{"userId": dataset.UserID, "datasetId": dataset.UploadID, "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("CreateDataset")

	if err != nil {
		return errors.Wrap(err, "mongo", "unable to create dataset")
	}
	return nil
}

func (s *Session) UpdateDataset(dataset *upload.Upload) error {
	if err := s.validateDataset(dataset); err != nil {
		return err
	}

	if s.IsClosed() {
		return errors.New("mongo", "session closed")
	}

	startTime := time.Now()

	dataset.ModifiedTime = s.Timestamp()
	dataset.ModifiedUserID = s.AgentUserID()

	selector := bson.M{
		"_userId":  dataset.UserID,
		"_groupId": dataset.GroupID,
		"uploadId": dataset.UploadID,
		"type":     dataset.Type,
	}
	err := s.C().Update(selector, dataset)

	loggerFields := log.Fields{"datasetId": dataset.UploadID, "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("UpdateDataset")

	if err != nil {
		return errors.Wrap(err, "mongo", "unable to update dataset")
	}
	return nil
}

func (s *Session) DeleteDataset(dataset *upload.Upload) error {
	if err := s.validateDataset(dataset); err != nil {
		return err
	}

	if s.IsClosed() {
		return errors.New("mongo", "session closed")
	}

	startTime := time.Now()

	timestamp := s.Timestamp()
	agentUserID := s.AgentUserID()

	var err error
	var removeInfo *mgo.ChangeInfo
	var updateInfo *mgo.ChangeInfo

	selector := bson.M{
		"_userId":  dataset.UserID,
		"_groupId": dataset.GroupID,
		"uploadId": dataset.UploadID,
		"type":     bson.M{"$ne": "upload"},
	}
	removeInfo, err = s.C().RemoveAll(selector)
	if err == nil {
		selector = bson.M{
			"_userId":       dataset.UserID,
			"_groupId":      dataset.GroupID,
			"uploadId":      dataset.UploadID,
			"type":          "upload",
			"deletedTime":   bson.M{"$exists": false},
			"deletedUserId": bson.M{"$exists": false},
		}
		set := bson.M{
			"deletedTime": timestamp,
		}
		unset := bson.M{}
		if agentUserID != "" {
			set["deletedUserId"] = agentUserID
		} else {
			unset["deletedUserId"] = true
		}
		updateInfo, err = s.C().UpdateAll(selector, s.constructUpdate(set, unset))
	}

	loggerFields := log.Fields{"datasetId": dataset.UploadID, "removeInfo": removeInfo, "updateInfo": updateInfo, "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("DeleteDataset")

	if err != nil {
		return errors.Wrap(err, "mongo", "unable to delete dataset")
	}

	dataset.SetDeletedTime(timestamp)
	dataset.SetDeletedUserID(agentUserID)
	return nil
}

func (s *Session) CreateDatasetData(dataset *upload.Upload, datasetData []data.Datum) error {
	if err := s.validateDataset(dataset); err != nil {
		return err
	}
	if datasetData == nil {
		return errors.New("mongo", "dataset data is missing")
	}

	if s.IsClosed() {
		return errors.New("mongo", "session closed")
	}

	startTime := time.Now()

	timestamp := s.Timestamp()
	agentUserID := s.AgentUserID()

	insertData := make([]interface{}, len(datasetData))
	for index, datum := range datasetData {
		datum.SetUserID(dataset.UserID)
		datum.SetGroupID(dataset.GroupID)
		datum.SetDatasetID(dataset.UploadID)
		datum.SetCreatedTime(timestamp)
		datum.SetCreatedUserID(agentUserID)
		insertData[index] = datum
	}

	bulk := s.C().Bulk()
	bulk.Unordered()
	bulk.Insert(insertData...)

	_, err := bulk.Run()

	loggerFields := log.Fields{"datasetId": dataset.UploadID, "dataCount": len(datasetData), "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("CreateDatasetData")

	if err != nil {
		return errors.Wrap(err, "mongo", "unable to create dataset data")
	}
	return nil
}

func (s *Session) ActivateDatasetData(dataset *upload.Upload) error {
	if err := s.validateDataset(dataset); err != nil {
		return err
	}

	if s.IsClosed() {
		return errors.New("mongo", "session closed")
	}

	startTime := time.Now()

	timestamp := s.Timestamp()
	agentUserID := s.AgentUserID()

	selector := bson.M{
		"_userId":  dataset.UserID,
		"_groupId": dataset.GroupID,
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
	if agentUserID != "" {
		set["modifiedUserId"] = agentUserID
	} else {
		unset["modifiedUserId"] = true
	}
	updateInfo, err := s.C().UpdateAll(selector, s.constructUpdate(set, unset))

	loggerFields := log.Fields{"datasetId": dataset.UploadID, "updateInfo": updateInfo, "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("ActivateDatasetData")

	if err != nil {
		return errors.Wrap(err, "mongo", "unable to activate dataset data")
	}

	dataset.SetActive(true)
	dataset.SetModifiedTime(timestamp)
	dataset.SetModifiedUserID(agentUserID)
	return nil
}

func (s *Session) ArchiveDeviceDataUsingHashesFromDataset(dataset *upload.Upload) error {
	if err := s.validateDataset(dataset); err != nil {
		return err
	}

	if s.IsClosed() {
		return errors.New("mongo", "session closed")
	}

	startTime := time.Now()

	timestamp := s.Timestamp()
	agentUserID := s.AgentUserID()

	var updateInfo *mgo.ChangeInfo

	var hashes []string
	query := bson.M{
		"uploadId": dataset.UploadID,
		"type":     bson.M{"$ne": "upload"},
	}
	err := s.C().Find(query).Distinct("_deduplicator.hash", &hashes)
	if err == nil && len(hashes) > 0 {
		selector := bson.M{
			"_userId":            dataset.UserID,
			"_groupId":           dataset.GroupID,
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
		if agentUserID != "" {
			set["modifiedUserId"] = agentUserID
		} else {
			unset["modifiedUserId"] = true
		}
		updateInfo, err = s.C().UpdateAll(selector, s.constructUpdate(set, unset))
	}

	loggerFields := log.Fields{"userId": dataset.UserID, "deviceId": *dataset.DeviceID, "updateInfo": updateInfo, "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("ArchiveDeviceDataUsingHashesFromDataset")

	if err != nil {
		return errors.Wrap(err, "mongo", "unable to archive device data using hashes from dataset")
	}
	return nil
}

func (s *Session) UnarchiveDeviceDataUsingHashesFromDataset(dataset *upload.Upload) error {
	if err := s.validateDataset(dataset); err != nil {
		return err
	}

	if s.IsClosed() {
		return errors.New("mongo", "session closed")
	}

	startTime := time.Now()

	timestamp := s.Timestamp()
	agentUserID := s.AgentUserID()

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
	pipe := s.C().Pipe(pipeline)
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
			loggerFields := log.Fields{"datasetId": dataset.UploadID, "result": result}
			s.Logger().WithFields(loggerFields).Error("Unexpected pipe result for UnarchiveDeviceDataUsingHashesFromDataset")
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
		if agentUserID != "" {
			set["modifiedUserId"] = agentUserID
		} else {
			unset["modifiedUserId"] = true
		}
		if result.ID.Active {
			unset["archivedDatasetId"] = true
			unset["archivedTime"] = true
		} else {
			set["archivedDatasetId"] = result.ID.ArchivedDatasetID
			set["archivedTime"] = result.ID.ArchivedTime
		}
		updateInfo, err := s.C().UpdateAll(selector, s.constructUpdate(set, unset))
		if err != nil {
			loggerFields := log.Fields{"datasetId": dataset.UploadID, "result": result}
			s.Logger().WithFields(loggerFields).WithError(err).Error("Unable to update result for UnarchiveDeviceDataUsingHashesFromDataset")
			if overallErr == nil {
				overallErr = errors.Wrap(err, "mongo", "unable to transfer device data active")
			}
		} else {
			overallUpdateInfo.Updated += updateInfo.Updated
			overallUpdateInfo.Removed += updateInfo.Removed
		}
	}

	if err := iter.Err(); err != nil {
		if overallErr == nil {
			overallErr = errors.Wrap(err, "mongo", "unable to iterate to transfer device data active")
		}
	}

	loggerFields := log.Fields{"datasetId": dataset.UploadID, "updateInfo": overallUpdateInfo, "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(overallErr).Debug("UnarchiveDeviceDataUsingHashesFromDataset")

	return overallErr
}

func (s *Session) DeleteOtherDatasetData(dataset *upload.Upload) error {
	if err := s.validateDataset(dataset); err != nil {
		return err
	}

	if s.IsClosed() {
		return errors.New("mongo", "session closed")
	}

	startTime := time.Now()

	timestamp := s.Timestamp()
	agentUserID := s.AgentUserID()

	var err error
	var removeInfo *mgo.ChangeInfo
	var updateInfo *mgo.ChangeInfo

	selector := bson.M{
		"_userId":  dataset.UserID,
		"_groupId": dataset.GroupID,
		"deviceId": *dataset.DeviceID,
		"uploadId": bson.M{"$ne": dataset.UploadID},
		"type":     bson.M{"$ne": "upload"},
	}
	removeInfo, err = s.C().RemoveAll(selector)
	if err == nil {
		selector = bson.M{
			"_userId":       dataset.UserID,
			"_groupId":      dataset.GroupID,
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
		if agentUserID != "" {
			set["deletedUserId"] = agentUserID
		} else {
			unset["deletedUserId"] = true
		}
		updateInfo, err = s.C().UpdateAll(selector, s.constructUpdate(set, unset))
	}

	loggerFields := log.Fields{"datasetId": dataset.UploadID, "removeInfo": removeInfo, "updateInfo": updateInfo, "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("DeleteOtherDatasetData")

	if err != nil {
		return errors.Wrap(err, "mongo", "unable to remove other dataset data")
	}
	return nil
}

func (s *Session) DestroyDataForUserByID(userID string) error {
	if userID == "" {
		return errors.New("mongo", "user id is missing")
	}

	if s.IsClosed() {
		return errors.New("mongo", "session closed")
	}

	startTime := time.Now()

	selector := bson.M{
		"_userId": userID,
	}
	removeInfo, err := s.C().RemoveAll(selector)

	loggerFields := log.Fields{"userId": userID, "removeInfo": removeInfo, "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("DestroyDataForUserByID")

	if err != nil {
		return errors.Wrap(err, "mongo", "unable to destroy data for user by id")
	}

	return nil
}

func (s *Session) validateDataset(dataset *upload.Upload) error {
	if dataset == nil {
		return errors.New("mongo", "dataset is missing")
	}
	if dataset.UserID == "" {
		return errors.New("mongo", "dataset user id is missing")
	}
	if dataset.GroupID == "" {
		return errors.New("mongo", "dataset group id is missing")
	}
	if dataset.UploadID == "" {
		return errors.New("mongo", "dataset upload id is missing")
	}
	if dataset.DeviceID == nil || *dataset.DeviceID == "" {
		return errors.New("mongo", "dataset device id is missing")
	}

	return nil
}

func (s *Session) constructUpdate(set bson.M, unset bson.M) bson.M {
	update := bson.M{}
	if len(set) > 0 {
		update["$set"] = set
	}
	if len(unset) > 0 {
		update["$unset"] = unset
	}
	return update
}
