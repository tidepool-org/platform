package application

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"

	applicationVersion "github.com/tidepool-org/platform/application/version"
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/config/env"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/json"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/sync"
	"github.com/tidepool-org/platform/version"
)

type Application struct {
	name            string
	prefix          string
	scopes          []string
	versionReporter version.Reporter
	configReporter  config.Reporter
	logger          log.Logger
}

func New(prefix string, scopes ...string) (*Application, error) {
	if prefix == "" {
		return nil, errors.New("prefix is missing")
	}

	name := filepath.Base(os.Args[0])

	if strings.EqualFold(name, "debug") {
		if debugName, found := syscall.Getenv(env.GetKey(prefix, []string{name}, "name")); found {
			name = debugName
		}
	}

	return &Application{
		name:   name,
		prefix: prefix,
		scopes: scopes,
	}, nil
}

func (a *Application) Initialize() error {
	if err := a.initializeVersionReporter(); err != nil {
		return err
	}
	if err := a.initializeConfigReporter(); err != nil {
		return err
	}
	return a.initializeLogger()
}

func (a *Application) Terminate() {
	a.terminateLogger()
	a.terminateConfigReporter()
	a.terminateVersionReporter()
}

func (a *Application) Name() string {
	return a.name
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

func (a *Application) SetLogger(logger log.Logger) {
	if logger == nil {
		logger = null.NewLogger()
	}

	a.logger = logger
}

func (a *Application) initializeVersionReporter() error {
	versionReporter, err := applicationVersion.NewReporter()
	if err != nil {
		return errors.Wrap(err, "unable to create version reporter")
	}

	a.versionReporter = versionReporter

	return nil
}

func (a *Application) terminateVersionReporter() {
	a.versionReporter = nil
}

func (a *Application) initializeConfigReporter() error {
	configReporter, err := env.NewReporter(a.prefix)
	if err != nil {
		return errors.Wrap(err, "unable to create config reporter")
	}

	a.configReporter = configReporter.WithScopes(a.Name()).WithScopes(a.scopes...)

	return nil
}

func (a *Application) terminateConfigReporter() {
	a.configReporter = nil
}

func (a *Application) initializeLogger() error {
	writer, err := sync.NewWriter(os.Stdout)
	if err != nil {
		return errors.Wrap(err, "unable to create writer")
	}

	level := a.ConfigReporter().WithScopes("logger").GetWithDefault("level", "warn")

	logger, err := json.NewLogger(writer, log.DefaultLevelRanks(), log.Level(level))
	if err != nil {
		return errors.Wrap(err, "unable to create logger")
	}

	logger = logger.WithField("process", map[string]interface{}{
		"name":    a.Name(),
		"id":      os.Getpid(),
		"version": a.VersionReporter().Short(),
	})

	a.logger = logger

	a.logger.Infof("Logger level is %s", a.logger.Level())

	return nil
}

func (a *Application) terminateLogger() {
	if a.logger != nil {
		a.logger.Info("Destroying logger")
		a.logger = nil

		os.Stdout.Sync()
	}
}
