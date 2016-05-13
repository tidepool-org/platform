package data

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

type Normalizer interface {
	AppendDatum(datum Datum)

	NewChildNormalizer(reference interface{}) Normalizer
}
