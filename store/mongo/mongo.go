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
	"strings"
	"time"

	mgo "gopkg.in/mgo.v2"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/store"
)

// TODO: Consider SetStats, GetStats
// TODO: Consider SetDebug and SetLogger
// TODO: Consider findAndModify via Query.Apply

type Config struct {
	Addresses  string         `yaml:"addresses"` // TODO: This should be an array, but configor doesn't support that. Bleech! Fix?
	Database   string         `yaml:"database"`
	Collection string         `yaml:"collection"`
	Username   *string        `yaml:"username"`
	Password   *string        `yaml:"password"`
	Timeout    *time.Duration `yaml:"timeout"`
	SSL        bool           `yaml:"ssl"`
}

type Status struct {
	State       string
	BuildInfo   *mgo.BuildInfo
	LiveServers []string
	Mode        mgo.Mode
	Safe        *mgo.Safe
	Ping        string
}

func (c *Config) Validate() error {
	addresses := strings.Split(c.Addresses, ",")
	if len(addresses) < 1 {
		return app.Error("mongo", "addresses is missing")
	}
	if c.Database == "" {
		return app.Error("mongo", "database is missing")
	}
	if c.Collection == "" {
		return app.Error("mongo", "collection is missing")
	}
	return nil
}

func NewStore(config *Config, logger log.Logger) (*Store, error) {
	if err := config.Validate(); err != nil {
		return nil, app.ExtError(err, "mongo", "config is not valid")
	}

	loggerFields := map[string]interface{}{
		"database":   config.Database,
		"collection": config.Collection,
	}
	logger = logger.WithFields(loggerFields)

	dialInfo := mgo.DialInfo{}
	dialInfo.Addrs = strings.Split(config.Addresses, ",")
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
		return nil, app.ExtError(err, "mongo", "unsupported mongo build version")
	}

	logger.Debug("Setting Mongo consistency mode to Strong")

	session.SetMode(mgo.Strong, true)

	// TODO: Do we need to set Safe so we get write > 1?

	return &Store{
		Config:    config,
		Session:   session,
		BuildInfo: &buildInfo,
	}, nil
}

type Store struct {
	Config    *Config
	Session   *mgo.Session
	BuildInfo *mgo.BuildInfo
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
	state := "OPEN"
	ping := "OK"
	if s.IsClosed() {
		state = "CLOSED"
	} else if s.Session.Ping() != nil {
		ping = "FAILURE"
	}

	return &Status{
		State:       state,
		BuildInfo:   s.BuildInfo,
		LiveServers: s.Session.LiveServers(),
		Mode:        s.Session.Mode(),
		Safe:        s.Session.Safe(),
		Ping:        ping,
	}
}

func (s *Store) NewSession(logger log.Logger) (store.Session, error) {
	if s.IsClosed() {
		return nil, app.Error("mongo", "store closed")
	}

	loggerFields := map[string]interface{}{
		"database":   s.Config.Database,
		"collection": s.Config.Collection,
	}

	return &Session{
		config:  s.Config,
		logger:  logger.WithFields(loggerFields),
		session: s.Session.Copy(),
	}, nil
}

type Session struct {
	config  *Config
	session *mgo.Session
	logger  log.Logger
}

func (s *Session) IsClosed() bool {
	return s.session == nil
}

func (s *Session) Close() {
	if s.session != nil {
		s.session.Close()
		s.session = nil
	}
}

func (s *Session) Find(query store.Query, result interface{}) error {
	if s.IsClosed() {
		return app.Error("mongo", "session closed")
	}

	s.logger.WithField("query", query).Debug("Find")

	if err := s.C().Find(query).One(result); err != nil {
		return app.ExtError(err, "mongo", "unable to find")
	}
	return nil
}

func (s *Session) FindAll(query store.Query, sort []string, filter store.Filter) store.Iterator {
	if s.IsClosed() {
		return &Iterator{}
	}

	s.logger.WithField("query", query).WithField("sort", sort).WithField("filter", filter).Debug("FindAll")

	return &Iterator{s.logger, s.C().Find(query).Sort(sort...).Select(filter).Iter()}
}

func (s *Session) Insert(document interface{}) error {
	if s.IsClosed() {
		return app.Error("mongo", "session closed")
	}

	s.logger.WithField("document", document).Debug("Insert")

	if err := s.C().Insert(document); err != nil {
		return app.ExtError(err, "mongo", "unable to insert")
	}
	return nil
}

func (s *Session) InsertAll(documents ...interface{}) error {
	if s.IsClosed() {
		return app.Error("mongo", "session closed")
	}

	s.logger.WithField("documents", len(documents)).Debug("InsertAll")

	bulk := s.C().Bulk()
	bulk.Unordered()
	bulk.Insert(documents...)

	if _, err := bulk.Run(); err != nil {
		return err
	}
	// s.logger.Warn("BulkResult=", result)		// TODO: Is there anything we can do with this?
	return nil
}

func (s *Session) Update(selector interface{}, update interface{}) error {
	if s.IsClosed() {
		return app.Error("mongo", "session closed")
	}

	s.logger.WithField("selector", selector).WithField("update", update).Debug("Update")

	if err := s.C().Update(selector, update); err != nil {
		return app.ExtError(err, "mongo", "unable to update")
	}

	return nil
}

func (s *Session) UpdateAll(selector interface{}, update interface{}) error {
	if s.IsClosed() {
		return app.Error("mongo", "session closed")
	}

	s.logger.WithField("selector", selector).WithField("update", update).Debug("UpdateAll")

	if _, err := s.C().UpdateAll(selector, update); err != nil {
		return app.ExtError(err, "mongo", "unable to update all")
	}
	return nil
}

func (s *Session) RemoveAll(selector interface{}) error {
	if s.IsClosed() {
		return app.Error("mongo", "session closed")
	}

	s.logger.WithField("selector", selector).Debug("RemoveAll")

	if _, err := s.C().RemoveAll(selector); err != nil {
		return app.ExtError(err, "mongo", "unable to remove all")
	}
	return nil
}

func (s *Session) C() *mgo.Collection {
	return s.session.DB(s.config.Database).C(s.config.Collection)
}

type Iterator struct {
	logger   log.Logger
	iterator *mgo.Iter
}

func (i *Iterator) IsClosed() bool {
	return i.iterator == nil
}

func (i *Iterator) Close() (err error) {
	if i.iterator != nil {
		err = i.iterator.Close()
		i.iterator = nil
	}

	if err != nil {
		err = app.ExtError(err, "mongo", "unable to close iterator")
	}
	return err
}

func (i *Iterator) Err() error {
	if i.IsClosed() {
		return app.Error("mongo", "iterator closed")
	}

	err := i.iterator.Err()
	if err != nil {
		err = app.ExtError(err, "mongo", "error while iterating")
	}
	return err
}

func (i *Iterator) Next(result interface{}) bool {
	if i.IsClosed() {
		return false
	}

	return i.iterator.Next(result)
}

func (i *Iterator) All(result interface{}) error {
	if i.IsClosed() {
		return app.Error("mongo", "iterator closed")
	}

	err := i.iterator.All(result)
	if err != nil {
		err = app.ExtError(err, "mongo", "unable to get all from iterator")
	}
	return err
}
