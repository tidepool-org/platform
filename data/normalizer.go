package data

import "github.com/tidepool-org/platform/structure"

type Normalizable interface {
	Normalize(normalizer Normalizer)
}

type Normalizer interface {
	structure.OriginReporter
	structure.SourceReporter
	structure.MetaReporter

	structure.ErrorReporter

	Normalize(normalizable Normalizable) error

	AddData(data ...Datum)

	WithOrigin(origin structure.Origin) Normalizer
	WithSource(source structure.Source) Normalizer
	WithMeta(meta interface{}) Normalizer
	WithReference(reference string) Normalizer
}
