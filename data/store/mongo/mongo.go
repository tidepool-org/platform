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
	"crypto/tls"
	"net"
	"strconv"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/types/base/upload"
	"github.com/tidepool-org/platform/log"
)

// TODO: Consider SetStats, GetStats
// TODO: Consider SetDebug and SetLogger
// TODO: Consider findAndModify via Query.Apply

type Status struct {
	State       string
	BuildInfo   *mgo.BuildInfo
	LiveServers []string
	Mode        mgo.Mode
	Safe        *mgo.Safe
	Ping        string
}

func New(logger log.Logger, config *Config) (*Store, error) {
	if logger == nil {
		return nil, app.Error("mongo", "logger is missing")
	}
	if config == nil {
		return nil, app.Error("mongo", "config is missing")
	}

	if err := config.Validate(); err != nil {
		return nil, app.ExtError(err, "mongo", "config is invalid")
	}

	loggerFields := map[string]interface{}{
		"database":   config.Database,
		"collection": config.Collection,
	}
	logger = logger.WithFields(loggerFields)

	dialInfo := mgo.DialInfo{}
	dialInfo.Addrs = app.SplitStringAndRemoveWhitespace(config.Addresses, ",")
	dialInfo.Database = config.Database
	if config.Username != nil {
		dialInfo.Username = *config.Username
	}
	if config.Password != nil {
		dialInfo.Password = *config.Password
	}
	if config.Timeout != nil {
		dialInfo.Timeout = *config.Timeout
	}
	if config.SSL {
		dialInfo.DialServer = func(serverAddr *mgo.ServerAddr) (net.Conn, error) {
			return tls.Dial("tcp", serverAddr.String(), &tls.Config{InsecureSkipVerify: true}) // TODO: Secure this connection
		}
	}

	logger.Debug("Dialing Mongo database")

	session, err := mgo.DialWithInfo(&dialInfo)
	if err != nil {
		return nil, app.ExtError(err, "mongo", "unable to dial database")
	}

	logger.Debug("Verifying Mongo build version is supported")

	buildInfo, err := session.BuildInfo()
	if err != nil {
		session.Close()
		return nil, app.ExtError(err, "mongo", "unable to determine build info")
	}

	if !buildInfo.VersionAtLeast(3) {
		session.Close()
		return nil, app.Errorf("mongo", "unsupported mongo build version %s", strconv.Quote(buildInfo.Version))
	}

	logger.Debug("Setting Mongo consistency mode to Strong")

	session.SetMode(mgo.Strong, true)

	// TODO: Do we need to set Safe so we get write > 1?

	return &Store{
		Config:  config,
		Session: session,
	}, nil
}

type Store struct {
	Config  *Config
	Session *mgo.Session
}

func (s *Store) IsClosed() bool {
	return s.Session == nil
}

func (s *Store) Close() {
	if s.Session != nil {
		s.Session.Close()
		s.Session = nil
	}
}

func (s *Store) GetStatus() interface{} {
	status := &Status{
		State: "CLOSED",
		Ping:  "FAILED",
	}

	if !s.IsClosed() {
		status.State = "OPEN"
		if buildInfo, err := s.Session.BuildInfo(); err == nil {
			status.BuildInfo = &buildInfo
		}
		status.LiveServers = s.Session.LiveServers()
		status.Mode = s.Session.Mode()
		status.Safe = s.Session.Safe()
		if s.Session.Ping() == nil {
			status.Ping = "OK"
		}
	}

	return status
}

func (s *Store) NewSession(logger log.Logger) (store.Session, error) {
	if logger == nil {
		return nil, app.Error("mongo", "logger is missing")
	}

	if s.IsClosed() {
		return nil, app.Error("mongo", "store closed")
	}

	loggerFields := map[string]interface{}{
		"database":   s.Config.Database,
		"collection": s.Config.Collection,
	}

	return &Session{
		logger:        logger.WithFields(loggerFields),
		config:        s.Config,
		sourceSession: s.Session,
	}, nil
}

