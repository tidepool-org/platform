package test

import (
	"fmt"
	"reflect"
	"regexp"
	"sync"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
)

type Serializer struct {
	SerializedFields []log.Fields
	mux              sync.Mutex
}

func NewSerializer() *Serializer {
	return &Serializer{}
}

func (s *Serializer) Serialize(fields log.Fields) error {
	if fields == nil {
		return errors.New("fields are missing")
	}
	s.mux.Lock()
	s.SerializedFields = append(s.SerializedFields, fields)
	s.mux.Unlock()
	return nil
}

func (s *Serializer) AssertLog(level log.Level, message string, containsFields ...log.Fields) {
	s.assertContainsFields(append(containsFields, log.Fields{"level": level, "message": message}), nil)
}

func (s *Serializer) AssertLogExpression(level log.Level, messageExpression *regexp.Regexp, containsFields ...log.Fields) {
	s.assertContainsFields(append(containsFields, log.Fields{"level": level}), func(serializedFields log.Fields) bool {
		if value, found := serializedFields["message"]; found {
			if message, ok := value.(string); ok {
				return messageExpression.MatchString(message)
			}
		}
		return false
	})
}

func (s *Serializer) AssertDebug(message string, containsFields ...log.Fields) {
	s.AssertLog(log.DebugLevel, message, containsFields...)
}

func (s *Serializer) AssertDebugExpression(messageExpression *regexp.Regexp, containsFields ...log.Fields) {
	s.AssertLogExpression(log.DebugLevel, messageExpression, containsFields...)
}

func (s *Serializer) AssertInfo(message string, containsFields ...log.Fields) {
	s.AssertLog(log.InfoLevel, message, containsFields...)
}

func (s *Serializer) AssertInfoExpression(messageExpression *regexp.Regexp, containsFields ...log.Fields) {
	s.AssertLogExpression(log.InfoLevel, messageExpression, containsFields...)
}

func (s *Serializer) AssertWarn(message string, containsFields ...log.Fields) {
	s.AssertLog(log.WarnLevel, message, containsFields...)
}

func (s *Serializer) AssertWarnExpression(messageExpression *regexp.Regexp, containsFields ...log.Fields) {
	s.AssertLogExpression(log.WarnLevel, messageExpression, containsFields...)
}

func (s *Serializer) AssertError(message string, containsFields ...log.Fields) {
	s.AssertLog(log.ErrorLevel, message, containsFields...)
}

func (s *Serializer) AssertErrorExpression(messageExpression *regexp.Regexp, containsFields ...log.Fields) {
	s.AssertLogExpression(log.ErrorLevel, messageExpression, containsFields...)
}

func (s *Serializer) assertContainsFields(containsFields []log.Fields, matcher func(serializedFields log.Fields) bool) {
	joinedContainsFields := s.joinContainsFields(containsFields)
	s.mux.Lock()
	defer s.mux.Unlock()
	for _, serializedFields := range s.SerializedFields {
		if s.serializedFieldsContainsFields(serializedFields, joinedContainsFields) && (matcher == nil || matcher(serializedFields)) {
			return
		}
	}
	panic(fmt.Sprintf("logger does not contain specified message and fields"))
}

func (s *Serializer) joinContainsFields(containsFields []log.Fields) log.Fields {
	joinedFields := log.Fields{}
	for _, fields := range containsFields {
		for key, value := range fields {
			if value != nil {
				if joinedValue, found := joinedFields[key]; !found {
					joinedFields[key] = value
				} else if !reflect.DeepEqual(joinedValue, value) {
					panic(fmt.Sprintf("duplicate log field found with key %q", key))
				}
			}
		}
	}
	return joinedFields
}

func (s *Serializer) serializedFieldsContainsFields(serializedFields log.Fields, containsFields log.Fields) bool {
	for containsKey, containsValue := range containsFields {
		if serializedValue, found := serializedFields[containsKey]; !found || !reflect.DeepEqual(serializedValue, containsValue) {
			return false
		}
	}
	return true
}
