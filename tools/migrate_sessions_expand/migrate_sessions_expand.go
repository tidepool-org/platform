package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/urfave/cli"
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/json"
	"github.com/tidepool-org/platform/session"
	"github.com/tidepool-org/platform/store/mongo"
	"github.com/tidepool-org/platform/version"
)

type Config struct {
	LogLevel log.Level
	Mongo    *mongo.Config
	Secret   string
	DryRun   bool
}

const (
	HelpFlag      = "help"
	VersionFlag   = "version"
	VerboseFlag   = "verbose"
	DryRunFlag    = "dry-run"
	AddressesFlag = "addresses"
	TLSFlag       = "tls"
	SecretFlag    = "secret"
)

func main() {
	application, err := initializeApplication()
	if err != nil {
		fmt.Println("ERROR: Unable to initialize application:", err)
		os.Exit(1)
	}

	if err = application.Run(os.Args); err != nil {
		fmt.Println("ERROR: Unable to run application:", err)
		os.Exit(1)
	}
}

func initializeApplication() (*cli.App, error) {
	versionReporter, err := initializeVersionReporter()
	if err != nil {
		return nil, err
	}

	application := cli.NewApp()
	application.Usage = "Migrate all sessions in database to expanded form"
	application.Version = versionReporter.Long()
	application.Authors = []cli.Author{{Name: "Darin Krauss", Email: "darin@tidepool.org"}}
	application.Copyright = "Copyright \u00A9 2016, Tidepool Project"
	application.HideHelp = true
	application.HideVersion = true
	application.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  fmt.Sprintf("%s,%s,%s", HelpFlag, "h", "?"),
			Usage: "print this page and exit",
		},
		cli.BoolFlag{
			Name:  VersionFlag,
			Usage: "print version and exit",
		},
		cli.BoolFlag{
			Name:  fmt.Sprintf("%s,%s", VerboseFlag, "v"),
			Usage: "increased verbosity",
		},
		cli.BoolFlag{
			Name:  fmt.Sprintf("%s,%s", DryRunFlag, "n"),
			Usage: "dry run only, do not update database",
		},
		cli.StringFlag{
			Name:  fmt.Sprintf("%s,%s", AddressesFlag, "a"),
			Usage: "comma-delimited list of address(es) to mongo database (host:port)",
		},
		cli.BoolFlag{
			Name:  fmt.Sprintf("%s,%s", TLSFlag, "t"),
			Usage: "use TLS to connect to mongo database",
		},
		cli.StringFlag{
			Name:   SecretFlag,
			Usage:  "authorization secret",
			EnvVar: "API_SECRET",
		},
	}
	application.Action = func(context *cli.Context) error {
		executeApplication(versionReporter, context)
		return nil
	}

	return application, nil
}

func initializeVersionReporter() (version.Reporter, error) {
	versionReporter, err := version.NewDefaultReporter()
	if err != nil {
		return nil, errors.Wrap(err, "migrate_sessions_expand", "unable to create version reporter")
	}

	return versionReporter, nil
}

func executeApplication(versionReporter version.Reporter, context *cli.Context) {
	if context.Bool(HelpFlag) {
		cli.ShowAppHelp(context)
		return
	}

	if context.Bool(VersionFlag) {
		fmt.Println(versionReporter.Long())
		return
	}

	config, err := buildConfigFromContext(context)
	if err != nil {
		fmt.Println("ERROR: Unable to build config from context:", err)
		os.Exit(1)
	}

	logger, err := initializeLogger(versionReporter, config)
	if err != nil {
		fmt.Println("ERROR: Unable to initialize logger:", err)
		os.Exit(1)
	}

	err = migrateSessionsToExpandedForm(logger, config)
	if err != nil {
		logger.WithError(err).Error("Unable to migrate sessions to expanded form")
		os.Exit(1)
	}
}

func buildConfigFromContext(context *cli.Context) (*Config, error) {
	config := &Config{
		LogLevel: log.InfoLevel,
		Mongo:    mongo.NewConfig(),
	}

	if context.Bool(VerboseFlag) {
		config.LogLevel = log.DebugLevel
	}
	config.Mongo.Addresses = mongo.SplitAddresses(context.String(AddressesFlag))
	config.Mongo.TLS = context.Bool(TLSFlag)
	config.DryRun = context.Bool(DryRunFlag)
	config.Secret = context.String(SecretFlag)
	if config.Secret == "" {
		return nil, errors.New("migrate_sessions_expand", "secret is missing")
	}

	return config, nil
}

func initializeLogger(versionReporter version.Reporter, config *Config) (log.Logger, error) {
	logger, err := json.NewLogger(os.Stdout, log.DefaultLevels(), config.LogLevel)
	if err != nil {
		return nil, errors.Wrap(err, "migrate_sessions_expand", "unable to create logger")
	}

	logger = logger.WithFields(log.Fields{
		"process": filepath.Base(os.Args[0]),
		"pid":     os.Getpid(),
		"version": versionReporter.Short(),
	})

	return logger, nil
}

