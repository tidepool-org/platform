package log

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
