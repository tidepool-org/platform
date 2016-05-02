package log

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
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Sirupsen/logrus"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/version"
)

type Config struct {
	Level string `default:"warn"`
}

func (c *Config) Validate() error {
	if _, err := logrus.ParseLevel(c.Level); err != nil {
		return app.ExtError(err, "log", "level is not valid")
	}
	return nil
}

var _logger *loggerWithFields

type loggerWithFields struct {
	logger              *logrus.Logger
	fields              map[string]interface{}
	ignoredFileSegments int
}

func (l *loggerWithFields) Debug(message string) {
	l.finalizeFields().Debug(message)
}

func (l *loggerWithFields) Info(message string) {
	l.finalizeFields().Info(message)
}

func (l *loggerWithFields) Warn(message string) {
	l.finalizeFields().Warn(message)
}

func (l *loggerWithFields) Error(message string) {
	l.finalizeFields().Error(message)
}

// TODO: Remove Fatal

func (l *loggerWithFields) Fatal(message string) {
	l.finalizeFields().Fatal(message)
}

func (l *loggerWithFields) WithError(err error) Logger {
	return l.WithFields(map[string]interface{}{"error": err.Error()})
}

func (l *loggerWithFields) WithField(key string, value interface{}) Logger {
	return l.WithFields(map[string]interface{}{key: value})
}

func (l *loggerWithFields) WithFields(fields map[string]interface{}) Logger {
	withFields := make(map[string]interface{})
	for k, v := range l.fields {
		withFields[k] = v
	}
	for k, v := range fields {
		withFields[k] = v
	}
	return &loggerWithFields{l.logger, withFields, l.ignoredFileSegments}
}

func (l *loggerWithFields) finalizeFields() *logrus.Entry {
	return l.logger.WithFields(l.fields).WithFields(l.locationFields())
}

func (l *loggerWithFields) locationFields() map[string]interface{} {
	fields := map[string]interface{}{}
	if _, file, line, ok := runtime.Caller(l.skip()); ok {
		fileSegments := strings.SplitN(file, "/", l.ignoredFileSegments+1)
		fields["file"] = fileSegments[len(fileSegments)-1]
		fields["line"] = line
	}
	return fields
}

func (l *loggerWithFields) skip() int {
	if l == _logger {
		return 4
	}
	return 3
}

func init() {
	ignoredFileSegments := 0
	if _, file, _, ok := runtime.Caller(0); ok {
		ignoredFileSegments = len(strings.Split(file, "/")) - 2
	}

	_logger = &loggerWithFields{
		&logrus.Logger{
			Out:       os.Stderr,
			Formatter: &logrus.JSONFormatter{},
			Level:     logrus.WarnLevel,
		},
		map[string]interface{}{
			"process": filepath.Base(os.Args[0]),
			"pid":     os.Getpid(),
			"version": version.Current().Short(),
		},
		ignoredFileSegments,
	}

	loggerConfig := &Config{}
	if err := config.Load("logger", loggerConfig); err != nil {
		_logger.WithError(err).Fatal("unable to load logger config")
	}
	if err := loggerConfig.Validate(); err != nil {
		_logger.WithError(err).Fatal("logger config is not valid")
	}

	level, err := logrus.ParseLevel(loggerConfig.Level)
	if err != nil {
		_logger.WithError(err).Fatal("unable to parse logger level")
	}

	_logger.logger.Level = level

	_logger.Info(fmt.Sprintf("Logger level is %s", level.String()))
}
