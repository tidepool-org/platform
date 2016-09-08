package mongo

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import (
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/types/base/upload"
	"github.com/tidepool-org/platform/log"
	commonMongo "github.com/tidepool-org/platform/store/mongo"
)

func New(logger log.Logger, config *commonMongo.Config) (*Store, error) {
	mongoStore, err := commonMongo.New(logger, config)
	if err != nil {
		return nil, err
	}

	return &Store{
		Store: mongoStore,
	}, nil
}

type Store struct {
	*commonMongo.Store
}

func (s *Store) NewSession(logger log.Logger) (store.Session, error) {
	mongoSession, err := s.Store.NewSession(logger)
	if err != nil {
		return nil, err
	}

	return &Session{
		Session: mongoSession,
	}, nil
}

type Session struct {
	*commonMongo.Session
	agent store.Agent
}

func (s *Session) SetAgent(agent store.Agent) {
	s.agent = agent
}

func (s *Session) GetDatasetsForUser(userID string, filter *store.Filter, pagination *store.Pagination) ([]*upload.Upload, error) {
	if userID == "" {
		return nil, app.Error("mongo", "user id is missing")
	}
	if filter == nil {
		filter = store.NewFilter()
	} else if err := filter.Validate(); err != nil {
		return nil, app.ExtError(err, "mongo", "filter is invalid")
	}
	if pagination == nil {
		pagination = store.NewPagination()
	} else if err := pagination.Validate(); err != nil {
		return nil, app.ExtError(err, "mongo", "pagination is invalid")
	}

	if s.IsClosed() {
		return nil, app.Error("mongo", "session closed")
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
	err := s.C().Find(selector).Sort("-createdTime").Skip(pagination.Page * pagination.Size).Limit(pagination.Size).All(&datasets)

	loggerFields := log.Fields{"userID": userID, "datasets-count": len(datasets), "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("GetDatasetsForUser")

	if err != nil {
		return nil, app.ExtError(err, "mongo", "unable to get datasets for user")
	}

	if datasets == nil {
		datasets = []*upload.Upload{}
	}
	return datasets, nil
}

func (s *Session) GetDataset(datasetID string) (*upload.Upload, error) {
	if datasetID == "" {
		return nil, app.Error("mongo", "dataset id is missing")
	}

	if s.IsClosed() {
		return nil, app.Error("mongo", "session closed")
	}

	startTime := time.Now()

	var dataset upload.Upload
	selector := bson.M{
		"uploadId": datasetID,
		"type":     "upload",
	}
	err := s.C().Find(selector).One(&dataset)

	loggerFields := log.Fields{"datasetID": datasetID, "dataset": dataset, "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("GetDataset")

	if err != nil {
		return nil, app.ExtError(err, "mongo", "unable to get dataset")
	}
	return &dataset, nil
}

func (s *Session) CreateDataset(dataset *upload.Upload) error {
	if dataset == nil {
		return app.Error("mongo", "dataset is missing")
	}
	if dataset.UserID == "" {
		return app.Error("mongo", "dataset user id is missing")
	}
	if dataset.GroupID == "" {
		return app.Error("mongo", "dataset group id is missing")
	}
	if dataset.UploadID == "" {
		return app.Error("mongo", "dataset upload id is missing")
	}

	if s.IsClosed() {
		return app.Error("mongo", "session closed")
	}

	startTime := time.Now()

	dataset.CreatedTime = newTimestamp()
	dataset.CreatedUserID = s.agentUserID()

	dataset.ByUser = dataset.CreatedUserID

	// TODO: Consider upsert instead to prevent multiples being created?

	selector := bson.M{
		"_userId":  dataset.UserID,
		"_groupId": dataset.GroupID,
		"uploadId": dataset.UploadID,
		"type":     dataset.Type,
	}
	count, err := s.C().Find(selector).Count()
	if err == nil {
		if count > 0 {
			err = app.Error("mongo", "dataset already exists")
		} else {
			err = s.C().Insert(dataset)
		}
	}

	loggerFields := log.Fields{"dataset": dataset, "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("CreateDataset")

	if err != nil {
		return app.ExtError(err, "mongo", "unable to create dataset")
	}
	return nil
}

func (s *Session) UpdateDataset(dataset *upload.Upload) error {
	if dataset == nil {
		return app.Error("mongo", "dataset is missing")
	}
	if dataset.UserID == "" {
		return app.Error("mongo", "dataset user id is missing")
	}
	if dataset.GroupID == "" {
		return app.Error("mongo", "dataset group id is missing")
	}
	if dataset.UploadID == "" {
		return app.Error("mongo", "dataset upload id is missing")
	}

	if s.IsClosed() {
		return app.Error("mongo", "session closed")
	}

	startTime := time.Now()

	dataset.ModifiedTime = newTimestamp()
	dataset.ModifiedUserID = s.agentUserID()

	selector := bson.M{
		"_userId":  dataset.UserID,
		"_groupId": dataset.GroupID,
		"uploadId": dataset.UploadID,
		"type":     dataset.Type,
	}
	err := s.C().Update(selector, dataset)

	loggerFields := log.Fields{"dataset": dataset, "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("UpdateDataset")

	if err != nil {
		return app.ExtError(err, "mongo", "unable to update dataset")
	}
	return nil
}

func (s *Session) DeleteDataset(dataset *upload.Upload) error {
	if dataset == nil {
		return app.Error("mongo", "dataset is missing")
	}
	if dataset.UserID == "" {
		return app.Error("mongo", "dataset user id is missing")
	}
	if dataset.GroupID == "" {
		return app.Error("mongo", "dataset group id is missing")
	}
	if dataset.UploadID == "" {
		return app.Error("mongo", "dataset upload id is missing")
	}

	if s.IsClosed() {
		return app.Error("mongo", "session closed")
	}

	startTime := time.Now()

	deletedTimestamp := newTimestamp()
	deletedUserID := s.agentUserID()

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

	loggerFields := log.Fields{"datasetID": dataset.UploadID, "remove-info": removeInfo, "update-info": updateInfo, "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("DeleteDataset")

	if err != nil {
		return app.ExtError(err, "mongo", "unable to delete dataset")
	}

	dataset.SetDeletedTime(deletedTimestamp)
	dataset.SetDeletedUserID(deletedUserID)
	return nil
}

func (s *Session) CreateDatasetData(dataset *upload.Upload, datasetData []data.Datum) error {
	if dataset == nil {
		return app.Error("mongo", "dataset is missing")
	}
	if datasetData == nil {
		return app.Error("mongo", "dataset data is missing")
	}
	if dataset.UserID == "" {
		return app.Error("mongo", "dataset user id is missing")
	}
	if dataset.GroupID == "" {
		return app.Error("mongo", "dataset group id is missing")
	}
	if dataset.UploadID == "" {
		return app.Error("mongo", "dataset upload id is missing")
	}

	if s.IsClosed() {
		return app.Error("mongo", "session closed")
	}

	startTime := time.Now()

	createdTimestamp := newTimestamp()
	createdUserID := s.agentUserID()

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

	loggerFields := log.Fields{"dataset": dataset, "dataset-data-length": len(datasetData), "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("CreateDatasetData")

	if err != nil {
		return app.ExtError(err, "mongo", "unable to create dataset data")
	}
	return nil
}

func (s *Session) ActivateDatasetData(dataset *upload.Upload) error {
	if dataset == nil {
		return app.Error("mongo", "dataset is missing")
	}
	if dataset.UserID == "" {
		return app.Error("mongo", "dataset user id is missing")
	}
	if dataset.GroupID == "" {
		return app.Error("mongo", "dataset group id is missing")
	}
	if dataset.UploadID == "" {
		return app.Error("mongo", "dataset upload id is missing")
	}

	if s.IsClosed() {
		return app.Error("mongo", "session closed")
	}

	startTime := time.Now()

	modifiedTimestamp := newTimestamp()
	modifiedUserID := s.agentUserID()

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

	loggerFields := log.Fields{"dataset": dataset, "update-info": updateInfo, "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("ActivateDatasetData")

	if err != nil {
		return app.ExtError(err, "mongo", "unable to activate dataset data")
	}

	dataset.SetActive(true)
	dataset.SetModifiedTime(modifiedTimestamp)
	dataset.SetModifiedUserID(modifiedUserID)
	return nil
}

func (s *Session) DeleteOtherDatasetData(dataset *upload.Upload) error {
	if dataset == nil {
		return app.Error("mongo", "dataset is missing")
	}
	if dataset.UserID == "" {
		return app.Error("mongo", "dataset user id is missing")
	}
	if dataset.GroupID == "" {
		return app.Error("mongo", "dataset group id is missing")
	}
	if dataset.UploadID == "" {
		return app.Error("mongo", "dataset upload id is missing")
	}
	if dataset.DeviceID == nil || *dataset.DeviceID == "" {
		return app.Error("mongo", "dataset device id is missing")
	}

	if s.IsClosed() {
		return app.Error("mongo", "session closed")
	}

	startTime := time.Now()

	deletedTimestamp := newTimestamp()
	deletedUserID := s.agentUserID()

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

	loggerFields := log.Fields{"dataset": dataset, "remove-info": removeInfo, "update-info": updateInfo, "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("DeleteOtherDatasetData")

	if err != nil {
		return app.ExtError(err, "mongo", "unable to remove other dataset data")
	}
	return nil
}

func (s *Session) agentUserID() string {
	if s.agent == nil {
		return ""
	}
	return s.agent.UserID()
}

func newTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}
