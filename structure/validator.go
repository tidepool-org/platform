package structure

import (
	"regexp"
	"time"
)

type Validatable interface {
	Validate(validator Validator)
}

type Validator interface {
	LoggerReporter
	OriginReporter
	SourceReporter
	MetaReporter

	ErrorReporter

	Validate(validatable Validatable) error

	Bool(reference string, value *bool) Bool
	Float64(reference string, value *float64) Float64
	Int(reference string, value *int) Int
	String(reference string, value *string) String
	StringArray(reference string, value *[]string) StringArray
	Time(reference string, value *time.Time) Time
	Duration(reference string, value *time.Duration) Duration

	Object(reference string, value *map[string]interface{}) Object
	Array(reference string, value *[]interface{}) Array
	Bytes(reference string, value []byte) Bytes

	WithOrigin(origin Origin) Validator
	WithSource(source Source) Validator
	WithMeta(meta interface{}) Validator
	WithReference(reference string) Validator
}

type BoolUsingFunc func(value bool, errorReporter ErrorReporter)

type Bool interface {
	Exists() Bool
	NotExists() Bool

	True() Bool
	False() Bool

	Using(usingFunc BoolUsingFunc) Bool
}

type Float64UsingFunc func(value float64, errorReporter ErrorReporter)

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

	Using(usingFunc Float64UsingFunc) Float64
}

type IntUsingFunc func(value int, errorReporter ErrorReporter)

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

	Using(usingFunc IntUsingFunc) Int
}

type StringUsingFunc func(value string, errorReporter ErrorReporter)

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

	Using(usingFunc StringUsingFunc) String

	AsTime(layout string) Time

	Email() String

	Alphanumeric() String
	Hexadecimal() String
	UUID() String
}

type StringArrayEachFunc func(stringValidator String)
type StringArrayEachUsingFunc func(value string, errorReporter ErrorReporter)
type StringArrayUsingFunc func(value []string, errorReporter ErrorReporter)

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

	Each(eachFunc StringArrayEachFunc) StringArray

	EachNotEmpty() StringArray

	EachOneOf(allowedValues ...string) StringArray
	EachNotOneOf(disallowedValues ...string) StringArray

	EachMatches(expression *regexp.Regexp) StringArray
	EachNotMatches(expression *regexp.Regexp) StringArray

	EachUsing(eachUsingFunc StringArrayEachUsingFunc) StringArray

	EachUnique() StringArray

	Using(usingFunc StringArrayUsingFunc) StringArray
}

type TimeUsingFunc func(value time.Time, errorReporter ErrorReporter)

type Time interface {
	Exists() Time
	NotExists() Time

	Zero() Time
	NotZero() Time

	After(limit time.Time) Time
	AfterNow(threshold time.Duration) Time
	Before(limit time.Time) Time
	BeforeNow(threshold time.Duration) Time

	Using(usingFunc TimeUsingFunc) Time
}

type DurationUsingFunc func(value time.Duration, errorReporter ErrorReporter)

type Duration interface {
	Exists() Duration
	NotExists() Duration

	EqualTo(value time.Duration) Duration
	NotEqualTo(value time.Duration) Duration

	LessThan(limit time.Duration) Duration
	LessThanOrEqualTo(limit time.Duration) Duration
	GreaterThan(limit time.Duration) Duration
	GreaterThanOrEqualTo(limit time.Duration) Duration
	InRange(lowerLimit time.Duration, upperLimit time.Duration) Duration

	OneOf(allowedValues ...time.Duration) Duration
	NotOneOf(disallowedValues ...time.Duration) Duration

	Using(usingFunc DurationUsingFunc) Duration
}

type ObjectUsingFunc func(value map[string]interface{}, errorReporter ErrorReporter)

type Object interface {
	Exists() Object
	NotExists() Object

	Empty() Object
	NotEmpty() Object

	SizeLessThanOrEqualTo(limit int) Object

	Using(usingFunc ObjectUsingFunc) Object
}

type ArrayUsingFunc func(value []interface{}, errorReporter ErrorReporter)

type Array interface {
	Exists() Array
	NotExists() Array

	Empty() Array
	NotEmpty() Array

	Using(usingFunc ArrayUsingFunc) Array
}

type Bytes interface {
	NotEmpty() Bytes

	LengthLessThanOrEqualTo(limit int) Bytes
}
