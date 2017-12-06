package structure

import (
	"regexp"
	"time"
)

type Validatable interface {
	Validate(validator Validator)
}

type Validator interface {
	Error() error
	ReportError(err error)

	Validate(validatable Validatable) error

	Bool(reference string, value *bool) Bool
	Float64(reference string, value *float64) Float64
	Int(reference string, value *int) Int
	String(reference string, value *string) String
	StringArray(reference string, value *[]string) StringArray
	Time(reference string, value *time.Time) Time

	WithSource(source Source) Validator
	WithMeta(meta interface{}) Validator
	WithReference(reference string) Validator
}

type Bool interface {
	Exists() Bool
	NotExists() Bool

	True() Bool
	False() Bool
}

type Float64 interface {
	Exists() Float64
	NotExists() Float64

	EqualTo(value float64) Float64
	NotEqualTo(value float64) Float64

	LessThan(limit float64) Float64
	LessThanOrEqualTo(limit float64) Float64
	GreaterThan(limit float64) Float64
	GreaterThanOrEqualTo(limit float64) Float64
	InRange(lowerLimit float64, upperLimit float64) Float64

	OneOf(allowedValues ...float64) Float64
	NotOneOf(disallowedValues ...float64) Float64
}

type Int interface {
	Exists() Int
	NotExists() Int

	EqualTo(value int) Int
	NotEqualTo(value int) Int

	LessThan(limit int) Int
	LessThanOrEqualTo(limit int) Int
	GreaterThan(limit int) Int
	GreaterThanOrEqualTo(limit int) Int
	InRange(lowerLimit int, upperLimit int) Int

	OneOf(allowedValues ...int) Int
	NotOneOf(disallowedValues ...int) Int
}

type String interface {
	Exists() String
	NotExists() String

	Empty() String
	NotEmpty() String

	EqualTo(value string) String
	NotEqualTo(value string) String

	LengthEqualTo(limit int) String
	LengthNotEqualTo(limit int) String
	LengthLessThan(limit int) String
	LengthLessThanOrEqualTo(limit int) String
	LengthGreaterThan(limit int) String
	LengthGreaterThanOrEqualTo(limit int) String
	LengthInRange(lowerLimit int, upperLimit int) String

	OneOf(allowedValues ...string) String
	NotOneOf(disallowedValues ...string) String

	Matches(expression *regexp.Regexp) String
	NotMatches(expression *regexp.Regexp) String
}

type StringArray interface {
	Exists() StringArray
	NotExists() StringArray

	Empty() StringArray
	NotEmpty() StringArray

	LengthEqualTo(limit int) StringArray
	LengthNotEqualTo(limit int) StringArray
	LengthLessThan(limit int) StringArray
	LengthLessThanOrEqualTo(limit int) StringArray
	LengthGreaterThan(limit int) StringArray
	LengthGreaterThanOrEqualTo(limit int) StringArray
	LengthInRange(lowerLimit int, upperLimit int) StringArray

	EachOneOf(allowedValues ...string) StringArray
	EachNotOneOf(disallowedValues ...string) StringArray

	EachMatches(expression *regexp.Regexp) StringArray
	EachNotMatches(expression *regexp.Regexp) StringArray
}

type Time interface {
	Exists() Time
	NotExists() Time

	Zero() Time
	NotZero() Time

	After(limit time.Time) Time
	AfterNow(threshold time.Duration) Time
	Before(limit time.Time) Time
	BeforeNow(threshold time.Duration) Time
}
