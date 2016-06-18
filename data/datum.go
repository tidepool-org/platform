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

type Datum interface {
	Meta() interface{}

	Parse(parser ObjectParser) error
	Validate(validator Validator) error
	Normalize(normalizer Normalizer) error

	SetUserID(userID string)
	SetGroupID(groupID string)
	SetDatasetID(datasetID string)
	SetActive(active bool)
}
