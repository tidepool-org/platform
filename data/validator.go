package data

import (
	"time"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service"
)

type Validator interface {
	Logger() log.Logger

	SetMeta(meta interface{})

	AppendError(reference interface{}, err *service.Error)

	ValidateBoolean(reference interface{}, value *bool) Boolean
	ValidateInteger(reference interface{}, value *int) Integer
	ValidateFloat(reference interface{}, value *float64) Float
	ValidateString(reference interface{}, value *string) String
	ValidateStringArray(reference interface{}, value *[]string) StringArray
	ValidateObject(reference interface{}, value *map[string]interface{}) Object
	ValidateObjectArray(reference interface{}, value *[]map[string]interface{}) ObjectArray

	ValidateStringAsTime(reference interface{}, stringValue *string, timeLayout string) Time

	NewChildValidator(reference interface{}) Validator
}

type Boolean interface {
	Exists() Boolean
	NotExists() Boolean

	True() Boolean
	False() Boolean
}

type Integer interface {
	Exists() Integer
	NotExists() Integer

	EqualTo(value int) Integer
	NotEqualTo(value int) Integer

	LessThan(limit int) Integer
	LessThanOrEqualTo(limit int) Integer
	GreaterThan(limit int) Integer
	GreaterThanOrEqualTo(limit int) Integer
	InRange(lowerLimit int, upperLimit int) Integer

	OneOf(allowedValues []int) Integer
	NotOneOf(disallowedValues []int) Integer
}

type Float interface {
	Exists() Float
	NotExists() Float

	EqualTo(value float64) Float
	NotEqualTo(value float64) Float

	LessThan(limit float64) Float
	LessThanOrEqualTo(limit float64) Float
	GreaterThan(limit float64) Float
	GreaterThanOrEqualTo(limit float64) Float
	InRange(lowerLimit float64, upperLimit float64) Float

	OneOf(allowedValues []float64) Float
	NotOneOf(disallowedValues []float64) Float
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

	OneOf(allowedValues []string) String
	NotOneOf(disallowedValues []string) String
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

	EachOneOf(allowedValues []string) StringArray
	EachNotOneOf(disallowedValues []string) StringArray
}

type Object interface {
	Exists() Object
	NotExists() Object

	Empty() Object
	NotEmpty() Object
}

type ObjectArray interface {
	Exists() ObjectArray
	NotExists() ObjectArray

	Empty() ObjectArray
	NotEmpty() ObjectArray

	LengthEqualTo(limit int) ObjectArray
	LengthNotEqualTo(limit int) ObjectArray
	LengthLessThan(limit int) ObjectArray
	LengthLessThanOrEqualTo(limit int) ObjectArray
	LengthGreaterThan(limit int) ObjectArray
	LengthGreaterThanOrEqualTo(limit int) ObjectArray
	LengthInRange(lowerLimit int, upperLimit int) ObjectArray

	// TODO: SizeLessThanOrEqualTo(limit int) ObjectArray
}

type Time interface {
	Exists() Time
	NotExists() Time

	After(limit time.Time) Time
	AfterNow(threshold time.Duration) Time
	Before(limit time.Time) Time
	BeforeNow(threshold time.Duration) Time
}
