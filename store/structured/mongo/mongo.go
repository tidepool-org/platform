package mongo

import (
	"crypto/tls"
	"net"
	"sync"
	"time"

	mgo "github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"

	"github.com/tidepool-org/platform/errors"
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

//NewStore constructs a Store from a Config, using the given logger
func NewStore(config *Config, logger log.Logger) (*Store, error) {
	if config == nil {
		return nil, errors.New("config is missing")
	} else if err := config.Validate(); err != nil {
		return nil, errors.Wrap(err, "config is invalid")
	}
	if logger == nil {
		return nil, errors.New("logger is missing")
	}

	loggerFields := map[string]interface{}{
		"database":         config.Database,
		"collectionPrefix": config.CollectionPrefix,
	}
	logger = logger.WithFields(loggerFields)

	_, err := mgo.ParseURL(config.AsConnectionString())
	if err != nil {
		return nil, errors.Wrap(err, "URL is unparseable")
	}

	store := &Store{
		Config:  config,
		session: nil,
		logger:  logger,
	}

	store.Start()
	return store, nil
}
func (s *Store) getDialInfo() (*mgo.DialInfo, error) {
	dialInfo, err := mgo.ParseURL(s.Config.AsConnectionString())
	if err != nil {
		return nil, err
	}

	if s.Config.TLS {
		dialInfo.DialServer = func(serverAddr *mgo.ServerAddr) (net.Conn, error) {
			return tls.Dial("tcp", serverAddr.String(), &tls.Config{InsecureSkipVerify: true}) // TODO: Secure this connection
		}
	}
	dialInfo.Timeout = s.Config.Timeout
	return dialInfo, nil
}

func (s *Store) Start() {
	if s.Session() == nil && s.closingChannel == nil {
		s.initializeGroup.Add(1)
		go s.connectionRoutine()
	} else if s.Session() != nil {
		close(s.closingChannel)
		s.closingChannel = nil
	}
}

func (s *Store) connectionRoutine() {
	err := s.initializeSession()
	var attempts int64
	if err != nil {
		s.logger.Errorf("Unable to open inital store session : %v", err)
		s.closingChannel = make(chan bool, 1)
		for {
			timer := time.After(s.Config.WaitConnectionInterval)
			select {
			case <-s.closingChannel:
				close(s.closingChannel)
				s.closingChannel = nil
				s.initializeGroup.Done()
				return
			case <-timer:
				err := s.initializeSession()
				if err == nil {
					s.logger.Debug("Store session opened succesfully")
					s.closingChannel <- true
				} else {
					if s.Config.MaxConnectionAttempts > 0 && s.Config.MaxConnectionAttempts > attempts {
						s.logger.Errorf("Unable to open store session, maximum connection attempts reached : %v", err)
						s.closingChannel <- true
						panic(err)
					} else if s.Config.MaxConnectionAttempts > 0 {
						s.logger.Errorf("Unable to open store session : %v", err)
						attempts++
					}
				}
			}
		}
	} else {
		s.createIndexesFromConfig()
		if s.closingChannel != nil {
			close(s.closingChannel)
			s.closingChannel = nil
		}
		s.initializeGroup.Done()
		return
	}
}

func (s *Store) createIndexesFromConfig() {
	if s.Config.Indexes != nil {
		for collection, idxs := range s.Config.Indexes {
			session := s.NewSession(collection)
			defer session.Close()
			err := session.EnsureAllIndexes(idxs)
			if err != nil {
				s.logger.Errorf("unable to ensure indexes on %s : %v", collection, err)
			}
		}
	}
}

func (s *Store) WaitUntilStarted() {
	s.initializeGroup.Wait()
}

func (s *Store) initializeSession() error {
	dialInfo, err := s.getDialInfo()
	if err != nil {
		return errors.Wrap(err, "URL is unparseable")
	}

	s.logger.WithField("config", s.Config).Debug("Dialing Mongo database")
	session, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		return errors.Wrap(err, "unable to dial database")
	}

	s.logger.Debug("Verifying Mongo build version is supported")

	buildInfo, err := session.BuildInfo()
	if err != nil {
		session.Close()
		return errors.Wrap(err, "unable to determine build info")
	}

	if !buildInfo.VersionAtLeast(3) {
		session.Close()
		return errors.Newf("unsupported mongo build version %q", buildInfo.Version)
	}

	s.logger.Debug("Setting Mongo consistency mode to Strong")
	session.SetMode(mgo.Strong, true)
	s.sessionMux.Lock()
	s.session = session
	s.sessionMux.Unlock()
	return nil
}

