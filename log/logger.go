package log

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/tidepool-org/platform/errors"
)

func NewLogger(serializer Serializer, levels Levels, level Level) (Logger, error) {
	if serializer == nil {
		return nil, errors.New("log", "serializer is missing")
	}
	if levels == nil {
		return nil, errors.New("log", "levels is missing")
	}

	fields := Fields{}

	rank, found := levels[level]
	if !found {
		return nil, errors.New("log", "level not found")
	}

	ignoredFileSegments := 1
	if _, file, _, ok := runtime.Caller(0); ok {
		ignoredFileSegments = len(strings.Split(file, "/")) - 1
	}

	return &logger{
		serializer:          serializer,
		fields:              fields,
		levels:              joinLevels(levels),
		level:               level,
		rank:                rank,
		ignoredFileSegments: ignoredFileSegments,
	}, nil
}

type logger struct {
	serializer          Serializer
	fields              Fields
	levels              Levels
	level               Level
	rank                Rank
	ignoredFileSegments int
}

func (l *logger) Log(level Level, message string) {
	l.log(level, message)
}

func (l *logger) Debug(message string) {
	l.log(DebugLevel, message)
}

func (l *logger) Info(message string) {
	l.log(InfoLevel, message)
}

func (l *logger) Warn(message string) {
	l.log(WarnLevel, message)
}

func (l *logger) Error(message string) {
	l.log(ErrorLevel, message)
}

func (l *logger) Debugf(message string, args ...interface{}) {
	l.log(DebugLevel, fmt.Sprintf(message, args...))
}

func (l *logger) Infof(message string, args ...interface{}) {
	l.log(InfoLevel, fmt.Sprintf(message, args...))
}

func (l *logger) Warnf(message string, args ...interface{}) {
	l.log(WarnLevel, fmt.Sprintf(message, args...))
}

func (l *logger) Errorf(message string, args ...interface{}) {
	l.log(ErrorLevel, fmt.Sprintf(message, args...))
}

func (l *logger) WithError(err error) Logger {
	fields := Fields{}

	if err != nil {
		fields["error"] = err.Error()
	}

	return l.WithFields(fields)
}

func (l *logger) WithField(key string, value interface{}) Logger {
	return l.WithFields(Fields{key: value})
}

func (l *logger) WithFields(fields Fields) Logger {
	return &logger{
		serializer:          l.serializer,
		fields:              joinFields(l.fields, fields),
		levels:              l.levels,
		level:               l.level,
		rank:                l.rank,
		ignoredFileSegments: l.ignoredFileSegments,
	}
}

func (l *logger) WithLevel(level Level, rank Rank) Logger {
	return l.WithLevels(Levels{level: rank})
}

func (l *logger) WithLevels(levels Levels) Logger {
	return &logger{
		serializer:          l.serializer,
		fields:              l.fields,
		levels:              joinLevels(l.levels, levels),
		level:               l.level,
		rank:                l.rank,
		ignoredFileSegments: l.ignoredFileSegments,
	}
}
func (l *logger) Level() Level {
	return l.level
}

func (l *logger) SetLevel(level Level) error {
	rank, ok := l.levels[level]
	if !ok {
		return errors.New("log", "level not found")
	}

	l.level = level
	l.rank = rank
	return nil
}

func (l *logger) log(level Level, message string) {
	rank, found := l.levels[level]
	if !found {
		return
	}

	if rank < l.rank {
		return
	}

	fields := Fields{
		"level": level,
		"time":  time.Now().UTC().Format("2006-01-02T15:04:05.999Z07:00"),
	}

	if message != "" {
		fields["message"] = message
	}

	if _, file, line, ok := runtime.Caller(2); ok {
		fileSegments := strings.SplitN(file, "/", l.ignoredFileSegments)
		fields["file"] = fileSegments[len(fileSegments)-1]
		fields["line"] = line
	}

	if err := l.serializer.Serialize(joinFields(l.fields, fields)); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failure to serialize log fields: %s", err)
	}
}

func joinLevels(levels ...Levels) Levels {
	joined := Levels{}
	for _, inner := range levels {
		for level, rank := range inner {
			joined[level] = rank
		}
	}
	return joined
}

func joinFields(fields ...Fields) Fields {
	joined := Fields{}
	for _, inner := range fields {
		for key, value := range inner {
			if key != "" && value != nil {
				joined[key] = value
			}
		}
	}
	return joined
}
