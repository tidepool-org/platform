package application

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	applicationVersion "github.com/tidepool-org/platform/application/version"
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/config/env"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	logJSON "github.com/tidepool-org/platform/log/json"
	"github.com/tidepool-org/platform/sync"
	"github.com/tidepool-org/platform/version"
)

type Application struct {
	prefix          string
	name            string
	userAgent       string
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

	userAgent := fmt.Sprintf("%s-%s", userAgentTitle(prefix), userAgentTitle(name))

	return &Application{
		prefix:    prefix,
		name:      name,
		userAgent: userAgent,
		scopes:    scopes,
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

func (a *Application) Prefix() string {
	return a.prefix
}

func (a *Application) Name() string {
	return a.name
}

func (a *Application) UserAgent() string {
	return a.userAgent
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
	versionReporter, err := applicationVersion.NewReporter()
	if err != nil {
		return errors.Wrap(err, "unable to create version reporter")
	}

	a.versionReporter = versionReporter

	a.userAgent = fmt.Sprintf("%s/%s", a.userAgent, a.versionReporter.Base())

	return nil
}

func (a *Application) terminateVersionReporter() {
	a.versionReporter = nil
}

func (a *Application) initializeConfigReporter() error {
	configReporter, err := env.NewReporter(a.Prefix())
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

	logger, err := logJSON.NewLogger(writer, log.DefaultLevelRanks(), log.Level(level))
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

var userAgentTitleExpression = regexp.MustCompile("[^a-zA-Z0-9]+")

func userAgentTitle(s string) string {
	return strings.Replace(strings.Title(strings.ToLower(userAgentTitleExpression.ReplaceAllString(s, " "))), " ", "", -1)
}
