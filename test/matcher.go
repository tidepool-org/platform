package test

import (
	"fmt"
	"reflect"
	"time"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	gomegaFormat "github.com/onsi/gomega/format"
	gomegaGstruct "github.com/onsi/gomega/gstruct"
	gomegaMatchers "github.com/onsi/gomega/matchers"
	gomegaTypes "github.com/onsi/gomega/types"
)

func MatchTime(datum *time.Time) gomegaTypes.GomegaMatcher {
	if datum == nil {
		return gomega.BeNil()
	}
	return gomegaGstruct.PointTo(gomega.BeTemporally("==", *datum))
}

func MatchArray(elements ...interface{}) gomegaTypes.GomegaMatcher {
	return &MatchArrayMatcher{
		Elements: elements,
	}
}

type MatchArrayMatcher struct {
	Elements []interface{}
}

func (m *MatchArrayMatcher) Match(actual interface{}) (bool, error) {
	if !isArrayOrSlice(actual) {
		return false, fmt.Errorf("MatchArray matcher expects an array/slice.  Got:\n%s", gomegaFormat.Object(actual, 1))
	}

	elements := m.Elements
	if len(elements) == 1 && isArrayOrSlice(elements[0]) {
		element := reflect.ValueOf(elements[0])
		elements = []interface{}{}
		for index := 0; index < element.Len(); index++ {
			elements = append(elements, element.Index(index).Interface())
		}
	}

	matchers := []gomegaTypes.GomegaMatcher{}
	for _, element := range elements {
		matcher, isMatcher := element.(gomegaTypes.GomegaMatcher)
		if !isMatcher {
			matcher = &gomegaMatchers.EqualMatcher{Expected: element}
		}
		matchers = append(matchers, matcher)
	}

	value := reflect.ValueOf(actual)
	values := []interface{}{}
	for index := 0; index < value.Len(); index++ {
		values = append(values, value.Index(index).Interface())
	}

	if len(values) != len(matchers) {
		return false, nil
	}

	for index := 0; index < len(values); index++ {
		if success, err := matchers[index].Match(values[index]); err != nil || !success {
			return success, err
		}
	}

	return true, nil
}

func (m *MatchArrayMatcher) FailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "to match array", m.Elements)
}

func (m *MatchArrayMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to match array", m.Elements)
}

func isArrayOrSlice(actual interface{}) bool {
	if actual == nil {
		return false
	}
	switch reflect.TypeOf(actual).Kind() {
	case reflect.Array, reflect.Slice:
		return true
	default:
		return false
	}
}
