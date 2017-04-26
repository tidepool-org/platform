package service

import (
	"fmt"
	"path/filepath"

	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/environment"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/version"
)

type Standard struct {
	name                string
	prefix              string
	versionReporter     version.Reporter
	environmentReporter environment.Reporter
	configLoader        config.Loader
	logger              log.Logger
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
	if err := s.initializeEnvironmentReporter(); err != nil {
		return err
	}
	if err := s.initializeConfigLoader(); err != nil {
		return err
	}
	if err := s.initializeLogger(); err != nil {
		return err
	}

	return nil
}

func (s *Standard) Terminate() {
	s.logger = nil
	s.configLoader = nil
	s.environmentReporter = nil
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

func (s *Standard) EnvironmentReporter() environment.Reporter {
	return s.environmentReporter
}

func (s *Standard) ConfigLoader() config.Loader {
	return s.configLoader
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

func (s *Standard) initializeEnvironmentReporter() error {
	environmentReporter, err := environment.NewDefaultReporter(s.prefix)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create environment reporter")
	}
	s.environmentReporter = environmentReporter

	return nil
}

func (s *Standard) initializeConfigLoader() error {
	configLoader, err := config.NewLoader(s.environmentReporter, filepath.Join(s.environmentReporter.GetValue("CONFIG_DIRECTORY"), s.name))
	if err != nil {
		return errors.Wrap(err, "service", "unable to create config loader")
	}
	s.configLoader = configLoader

	return nil
}

func (s *Standard) initializeLogger() error {
	loggerConfig := &log.Config{}
	if err := s.configLoader.Load("logger", loggerConfig); err != nil {
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
