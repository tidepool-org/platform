package service

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
	level := s.ConfigReporter().WithScopes("logger").GetWithDefault("level", "warn")

	logger, err := json.NewLogger(os.Stdout, log.DefaultLevels(), log.Level(level))
	if err != nil {
		return errors.Wrap(err, "service", "unable to create logger")
	}

	logger = logger.WithFields(log.Fields{
		"process": filepath.Base(os.Args[0]),
		"pid":     os.Getpid(),
		"version": s.VersionReporter().Short(),
	})

	s.logger = logger

	s.Logger().Warnf("Logger level is %s", s.Logger().Level())

	return nil
}
