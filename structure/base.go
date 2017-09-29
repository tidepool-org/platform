package structure

type Base interface {
	Structure

	Error() error
	ReportError(err error)

	WithSource(source Source) Base
	WithMeta(meta interface{}) Base
	WithReference(reference string) Base
}
