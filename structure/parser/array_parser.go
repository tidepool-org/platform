package parser

import (
	"math"
	"strconv"
	"time"

	"github.com/tidepool-org/platform/structure"
	structureBase "github.com/tidepool-org/platform/structure/base"
)

type Array struct {
	base   *structureBase.Base
	array  *[]interface{}
	parsed []bool
}

func NewArray(array *[]interface{}) *Array {
	return NewArrayParser(structureBase.New().WithSource(structure.NewPointerSource()), array)
}

func NewArrayParser(base *structureBase.Base, array *[]interface{}) *Array {
	var parsed []bool
	if array != nil {
		parsed = make([]bool, len(*array))
	}

	return &Array{
		base:   base,
		array:  array,
		parsed: parsed,
	}
}

func (a *Array) Error() error {
	return a.base.Error()
}

func (a *Array) ReportError(err error) {
	a.base.ReportError(err)
}

func (a *Array) Exists() bool {
	return a.array != nil
}

func (a *Array) Parse(arrayParsable structure.ArrayParsable) error {
	arrayParsable.Parse(a)
	return a.Error()
}

func (a *Array) References() []int {
	if a.array == nil {
		return nil
	}

	references := []int{}
	for reference := range *a.array {
		references = append(references, reference)
	}

	return references
}

func (a *Array) ReferenceExists(reference int) bool {
	if a.array == nil {
		return false
	}

	return reference >= 0 && reference < len(*a.array)
}

func (a *Array) Bool(reference int) *bool {
	rawValue, ok := a.raw(reference)
	if !ok {
		return nil
	}

	boolValue, ok := rawValue.(bool)
	if !ok {
		a.base.WithReference(strconv.Itoa(reference)).ReportError(ErrorTypeNotBool(rawValue))
		return nil
	}

	return &boolValue
}

func (a *Array) Float64(reference int) *float64 {
	rawValue, ok := a.raw(reference)
	if !ok {
		return nil
	}

	float64Value, float64ValueOk := rawValue.(float64)
	if !float64ValueOk {
		intValue, intValueOk := rawValue.(int)
		if !intValueOk {
			a.base.WithReference(strconv.Itoa(reference)).ReportError(ErrorTypeNotFloat64(rawValue))
			return nil
		}
		float64Value = float64(intValue)
	}

	return &float64Value
}

func (a *Array) Int(reference int) *int {
	rawValue, ok := a.raw(reference)
	if !ok {
		return nil
	}

	intValue, intValueOk := rawValue.(int)
	if !intValueOk {
		float64Value, float64ValueOk := rawValue.(float64)
		if !float64ValueOk {
			a.base.WithReference(strconv.Itoa(reference)).ReportError(ErrorTypeNotInt(rawValue))
			return nil
		}
		if math.Trunc(float64Value) != float64Value {
			a.base.WithReference(strconv.Itoa(reference)).ReportError(ErrorTypeNotInt(rawValue))
			return nil
		}
		intValue = int(float64Value)
	}

	return &intValue
}

func (a *Array) String(reference int) *string {
	rawValue, ok := a.raw(reference)
	if !ok {
		return nil
	}

	stringValue, ok := rawValue.(string)
	if !ok {
		a.base.WithReference(strconv.Itoa(reference)).ReportError(ErrorTypeNotString(rawValue))
		return nil
	}

	return &stringValue
}

func (a *Array) StringArray(reference int) *[]string {
	rawValue, ok := a.raw(reference)
	if !ok {
		return nil
	}

	stringArrayValue, stringArrayValueOk := rawValue.([]string)
	if !stringArrayValueOk {
		arrayValue, arrayValueOk := rawValue.([]interface{})
		if !arrayValueOk {
			a.base.WithReference(strconv.Itoa(reference)).ReportError(ErrorTypeNotArray(rawValue))
			return nil
		}

		stringArrayValue = []string{}
		parser := NewArrayParser(a.base.WithReference(strconv.Itoa(reference)), &arrayValue)
		for arrayIndex := range arrayValue {
			var stringElement string
			if stringParsed := parser.String(arrayIndex); stringParsed != nil {
				stringElement = *stringParsed
			}
			stringArrayValue = append(stringArrayValue, stringElement)
		}
	}

	return &stringArrayValue
}

func (a *Array) Time(reference int, layout string) *time.Time {
	rawValue, ok := a.raw(reference)
	if !ok {
		return nil
	}

	stringValue, ok := rawValue.(string)
	if !ok {
		a.base.WithReference(strconv.Itoa(reference)).ReportError(ErrorTypeNotTime(rawValue))
		return nil
	}

	timeValue, err := time.Parse(layout, stringValue)
	if err != nil {
		a.base.WithReference(strconv.Itoa(reference)).ReportError(ErrorTimeNotParsable(stringValue, layout))
		return nil
	}

	return &timeValue
}

func (a *Array) Object(reference int) *map[string]interface{} {
	rawValue, ok := a.raw(reference)
	if !ok {
		return nil
	}

	objectValue, ok := rawValue.(map[string]interface{})
	if !ok {
		a.base.WithReference(strconv.Itoa(reference)).ReportError(ErrorTypeNotObject(rawValue))
		return nil
	}

	return &objectValue
}

func (a *Array) Array(reference int) *[]interface{} {
	rawValue, ok := a.raw(reference)
	if !ok {
		return nil
	}

	arrayValue, ok := rawValue.([]interface{})
	if !ok {
		a.base.WithReference(strconv.Itoa(reference)).ReportError(ErrorTypeNotArray(rawValue))
		return nil
	}

	return &arrayValue
}

func (a *Array) Interface(reference int) *interface{} {
	rawValue, ok := a.raw(reference)
	if !ok {
		return nil
	}

	return &rawValue
}

func (a *Array) NotParsed() error {
	if a.array == nil {
		return a.Error()
	}

	for reference := range *a.array {
		if !a.parsed[reference] {
			a.base.WithReference(strconv.Itoa(reference)).ReportError(ErrorNotParsed())
		}
	}

	return a.Error()
}

func (a *Array) WithSource(source structure.Source) structure.ArrayParser {
	return &Array{
		base:   a.base.WithSource(source),
		array:  a.array,
		parsed: a.parsed,
	}
}

func (a *Array) WithMeta(meta interface{}) structure.ArrayParser {
	return &Array{
		base:   a.base.WithMeta(meta),
		array:  a.array,
		parsed: a.parsed,
	}
}

func (a *Array) WithReferenceObjectParser(reference int) structure.ObjectParser {
	return NewObjectParser(a.base.WithReference(strconv.Itoa(reference)), a.Object(reference))
}

func (a *Array) WithReferenceArrayParser(reference int) structure.ArrayParser {
	return NewArrayParser(a.base.WithReference(strconv.Itoa(reference)), a.Array(reference))
}

func (a *Array) raw(reference int) (interface{}, bool) {
	if a.array == nil {
		return nil, false
	}

	if reference < 0 || reference >= len(*a.array) {
		return nil, false
	}

	a.parsed[reference] = true

	rawValue := (*a.array)[reference]
	if rawValue == nil {
		return nil, false
	}

	return rawValue, true
}
