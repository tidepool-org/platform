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

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/store/mongo"
	"github.com/tidepool-org/platform/dataservices/server"
	"github.com/tidepool-org/platform/dataservices/server/api"
	standardServer "github.com/tidepool-org/platform/dataservices/server/server"
	"github.com/tidepool-org/platform/environment"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/userservices/client"
	"github.com/tidepool-org/platform/version"
)

var (
	VersionBase        string
	VersionShortCommit string
	VersionFullCommit  string
)

func main() {
	environmentReporter, err := initializeEnvironmentReporter()
	if err != nil {
		fmt.Printf("ERROR: Failure initializing environment reporter: %s\n", err.Error())
		os.Exit(1)
	}

	versionReporter, err := initializeVersionReporter()
	if err != nil {
		fmt.Printf("ERROR: Failure initializing version reporter: %s\n", err.Error())
		os.Exit(1)
	}

	configLoader, err := initializeConfigLoader(environmentReporter)
	if err != nil {
		fmt.Printf("ERROR: Failure initializing config loader: %s\n", err.Error())
		os.Exit(1)
	}

	logger, err := initializeLogger(configLoader, versionReporter)
	if err != nil {
		fmt.Printf("ERROR: Failure initializing logger: %s\n", err.Error())
		os.Exit(1)
	}

	dataStore, err := initializeDataStore(configLoader, logger)
	if err != nil {
		logger.WithError(err).Error("Failure initializing data store")
		os.Exit(1)
	}
	defer dataStore.Close()

	userServicesClient, err := initializeUserServicesClient(configLoader, logger)
	if err != nil {
		logger.WithError(err).Error("Failure initializing userservices client")
		os.Exit(1)
	}
	defer userServicesClient.Close()

	api, err := initializeAPI(configLoader, logger, dataStore, userServicesClient, versionReporter)
	if err != nil {
		logger.WithError(err).Error("Failure initializing API")
		os.Exit(1)
	}

	server, err := initializeServer(configLoader, logger, api)
	if err != nil {
		logger.WithError(err).Error("Failure initializing server")
		os.Exit(1)
	}

	if err = server.Serve(); err != nil {
		logger.WithError(err).Error("Failure running server")
		os.Exit(1)
	}
}

// TODO: Wrap this up into an object

func initializeEnvironmentReporter() (environment.Reporter, error) {
	return environment.NewReporter(os.Getenv("TIDEPOOL_ENV"))
}

func initializeVersionReporter() (version.Reporter, error) {
	return version.NewReporter(VersionBase, VersionShortCommit, VersionFullCommit)
}

func initializeConfigLoader(environmentReporter environment.Reporter) (config.Loader, error) {
	return config.NewLoader(os.Getenv("TIDEPOOL_CONFIG_DIRECTORY"), "TIDEPOOL", environmentReporter)
}

func initializeLogger(configLoader config.Loader, versionReporter version.Reporter) (log.Logger, error) {
	loggerConfig := &log.Config{}
	if err := configLoader.Load("logger", loggerConfig); err != nil {
		return nil, app.ExtError(err, "dataservices", "unable to load logger config")
	}

	logger, err := log.NewLogger(loggerConfig, versionReporter)
	if err != nil {
		return nil, app.ExtError(err, "dataservices", "unable to initialize logger")
	}

	logger.Info(fmt.Sprintf("Logger level is %s", loggerConfig.Level))

	return logger, nil
}

func initializeDataStore(configLoader config.Loader, logger log.Logger) (store.Store, error) {

	// TODO: Consider alternate data stores

	logger.Debug("Loading mongo data store config")

	mongoDataStoreConfig := &mongo.Config{}
	if err := configLoader.Load("data_store", mongoDataStoreConfig); err != nil {
		return nil, app.ExtError(err, "dataservices", "unable to load mongo data store config")
	}
	mongoDataStoreConfig.Collection = "deviceData"

	logger.Debug("Creating mongo data store")

	mongoDataStore, err := mongo.New(logger, mongoDataStoreConfig)
	if err != nil {
		return nil, app.ExtError(err, "dataservices", "unable to create mongo data store")
	}

	return mongoDataStore, nil
}

func initializeUserServicesClient(configLoader config.Loader, logger log.Logger) (client.Client, error) {

	logger.Debug("Loading userservices client config")

	userServicesClientConfig := &client.Config{}
	if err := configLoader.Load("userservices_client", userServicesClientConfig); err != nil {
		return nil, app.ExtError(err, "dataservices", "unable to load userservices client config")
	}

	logger.Debug("Creating userservices client")

	userServicesClient, err := client.NewStandard(logger, userServicesClientConfig)
	if err != nil {
		return nil, app.ExtError(err, "dataservices", "unable to create userservices client")
	}

	logger.Debug("Starting userservices client")
	if err = userServicesClient.Start(); err != nil {
		return nil, app.ExtError(err, "dataservices", "unable to start userservices client")
	}

	return userServicesClient, nil
}

func initializeAPI(configLoader config.Loader, logger log.Logger, dataStore store.Store, userServicesClient client.Client, reporter version.Reporter) (server.API, error) {
	return api.NewStandard(logger, dataStore, userServicesClient, reporter)
}

func initializeServer(configLoader config.Loader, logger log.Logger, api server.API) (server.Server, error) {

	logger.Debug("Loading dataservices server config")

	dataservicesServerConfig := &standardServer.Config{}
	if err := configLoader.Load("dataservices_server", dataservicesServerConfig); err != nil {
		return nil, app.ExtError(err, "dataservices", "unable to load dataservices server config")
	}

	logger.Debug("Creating dataservices server")

	dataservicesServer, err := standardServer.NewStandard(logger, api, dataservicesServerConfig)
	if err != nil {
		return nil, app.ExtError(err, "dataservices", "unable to create dataservices server")
	}

	return dataservicesServer, nil
}
