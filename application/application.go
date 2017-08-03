package application

import (
	"os"
	"path/filepath"

	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/config/env"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/json"
	"github.com/tidepool-org/platform/version"
)

type Application struct {
	name            string
	prefix          string
	versionReporter version.Reporter
	configReporter  config.Reporter
	logger          log.Logger
}

func New(name string, prefix string) (*Application, error) {
	if name == "" {
		return nil, errors.New("application", "name is missing")
	}
	if prefix == "" {
		return nil, errors.New("application", "prefix is missing")
	}

	return &Application{
		name:   name,
		prefix: prefix,
	}, nil
}

func (a *Application) Initialize() error {
	if err := a.initializeVersionReporter(); err != nil {
		return err
	}
	if err := a.initializeConfigReporter(); err != nil {
		return err
	}
	if err := a.initializeLogger(); err != nil {
		return err
	}

	return nil
}

func (a *Application) Terminate() {
	a.logger = nil
	a.configReporter = nil
	a.versionReporter = nil
}

func (a *Application) Name() string {
	return a.name
}

func (a *Application) Prefix() string {
	return a.prefix
}

func (a *Application) VersionReporter() version.Reporter {
	return a.versionReporter
}

func (a *Application) ConfigReporter() config.Reporter {
	return a.configReporter
}

func (a *Application) Logger() log.Logger {
	return a.logger
}

func (a *Application) initializeVersionReporter() error {
	versionReporter, err := version.NewDefaultReporter()
	if err != nil {
		return errors.Wrap(err, "application", "unable to create version reporter")
	}

	a.versionReporter = versionReporter

	return nil
}

func (a *Application) initializeConfigReporter() error {
	configReporter, err := env.NewReporter(a.Prefix())
	if err != nil {
		return errors.Wrap(err, "application", "unable to create config reporter")
	}

	a.configReporter = configReporter.WithScopes(a.Name())

	return nil
}

func (a *Application) initializeLogger() error {
	level := a.ConfigReporter().WithScopes("logger").GetWithDefault("level", "warn")

	logger, err := json.NewLogger(os.Stdout, log.DefaultLevels(), log.Level(level))
	if err != nil {
		return errors.Wrap(err, "application", "unable to create logger")
	}

	logger = logger.WithFields(log.Fields{
		"process": filepath.Base(os.Args[0]),
		"pid":     os.Getpid(),
		"version": a.VersionReporter().Short(),
	})

	a.logger = logger

	a.Logger().Infof("Logger level is %s", a.Logger().Level())

	return nil
}
