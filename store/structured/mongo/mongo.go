package mongo

import (
	"crypto/tls"
	"net"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

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

func NewStore(cfg *Config, lgr log.Logger) (*Store, error) {
	if cfg == nil {
		return nil, errors.New("config is missing")
	} else if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "config is invalid")
	}
	if lgr == nil {
		return nil, errors.New("logger is missing")
	}

	loggerFields := map[string]interface{}{
		"database":         cfg.Database,
		"collectionPrefix": cfg.CollectionPrefix,
	}
	lgr = lgr.WithFields(loggerFields)

	dialInfo := mgo.DialInfo{}
	dialInfo.Addrs = cfg.Addresses
	if cfg.TLS {
		dialInfo.DialServer = func(serverAddr *mgo.ServerAddr) (net.Conn, error) {
			return tls.Dial("tcp", serverAddr.String(), &tls.Config{InsecureSkipVerify: true}) // TODO: Secure this connection
		}
	}
	dialInfo.Database = cfg.Database
	if cfg.Username != nil {
		dialInfo.Username = *cfg.Username
	}
	if cfg.Password != nil {
		dialInfo.Password = *cfg.Password
	}
	dialInfo.Timeout = cfg.Timeout

	lgr.WithField("config", cfg).Debug("Dialing Mongo database")

	session, err := mgo.DialWithInfo(&dialInfo)
	if err != nil {
		return nil, errors.Wrap(err, "unable to dial database")
	}

	lgr.Debug("Verifying Mongo build version is supported")

	buildInfo, err := session.BuildInfo()
	if err != nil {
		session.Close()
		return nil, errors.Wrap(err, "unable to determine build info")
	}

	if !buildInfo.VersionAtLeast(3) {
		session.Close()
		return nil, errors.Newf("unsupported mongo build version %q", buildInfo.Version)
	}

	lgr.Debug("Setting Mongo consistency mode to Strong")

	session.SetMode(mgo.Strong, true)

	// TODO: Do we need to set Safe so we get write > 1?

	return &Store{
		Config:  cfg,
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

func (s *Store) Close() error {
	if s.Session != nil {
		s.Session.Close()
		s.Session = nil
	}
	return nil
}

func (s *Store) Status() interface{} {
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

func (s *Store) NewSession(collection string) *Session {
	return &Session{
		sourceSession: s.Session,
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

func (s *Session) ConstructUpdate(set bson.M, unset bson.M) bson.M {
	update := bson.M{}
	if len(set) != 0 {
		update["$set"] = set
	}
	if len(unset) != 0 {
		update["$unset"] = unset
	}
	if len(update) != 0 {
		update["$inc"] = bson.M{
			"revision": 1,
		}
		return update
	}
	return nil
}
