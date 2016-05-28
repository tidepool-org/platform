package parser

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
	"math"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/service"
)

type StandardObject struct {
	context data.Context
	object  *map[string]interface{}
}

func NewStandardObject(context data.Context, object *map[string]interface{}) (*StandardObject, error) {
	if context == nil {
		return nil, app.Error("parser", "context is missing")
	}

	return &StandardObject{
		context: context,
		object:  object,
	}, nil
}

func (s *StandardObject) SetMeta(meta interface{}) {
	s.context.SetMeta(meta)
}

func (s *StandardObject) AppendError(reference interface{}, err *service.Error) {
	s.context.AppendError(reference, err)
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
		if math.Trunc(floatValue) != floatValue {
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
		parser, _ := NewStandardArray(s.context.NewChildContext(key), &arrayValue)
		for arrayIndex := range arrayValue {
			var stringElement string
			if stringParsed := parser.ParseString(arrayIndex); stringParsed != nil {
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

		parser, _ := NewStandardArray(s.context.NewChildContext(key), &arrayValue)
		for arrayIndex := range arrayValue {
			var objectElement map[string]interface{}
			if objectParsed := parser.ParseObject(arrayIndex); objectParsed != nil {
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
	standardObject, _ := NewStandardObject(s.context.NewChildContext(key), s.ParseObject(key))
	return standardObject
}

func (s *StandardObject) NewChildArrayParser(key string) data.ArrayParser {
	standardArray, _ := NewStandardArray(s.context.NewChildContext(key), s.ParseInterfaceArray(key))
	return standardArray
}
