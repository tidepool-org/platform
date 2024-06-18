package parser

import "github.com/tidepool-org/platform/errors"

const (
	ErrorCodeTypeNotBool      = "type-not-bool"
	ErrorCodeTypeNotFloat64   = "type-not-float64"
	ErrorCodeTypeNotInt       = "type-not-int"
	ErrorCodeTypeNotString    = "type-not-string"
	ErrorCodeTypeNotTime      = "type-not-time"
	ErrorCodeTypeNotObject    = "type-not-object"
	ErrorCodeTypeNotArray     = "type-not-array"
	ErrorCodeTypeNotJSON      = "type-not-json"
	ErrorCodeValueNotParsable = "value-not-parsable"
	ErrorCodeNotParsed        = "not-parsed"
)

func ErrorTypeNotBool(value interface{}) error {
	return errors.Preparedf(ErrorCodeTypeNotBool, "type is not bool", "type is not bool, but %T", value)
}

func ErrorTypeNotFloat64(value interface{}) error {
	return errors.Preparedf(ErrorCodeTypeNotFloat64, "type is not float64", "type is not float64, but %T", value)
}

func ErrorTypeNotInt(value interface{}) error {
	return errors.Preparedf(ErrorCodeTypeNotInt, "type is not int", "type is not int, but %T", value)
}

func ErrorTypeNotString(value interface{}) error {
	return errors.Preparedf(ErrorCodeTypeNotString, "type is not string", "type is not string, but %T", value)
}

func ErrorTypeNotJSON(value interface{}, err error) error {
	return errors.Preparedf(ErrorCodeTypeNotJSON, "type is not parsable json", "type is not parsable json, error: %s, original value: %s", err, value)
}

func ErrorTypeNotTime(value interface{}) error {
	return errors.Preparedf(ErrorCodeTypeNotTime, "type is not time", "type is not time, but %T", value)
}

func ErrorTypeNotObject(value interface{}) error {
	return errors.Preparedf(ErrorCodeTypeNotObject, "type is not object", "type is not object, but %T", value)
}

func ErrorTypeNotArray(value interface{}) error {
	return errors.Preparedf(ErrorCodeTypeNotArray, "type is not array", "type is not array, but %T", value)
}

func ErrorValueTimeNotParsable(value string, layout string) error {
	return errors.Preparedf(ErrorCodeValueNotParsable, "value is not a parsable time", "value %q is not a parsable time of format %q", value, layout)
}

func ErrorNotParsed() error {
	return errors.Prepared(ErrorCodeNotParsed, "not parsed", "not parsed")
}
