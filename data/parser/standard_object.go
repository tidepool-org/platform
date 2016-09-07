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
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service"
)

type StandardObject struct {
	context         data.Context
	factory         data.Factory
	object          *map[string]interface{}
	parsed          map[string]bool
	notParsedPolicy NotParsedPolicy
}

func NewStandardObject(context data.Context, factory data.Factory, object *map[string]interface{}, notParsedPolicy NotParsedPolicy) (*StandardObject, error) {
	if context == nil {
		return nil, app.Error("parser", "context is missing")
	}
	if factory == nil {
		return nil, app.Error("parser", "factory is missing")
	}

	var parsed map[string]bool
	if object != nil {
		parsed = make(map[string]bool, len(*object))
	}

	return &StandardObject{
		context:         context,
		factory:         factory,
		object:          object,
		parsed:          parsed,
		notParsedPolicy: notParsedPolicy,
	}, nil
}

func (s *StandardObject) Logger() log.Logger {
	return s.context.Logger()
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

	s.parsed[key] = true

	rawValue, ok := (*s.object)[key]
	if !ok {
		return nil
	}

	booleanValue, ok := rawValue.(bool)
	if !ok {
		s.AppendError(key, service.ErrorTypeNotBoolean(rawValue))
		return nil
	}

	return &booleanValue
}

func (s *StandardObject) ParseInteger(key string) *int {
	if s.object == nil {
		return nil
	}

	s.parsed[key] = true

	rawValue, ok := (*s.object)[key]
	if !ok {
		return nil
	}

	integerValue, integerValueOk := rawValue.(int)
	if !integerValueOk {
		floatValue, floatValueOk := rawValue.(float64)
		if !floatValueOk {
			s.AppendError(key, service.ErrorTypeNotInteger(rawValue))
			return nil
		}
		if math.Trunc(floatValue) != floatValue {
			s.AppendError(key, service.ErrorTypeNotInteger(rawValue))
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

	s.parsed[key] = true

	rawValue, ok := (*s.object)[key]
	if !ok {
		return nil
	}

	floatValue, floatValueOk := rawValue.(float64)
	if !floatValueOk {
		integerValue, integerValueOk := rawValue.(int)
		if !integerValueOk {
			s.AppendError(key, service.ErrorTypeNotFloat(rawValue))
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

	s.parsed[key] = true

	rawValue, ok := (*s.object)[key]
	if !ok {
		return nil
	}

	stringValue, ok := rawValue.(string)
	if !ok {
		s.AppendError(key, service.ErrorTypeNotString(rawValue))
		return nil
	}

	return &stringValue
}

func (s *StandardObject) ParseStringArray(key string) *[]string {
	if s.object == nil {
		return nil
	}

	s.parsed[key] = true

	rawValue, ok := (*s.object)[key]
	if !ok {
		return nil
	}

	stringArrayValue, stringArrayValueOk := rawValue.([]string)
	if !stringArrayValueOk {
		arrayValue, arrayValueOk := rawValue.([]interface{})
		if !arrayValueOk {
			s.AppendError(key, service.ErrorTypeNotArray(rawValue))
			return nil
		}

		stringArrayValue = []string{}
		parser, _ := NewStandardArray(s.context.NewChildContext(key), s.factory, &arrayValue, IgnoreNotParsed)
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

	s.parsed[key] = true

	rawValue, ok := (*s.object)[key]
	if !ok {
		return nil
	}

	objectValue, ok := rawValue.(map[string]interface{})
	if !ok {
		s.AppendError(key, service.ErrorTypeNotObject(rawValue))
		return nil
	}

	return &objectValue
}

func (s *StandardObject) ParseObjectArray(key string) *[]map[string]interface{} {
	if s.object == nil {
		return nil
	}

	s.parsed[key] = true

	rawValue, ok := (*s.object)[key]
	if !ok {
		return nil
	}

	objectArrayValue, objectArrayValueOk := rawValue.([]map[string]interface{})
	if !objectArrayValueOk {
		arrayValue, arrayValueOk := rawValue.([]interface{})
		if !arrayValueOk {
			s.AppendError(key, service.ErrorTypeNotArray(rawValue))
			return nil
		}

		parser, _ := NewStandardArray(s.context.NewChildContext(key), s.factory, &arrayValue, IgnoreNotParsed)
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

	s.parsed[key] = true

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

	s.parsed[key] = true

	rawValue, ok := (*s.object)[key]
	if !ok {
		return nil
	}

	arrayValue, ok := rawValue.([]interface{})
	if !ok {
		s.AppendError(key, service.ErrorTypeNotArray(rawValue))
		return nil
	}

	return &arrayValue
}

func (s *StandardObject) ParseDatum(key string) *data.Datum {
	parser := s.NewChildObjectParser(key)

	datum, err := ParseDatum(parser, s.factory)
	if err != nil || datum == nil {
		return nil
	}

	if datum != nil && *datum != nil {
		parser.ProcessNotParsed()
	}

	return datum
}

func (s *StandardObject) ParseDatumArray(key string) *[]data.Datum {
	if s.object == nil {
		return nil
	}

	s.parsed[key] = true

	rawValue, ok := (*s.object)[key]
	if !ok {
		return nil
	}

	arrayValue, arrayValueOk := rawValue.([]interface{})
	if !arrayValueOk {
		s.AppendError(key, service.ErrorTypeNotArray(rawValue))
		return nil
	}

	parser, err := NewStandardArray(s.context.NewChildContext(key), s.factory, &arrayValue, IgnoreNotParsed)
	if err != nil {
		return nil
	}

	datumArray, err := ParseDatumArray(parser)
	if err != nil {
		return nil
	}

	if datumArray != nil && *datumArray != nil {
		parser.ProcessNotParsed()
	}

	return datumArray
}

func (s *StandardObject) ProcessNotParsed() {
	if s.object == nil {
		return
	}

	switch s.notParsedPolicy {
	case WarnLoggerNotParsed:
		for key := range *s.object {
			if !s.parsed[key] {
				s.Logger().WithField("reference", s.context.ResolveReference(key)).Warn("Reference not parsed")
			}
		}
	case AppendErrorNotParsed:
		for key := range *s.object {
			if !s.parsed[key] {
				s.AppendError(key, ErrorNotParsed())
			}
		}
	}
}

func (s *StandardObject) NewChildObjectParser(key string) data.ObjectParser {
	standardObject, _ := NewStandardObject(s.context.NewChildContext(key), s.factory, s.ParseObject(key), s.notParsedPolicy)
	return standardObject
}

func (s *StandardObject) NewChildArrayParser(key string) data.ArrayParser {
	standardArray, _ := NewStandardArray(s.context.NewChildContext(key), s.factory, s.ParseInterfaceArray(key), s.notParsedPolicy)
	return standardArray
}
