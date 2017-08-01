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
	application.Usage = "Migrate all data deduplicators to latest format"
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
		return nil, errors.Wrap(err, "main", "unable to create version reporter")
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

	err = migrateDataDeduplicatorDescriptors(logger, config)
	if err != nil {
		logger.WithError(err).Error("Unable to migrate group id to user id for data")
		os.Exit(1)
	}
}

func buildConfigFromContext(context *cli.Context) (*Config, error) {
	config := &Config{
		Log:   log.NewConfig(),
		Mongo: mongo.NewConfig(),
	}

	if context.Bool(VerboseFlag) {
		config.Log.Level = "debug"
	} else {
		config.Log.Level = "info"
	}
	config.Mongo.Addresses = app.SplitStringAndRemoveWhitespace(context.String(AddressesFlag), ",")
	config.Mongo.TLS = context.Bool(TLSFlag)
	config.DryRun = context.Bool(DryRunFlag)

	return config, nil
}

func initializeLogger(versionReporter version.Reporter, config *Config) (log.Logger, error) {
	logger, err := log.NewStandard(versionReporter, config.Log)
	if err != nil {
		return nil, errors.Wrap(err, "main", "unable to create logger")
	}

	return logger, nil
}

func migrateDataDeduplicatorDescriptors(logger log.Logger, config *Config) error {
	logger.Debug("Migrating data deduplicator descriptors")

	logger.Debug("Creating data store")

	mongoConfig := config.Mongo.Clone()
	mongoConfig.Database = "data"
	mongoConfig.Collection = "deviceData"
	mongoConfig.Timeout = 60 * time.Minute
	dataStore, err := mongo.New(logger, mongoConfig)
	if err != nil {
		return errors.Wrap(err, "main", "unable to create data store")
	}
	defer dataStore.Close()

	logger.Debug("Creating data session")

	dataStoreSession := dataStore.NewSession(logger)
	defer dataStoreSession.Close()

	var count int
	count += migrateUploadDataDeduplicatorDescriptor(logger, config, dataStoreSession, "truncate", "org.tidepool.truncate")
	count += migrateUploadDataDeduplicatorDescriptor(logger, config, dataStoreSession, "hash-deactivate-old", "org.tidepool.hash-deactivate-old")
	count += migrateUploadDataDeduplicatorDescriptor(logger, config, dataStoreSession, "hash", "org.tidepool.hash-drop-new")
	count += migrateNonUploadDataDeduplicatorDescriptor(logger, config, dataStoreSession)

	logger.Info(fmt.Sprintf("Migrated %d data duplicator descriptors", count))

	return nil
}

func migrateUploadDataDeduplicatorDescriptor(logger log.Logger, config *Config, dataStoreSession *mongo.Session, fromName string, toName string) int {
	logger = logger.WithFields(log.Fields{"fromName": fromName, "toName": toName})

	logger.Debug("Migrating upload data deduplicator descriptors")

	var count int
	var err error

	selector := bson.M{
		"type":               "upload",
		"_deduplicator.name": fromName,
	}

	if config.DryRun {
		count, err = dataStoreSession.C().Find(selector).Count()
	} else {
		update := bson.M{
			"$set": bson.M{
				"_deduplicator.name":    toName,
				"_deduplicator.version": "0.0.0",
			},
		}

		var changeInfo *mgo.ChangeInfo
		changeInfo, err = dataStoreSession.C().UpdateAll(selector, update)
		if changeInfo != nil {
			count = changeInfo.Updated
		}
	}

	if err != nil {
		logger.WithError(err).Error("Unable to migrate upload data deduplicator descriptors")
	}

	logger.Debug(fmt.Sprintf("Migrated %d upload data deduplicator descriptors", count))

	return count
}

func migrateNonUploadDataDeduplicatorDescriptor(logger log.Logger, config *Config, dataStoreSession *mongo.Session) int {
	logger.Debug("Migrating non-upload data deduplicator descriptors")

	var count int
	var err error

	selector := bson.M{
		"type": bson.M{
			"$ne": "upload",
		},
		"_deduplicator.name": bson.M{
			"$exists": true,
		},
	}

	if config.DryRun {
		count, err = dataStoreSession.C().Find(selector).Count()
	} else {
		update := bson.M{
			"$unset": bson.M{
				"_deduplicator.name": 1,
			},
		}

		var changeInfo *mgo.ChangeInfo
		changeInfo, err = dataStoreSession.C().UpdateAll(selector, update)
		if changeInfo != nil {
			count = changeInfo.Updated
		}
	}

	if err != nil {
		logger.WithError(err).Error("Unable to migrate non-upload data deduplicator descriptors")
	}

	logger.Debug(fmt.Sprintf("Migrated %d non-upload data deduplicator descriptors", count))

	return count
}
