package structure

type Normalizable interface {
	Normalize(normalizer Normalizer)
}

type Normalizer interface {
	Error() error

	Normalize(normalizable Normalizable) error

	WithSource(source Source) Normalizer
	WithMeta(meta interface{}) Normalizer
	WithReference(reference string) Normalizer
}
