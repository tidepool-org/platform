package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
)

// CONCURRENCY: SAFE IFF serializer is safe

func New(serializer log.Serializer, levelRanks log.LevelRanks, level log.Level) (log.Logger, error) {
	if serializer == nil {
		return nil, errors.New("serializer is missing")
	}
	if levelRanks == nil {
		return nil, errors.New("level ranks is missing")
	}

	fields := log.Fields{}

	rank, found := levelRanks[level]
	if !found {
		return nil, errors.New("level not found")
	}

	return &logger{
		serializer: serializer,
		fields:     fields,
		levelRanks: joinLevelRanks(levelRanks),
		level:      level,
		rank:       rank,
	}, nil
}

type logger struct {
	serializer log.Serializer
	fields     log.Fields
	levelRanks log.LevelRanks
	level      log.Level
	rank       log.Rank
}

func (l *logger) Log(level log.Level, message string) {
	l.log(level, message)
}

func (l *logger) Debug(message string) {
	l.log(log.DebugLevel, message)
}

func (l *logger) Info(message string) {
	l.log(log.InfoLevel, message)
}

func (l *logger) Warn(message string) {
	l.log(log.WarnLevel, message)
}

func (l *logger) Error(message string) {
	l.log(log.ErrorLevel, message)
}

func (l *logger) Debugf(message string, args ...interface{}) {
	l.log(log.DebugLevel, fmt.Sprintf(message, args...))
}

func (l *logger) Infof(message string, args ...interface{}) {
	l.log(log.InfoLevel, fmt.Sprintf(message, args...))
}

func (l *logger) Warnf(message string, args ...interface{}) {
	l.log(log.WarnLevel, fmt.Sprintf(message, args...))
}

func (l *logger) Errorf(message string, args ...interface{}) {
	l.log(log.ErrorLevel, fmt.Sprintf(message, args...))
}

func (l *logger) WithError(err error) log.Logger {
	var value interface{}
	if err != nil {
		value = errors.NewSerializable(err)
	}
	return l.WithField("error", value)
}

func (l *logger) WithField(key string, value interface{}) log.Logger {
	return l.WithFields(log.Fields{key: value})
}

func (l *logger) WithFields(fields log.Fields) log.Logger {
	return &logger{
		serializer: l.serializer,
		fields:     joinFields(l.fields, fields),
		levelRanks: l.levelRanks,
		level:      l.level,
		rank:       l.rank,
	}
}

func (l *logger) WithLevelRank(level log.Level, rank log.Rank) log.Logger {
	return l.WithLevelRanks(log.LevelRanks{level: rank})
}

func (l *logger) WithLevelRanks(levelRanks log.LevelRanks) log.Logger {
	return &logger{
		serializer: l.serializer,
		fields:     l.fields,
		levelRanks: joinLevelRanks(l.levelRanks, levelRanks),
		level:      l.level,
		rank:       l.rank,
	}
}

func (l *logger) WithLevel(level log.Level) log.Logger {
	rank, ok := l.levelRanks[level]
	if !ok {
		level = l.level
		rank = l.rank
	}

	return &logger{
		serializer: l.serializer,
		fields:     l.fields,
		levelRanks: l.levelRanks,
		level:      level,
		rank:       rank,
	}
}

func (l *logger) Level() log.Level {
	return l.level
}

func (l *logger) log(level log.Level, message string) {
	rank, found := l.levelRanks[level]
	if !found {
		return
	}

	if rank < l.rank {
		return
	}

	fields := log.Fields{
		"caller": errors.GetCaller(2),
		"level":  level,
		"time":   time.Now().Truncate(time.Microsecond).Format(time.RFC3339Nano),
	}

	if message != "" {
		fields["message"] = message
	}

	if err := l.serializer.Serialize(joinFields(l.fields, fields)); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failure to serialize log fields: %s", err)
	}
}

func joinLevelRanks(levelRanks ...log.LevelRanks) log.LevelRanks {
	joined := log.LevelRanks{}
	for _, inner := range levelRanks {
		for level, rank := range inner {
			joined[level] = rank
		}
	}
	return joined
}

func joinFields(fields ...log.Fields) log.Fields {
	joined := log.Fields{}
	for _, inner := range fields {
		for key, value := range inner {
			if key != "" {
				if value != nil {
					joined[key] = value
				} else {
					delete(joined, key)
				}
			}
		}
	}
	return joined
}
