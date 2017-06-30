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

func (s *Store) NewSession(logger log.Logger) (store.Session, error) {
	baseSession, err := s.Store.NewSession(logger)
	if err != nil {
		return nil, err
	}

	return &Session{
		Session: baseSession,
	}, nil
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

func (s *Session) FindPreviousActiveDatasetForDevice(dataset *upload.Upload) (*upload.Upload, error) {
	if dataset == nil {
		return nil, errors.New("mongo", "dataset is missing")
	}
	if dataset.UserID == "" {
		return nil, errors.New("mongo", "dataset user id is missing")
	}
	if dataset.GroupID == "" {
		return nil, errors.New("mongo", "dataset group id is missing")
	}
	if dataset.DeviceID == nil || *dataset.DeviceID == "" {
		return nil, errors.New("mongo", "dataset device id is missing")
	}
	if dataset.Deduplicator == nil || dataset.Deduplicator.Name == "" {
		return nil, errors.New("mongo", "dataset deduplicator name is missing")
	}
	if dataset.CreatedTime == "" {
		return nil, errors.New("mongo", "dataset created time is missing")
	}

	if s.IsClosed() {
		return nil, errors.New("mongo", "session closed")
	}

	startTime := time.Now()

	var previousDataset *upload.Upload
	query := bson.M{
		"_userId":            dataset.UserID,
		"_groupId":           dataset.GroupID,
		"deviceId":           *dataset.DeviceID,
		"type":               "upload",
		"_deduplicator.name": dataset.Deduplicator.Name,
		"createdTime": bson.M{
			"$exists": true,
			"$ne":     "",
			"$lt":     dataset.CreatedTime,
		},
	}
	err := s.C().Find(query).Sort("-createdTime").Limit(1).One(&previousDataset)
	if err == mgo.ErrNotFound {
		err = nil
		previousDataset = nil
	}

	loggerFields := log.Fields{"datasetId": dataset.UploadID, "duration": time.Since(startTime) / time.Microsecond}
	if previousDataset != nil {
		loggerFields["previousDatasetID"] = previousDataset.UploadID
	}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("FindPreviousActiveDatasetForDevice")

	if err != nil {
		return nil, errors.Wrap(err, "mongo", "unable to find previous active dataset for device")
	}

	return previousDataset, nil
}

func (s *Session) CreateDataset(dataset *upload.Upload) error {
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

	if s.IsClosed() {
		return errors.New("mongo", "session closed")
	}

	startTime := time.Now()

	deletedTimestamp := s.Timestamp()
	deletedUserID := s.AgentUserID()

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
			"deletedTime": deletedTimestamp,
		}
		if deletedUserID != "" {
			set["deletedUserId"] = deletedUserID
		}
		update := bson.M{
			"$set": set,
		}
		updateInfo, err = s.C().UpdateAll(selector, update)
	}

	loggerFields := log.Fields{"datasetId": dataset.UploadID, "removeInfo": removeInfo, "updateInfo": updateInfo, "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("DeleteDataset")

	if err != nil {
		return errors.Wrap(err, "mongo", "unable to delete dataset")
	}

	dataset.SetDeletedTime(deletedTimestamp)
	dataset.SetDeletedUserID(deletedUserID)
	return nil
}

