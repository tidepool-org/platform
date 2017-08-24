package log

type (
	Fields     map[string]interface{}
	Level      string
	Rank       int
	LevelRanks map[Level]Rank
)

type Serializer interface {
	Serialize(fields Fields) error
}

type Logger interface {
	Log(level Level, message string)

	Debug(message string)
	Info(message string)
	Warn(message string)
	Error(message string)

	Debugf(message string, args ...interface{})
	Infof(message string, args ...interface{})
	Warnf(message string, args ...interface{})
	Errorf(message string, args ...interface{})

	WithError(err error) Logger

	WithField(key string, value interface{}) Logger
	WithFields(fields Fields) Logger

	WithLevelRank(level Level, rank Rank) Logger
	WithLevelRanks(levelRanks LevelRanks) Logger

	WithLevel(level Level) Logger

	Level() Level
}
