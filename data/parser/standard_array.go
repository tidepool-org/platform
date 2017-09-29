package parser

import (
	"math"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service"
)

type StandardArray struct {
	context         data.Context
	factory         data.Factory
	array           *[]interface{}
	parsed          []bool
	notParsedPolicy NotParsedPolicy
}

func NewStandardArray(context data.Context, factory data.Factory, array *[]interface{}, notParsedPolicy NotParsedPolicy) (*StandardArray, error) {
	if context == nil {
		return nil, errors.New("context is missing")
	}
	if factory == nil {
		return nil, errors.New("factory is missing")
	}

	var parsed []bool
	if array != nil {
		parsed = make([]bool, len(*array))
	}

	return &StandardArray{
		context:         context,
		factory:         factory,
		array:           array,
		parsed:          parsed,
		notParsedPolicy: notParsedPolicy,
	}, nil
}

func (s *StandardArray) Logger() log.Logger {
	return s.context.Logger()
}

func (s *StandardArray) SetMeta(meta interface{}) {
	s.context.SetMeta(meta)
}

func (s *StandardArray) AppendError(reference interface{}, err *service.Error) {
	s.context.AppendError(reference, err)
}

func (s *StandardArray) Array() *[]interface{} {
	return s.array
}

func (s *StandardArray) ParseBoolean(index int) *bool {
	if s.array == nil {
		return nil
	}

	if index < 0 || index >= len(*s.array) {
		return nil
	}

	s.parsed[index] = true

	rawValue := (*s.array)[index]

	booleanValue, ok := rawValue.(bool)
	if !ok {
		s.AppendError(index, service.ErrorTypeNotBoolean(rawValue))
		return nil
	}

	return &booleanValue
}

func (s *StandardArray) ParseInteger(index int) *int {
	if s.array == nil {
		return nil
	}

	if index < 0 || index >= len(*s.array) {
		return nil
	}

	s.parsed[index] = true

	rawValue := (*s.array)[index]

	integerValue, integerValueOk := rawValue.(int)
	if !integerValueOk {
		floatValue, floatValueOk := rawValue.(float64)
		if !floatValueOk {
			s.AppendError(index, service.ErrorTypeNotInteger(rawValue))
			return nil
		}
		if math.Trunc(floatValue) != floatValue {
			s.AppendError(index, service.ErrorTypeNotInteger(rawValue))
			return nil
		}
		integerValue = int(floatValue)
	}

	return &integerValue
}

func (s *StandardArray) ParseFloat(index int) *float64 {
	if s.array == nil {
		return nil
	}

	if index < 0 || index >= len(*s.array) {
		return nil
	}

	s.parsed[index] = true

	rawValue := (*s.array)[index]

	floatValue, floatValueOk := rawValue.(float64)
	if !floatValueOk {
		integerValue, integerValueOk := rawValue.(int)
		if !integerValueOk {
			s.AppendError(index, service.ErrorTypeNotFloat(rawValue))
			return nil
		}
		floatValue = float64(integerValue)
	}

	return &floatValue
}

func (s *StandardArray) ParseString(index int) *string {
	if s.array == nil {
		return nil
	}

	if index < 0 || index >= len(*s.array) {
		return nil
	}

	s.parsed[index] = true

	rawValue := (*s.array)[index]

	stringValue, ok := rawValue.(string)
	if !ok {
		s.AppendError(index, service.ErrorTypeNotString(rawValue))
		return nil
	}

	return &stringValue
}

