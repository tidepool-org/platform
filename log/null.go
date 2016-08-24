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

type Null struct{}

func NewNull() *Null {
	return &Null{}
}

func (n *Null) Debug(message string)                           {}
func (n *Null) Info(message string)                            {}
func (n *Null) Warn(message string)                            {}
func (n *Null) Error(message string)                           {}
func (n *Null) WithError(err error) Logger                     { return n }
func (n *Null) WithField(key string, value interface{}) Logger { return n }
func (n *Null) WithFields(fields Fields) Logger                { return n }
