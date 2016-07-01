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

type NullLogger struct{}

func NewNullLogger() *NullLogger {
	return &NullLogger{}
}

func (n *NullLogger) Debug(message string)                           {}
func (n *NullLogger) Info(message string)                            {}
func (n *NullLogger) Warn(message string)                            {}
func (n *NullLogger) Error(message string)                           {}
func (n *NullLogger) WithError(err error) Logger                     { return n }
func (n *NullLogger) WithField(key string, value interface{}) Logger { return n }
func (n *NullLogger) WithFields(fields Fields) Logger                { return n }