//Store represents a live session to a Mongo database
type Store struct {
	Config          *Config
	logger          log.Logger
	closingChannel  chan bool
	initializeGroup sync.WaitGroup
	session         *mgo.Session
	sessionMux      sync.Mutex
}

func (s *Store) Session() *mgo.Session {
	s.sessionMux.Lock()
	defer s.sessionMux.Unlock()
	return s.session
}

//IsClosed returns true if the session is closed
func (s *Store) IsClosed() bool {
	return s.Session() == nil
}

//Close the session to the Mongo database
func (s *Store) Close() error {
	if s.closingChannel != nil {
		s.closingChannel <- true
	}
	s.initializeGroup.Wait()
	if s.Session() != nil {
		s.sessionMux.Lock()
		s.session.Close()
		s.session = nil
		s.sessionMux.Unlock()
	}
	return nil
}

//Status returns the current state of the sessions
func (s *Store) Status() interface{} {
	status := &Status{
		State: "CLOSED",
		Ping:  "FAILED",
	}

	if !s.IsClosed() {
		status.State = "OPEN"
		if buildInfo, err := s.Session().BuildInfo(); err == nil {
			status.BuildInfo = &buildInfo
		}
		status.LiveServers = s.Session().LiveServers()
		status.Mode = s.Session().Mode()
		status.Safe = s.Session().Safe()
		if s.Session().Ping() == nil {
			status.Ping = "OK"
		}
	}

	return status
}

func (s *Store) NewSession(collection string) *Session {
	return &Session{
		sourceSession: s.Session(),
		database:      s.Config.Database,
		collection:    s.Config.CollectionPrefix + collection,
	}
}

type Session struct {
	sourceSession *mgo.Session
	targetSession *mgo.Session
	database      string
	collection    string
}

func (s *Session) IsClosed() bool {
	return s.sourceSession == nil
}

func (s *Session) Close() error {
	if s.targetSession != nil {
		s.targetSession.Close()
		s.targetSession = nil
	}
	s.sourceSession = nil
	return nil
}

func (s *Session) EnsureAllIndexes(indexes []mgo.Index) error {
	for _, index := range indexes {
		if err := s.C().EnsureIndex(index); err != nil {
			return errors.Wrapf(err, "unable to ensure index with key %v", index.Key)
		}
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

	return s.targetSession.DB(s.database).C(s.collection)
}

func (s *Session) ConstructUpdate(set bson.M, unset bson.M, operators ...map[string]bson.M) bson.M {
	update := bson.M{}
	if len(set) > 0 {
		update["$set"] = set
	}
	if len(unset) > 0 {
		update["$unset"] = unset
	}
	for _, operator := range operators {
		for fieldKey, fieldValues := range operator {
			update = mergeUpdateField(update, fieldKey, fieldValues)
		}
	}
	if len(update) > 0 {
		return mergeUpdateField(update, "$inc", bson.M{"revision": 1})
	}
	return nil
}

func mergeUpdateField(update bson.M, fieldKey string, fieldValues bson.M) bson.M {
	var mergedFieldValues bson.M
	if raw, ok := update[fieldKey]; ok {
		mergedFieldValues, _ = raw.(bson.M)
	}
	if mergedFieldValues == nil {
		mergedFieldValues = bson.M{}
	}
	for key, value := range fieldValues {
		mergedFieldValues[key] = value
	}
	if len(mergedFieldValues) > 0 {
		update[fieldKey] = mergedFieldValues
	} else {
		delete(update, fieldKey)
	}
	return update
}

type QueryModifier func(query bson.M) bson.M

func ModifyQuery(query bson.M, queryModifiers ...QueryModifier) bson.M {
	if query == nil {
		return nil
	}
	for _, queryModifier := range queryModifiers {
		query = queryModifier(query)
	}
	return query
}

func NotDeleted(query bson.M) bson.M {
	if query == nil {
		return nil
	}
	query["deletedTime"] = bson.M{
		"$exists": false,
	}
	return query
}
