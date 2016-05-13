package data

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

type ObjectParser interface {
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