func (s *Session) GetDatasetDataDeduplicatorHashes(dataset *upload.Upload, active bool) ([]string, error) {
	if dataset == nil {
		return nil, errors.New("mongo", "dataset is missing")
	}
	if dataset.UserID == "" {
		return nil, errors.New("mongo", "dataset user id is missing")
	}
	if dataset.GroupID == "" {
		return nil, errors.New("mongo", "dataset group id is missing")
	}
	if dataset.UploadID == "" {
		return nil, errors.New("mongo", "dataset upload id is missing")
	}

	if s.IsClosed() {
		return nil, errors.New("mongo", "session closed")
	}

	startTime := time.Now()

	var foundHashes []string
	query := bson.M{
		"_userId":  dataset.UserID,
		"_groupId": dataset.GroupID,
		"uploadId": dataset.UploadID,
		"type":     bson.M{"$ne": "upload"},
		"_active":  active,
	}
	err := s.C().Find(query).Distinct("_deduplicator.hash", &foundHashes)

	loggerFields := log.Fields{"datasetId": dataset.UploadID, "foundHashesCount": len(foundHashes), "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("GetDatasetDataDeduplicatorHashes")

	if err != nil {
		return nil, errors.Wrap(err, "mongo", "unable to get dataset data deduplicator hashes")
	}

	if len(foundHashes) == 0 {
		foundHashes = nil
	}

	return foundHashes, nil
}

func (s *Session) CreateDatasetData(dataset *upload.Upload, datasetData []data.Datum) error {
	if dataset == nil {
		return errors.New("mongo", "dataset is missing")
	}
	if datasetData == nil {
		return errors.New("mongo", "dataset data is missing")
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

	if s.IsClosed() {
		return errors.New("mongo", "session closed")
	}

	startTime := time.Now()

	createdTimestamp := s.Timestamp()
	createdUserID := s.AgentUserID()

	insertData := make([]interface{}, len(datasetData))
	for index, datum := range datasetData {
		datum.SetUserID(dataset.UserID)
		datum.SetGroupID(dataset.GroupID)
		datum.SetDatasetID(dataset.UploadID)
		datum.SetCreatedTime(createdTimestamp)
		datum.SetCreatedUserID(createdUserID)
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

	if s.IsClosed() {
		return errors.New("mongo", "session closed")
	}

	startTime := time.Now()

	modifiedTimestamp := s.Timestamp()
	modifiedUserID := s.AgentUserID()

	selector := bson.M{
		"_userId":  dataset.UserID,
		"_groupId": dataset.GroupID,
		"uploadId": dataset.UploadID,
	}
	set := bson.M{
		"_active":      true,
		"modifiedTime": modifiedTimestamp,
	}
	if modifiedUserID != "" {
		set["modifiedUserId"] = modifiedUserID
	}
	update := bson.M{
		"$set": set,
	}
	updateInfo, err := s.C().UpdateAll(selector, update)

	loggerFields := log.Fields{"datasetId": dataset.UploadID, "updateInfo": updateInfo, "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("ActivateDatasetData")

	if err != nil {
		return errors.Wrap(err, "mongo", "unable to activate dataset data")
	}

	dataset.SetActive(true)
	dataset.SetModifiedTime(modifiedTimestamp)
	dataset.SetModifiedUserID(modifiedUserID)
	return nil
}

func (s *Session) SetDatasetDataActiveUsingHashes(dataset *upload.Upload, queryHashes []string, active bool) error {
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

	if s.IsClosed() {
		return errors.New("mongo", "session closed")
	}

	if len(queryHashes) == 0 {
		return nil
	}

	startTime := time.Now()

	modifiedTimestamp := s.Timestamp()
	modifiedUserID := s.AgentUserID()

	var err error
	var updateInfo *mgo.ChangeInfo

	selector := bson.M{
		"_userId":  dataset.UserID,
		"_groupId": dataset.GroupID,
		"uploadId": dataset.UploadID,
		"type":     bson.M{"$ne": "upload"},
		"_active":  !active,
		"_deduplicator.hash": bson.M{
			"$in": queryHashes,
		},
	}
	update := map[string]bson.M{
		"$set": {
			"_active":      active,
			"modifiedTime": modifiedTimestamp,
		},
	}
	if modifiedUserID != "" {
		update["$set"]["modifiedUserId"] = modifiedUserID
	} else {
		update["$unset"] = bson.M{"modifiedUserId": ""}
	}
	updateInfo, err = s.C().UpdateAll(selector, update)

	loggerFields := log.Fields{"datasetId": dataset.UploadID, "updateInfo": updateInfo, "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("SetDatasetDataActiveUsingHashes")

	if err != nil {
		return errors.Wrap(err, "mongo", "unable to set dataset data active using hashes")
	}
	return nil
}

func (s *Session) SetDeviceDataActiveUsingHashes(dataset *upload.Upload, queryHashes []string, active bool) error {
	if dataset == nil {
		return errors.New("mongo", "dataset is missing")
	}
	if dataset.UserID == "" {
		return errors.New("mongo", "dataset user id is missing")
	}
	if dataset.GroupID == "" {
		return errors.New("mongo", "dataset group id is missing")
	}
	if dataset.DeviceID == nil || *dataset.DeviceID == "" {
		return errors.New("mongo", "dataset device id is missing")
	}

	if s.IsClosed() {
		return errors.New("mongo", "session closed")
	}

	if len(queryHashes) == 0 {
		return nil
	}

	startTime := time.Now()

	modifiedTimestamp := s.Timestamp()
	modifiedUserID := s.AgentUserID()

	var err error
	var updateInfo *mgo.ChangeInfo

	selector := bson.M{
		"_userId":  dataset.UserID,
		"_groupId": dataset.GroupID,
		"deviceId": *dataset.DeviceID,
		"type":     bson.M{"$ne": "upload"},
		"_active":  !active,
		"_deduplicator.hash": bson.M{
			"$in": queryHashes,
		},
	}
	update := map[string]bson.M{
		"$set": {
			"_active":      active,
			"modifiedTime": modifiedTimestamp,
		},
	}
	if modifiedUserID != "" {
		update["$set"]["modifiedUserId"] = modifiedUserID
	} else {
		update["$unset"] = bson.M{"modifiedUserId": ""}
	}
	updateInfo, err = s.C().UpdateAll(selector, update)

	loggerFields := log.Fields{"deviceID": *dataset.DeviceID, "updateInfo": updateInfo, "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("SetDeviceDataActiveUsingHashes")

	if err != nil {
		return errors.Wrap(err, "mongo", "unable to set device data active using hashes")
	}
	return nil
}

func (s *Session) DeleteOtherDatasetData(dataset *upload.Upload) error {
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

	if s.IsClosed() {
		return errors.New("mongo", "session closed")
	}

	startTime := time.Now()

	deletedTimestamp := s.Timestamp()
	deletedUserID := s.AgentUserID()

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
			"deletedTime": deletedTimestamp,
		}
		if deletedUserID != "" {
			set["deletedUserId"] = deletedUserID
		}
		update := bson.M{
			"$set": set,
		}
		updateInfo, err = s.C().UpdateAll(selector, update)
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
