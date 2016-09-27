package mongo

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

import (
	"crypto/tls"
	"net"
	"strconv"
	"time"

	mgo "gopkg.in/mgo.v2"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/store"
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

	config = config.Clone()
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

func (s *Store) NewSession(logger log.Logger) (*Session, error) {
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
	agent         store.Agent
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

func (s *Session) Logger() log.Logger {
	return s.logger
}

func (s *Session) SetAgent(agent store.Agent) {
	s.agent = agent
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

func (s *Session) AgentUserID() string {
	if s.agent == nil {
		return ""
	}
	return s.agent.UserID()
}

func (s *Session) Timestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}
