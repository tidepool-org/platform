package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	jsonLog "github.com/tidepool-org/platform/log/json"
	"github.com/tidepool-org/platform/store/mongo"
	"github.com/tidepool-org/platform/version"
)

type Config struct {
	LogLevel log.Level
	Mongo    *mongo.Config
	DryRun   bool
}

const (
	HelpFlag      = "help"
	VersionFlag   = "version"
	VerboseFlag   = "verbose"
	DryRunFlag    = "dry-run"
	AddressesFlag = "addresses"
	TLSFlag       = "tls"
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
	application.Usage = "Migrate all device data to include user id derived from group id"
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
		return nil, errors.Wrap(err, "migrate_gid_to_uid", "unable to create version reporter")
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

	groupIDToUserIDMap, err := buildGroupIDToUserIDMap(logger, config, metaIDToUserIDMap)
	if err != nil {
		logger.WithError(err).Error("Unable to build group id to user id map")
		os.Exit(1)
	}

	err = migrateGroupIDToUserIDForDeviceData(logger, config, groupIDToUserIDMap)
	if err != nil {
		logger.WithError(err).Error("Unable to migrate group id to user id for device data")
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

	return config, nil
}

func initializeLogger(versionReporter version.Reporter, config *Config) (log.Logger, error) {
	logger, err := jsonLog.NewLogger(os.Stdout, log.DefaultLevels(), config.LogLevel)
	if err != nil {
		return nil, errors.Wrap(err, "migrate_gid_to_uid", "unable to create logger")
	}

	logger = logger.WithFields(log.Fields{
		"process": filepath.Base(os.Args[0]),
		"pid":     os.Getpid(),
		"version": versionReporter.Short(),
	})

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
		return nil, errors.Wrap(err, "migrate_gid_to_uid", "unable to create users store")
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
		return nil, errors.Wrap(err, "migrate_gid_to_uid", "unable to iterate users")
	}

	logger.Debugf("Found %d users with meta", len(metaIDToUserIDMap))

	return metaIDToUserIDMap, nil
}

func buildGroupIDToUserIDMap(logger log.Logger, config *Config, metaIDToUserIDMap map[string]string) (map[string]string, error) {
	logger.Debug("Building group id to user id map")

	metaIDMap := map[string]bool{}
	groupIDToUserIDMap := map[string]string{}

	logger.Debug("Creating meta store")

	mongoConfig := config.Mongo.Clone()
	mongoConfig.Database = "seagull"
	mongoConfig.Collection = "seagull"
	metaStore, err := mongo.New(logger, mongoConfig)
	if err != nil {
		return nil, errors.Wrap(err, "migrate_gid_to_uid", "unable to create meta store")
	}
	defer metaStore.Close()

	logger.Debug("Creating meta session")

	metaSession := metaStore.NewSession(logger)
	defer metaSession.Close()

	logger.Debug("Iterating meta")

	iter := metaSession.C().Find(bson.M{}).Iter()

	var result struct {
		MetaID string `bson:"_id"`
		Value  string `bson:"value"`
	}
	for iter.Next(&result) {
		metaLogger := logger

		metaID := result.MetaID
		if metaID == "" {
			metaLogger.Warn("Missing meta id in result from meta query")
			continue
		}

		metaLogger = metaLogger.WithField("metaId", metaID)

		if _, ok := metaIDMap[metaID]; ok {
			metaLogger.Error("Found multiple metas with same meta id")
			continue
		}
		metaIDMap[metaID] = true

		userID, ok := metaIDToUserIDMap[metaID]
		if !ok {
			metaLogger.Error("Missing user id for meta id")
			continue
		}

		metaLogger = metaLogger.WithField("userId", userID)

		if result.Value == "" {
			metaLogger.Warn("Missing value in result from meta query for meta id")
			continue
		}

		var value struct {
			Private struct {
				Uploads struct {
					ID string `json:"id"`
				} `json:"uploads"`
			} `json:"private"`
		}
		if err = json.Unmarshal([]byte(result.Value), &value); err != nil {
			metaLogger.WithError(err).Warn("Unable to unmarshal value from meta query for meta id")
			continue
		}

		groupID := value.Private.Uploads.ID
		if groupID == "" {
			metaLogger.Debug("Missing group id in value in result from meta query for meta id")
			continue
		}

		metaLogger = metaLogger.WithField("groupId", groupID)

		if _, ok = groupIDToUserIDMap[groupID]; ok {
			metaLogger.Error("Found multiple metas with same group id")
			continue
		}
		groupIDToUserIDMap[groupID] = userID
	}
	if err = iter.Close(); err != nil {
		return nil, errors.Wrap(err, "migrate_gid_to_uid", "unable to iterate meta")
	}

	logger.Debugf("Found %d groups with user", len(groupIDToUserIDMap))

	return groupIDToUserIDMap, nil
}

