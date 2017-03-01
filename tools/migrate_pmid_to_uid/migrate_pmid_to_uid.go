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

	"github.com/urfave/cli"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/store/mongo"
	"github.com/tidepool-org/platform/version"
)

type Config struct {
	Log    *log.Config
	Mongo  *mongo.Config
	DryRun bool
}

const (
	HelpFlag      = "help"
	VersionFlag   = "version"
	VerboseFlag   = "verbose"
	DryRunFlag    = "dry-run"
	AddressesFlag = "addresses"
	SSLFlag       = "ssl"
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
	application.Usage = "Migrate all metadata to include user id derived from _id"
	application.Version = versionReporter.Long()
	application.Authors = []cli.Author{{"Darin Krauss", "darin@tidepool.org"}}
	application.Copyright = "Copyright \u00A9 2017, Tidepool Project"
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
		return nil, app.ExtError(err, "migrate_pmid_to_uid", "unable to create version reporter")
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

	metaIDToUserIDMap, err := buildMetaIDToUserIDMap(logger, config)
	if err != nil {
		logger.WithError(err).Error("Unable to build meta id to user id map")
		os.Exit(1)
	}

	err = migrateMetaIDToUserIDForMetadata(logger, config, metaIDToUserIDMap)
	if err != nil {
		logger.WithError(err).Error("Unable to migrate meta id to user id for metadata")
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

	return config, nil
}

func initializeLogger(versionReporter version.Reporter, config *Config) (log.Logger, error) {
	logger, err := log.NewStandard(versionReporter, config.Log)
	if err != nil {
		return nil, app.ExtError(err, "migrate_pmid_to_uid", "unable to create logger")
	}

	return logger, nil
}

func buildMetaIDToUserIDMap(logger log.Logger, config *Config) (map[string]string, error) {
	logger.Debug("Building meta id to user id map")

	userIDMap := map[string]bool{}
	metaIDToUserIDMap := map[string]string{}

	logger.Debug("Creating users store")

	mongoConfig := config.Mongo.Clone()
	mongoConfig.Database = "user"
	mongoConfig.Collection = "users"
	usersStore, err := mongo.New(logger, mongoConfig)
	if err != nil {
		return nil, app.ExtError(err, "migrate_pmid_to_uid", "unable to create users store")
	}
	defer usersStore.Close()

	logger.Debug("Creating users session")

	usersSession, err := usersStore.NewSession(logger)
	if err != nil {
		return nil, app.ExtError(err, "migrate_pmid_to_uid", "unable to create users session")
	}
	defer usersSession.Close()

	logger.Debug("Iterating users")

	iter := usersSession.C().Find(bson.M{}).Select(bson.M{"_id": 0, "userid": 1, "private.meta.id": 1}).Iter()

	var result struct {
		UserID  string `bson:"userid"`
		Private struct {
			Meta struct {
				ID string `bson:"id"`
			} `bson:"meta"`
		} `bson:"private"`
	}
	for iter.Next(&result) {
		userLogger := logger

		userID := result.UserID
		if userID == "" {
			userLogger.Warn("Missing user id in result from users query")
			continue
		}

		userLogger = userLogger.WithField("userId", userID)

		if _, ok := userIDMap[userID]; ok {
			userLogger.Error("Found multiple users with same user id")
			continue
		}
		userIDMap[userID] = true

		metaID := result.Private.Meta.ID
		if metaID == "" {
			userLogger.Warn("Missing private meta id in result from users query for user id")
			continue
		}

		userLogger = userLogger.WithField("metaId", metaID)

		if _, ok := metaIDToUserIDMap[metaID]; ok {
			userLogger.Error("Found multiple users with same meta id")
			continue
		}
		metaIDToUserIDMap[metaID] = userID
	}
	if err = iter.Close(); err != nil {
		return nil, app.ExtError(err, "migrate_pmid_to_uid", "unable to iterate users")
	}

	logger.Debug(fmt.Sprintf("Found %d users with meta", len(metaIDToUserIDMap)))

	return metaIDToUserIDMap, nil
}

func migrateMetaIDToUserIDForMetadata(logger log.Logger, config *Config, metaIDToUserIDMap map[string]string) error {
	logger.Debug("Migrating meta id to user id for metadata")

	var migrateMetaCount int
	var migrateMetadataCount int

	logger.Debug("Creating metadata data store")

	mongoConfig := config.Mongo.Clone()
	mongoConfig.Database = "seagull"
	mongoConfig.Collection = "seagull"
	metadataStore, err := mongo.New(logger, mongoConfig)
	if err != nil {
		return app.ExtError(err, "migrate_pmid_to_uid", "unable to create metadata store")
	}
	defer metadataStore.Close()

	logger.Debug("Creating metadata session")

	metadataSession, err := metadataStore.NewSession(logger)
	if err != nil {
		return app.ExtError(err, "migrate_pmid_to_uid", "unable to create metadata session")
	}
	defer metadataSession.Close()

	if !config.DryRun {
		logger.Info("Creating unique index on user id")

		index := mgo.Index{
			Key:        []string{"userId"},
			Unique:     true,
			Background: true,
		}
		err = metadataSession.C().EnsureIndex(index)
		if err != nil {
			return app.ExtError(err, "migrate_pmid_to_uid", "unable to create metadata index on user id")
		}
	}

	logger.Debug("Walking meta id to user id map")

	var count int
	for metaID, userID := range metaIDToUserIDMap {
		metadataLogger := logger.WithFields(log.Fields{"metaID": metaID, "userId": userID})

		metadataLogger.Debug("Finding metadata for meta id with incorrect existing user id")

		query := bson.M{
			"$and": []bson.M{
				{"_id": metaID},
				{"userId": bson.M{"$exists": true}},
				{"userId": bson.M{"$ne": userID}},
			},
		}
		count, err = metadataSession.C().Find(query).Count()
		if err != nil {
			metadataLogger.WithError(err).Error("Unable to query for incorrect metadata")
			continue
		}

		if count != 0 {
			metadataLogger.WithField("count", count).Error("Found metadata for meta id with incorrect existing user id")
			continue
		}

		metadataLogger.Debug("Updating metadata for meta id with user id")

		selector := bson.M{
			"_id":    metaID,
			"userId": bson.M{"$exists": false},
		}

		if config.DryRun {
			count, err = metadataSession.C().Find(selector).Count()
		} else {
			update := bson.M{
				"$set": bson.M{"userId": userID},
			}

			var changeInfo *mgo.ChangeInfo
			changeInfo, err = metadataSession.C().UpdateAll(selector, update)
			if changeInfo != nil {
				count = changeInfo.Updated
			}
		}

		if err != nil {
			metadataLogger.WithError(err).Error("Unable to update metadata for meta id with user id")
			continue
		}

		if count > 0 {
			metadataLogger.Info(fmt.Sprintf("Migrated %d metadata", count))
			migrateMetaCount++
			migrateMetadataCount += count
		}
	}

	if !config.DryRun {
		if count, err = metadataSession.C().Find(bson.M{"userId": bson.M{"$exists": false}}).Count(); err != nil {
			logger.WithError(err).Error("Unable to query for metadata without user id")
		} else if count != 0 {
			logger.WithField("count", count).Error("Found metadata without user id")
		}
	}

	logger.Info(fmt.Sprintf("Migrated %d metadata for %d meta", migrateMetadataCount, migrateMetaCount))
	return nil
}
