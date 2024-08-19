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
	ForgivingTime(reference string, layout string) *time.Time
	JSON(reference string, target any)

	Object(reference string) *map[string]interface{}
	Array(reference string) *[]interface{}

	Interface(reference string) *interface{}

	NotParsed() error

	WithOrigin(origin Origin) ObjectParser
	WithSource(source Source) ObjectParser
	WithMeta(meta interface{}) ObjectParser
	WithReferenceObjectParser(reference string) ObjectParser
	WithReferenceArrayParser(reference string) ArrayParser
	WithReferenceErrorReporter(reference string) ErrorReporter
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
	WithReferenceErrorReporter(reference int) ErrorReporter
}

// ForgivingTimeString is a helper function added specifically to handle https://tidepool.atlassian.net/browse/BACK-1161
// It should be deprecated once Dexcom fixes their API.
func ForgivingTimeString(stringValue string) (forgivingTime string) {
	if len(stringValue) < 19 {
		forgivingBytes := []byte("0000-01-01T00:00:00")
		for i := range stringValue {
			forgivingBytes[i] = stringValue[i]
		}
		forgivingTime = string(forgivingBytes)
	} else {
		forgivingTime = stringValue
	}
	return forgivingTime
}
