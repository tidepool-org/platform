package data

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

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
	ValidateInterface(reference interface{}, value *interface{}) Interface
	ValidateInterfaceArray(reference interface{}, value *[]interface{}) InterfaceArray

	ValidateStringAsTime(reference interface{}, stringValue *string, timeLayout string) Time

	ValidateStringAsBloodGlucoseUnits(reference interface{}, stringValue *string) BloodGlucoseUnits
	ValidateFloatAsBloodGlucoseValue(reference interface{}, floatValue *float64) BloodGlucoseValue

	NewChildValidator(reference interface{}) Validator
}

type Boolean interface {
	Exists() Boolean

	True() Boolean
	False() Boolean
}

type Integer interface {
	Exists() Integer

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
}

type ObjectArray interface {
	Exists() ObjectArray

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

type Interface interface {
	Exists() Interface
}

type InterfaceArray interface {
	Exists() InterfaceArray

	NotEmpty() InterfaceArray

	LengthEqualTo(limit int) InterfaceArray
	LengthNotEqualTo(limit int) InterfaceArray
	LengthLessThan(limit int) InterfaceArray
	LengthLessThanOrEqualTo(limit int) InterfaceArray
	LengthGreaterThan(limit int) InterfaceArray
	LengthGreaterThanOrEqualTo(limit int) InterfaceArray
	LengthInRange(lowerLimit int, upperLimit int) InterfaceArray

	// TODO: SizeLessThanOrEqualTo(limit int) InterfaceArray
}

type Time interface {
	Exists() Time

	After(limit time.Time) Time
	AfterNow() Time
	Before(limit time.Time) Time
	BeforeNow() Time
}

type BloodGlucoseUnits interface {
	Exists() BloodGlucoseUnits
}

type BloodGlucoseValue interface {
	Exists() BloodGlucoseValue

	InRange(lowerLimit float64, upperLimit float64) BloodGlucoseValue
	InRangeForUnits(units *string) BloodGlucoseValue
}
