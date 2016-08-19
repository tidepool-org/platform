package version

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

var (
	Base        string
	ShortCommit string
	FullCommit  string
)

func NewDefaultReporter() (Reporter, error) {
	return NewReporter(Base, ShortCommit, FullCommit)
}
