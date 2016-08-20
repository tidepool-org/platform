package log

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Sirupsen/logrus"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/version"
)

type Fields map[string]interface{}

type Logger interface {
	Debug(message string)
	Info(message string)
	Warn(message string)
	Error(message string)

	WithError(err error) Logger
	WithField(key string, value interface{}) Logger
	WithFields(fields Fields) Logger
}

func NewLogger(versionReporter version.Reporter, config *Config) (Logger, error) {
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

	return &logger{
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

type logger struct {
	logger              *logrus.Logger
	fields              logrus.Fields
	ignoredFileSegments int
}

func (l *logger) Debug(message string) {
	l.finalizeFields().Debug(message)
}

func (l *logger) Info(message string) {
	l.finalizeFields().Info(message)
}

func (l *logger) Warn(message string) {
	l.finalizeFields().Warn(message)
}

func (l *logger) Error(message string) {
	l.finalizeFields().Error(message)
}

func (l *logger) WithError(err error) Logger {
	if err == nil {
		return l
	}

	return l.WithFields(Fields{"error": err.Error()})
}

func (l *logger) WithField(key string, value interface{}) Logger {
	if key == "" || value == nil {
		return l
	}

	return l.WithFields(Fields{key: value})
}

func (l *logger) WithFields(fields Fields) Logger {
	if len(fields) == 0 {
		return l
	}

	withFields := logrus.Fields{}
	for k, v := range l.fields {
		if k != "" && v != nil {
			withFields[k] = v
		}
	}
	for k, v := range fields {
		if k != "" && v != nil {
			withFields[k] = v
		}
	}

	return &logger{l.logger, withFields, l.ignoredFileSegments}
}

func (l *logger) finalizeFields() *logrus.Entry {
	return l.logger.WithFields(l.fields).WithFields(l.locationFields())
}

func (l *logger) locationFields() logrus.Fields {
	fields := logrus.Fields{}
	if _, file, line, ok := runtime.Caller(3); ok {
		fileSegments := strings.SplitN(file, "/", l.ignoredFileSegments)
		fields["file"] = fileSegments[len(fileSegments)-1]
		fields["line"] = line
	}
	return fields
}
