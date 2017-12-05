package data

import "github.com/tidepool-org/platform/structure"

type Normalizable interface {
	Normalize(normalizer Normalizer)
}

type Normalizer interface {
	Error() error
	ReportError(err error)

	Normalize(normalizable Normalizable) error

	AddData(data ...Datum)

	WithSource(source structure.Source) Normalizer
	WithMeta(meta interface{}) Normalizer
	WithReference(reference string) Normalizer
}
