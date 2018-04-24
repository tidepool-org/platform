package structure

type Normalizable interface {
	Normalize(normalizer Normalizer)
}

type Normalizer interface {
	OriginReporter
	SourceReporter
	MetaReporter

	ErrorReporter

	Normalize(normalizable Normalizable) error

	WithOrigin(origin Origin) Normalizer
	WithSource(source Source) Normalizer
	WithMeta(meta interface{}) Normalizer
	WithReference(reference string) Normalizer
}
