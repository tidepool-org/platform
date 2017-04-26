package log

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Sirupsen/logrus"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/version"
)

type Standard struct {
	logger              *logrus.Logger
	fields              logrus.Fields
	ignoredFileSegments int
}

func NewStandard(versionReporter version.Reporter, config *Config) (*Standard, error) {
	if versionReporter == nil {
		return nil, app.Error("log", "version reporter is missing")
	}
	if config == nil {
		return nil, app.Error("log", "config is missing")
	}

	if err := config.Validate(); err != nil {
		return nil, app.Error("log", "config is invalid")
	}

	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		return nil, app.Error("log", "unable to parse level")
	}

	ignoredFileSegments := 1
	if _, file, _, ok := runtime.Caller(0); ok {
		ignoredFileSegments = len(strings.Split(file, "/")) - 1
	}

	return &Standard{
		&logrus.Logger{
			Out:       os.Stderr,
			Formatter: &logrus.JSONFormatter{},
			Level:     level,
		},
		logrus.Fields{
			"process": filepath.Base(os.Args[0]),
			"pid":     os.Getpid(),
			"version": versionReporter.Short(),
		},
		ignoredFileSegments,
	}, nil
}

func (s *Standard) Debug(message string) {
	s.finalizeFields().Debug(message)
}

func (s *Standard) Info(message string) {
	s.finalizeFields().Info(message)
}

func (s *Standard) Warn(message string) {
	s.finalizeFields().Warn(message)
}

func (s *Standard) Error(message string) {
	s.finalizeFields().Error(message)
}

func (s *Standard) WithError(err error) Logger {
	if err == nil {
		return s
	}

	return s.WithFields(Fields{"error": err.Error()})
}

func (s *Standard) WithField(key string, value interface{}) Logger {
	if key == "" || value == nil {
		return s
	}

	return s.WithFields(Fields{key: value})
}

func (s *Standard) WithFields(fields Fields) Logger {
	if len(fields) == 0 {
		return s
	}

	withFields := logrus.Fields{}
	for k, v := range s.fields {
		if k != "" && v != nil {
			withFields[k] = v
		}
	}
	for k, v := range fields {
		if k != "" && v != nil {
			withFields[k] = v
		}
	}

	return &Standard{s.logger, withFields, s.ignoredFileSegments}
}

func (s *Standard) finalizeFields() *logrus.Entry {
	return s.logger.WithFields(s.fields).WithFields(s.locationFields())
}

func (s *Standard) locationFields() logrus.Fields {
	fields := logrus.Fields{}
	if _, file, line, ok := runtime.Caller(3); ok {
		fileSegments := strings.SplitN(file, "/", s.ignoredFileSegments)
		fields["file"] = fileSegments[len(fileSegments)-1]
		fields["line"] = line
	}
	return fields
}
