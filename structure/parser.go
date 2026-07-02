package structure

import "time"

type ObjectParsable interface {
	Parse(parser ObjectParser)
}

type ObjectParser interface {
	LoggerReporter
	OriginReporter
	SourceReporter
	MetaReporter

	ErrorReporter

	Exists() bool

	Parse(objectParsable ObjectParsable) error

	References() []string
	ReferenceExists(reference string) bool

	Bool(reference string) *bool
	Float64(reference string) *float64
	Int(reference string) *int
	String(reference string) *string
	StringArray(reference string) *[]string
	Time(reference string, layout string) *time.Time
	JSON(reference string, target any)

	Object(reference string) *map[string]any
	Array(reference string) *[]any

	Interface(reference string) *any

	NotParsed() map[string]any
	ReportNotParsed()

	WithOrigin(origin Origin) ObjectParser
	WithSource(source Source) ObjectParser
	WithMeta(meta any) ObjectParser
	WithReferenceObjectParser(reference string) ObjectParser
	WithReferenceArrayParser(reference string) ArrayParser
	WithReferenceErrorReporter(reference string) ErrorReporter
}

type ArrayParsable interface {
	Parse(parser ArrayParser)
}

type ArrayParser interface {
	LoggerReporter
	OriginReporter
	SourceReporter
	MetaReporter

	ErrorReporter

	Exists() bool

	Parse(arrayParsable ArrayParsable) error

	References() []int
	ReferenceExists(reference int) bool

	Bool(reference int) *bool
	Float64(reference int) *float64
	Int(reference int) *int
	String(reference int) *string
	StringArray(reference int) *[]string
	Time(reference int, layout string) *time.Time

	Object(reference int) *map[string]any
	Array(reference int) *[]any

	Interface(reference int) *any

	NotParsed() map[string]any
	ReportNotParsed()

	WithOrigin(origin Origin) ArrayParser
	WithSource(source Source) ArrayParser
	WithMeta(meta any) ArrayParser
	WithReferenceObjectParser(reference int) ObjectParser
	WithReferenceArrayParser(reference int) ArrayParser
	WithReferenceErrorReporter(reference int) ErrorReporter
}
