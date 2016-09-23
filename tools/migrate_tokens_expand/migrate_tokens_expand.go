package main

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
	"fmt"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/urfave/cli"
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/session"
	"github.com/tidepool-org/platform/store/mongo"
	"github.com/tidepool-org/platform/version"
)

type Config struct {
	Log    *log.Config
	Mongo  *mongo.Config
	Secret string
	DryRun bool
}

const (
	HelpFlag      = "help"
	VersionFlag   = "version"
	VerboseFlag   = "verbose"
	DryRunFlag    = "dry-run"
	AddressesFlag = "addresses"
	SSLFlag       = "ssl"
	SecretFlag    = "secret"
)

func main() {
	application, err := initializeApplication()
	if err != nil {
		fmt.Println("ERROR: Unable to initialize application:", err)
		os.Exit(1)
	}

	if err := application.Run(os.Args); err != nil {
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
	application.Usage = "Migrate all tokens in database to expanded form"
	application.Version = versionReporter.Long()
	application.Authors = []cli.Author{{"Darin Krauss", "darin@tidepool.org"}}
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
			Name:  fmt.Sprintf("%s,%s", SSLFlag, "s"),
			Usage: "use SSL to connect to mongo database",
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
		return nil, app.ExtError(err, "migrate_tokens_expand", "unable to create version reporter")
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

	err = migrateTokensToExpandedForm(logger, config)
	if err != nil {
		logger.WithError(err).Error("Unable to migrate tokens to expanded form")
		os.Exit(1)
	}
}

func buildConfigFromContext(context *cli.Context) (*Config, error) {
	config := &Config{
		Log: &log.Config{
			Level: "info",
		},
		Mongo: &mongo.Config{
			Timeout: app.DurationAsPointer(60 * time.Second),
		},
	}

	if context.Bool(VerboseFlag) {
		config.Log.Level = "debug"
	}
	config.Mongo.Addresses = context.String(AddressesFlag)
	if context.Bool(SSLFlag) {
		config.Mongo.SSL = true
	}
	if context.Bool(DryRunFlag) {
		config.DryRun = true
	}
	config.Secret = context.String(SecretFlag)
	if config.Secret == "" {
		return nil, app.Error("migrate_tokens_expand", "secret is missing")
	}

	return config, nil
}

func initializeLogger(versionReporter version.Reporter, config *Config) (log.Logger, error) {
	logger, err := log.NewStandard(versionReporter, config.Log)
	if err != nil {
		return nil, app.ExtError(err, "migrate_tokens_expand", "unable to create logger")
	}

	return logger, nil
}

func migrateTokensToExpandedForm(logger log.Logger, config *Config) error {
	logger.Debug("Migrating tokens to expanded form")

	logger.Debug("Creating tokens store")

	mongoConfig := config.Mongo.Clone()
	mongoConfig.Database = "user"
	mongoConfig.Collection = "tokens"
	tokensStore, err := mongo.New(logger, mongoConfig)
	if err != nil {
		return app.ExtError(err, "migrate_tokens_expand", "unable to create tokens store")
	}
	defer tokensStore.Close()

	logger.Debug("Creating tokens sessions")

	iterateTokensSession, err := tokensStore.NewSession(logger)
	if err != nil {
		return app.ExtError(err, "migrate_tokens_expand", "unable to create iterate tokens session")
	}
	defer iterateTokensSession.Close()

	updateTokensSession, err := tokensStore.NewSession(logger)
	if err != nil {
		return app.ExtError(err, "migrate_tokens_expand", "unable to create update tokens session")
	}
	defer updateTokensSession.Close()

	logger.Debug("Iterating tokens")

	iter := iterateTokensSession.C().Find(bson.M{}).Iter()

	expiredTime := time.Now().Unix()
	expiredTokenCount := 0
	migratedTokenCount := 0
	token := &session.Token{}
	for iter.Next(token) {

		if IsTokenExpanded(token) {
			continue
		}

		tokenLogger := logger.WithField("token", token)

		tokenID := token.ID
		if tokenID == "" {
			tokenLogger.Warn("Missing token id in result from tokens query")
			continue
		}

		if err = ExpandToken(token, config.Secret); err != nil {
			tokenLogger.WithError(err).Error("Unable to expand token")
			continue
		}

		if token.ExpiresAt < expiredTime {
			if !config.DryRun {
				if err = updateTokensSession.C().RemoveId(tokenID); err != nil {
					tokenLogger.WithError(err).Error("Unable to remove token")
					continue
				}
			}

			expiredTokenCount++

			tokenLogger.Debug(fmt.Sprintf("Expired token (expired %d seconds ago)", expiredTime-token.ExpiresAt))
		} else {
			if !config.DryRun {
				if err = updateTokensSession.C().UpdateId(tokenID, token); err != nil {
					tokenLogger.WithError(err).Error("Unable to update token")
					continue
				}
			}

			migratedTokenCount++

			tokenLogger.Debug(fmt.Sprintf("Migrated token (expires %d second from now)", token.ExpiresAt-expiredTime))
		}
	}
	if err = iter.Close(); err != nil {
		return app.ExtError(err, "migrate_tokens_expand", "unable to iterate tokens")
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
		if count, err = iterateTokensSession.C().Find(selector).Count(); err != nil {
			logger.WithError(err).Error("Unable to query for unexpanded tokens")
		} else if count != 0 {
			logger.WithField("count", count).Error("Found unexpanded tokens")
		}
	}

	logger.Info(fmt.Sprintf("Deleted %d expired tokens and migrated %d tokens to expanded form", expiredTokenCount, migratedTokenCount))

	return nil
}

func IsTokenExpanded(token *session.Token) bool {
	return token.Duration != 0
}

func ExpandToken(token *session.Token, secret string) error {
	parsedClaims := struct {
		jwt.StandardClaims
		IsServer string  `json:"svr"`
		UserID   string  `json:"usr"`
		Duration float64 `json:"dur"`
	}{}

	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	}
	_, err := jwt.ParseWithClaims(token.ID, &parsedClaims, keyFunc)
	if err != nil {
		validationError, ok := err.(*jwt.ValidationError)
		if !ok {
			return app.ExtError(err, "migrate_tokens_expand", "unexpected error")
		}
		if validationError.Errors != jwt.ValidationErrorExpired {
			return app.ExtError(validationError, "migrate_tokens_expand", "unexpected validation error")
		}
	}

	token.IsServer = parsedClaims.IsServer == "yes"
	if token.IsServer {
		token.ServerID = parsedClaims.UserID
	} else {
		token.UserID = parsedClaims.UserID
	}
	token.Duration = int64(parsedClaims.Duration)
	token.ExpiresAt = parsedClaims.ExpiresAt

	token.CreatedAt = token.Time

	return nil
}
