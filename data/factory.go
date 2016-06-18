package data

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

type Inspector interface {
	GetProperty(key string) *string
	NewMissingPropertyError(key string) error
	NewInvalidPropertyError(key string, value string, allowedValues []string) error
}

type Factory interface {
	New(inspector Inspector) (Datum, error)
	Init(inspector Inspector) (Datum, error)
}
