package structure

type Normalizable interface {
	Normalize(normalizer Normalizer)
}

type Normalizer interface {
	Structure

	Normalize(normalizable Normalizable) error

	WithMeta(meta interface{}) Normalizer
	WithReference(reference string) Normalizer
}
