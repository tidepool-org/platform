package version

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

var (
	_base        string
	_shortCommit string
	_fullCommit  string
	_current     Version
)

func init() {
	_current = NewStandard(_base, _shortCommit, _fullCommit)
}
