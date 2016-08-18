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
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/deduplicator"
	"github.com/tidepool-org/platform/data/deduplicator/delegate"
	"github.com/tidepool-org/platform/data/deduplicator/truncate"
	"github.com/tidepool-org/platform/data/factory"
	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/store/mongo"
	"github.com/tidepool-org/platform/dataservices/service"
	"github.com/tidepool-org/platform/dataservices/service/api"
	standardServer "github.com/tidepool-org/platform/dataservices/service/server"
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

	dataFactory, err := initializeDataFactory(logger)
	if err != nil {
		logger.WithError(err).Error("Failure initializing data factory")
		os.Exit(1)
	}

	dataStore, err := initializeDataStore(configLoader, logger)
	if err != nil {
		logger.WithError(err).Error("Failure initializing data store")
		os.Exit(1)
	}
	defer dataStore.Close()

	dataDeduplicatorFactory, err := initializeDataDeduplicatorFactory(logger)
	if err != nil {
		logger.WithError(err).Error("Failure initializing data deduplicator factory")
		os.Exit(1)
	}

	userServicesClient, err := initializeUserServicesClient(configLoader, logger)
	if err != nil {
		logger.WithError(err).Error("Failure initializing userservices client")
		os.Exit(1)
	}
	defer userServicesClient.Close()

	api, err := initializeAPI(logger, dataFactory, dataStore, dataDeduplicatorFactory, userServicesClient, versionReporter)
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

func initializeDataFactory(logger log.Logger) (data.Factory, error) {

	logger.Debug("Creating data factory")

	standardDataFactory, err := factory.NewStandard()
	if err != nil {
		return nil, app.ExtError(err, "dataservices", "unable to create standard data factory")
	}

	return standardDataFactory, nil
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

func initializeDataDeduplicatorFactory(logger log.Logger) (deduplicator.Factory, error) {

	logger.Debug("Creating data deduplicator factory")

	truncateDeduplicatorFactory, err := truncate.NewFactory()
	if err != nil {
		return nil, app.ExtError(err, "dataservices", "unable to create truncate deduplicator factory")
	}

	factories := []deduplicator.Factory{
		truncateDeduplicatorFactory,
	}

	delegateDeduplicatorFactory, err := delegate.NewFactory(factories)
	if err != nil {
		return nil, app.ExtError(err, "dataservices", "unable to create delegate deduplicator factory")
	}

	return delegateDeduplicatorFactory, nil
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

func initializeAPI(logger log.Logger, dataFactory data.Factory, dataStore store.Store, dataDeduplicatorFactory deduplicator.Factory, userServicesClient client.Client, reporter version.Reporter) (service.API, error) {
	return api.NewStandard(logger, dataFactory, dataStore, dataDeduplicatorFactory, userServicesClient, reporter)
}

func initializeServer(configLoader config.Loader, logger log.Logger, api service.API) (service.Server, error) {

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