func migrateGroupIDToUserIDForDeviceData(logger log.Logger, config *Config, groupIDToUserIDMap map[string]string) error {
	logger.Debug("Migrating group id to user id for device data")

	var migrateGroupCount int
	var migrateDeviceDataCount int

	logger.Debug("Creating device data store")

	mongoConfig := config.Mongo.Clone()
	mongoConfig.Database = "data"
	mongoConfig.Collection = "deviceData"
	deviceDataStore, err := mongo.New(logger, mongoConfig)
	if err != nil {
		return errors.Wrap(err, "migrate_gid_to_uid", "unable to create device data store")
	}
	defer deviceDataStore.Close()

	logger.Debug("Creating device data session")

	deviceDataSession := deviceDataStore.NewSession(logger)
	defer deviceDataSession.Close()

	logger.Debug("Walking group id to user id map")

	var count int
	for groupID, userID := range groupIDToUserIDMap {
		dataLogger := logger.WithFields(log.Fields{"groupId": groupID, "userId": userID})

		dataLogger.Debug("Finding device data for group id with incorrect existing user id")

		query := bson.M{
			"$and": []bson.M{
				{"_groupId": groupID},
				{"_userId": bson.M{"$exists": true}},
				{"_userId": bson.M{"$ne": userID}},
			},
		}
		count, err = deviceDataSession.C().Find(query).Count()
		if err != nil {
			dataLogger.WithError(err).Error("Unable to query for incorrect device data")
			continue
		}

		if count != 0 {
			dataLogger.WithField("count", count).Error("Found device data for group id with incorrect existing user id")
			continue
		}

		dataLogger.Debug("Updating device data for group id with user id")

		selector := bson.M{
			"_groupId": groupID,
			"_userId":  bson.M{"$exists": false},
		}

		if config.DryRun {
			count, err = deviceDataSession.C().Find(selector).Count()
		} else {
			update := bson.M{
				"$set": bson.M{"_userId": userID},
			}

			var changeInfo *mgo.ChangeInfo
			changeInfo, err = deviceDataSession.C().UpdateAll(selector, update)
			if changeInfo != nil {
				count = changeInfo.Updated
			}
		}

		if err != nil {
			dataLogger.WithError(err).Error("Unable to update device data for group id with user id")
			continue
		}

		if count > 0 {
			dataLogger.Infof("Migrated %d device data", count)
			migrateGroupCount++
			migrateDeviceDataCount += count
		}
	}

	if !config.DryRun {
		if count, err = deviceDataSession.C().Find(bson.M{"_userId": bson.M{"$exists": false}}).Count(); err != nil {
			logger.WithError(err).Error("Unable to query for device data without user id")
		} else if count != 0 {
			logger.WithField("count", count).Error("Found device data without user id")
		}
	}

	logger.Infof("Migrated %d device data for %d groups", migrateDeviceDataCount, migrateGroupCount)
	return nil
}
