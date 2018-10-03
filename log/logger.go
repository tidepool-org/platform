package log

import (
	"fmt"
	"os"
	"time"

	"github.com/tidepool-org/platform/errors"
)

// CONCURRENCY: SAFE IFF serializer is safe

const TimeFormat = "2006-01-02T15:04:05.999Z07:00"

func NewLogger(serializer Serializer, levelRanks LevelRanks, level Level) (Logger, error) {
	if serializer == nil {
		return nil, errors.New("serializer is missing")
	}
	if levelRanks == nil {
		return nil, errors.New("level ranks is missing")
	}

	fields := Fields{}

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
	serializer Serializer
	fields     Fields
	levelRanks LevelRanks
	level      Level
	rank       Rank
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
	var value interface{}
	if err != nil {
		value = &errors.Serializable{Error: err}
	}
	return l.WithField("error", value)
}

func (l *logger) WithField(key string, value interface{}) Logger {
	return l.WithFields(Fields{key: value})
}

func (l *logger) WithFields(fields Fields) Logger {
	return &logger{
		serializer: l.serializer,
		fields:     joinFields(l.fields, fields),
		levelRanks: l.levelRanks,
		level:      l.level,
		rank:       l.rank,
	}
}

func (l *logger) WithLevelRank(level Level, rank Rank) Logger {
	return l.WithLevelRanks(LevelRanks{level: rank})
}

func (l *logger) WithLevelRanks(levelRanks LevelRanks) Logger {
	return &logger{
		serializer: l.serializer,
		fields:     l.fields,
		levelRanks: joinLevelRanks(l.levelRanks, levelRanks),
		level:      l.level,
		rank:       l.rank,
	}
}

func (l *logger) WithLevel(level Level) Logger {
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

func (l *logger) Level() Level {
	return l.level
}

func (l *logger) log(level Level, message string) {
	rank, found := l.levelRanks[level]
	if !found {
		return
	}

	if rank < l.rank {
		return
	}

	fields := Fields{
		"caller": errors.GetCaller(2),
		"level":  level,
		"time":   time.Now().Format(TimeFormat),
	}

	if message != "" {
		fields["message"] = message
	}

	if err := l.serializer.Serialize(joinFields(l.fields, fields)); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failure to serialize log fields: %s", err)
	}
}

func joinLevelRanks(levelRanks ...LevelRanks) LevelRanks {
	joined := LevelRanks{}
	for _, inner := range levelRanks {
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