func (s *StandardArray) ParseStringArray(index int) *[]string {
	if s.array == nil {
		return nil
	}

	if index < 0 || index >= len(*s.array) {
		return nil
	}

	s.parsed[index] = true

	rawValue := (*s.array)[index]

	stringArrayValue, stringArrayValueOk := rawValue.([]string)
	if !stringArrayValueOk {
		arrayValue, arrayValueOk := rawValue.([]interface{})
		if !arrayValueOk {
			s.AppendError(index, service.ErrorTypeNotArray(rawValue))
			return nil
		}

		stringArrayValue = []string{}
		parser, _ := NewStandardArray(s.context.NewChildContext(index), s.factory, &arrayValue, IgnoreNotParsed)
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

func (s *StandardArray) ParseObject(index int) *map[string]interface{} {
	if s.array == nil {
		return nil
	}

	if index < 0 || index >= len(*s.array) {
		return nil
	}

	s.parsed[index] = true

	rawValue := (*s.array)[index]

	objectValue, ok := rawValue.(map[string]interface{})
	if !ok {
		s.AppendError(index, service.ErrorTypeNotObject(rawValue))
		return nil
	}

	return &objectValue
}

func (s *StandardArray) ParseObjectArray(index int) *[]map[string]interface{} {
	if s.array == nil {
		return nil
	}

	if index < 0 || index >= len(*s.array) {
		return nil
	}

	s.parsed[index] = true

	rawValue := (*s.array)[index]

	objectArrayValue, objectArrayValueOk := rawValue.([]map[string]interface{})
	if !objectArrayValueOk {
		arrayValue, arrayValueOk := rawValue.([]interface{})
		if !arrayValueOk {
			s.AppendError(index, service.ErrorTypeNotArray(rawValue))
			return nil
		}

		parser, _ := NewStandardArray(s.context.NewChildContext(index), s.factory, &arrayValue, IgnoreNotParsed)
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

func (s *StandardArray) ParseInterface(index int) *interface{} {
	if s.array == nil {
		return nil
	}

	if index < 0 || index >= len(*s.array) {
		return nil
	}

	s.parsed[index] = true

	rawValue := (*s.array)[index]

	return &rawValue
}

func (s *StandardArray) ParseInterfaceArray(index int) *[]interface{} {
	if s.array == nil {
		return nil
	}

	if index < 0 || index >= len(*s.array) {
		return nil
	}

	s.parsed[index] = true

	rawValue := (*s.array)[index]

	arrayValue, ok := rawValue.([]interface{})
	if !ok {
		s.AppendError(index, service.ErrorTypeNotArray(rawValue))
		return nil
	}

	return &arrayValue
}

func (s *StandardArray) ParseDatum(index int) *data.Datum {
	parser := s.NewChildObjectParser(index)

	datum, err := ParseDatum(parser, s.factory)
	if err != nil || datum == nil {
		return nil
	}

	if datum != nil && *datum != nil {
		parser.ProcessNotParsed()
	}

	return datum
}

func (s *StandardArray) ParseDatumArray(index int) *[]data.Datum {
	if s.array == nil {
		return nil
	}

	if index < 0 || index >= len(*s.array) {
		return nil
	}

	s.parsed[index] = true

	rawValue := (*s.array)[index]

	arrayValue, arrayValueOk := rawValue.([]interface{})
	if !arrayValueOk {
		s.AppendError(index, service.ErrorTypeNotArray(rawValue))
		return nil
	}

	parser, err := NewStandardArray(s.context.NewChildContext(index), s.factory, &arrayValue, IgnoreNotParsed)
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

func (s *StandardArray) ProcessNotParsed() {
	if s.array == nil {
		return
	}

	switch s.notParsedPolicy {
	case WarnLoggerNotParsed:
		for index := range *s.array {
			if !s.parsed[index] {
				s.Logger().WithField("reference", s.context.ResolveReference(index)).Warn("Reference not parsed")
			}
		}
	case AppendErrorNotParsed:
		for index := range *s.array {
			if !s.parsed[index] {
				s.AppendError(index, ErrorNotParsed())
			}
		}
	}
}

func (s *StandardArray) NewChildObjectParser(index int) data.ObjectParser {
	standardObject, _ := NewStandardObject(s.context.NewChildContext(index), s.factory, s.ParseObject(index), s.notParsedPolicy)
	return standardObject
}

func (s *StandardArray) NewChildArrayParser(index int) data.ArrayParser {
	standardArray, _ := NewStandardArray(s.context.NewChildContext(index), s.factory, s.ParseInterfaceArray(index), s.notParsedPolicy)
	return standardArray
}