func migrateSessionsToExpandedForm(logger log.Logger, config *Config) error {
	logger.Debug("Migrating sessions to expanded form")

	logger.Debug("Creating sessions store")

	mongoConfig := config.Mongo.Clone()
	mongoConfig.Database = "user"
	mongoConfig.Collection = "tokens"
	sessionsStore, err := mongo.New(logger, mongoConfig)
	if err != nil {
		return errors.Wrap(err, "migrate_sessions_expand", "unable to create sessions store")
	}
	defer sessionsStore.Close()

	logger.Debug("Creating sessions sessions")

	iterateSessionsSession := sessionsStore.NewSession(logger)
	defer iterateSessionsSession.Close()

	updateSessionsSession := sessionsStore.NewSession(logger)
	defer updateSessionsSession.Close()

	logger.Debug("Iterating sessions")

	iter := iterateSessionsSession.C().Find(bson.M{}).Iter()

	expiredTime := time.Now().Unix()
	expiredSessionCount := 0
	migratedSessionCount := 0
	session := &session.Session{}
	for iter.Next(session) {

		if IsSessionExpanded(session) {
			continue
		}

		sessionLogger := logger.WithField("session", session)

		sessionID := session.ID
		if sessionID == "" {
			sessionLogger.Warn("Missing session id in result from sessions query")
			continue
		}

		if err = ExpandSession(session, config.Secret); err != nil {
			sessionLogger.WithError(err).Error("Unable to expand session")
			continue
		}

		if session.ExpiresAt < expiredTime {
			if !config.DryRun {
				if err = updateSessionsSession.C().RemoveId(sessionID); err != nil {
					sessionLogger.WithError(err).Error("Unable to remove session")
					continue
				}
			}

			expiredSessionCount++

			sessionLogger.Debugf("Expired session (expired %d seconds ago)", expiredTime-session.ExpiresAt)
		} else {
			if !config.DryRun {
				if err = updateSessionsSession.C().UpdateId(sessionID, session); err != nil {
					sessionLogger.WithError(err).Error("Unable to update session")
					continue
				}
			}

			migratedSessionCount++

			sessionLogger.Debugf("Migrated session (expires %d second from now)", session.ExpiresAt-expiredTime)
		}
	}
	if err = iter.Close(); err != nil {
		return errors.Wrap(err, "migrate_sessions_expand", "unable to iterate sessions")
	}

	if !config.DryRun {
		selector := bson.M{
			"$or": []bson.M{
				{"_id": bson.M{"$exists": false}},
				{"isServer": bson.M{"$exists": false}},
				{"$and": []bson.M{
					{"isServer": true},
					{"serverId": bson.M{"$exists": false}},
				}},
				{"$and": []bson.M{
					{"isServer": false},
					{"userId": bson.M{"$exists": false}},
				}},
				{"duration": bson.M{"$exists": false}},
				{"expiresAt": bson.M{"$exists": false}},
				{"time": bson.M{"$exists": false}},
				{"createdAt": bson.M{"$exists": false}},
			},
		}
		var count int
		if count, err = iterateSessionsSession.C().Find(selector).Count(); err != nil {
			logger.WithError(err).Error("Unable to query for unexpanded sessions")
		} else if count != 0 {
			logger.WithField("count", count).Error("Found unexpanded sessions")
		}
	}

	logger.Infof("Deleted %d expired sessions and migrated %d sessions to expanded form", expiredSessionCount, migratedSessionCount)

	return nil
}

func IsSessionExpanded(session *session.Session) bool {
	return session.Duration != 0
}

func ExpandSession(session *session.Session, secret string) error {
	parsedClaims := struct {
		jwt.StandardClaims
		IsServer string  `json:"svr"`
		UserID   string  `json:"usr"`
		Duration float64 `json:"dur"`
	}{}

	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	}
	_, err := jwt.ParseWithClaims(session.ID, &parsedClaims, keyFunc)
	if err != nil {
		validationError, ok := err.(*jwt.ValidationError)
		if !ok {
			return errors.Wrap(err, "migrate_sessions_expand", "unexpected error")
		}
		if validationError.Errors != jwt.ValidationErrorExpired {
			return errors.Wrap(validationError, "migrate_sessions_expand", "unexpected validation error")
		}
	}

	session.IsServer = parsedClaims.IsServer == "yes"
	if session.IsServer {
		session.ServerID = parsedClaims.UserID
	} else {
		session.UserID = parsedClaims.UserID
	}
	session.Duration = int64(parsedClaims.Duration)
	session.ExpiresAt = parsedClaims.ExpiresAt

	session.CreatedAt = session.Time

	return nil
}