type Session struct {
	logger        log.Logger
	config        *Config
	sourceSession *mgo.Session
	targetSession *mgo.Session
}

func (s *Session) IsClosed() bool {
	return s.sourceSession == nil
}

func (s *Session) Close() {
	if s.targetSession != nil {
		s.targetSession.Close()
		s.targetSession = nil
	}
	s.sourceSession = nil
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
	selector := bson.M{"type": "upload", "uploadId": datasetID}
	err := s.C().Find(selector).One(&dataset)

	loggerFields := log.Fields{"datasetID": datasetID, "dataset": dataset, "duration": time.Since(startTime) / time.Microsecond}
	s.logger.WithFields(loggerFields).WithError(err).Debug("GetDataset")

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

	selector := bson.M{"_userId": dataset.UserID, "_groupId": dataset.GroupID, "uploadId": dataset.UploadID, "type": dataset.Type}
	count, err := s.C().Find(selector).Count()
	if err == nil {
		if count > 0 {
			err = app.Error("mongo", "dataset already exists")
		} else {
			err = s.C().Insert(dataset)
		}
	}

	loggerFields := log.Fields{"dataset": dataset, "duration": time.Since(startTime) / time.Microsecond}
	s.logger.WithFields(loggerFields).WithError(err).Debug("CreateDataset")

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

	selector := bson.M{"_userId": dataset.UserID, "_groupId": dataset.GroupID, "uploadId": dataset.UploadID, "type": dataset.Type}
	err := s.C().Update(selector, dataset)

	loggerFields := log.Fields{"dataset": dataset, "duration": time.Since(startTime) / time.Microsecond}
	s.logger.WithFields(loggerFields).WithError(err).Debug("UpdateDataset")

	if err != nil {
		return app.ExtError(err, "mongo", "unable to update dataset")
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
	s.logger.WithFields(loggerFields).WithError(err).Debug("CreateDatasetData")

	if err != nil {
		return app.ExtError(err, "mongo", "unable to create dataset data")
	}
	return nil
}

func (s *Session) ActivateAllDatasetData(dataset *upload.Upload) error {
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

	selector := bson.M{"_userId": dataset.UserID, "_groupId": dataset.GroupID, "uploadId": dataset.UploadID}
	update := bson.M{"$set": bson.M{"_active": true}}
	changeInfo, err := s.C().UpdateAll(selector, update)

	loggerFields := log.Fields{"dataset": dataset, "change-info": changeInfo, "duration": time.Since(startTime) / time.Microsecond}
	s.logger.WithFields(loggerFields).WithError(err).Debug("ActivateAllDatasetData")

	if err != nil {
		return app.ExtError(err, "mongo", "unable to activate all dataset data")
	}

	dataset.SetActive(true)
	return nil
}

func (s *Session) DeleteAllOtherDatasetData(dataset *upload.Upload) error {
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

	selector := bson.M{"_userId": dataset.UserID, "_groupId": dataset.GroupID, "deviceId": *dataset.DeviceID, "type": bson.M{"$ne": "upload"}, "uploadId": bson.M{"$ne": dataset.UploadID}}
	changeInfo, err := s.C().RemoveAll(selector)

	loggerFields := log.Fields{"dataset": dataset, "change-info": changeInfo, "duration": time.Since(startTime) / time.Microsecond}
	s.logger.WithFields(loggerFields).WithError(err).Debug("DeleteAllOtherDatasetData")

	if err != nil {
		return app.ExtError(err, "mongo", "unable to remove all other dataset data")
	}
	return nil
}

func (s *Session) C() *mgo.Collection {
	if s.IsClosed() {
		return nil
	}

	if s.targetSession == nil {
		s.targetSession = s.sourceSession.Copy()
	}

	return s.targetSession.DB(s.config.Database).C(s.config.Collection)
}
