package main

import (
	"fmt"
	"os"
	"time"

	"github.com/urfave/cli"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/store/mongo"
	"github.com/tidepool-org/platform/version"
)

type Config struct {
	Log    *log.Config
	Mongo  *mongo.Config
	DryRun bool
	Index  bool
}

const (
	HelpFlag      = "help"
	VersionFlag   = "version"
	VerboseFlag   = "verbose"
	DryRunFlag    = "dry-run"
	IndexFlag     = "index"
	AddressesFlag = "addresses"
	SSLFlag       = "ssl"
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
	application.Usage = "Migrate all metadata to include user id derived from _id"
	application.Version = versionReporter.Long()
	application.Authors = []cli.Author{{Name: "Darin Krauss", Email: "darin@tidepool.org"}}
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
		cli.BoolFlag{
			Name:  fmt.Sprintf("%s,%s", IndexFlag, "i"),
			Usage: "add unique index after migration",
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
		return nil, errors.Wrap(err, "migrate_pmid_to_uid", "unable to create version reporter")
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
	if context.Bool(IndexFlag) {
		config.Index = true
	}

	if config.DryRun && config.Index {
		return nil, errors.New("migrate_pmid_to_uid", "cannot specify --index with --dry-run")
	}

	return config, nil
}

func initializeLogger(versionReporter version.Reporter, config *Config) (log.Logger, error) {
	logger, err := log.NewStandard(versionReporter, config.Log)
	if err != nil {
		return nil, errors.Wrap(err, "migrate_pmid_to_uid", "unable to create logger")
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
		return nil, errors.Wrap(err, "migrate_pmid_to_uid", "unable to create users store")
	}
	defer usersStore.Close()

	logger.Debug("Creating users session")

	usersSession := usersStore.NewSession(logger)
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
			userLogger.Error("Multiple users found with same user id")
			continue
		}
		userIDMap[userID] = true

		metaID := result.Private.Meta.ID
		if metaID == "" {
			userLogger.Warn("User found without meta id")
			continue
		}

		userLogger = userLogger.WithField("metaId", metaID)

		if _, ok := metaIDToUserIDMap[metaID]; ok {
			userLogger.Error("Multiple users found with same meta id")
			continue
		}
		metaIDToUserIDMap[metaID] = userID
	}
	if err = iter.Close(); err != nil {
		return nil, errors.Wrap(err, "migrate_pmid_to_uid", "unable to iterate users")
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
		return errors.Wrap(err, "migrate_pmid_to_uid", "unable to create metadata store")
	}
	defer metadataStore.Close()

	logger.Debug("Creating metadata session")

	metadataSession := metadataStore.NewSession(logger)
	defer metadataSession.Close()

	logger.Debug("Walking meta id to user id map")

	var count int
	for metaID, userID := range metaIDToUserIDMap {
		metadataLogger := logger.WithFields(log.Fields{"metaId": metaID, "userId": userID})

		metadataLogger.Debug("Finding metadata for meta id")

		var results []struct {
			ID     string  `bson:"_id"`
			UserID *string `bson:"userId"`
			Value  *string `bson:"value"`
		}
		err = metadataSession.C().Find(bson.M{"_id": metaID}).All(&results)
		if err != nil {
			metadataLogger.WithError(err).Error("Unable to query for metadata")
			continue
		}

		resultsCount := len(results)
		switch resultsCount {
		case 0:
			metadataLogger.Error("Metadata not found for meta id")
			continue
		case 1:
			break
		default:
			metadataLogger.WithField("count", resultsCount).Error("More than one metadata found for meta id")
			continue
		}

		if result := results[0]; result.UserID != nil {
			if existingUserID := *result.UserID; existingUserID != userID {
				metadataLogger.WithField("existingUserId", existingUserID).Error("Metadata found for meta id with incorrect existing user id")
			}
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
		iter := metadataSession.C().Find(bson.M{"userId": bson.M{"$exists": false}}).Iter()
		var result map[string]interface{}
		for iter.Next(&result) {
			logger.WithField("metaId", result["_id"]).Error("Metadata found without user id")
		}
		if err = iter.Close(); err != nil {
			return errors.Wrap(err, "migrate_pmid_to_uid", "unable to iterate metadata without user id")
		}
	}

	logger.Info(fmt.Sprintf("Migrated %d metadata for %d meta", migrateMetadataCount, migrateMetaCount))

	if config.Index {
		logger.Info("Creating unique index on user id")

		index := mgo.Index{
			Key:        []string{"userId"},
			Unique:     true,
			Background: false,
		}
		err = metadataSession.C().EnsureIndex(index)
		if err != nil {
			return errors.Wrap(err, "migrate_pmid_to_uid", "unable to create metadata index on user id")
		}
	}

	return nil
}
