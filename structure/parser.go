package structure

import "time"

type ObjectParsable interface {
	Parse(parser ObjectParser)
}

type ObjectParser interface {
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

	Object(reference string) *map[string]interface{}
	Array(reference string) *[]interface{}

	Interface(reference string) *interface{}

	NotParsed() error

	WithOrigin(origin Origin) ObjectParser
	WithSource(source Source) ObjectParser
	WithMeta(meta interface{}) ObjectParser
	WithReferenceObjectParser(reference string) ObjectParser
	WithReferenceArrayParser(reference string) ArrayParser
}

type ArrayParsable interface {
	Parse(parser ArrayParser)
}

type ArrayParser interface {
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

	Object(reference int) *map[string]interface{}
	Array(reference int) *[]interface{}

	Interface(reference int) *interface{}

	NotParsed() error

	WithOrigin(origin Origin) ArrayParser
	WithSource(source Source) ArrayParser
	WithMeta(meta interface{}) ArrayParser
	WithReferenceObjectParser(reference int) ObjectParser
	WithReferenceArrayParser(reference int) ArrayParser
}
