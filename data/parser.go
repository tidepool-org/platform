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

import "github.com/tidepool-org/platform/service"

type ObjectParser interface {
	SetMeta(meta interface{})
	AppendError(reference interface{}, err *service.Error)

	Object() *map[string]interface{}

	ParseBoolean(key string) *bool
	ParseInteger(key string) *int
	ParseFloat(key string) *float64
	ParseString(key string) *string
	ParseStringArray(key string) *[]string
	ParseObject(key string) *map[string]interface{}
	ParseObjectArray(key string) *[]map[string]interface{}
	ParseInterface(key string) *interface{}
	ParseInterfaceArray(key string) *[]interface{}

	NewChildObjectParser(key string) ObjectParser
	NewChildArrayParser(key string) ArrayParser
}

type ArrayParser interface {
	SetMeta(meta interface{})
	AppendError(reference interface{}, err *service.Error)

	Array() *[]interface{}

	ParseBoolean(index int) *bool
	ParseInteger(index int) *int
	ParseFloat(index int) *float64
	ParseString(index int) *string
	ParseStringArray(index int) *[]string
	ParseObject(index int) *map[string]interface{}
	ParseObjectArray(index int) *[]map[string]interface{}
	ParseInterface(index int) *interface{}
	ParseInterfaceArray(index int) *[]interface{}

	NewChildObjectParser(index int) ObjectParser
	NewChildArrayParser(index int) ArrayParser
}
