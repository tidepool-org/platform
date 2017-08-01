package service

import (
	"fmt"

	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/config/env"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/version"
)

type Standard struct {
	name            string
	prefix          string
	versionReporter version.Reporter
	configReporter  config.Reporter
	logger          log.Logger
}

func NewStandard(name string, prefix string) (*Standard, error) {
	if name == "" {
		return nil, errors.New("service", "name is missing")
	}

	return &Standard{
		name:   name,
		prefix: prefix,
	}, nil
}

func (s *Standard) Initialize() error {
	if err := s.initializeVersionReporter(); err != nil {
		return err
	}
	if err := s.initializeConfigReporter(); err != nil {
		return err
	}
	if err := s.initializeLogger(); err != nil {
		return err
	}

	return nil
}

func (s *Standard) Terminate() {
	s.logger = nil
	s.configReporter = nil
	s.versionReporter = nil
}

func (s *Standard) Name() string {
	return s.name
}

func (s *Standard) Prefix() string {
	return s.prefix
}

func (s *Standard) VersionReporter() version.Reporter {
	return s.versionReporter
}

func (s *Standard) ConfigReporter() config.Reporter {
	return s.configReporter
}

func (s *Standard) Logger() log.Logger {
	return s.logger
}

func (s *Standard) initializeVersionReporter() error {
	versionReporter, err := version.NewDefaultReporter()
	if err != nil {
		return errors.Wrap(err, "service", "unable to create version reporter")
	}
	s.versionReporter = versionReporter

	return nil
}

func (s *Standard) initializeConfigReporter() error {
	configReporter, err := env.NewReporter(s.Prefix())
	if err != nil {
		return errors.Wrap(err, "service", "unable to create config reporter")
	}

	s.configReporter = configReporter.WithScopes(s.Name())

	return nil
}

func (s *Standard) initializeLogger() error {
	loggerConfig := log.NewConfig()
	if err := loggerConfig.Load(s.ConfigReporter().WithScopes("logger")); err != nil {
		return errors.Wrap(err, "service", "unable to load logger config")
	}

	logger, err := log.NewStandard(s.versionReporter, loggerConfig)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create logger")
	}
	s.logger = logger

	s.logger.Warn(fmt.Sprintf("Logger level is %s", loggerConfig.Level))

	return nil
}
