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

	NormalizeBloodGlucose(units *string) BloodGlucoseNormalizer

	NewChildNormalizer(reference interface{}) Normalizer
}

type BloodGlucoseNormalizer interface {
	Units() *string
	Value(value *float64) *float64
	UnitsAndValue(value *float64) (*string, *float64)
}
