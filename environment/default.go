package environment

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

func NewDefaultReporter(prefix string) (Reporter, error) {
	return NewReporter(GetValue("ENV", prefix), prefix)
}
