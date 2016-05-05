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

// NOTE: Use RootLogger sparingly. Prefer using a derived logger based upon
// the RootLogger created at application start. If you aren't sure, don't use
// it and ask!

func RootLogger() Logger {
	return _logger
}
