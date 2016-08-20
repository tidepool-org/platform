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
	"github.com/tidepool-org/platform/dataservices/service/server"
	"github.com/tidepool-org/platform/environment"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/userservices/client"
	"github.com/tidepool-org/platform/version"
)

func main() {
	versionReporter, err := initializeVersionReporter()
	if err != nil {
		fmt.Printf("ERROR: Failure initializing version reporter: %s\n", err.Error())
		os.Exit(1)
	}

	environmentReporter, err := initializeEnvironmentReporter()
	if err != nil {
		fmt.Printf("ERROR: Failure initializing environment reporter: %s\n", err.Error())
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
		logger.WithError(err).Error("Failure initializing user services client")
		os.Exit(1)
	}
	defer userServicesClient.Close()

	dataServicesAPI, err := initializeDataServicesAPI(logger, dataFactory, dataStore, dataDeduplicatorFactory, userServicesClient, versionReporter, environmentReporter)
	if err != nil {
		logger.WithError(err).Error("Failure initializing data services API")
		os.Exit(1)
	}
	defer dataServicesAPI.Close()

	dataServicesServer, err := initializeDataServicesServer(configLoader, logger, dataServicesAPI)
	if err != nil {
		logger.WithError(err).Error("Failure initializing data services server")
		os.Exit(1)
	}
	defer dataServicesServer.Close()

	if err = dataServicesServer.Serve(); err != nil {
		logger.WithError(err).Error("Failure running data services server")
		os.Exit(1)
	}
}

// TODO: Wrap this up into an object

func initializeVersionReporter() (version.Reporter, error) {
	versionReporter, err := version.NewDefaultReporter()
	if err != nil {
		return nil, app.ExtError(err, "dataservices", "unable to create version reporter")
	}

	return versionReporter, nil
}

func initializeEnvironmentReporter() (environment.Reporter, error) {
	environmentReporter, err := environment.NewDefaultReporter()
	if err != nil {
		return nil, app.ExtError(err, "dataservices", "unable to create environment reporter")
	}

	return environmentReporter, nil
}

func initializeConfigLoader(environmentReporter environment.Reporter) (config.Loader, error) {
	configLoader, err := config.NewLoader(os.Getenv("TIDEPOOL_CONFIG_DIRECTORY"), "TIDEPOOL", environmentReporter)
	if err != nil {
		return nil, app.ExtError(err, "dataservices", "unable to create config loader")
	}

	return configLoader, nil
}

func initializeLogger(configLoader config.Loader, versionReporter version.Reporter) (log.Logger, error) {
	loggerConfig := &log.Config{}
	if err := configLoader.Load("logger", loggerConfig); err != nil {
		return nil, app.ExtError(err, "dataservices", "unable to load logger config")
	}

	logger, err := log.NewLogger(loggerConfig, versionReporter)
	if err != nil {
		return nil, app.ExtError(err, "dataservices", "unable to create logger")
	}

	logger.Info(fmt.Sprintf("Logger level is %s", loggerConfig.Level))

	return logger, nil
}

func initializeDataFactory(logger log.Logger) (data.Factory, error) {
	logger.Debug("Creating data factory")

	dataFactory, err := factory.NewStandard()
	if err != nil {
		return nil, app.ExtError(err, "dataservices", "unable to create data factory")
	}

	return dataFactory, nil
}

func initializeDataStore(configLoader config.Loader, logger log.Logger) (store.Store, error) {
	logger.Debug("Loading data store config")

	mongoDataStoreConfig := &mongo.Config{}
	if err := configLoader.Load("data_store", mongoDataStoreConfig); err != nil {
		return nil, app.ExtError(err, "dataservices", "unable to load data store config")
	}
	mongoDataStoreConfig.Collection = "deviceData"

	logger.Debug("Creating data store")

	mongoDataStore, err := mongo.New(logger, mongoDataStoreConfig)
	if err != nil {
		return nil, app.ExtError(err, "dataservices", "unable to create data store")
	}

	return mongoDataStore, nil
}

func initializeDataDeduplicatorFactory(logger log.Logger) (deduplicator.Factory, error) {
	logger.Debug("Creating truncate data deduplicator factory")

	truncateDeduplicatorFactory, err := truncate.NewFactory()
	if err != nil {
		return nil, app.ExtError(err, "dataservices", "unable to create truncate data deduplicator factory")
	}

	logger.Debug("Creating delegate data deduplicator factory")

	factories := []deduplicator.Factory{
		truncateDeduplicatorFactory,
	}

	delegateDeduplicatorFactory, err := delegate.NewFactory(factories)
	if err != nil {
		return nil, app.ExtError(err, "dataservices", "unable to create delegate data deduplicator factory")
	}

	return delegateDeduplicatorFactory, nil
}

func initializeUserServicesClient(configLoader config.Loader, logger log.Logger) (client.Client, error) {
	logger.Debug("Loading user services client config")

	userServicesClientConfig := &client.Config{}
	if err := configLoader.Load("userservices_client", userServicesClientConfig); err != nil {
		return nil, app.ExtError(err, "dataservices", "unable to load user services client config")
	}

	logger.Debug("Creating user services client")

	userServicesClient, err := client.NewStandard(logger, userServicesClientConfig)
	if err != nil {
		return nil, app.ExtError(err, "dataservices", "unable to create user services client")
	}

	logger.Debug("Starting user services client")
	if err = userServicesClient.Start(); err != nil {
		return nil, app.ExtError(err, "dataservices", "unable to start user services client")
	}

	return userServicesClient, nil
}

func initializeDataServicesAPI(logger log.Logger, dataFactory data.Factory, dataStore store.Store, dataDeduplicatorFactory deduplicator.Factory, userServicesClient client.Client, versionReporter version.Reporter, environmentReporter environment.Reporter) (service.API, error) {
	logger.Debug("Creating data services api")

	dataServicesAPI, err := api.NewStandard(logger, dataFactory, dataStore, dataDeduplicatorFactory, userServicesClient, versionReporter, environmentReporter)
	if err != nil {
		return nil, app.ExtError(err, "dataservices", "unable to create data services api")
	}

	return dataServicesAPI, nil
}

func initializeDataServicesServer(configLoader config.Loader, logger log.Logger, api service.API) (service.Server, error) {
	logger.Debug("Loading data services server config")

	dataServicesServerConfig := &server.Config{}
	if err := configLoader.Load("dataservices_server", dataServicesServerConfig); err != nil {
		return nil, app.ExtError(err, "dataservices", "unable to load data services server config")
	}

	logger.Debug("Creating data services server")

	dataServicesServer, err := server.NewStandard(logger, api, dataServicesServerConfig)
	if err != nil {
		return nil, app.ExtError(err, "dataservices", "unable to create data services server")
	}

	return dataServicesServer, nil
}
