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

type Normalizer interface {
	SetMeta(meta interface{})

	AppendDatum(datum Datum)

	NewChildNormalizer(reference interface{}) Normalizer

	NormalizeBloodGlucose(reference interface{}, units *string) BloodGlucoseNormalizer
}

type BloodGlucoseNormalizer interface {
	NormalizeValue(value *float64) *float64
	NormalizeUnits() *string
	NormalizeUnitsAndValue(value *float64) (*string, *float64)
}
