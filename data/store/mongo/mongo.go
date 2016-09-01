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

func (s *Session) GetDatasetsForUser(userID string) ([]*upload.Upload, error) {
	if userID == "" {
		return nil, app.Error("mongo", "user id is missing")
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
	err := s.C().Find(selector).All(&datasets)

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

func (s *Session) DeleteDataset(datasetID string) error {
	if datasetID == "" {
		return app.Error("mongo", "dataset id is missing")
	}

	if s.IsClosed() {
		return app.Error("mongo", "session closed")
	}

	startTime := time.Now()

	selector := bson.M{
		"uploadId": datasetID,
	}
	changeInfo, err := s.C().RemoveAll(selector)

	loggerFields := log.Fields{"datasetID": datasetID, "change-info": changeInfo, "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("DeleteDataset")

	if err != nil {
		return app.ExtError(err, "mongo", "unable to delete dataset")
	}
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

	insertData := make([]interface{}, len(datasetData))
	for index, datum := range datasetData {
		datum.SetUserID(dataset.UserID)
		datum.SetGroupID(dataset.GroupID)
		datum.SetDatasetID(dataset.UploadID)
		insertData[index] = datum
	}

	startTime := time.Now()

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

	selector := bson.M{
		"_userId":  dataset.UserID,
		"_groupId": dataset.GroupID,
		"uploadId": dataset.UploadID,
	}
	update := bson.M{
		"$set": bson.M{
			"_active": true,
		},
	}
	changeInfo, err := s.C().UpdateAll(selector, update)

	loggerFields := log.Fields{"dataset": dataset, "change-info": changeInfo, "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("ActivateDatasetData")

	if err != nil {
		return app.ExtError(err, "mongo", "unable to activate dataset data")
	}

	dataset.SetActive(true)
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

	selector := bson.M{
		"_userId":  dataset.UserID,
		"_groupId": dataset.GroupID,
		"deviceId": *dataset.DeviceID,
		"uploadId": bson.M{"$ne": dataset.UploadID},
		"type":     bson.M{"$ne": "upload"},
	}
	changeInfo, err := s.C().RemoveAll(selector)

	loggerFields := log.Fields{"dataset": dataset, "change-info": changeInfo, "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("DeleteOtherDatasetData")

	if err != nil {
		return app.ExtError(err, "mongo", "unable to remove other dataset data")
	}
	return nil
}
