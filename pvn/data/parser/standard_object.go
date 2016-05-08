package parser

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import "github.com/tidepool-org/platform/pvn/data"

type StandardObject struct {
	context data.Context
	object  *map[string]interface{}
}

func NewStandardObject(context data.Context, object *map[string]interface{}) *StandardObject {
	return &StandardObject{
		context: context,
		object:  object,
	}
}

func (s *StandardObject) Context() data.Context {
	return s.context
}

func (s *StandardObject) Object() *map[string]interface{} {
	return s.object
}

func (s *StandardObject) ParseBoolean(key string) *bool {
	if s.object == nil {
		return nil
	}

	rawValue, ok := (*s.object)[key]
	if !ok {
		return nil
	}

	booleanValue, ok := rawValue.(bool)
	if !ok {
		s.context.AppendError(key, ErrorTypeNotBoolean(rawValue))
		return nil
	}

	return &booleanValue
}

func (s *StandardObject) ParseInteger(key string) *int {
	if s.object == nil {
		return nil
	}

	rawValue, ok := (*s.object)[key]
	if !ok {
		return nil
	}

	integerValue, integerValueOk := rawValue.(int)
	if !integerValueOk {
		floatValue, floatValueOk := rawValue.(float64)
		if !floatValueOk {
			s.context.AppendError(key, ErrorTypeNotInteger(rawValue))
			return nil
		}
		integerValue = int(floatValue)
	}

	return &integerValue
}

func (s *StandardObject) ParseFloat(key string) *float64 {
	if s.object == nil {
		return nil
	}

	rawValue, ok := (*s.object)[key]
	if !ok {
		return nil
	}

	floatValue, floatValueOk := rawValue.(float64)
	if !floatValueOk {
		integerValue, integerValueOk := rawValue.(int)
		if !integerValueOk {
			s.context.AppendError(key, ErrorTypeNotFloat(rawValue))
			return nil
		}
		floatValue = float64(integerValue)
	}

	return &floatValue
}

func (s *StandardObject) ParseString(key string) *string {
	if s.object == nil {
		return nil
	}

	rawValue, ok := (*s.object)[key]
	if !ok {
		return nil
	}

	stringValue, ok := rawValue.(string)
	if !ok {
		s.context.AppendError(key, ErrorTypeNotString(rawValue))
		return nil
	}

	return &stringValue
}

func (s *StandardObject) ParseStringArray(key string) *[]string {
	if s.object == nil {
		return nil
	}

	rawValue, ok := (*s.object)[key]
	if !ok {
		return nil
	}

	stringArrayValue, stringArrayValueOk := rawValue.([]string)
	if !stringArrayValueOk {
		arrayValue, arrayValueOk := rawValue.([]interface{})
		if !arrayValueOk {
			s.context.AppendError(key, ErrorTypeNotArray(rawValue))
			return nil
		}

		stringArrayValue = []string{}
		parser := NewStandardArray(s.context.NewChildContext(key), &arrayValue)
		for index := range arrayValue {
			var stringElement string
			if stringParsed := parser.ParseString(index); stringParsed != nil {
				stringElement = *stringParsed
			}
			stringArrayValue = append(stringArrayValue, stringElement)
		}
	}

	return &stringArrayValue
}

func (s *StandardObject) ParseObject(key string) *map[string]interface{} {
	if s.object == nil {
		return nil
	}

	rawValue, ok := (*s.object)[key]
	if !ok {
		return nil
	}

	objectValue, ok := rawValue.(map[string]interface{})
	if !ok {
		s.context.AppendError(key, ErrorTypeNotObject(rawValue))
		return nil
	}

	return &objectValue
}

func (s *StandardObject) ParseObjectArray(key string) *[]map[string]interface{} {
	if s.object == nil {
		return nil
	}

	rawValue, ok := (*s.object)[key]
	if !ok {
		return nil
	}

	objectArrayValue, objectArrayValueOk := rawValue.([]map[string]interface{})
	if !objectArrayValueOk {
		arrayValue, arrayValueOk := rawValue.([]interface{})
		if !arrayValueOk {
			s.context.AppendError(key, ErrorTypeNotArray(rawValue))
			return nil
		}

		parser := NewStandardArray(s.context.NewChildContext(key), &arrayValue)
		for index := range arrayValue {
			var objectElement map[string]interface{}
			if objectParsed := parser.ParseObject(index); objectParsed != nil {
				objectElement = *objectParsed
			}
			objectArrayValue = append(objectArrayValue, objectElement)
		}
	}

	return &objectArrayValue
}

func (s *StandardObject) ParseInterface(key string) *interface{} {
	if s.object == nil {
		return nil
	}

	rawValue, ok := (*s.object)[key]
	if !ok {
		return nil
	}

	return &rawValue
}

func (s *StandardObject) ParseInterfaceArray(key string) *[]interface{} {
	if s.object == nil {
		return nil
	}

	rawValue, ok := (*s.object)[key]
	if !ok {
		return nil
	}

	arrayValue, ok := rawValue.([]interface{})
	if !ok {
		s.context.AppendError(key, ErrorTypeNotArray(rawValue))
		return nil
	}

	return &arrayValue
}

func (s *StandardObject) NewChildObjectParser(key string) data.ObjectParser {
	return NewStandardObject(s.context.NewChildContext(key), s.ParseObject(key))
}

func (s *StandardObject) NewChildArrayParser(key string) data.ArrayParser {
	return NewStandardArray(s.context.NewChildContext(key), s.ParseInterfaceArray(key))
}
