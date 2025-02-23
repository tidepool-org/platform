package parser

import (
	"encoding/json"
	"math"
	"sort"
	"time"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/structure"
	structureBase "github.com/tidepool-org/platform/structure/base"
)

type Object struct {
	base   *structureBase.Base
	object *map[string]interface{}
	parsed map[string]bool
}

func NewObject(logger log.Logger, object *map[string]interface{}) *Object {
	return NewObjectParser(structureBase.New(logger).WithSource(structure.NewPointerSource()), object)
}

func NewObjectParser(base *structureBase.Base, object *map[string]interface{}) *Object {
	var parsed map[string]bool
	if object != nil {
		parsed = make(map[string]bool, len(*object))
	}

	return &Object{
		base:   base,
		object: object,
		parsed: parsed,
	}
}

func (o *Object) Logger() log.Logger {
	return o.base.Logger()
}

func (o *Object) Origin() structure.Origin {
	return o.base.Origin()
}

func (o *Object) HasSource() bool {
	return o.base.HasSource()
}

func (o *Object) Source() structure.Source {
	return o.base.Source()
}

func (o *Object) HasMeta() bool {
	return o.base.HasMeta()
}

func (o *Object) Meta() interface{} {
	return o.base.Meta()
}

func (o *Object) HasError() bool {
	return o.base.HasError()
}

func (o *Object) Error() error {
	return o.base.Error()
}

func (o *Object) ReportError(err error) {
	o.base.ReportError(err)
}

func (o *Object) Exists() bool {
	return o.object != nil
}

func (o *Object) Parse(objectParsable structure.ObjectParsable) error {
	objectParsable.Parse(o)
	return o.Error()
}

func (o *Object) References() []string {
	if o.object == nil {
		return nil
	}

	references := []string{}
	for reference := range *o.object {
		references = append(references, reference)
	}

	return references
}

func (o *Object) ReferenceExists(reference string) bool {
	if o.object == nil {
		return false
	}

	_, ok := (*o.object)[reference]
	return ok
}

func (o *Object) Bool(reference string) *bool {
	rawValue, ok := o.raw(reference)
	if !ok {
		return nil
	}

	boolValue, ok := rawValue.(bool)
	if !ok {
		o.base.WithReference(reference).ReportError(ErrorTypeNotBool(rawValue))
		return nil
	}

	return &boolValue
}

func (o *Object) Float64(reference string) *float64 {
	rawValue, ok := o.raw(reference)
	if !ok {
		return nil
	}

	float64Value, float64ValueOk := rawValue.(float64)
	if !float64ValueOk {
		intValue, intValueOk := rawValue.(int)
		if !intValueOk {
			o.base.WithReference(reference).ReportError(ErrorTypeNotFloat64(rawValue))
			return nil
		}
		float64Value = float64(intValue)
	}

	return &float64Value
}

func (o *Object) Int(reference string) *int {
	rawValue, ok := o.raw(reference)
	if !ok {
		return nil
	}

	intValue, intValueOk := rawValue.(int)
	if !intValueOk {
		float64Value, float64ValueOk := rawValue.(float64)
		if !float64ValueOk {
			o.base.WithReference(reference).ReportError(ErrorTypeNotInt(rawValue))
			return nil
		}
		if math.Trunc(float64Value) != float64Value {
			o.base.WithReference(reference).ReportError(ErrorTypeNotInt(rawValue))
			return nil
		}
		intValue = int(float64Value)
	}

	return &intValue
}

func (o *Object) String(reference string) *string {
	rawValue, ok := o.raw(reference)
	if !ok {
		return nil
	}

	stringValue, ok := rawValue.(string)
	if !ok {
		o.base.WithReference(reference).ReportError(ErrorTypeNotString(rawValue))
		return nil
	}

	return &stringValue
}

func (o *Object) StringArray(reference string) *[]string {
	rawValue, ok := o.raw(reference)
	if !ok {
		return nil
	}

	stringArrayValue, stringArrayValueOk := rawValue.([]string)
	if !stringArrayValueOk {
		arrayValue, arrayValueOk := rawValue.([]interface{})
		if !arrayValueOk {
			o.base.WithReference(reference).ReportError(ErrorTypeNotArray(rawValue))
			return nil
		}

		stringArrayValue = []string{}
		parser := NewArrayParser(o.base.WithReference(reference), &arrayValue)
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

func (o *Object) Time(reference string, layout string) *time.Time {
	rawValue, ok := o.raw(reference)
	if !ok {
		return nil
	}

	timeValue, timeValueOk := rawValue.(time.Time)
	if !timeValueOk {
		if timeProvider, timeProviderOk := rawValue.(timeProvider); timeProviderOk {
			timeValue = timeProvider.Time()
		} else {
			stringValue, stringValueOk := rawValue.(string)
			if !stringValueOk {
				o.base.WithReference(reference).ReportError(ErrorTypeNotTime(rawValue))
				return nil
			}

			var err error
			timeValue, err = time.Parse(layout, stringValue)
			if err != nil {
				o.base.WithReference(reference).ReportError(ErrorValueTimeNotParsable(stringValue, layout))
				return nil
			}
		}
	}

	return &timeValue
}

func (o *Object) Object(reference string) *map[string]interface{} {
	rawValue, ok := o.raw(reference)
	if !ok {
		return nil
	}

	objectValue, ok := rawValue.(map[string]interface{})
	if !ok {
		o.base.WithReference(reference).ReportError(ErrorTypeNotObject(rawValue))
		return nil
	}

	return &objectValue
}

func (o *Object) Array(reference string) *[]interface{} {
	rawValue, ok := o.raw(reference)
	if !ok {
		return nil
	}

	arrayValue, ok := rawValue.([]interface{})
	if !ok {
		o.base.WithReference(reference).ReportError(ErrorTypeNotArray(rawValue))
		return nil
	}

	return &arrayValue
}

func (o *Object) JSON(reference string, target any) {
	rawValue, ok := o.raw(reference)
	if !ok {
		return
	}

	stringValue, ok := rawValue.(string)
	if !ok {
		o.base.WithReference(reference).ReportError(ErrorTypeNotString(rawValue))
		return
	}

	if stringValue == "" {
		return
	}

	err := json.Unmarshal([]byte(stringValue), target)
	if err != nil {
		o.base.WithReference(reference).ReportError(ErrorTypeNotJSON(rawValue, err))
	}
}

func (o *Object) Interface(reference string) *interface{} {
	rawValue, ok := o.raw(reference)
	if !ok {
		return nil
	}

	return &rawValue
}

func (o *Object) NotParsed() error {
	if o.object == nil {
		return o.Error()
	}

	var references []string
	for reference := range *o.object {
		if !o.parsed[reference] {
			references = append(references, reference)
		}
	}

	if len(references) > 0 {
		sort.Strings(references)
		for _, reference := range references {
			o.base.WithReference(reference).ReportError(ErrorNotParsed())
		}
	}

	return o.Error()
}

func (o *Object) WithOrigin(origin structure.Origin) structure.ObjectParser {
	return &Object{
		base:   o.base.WithOrigin(origin),
		object: o.object,
		parsed: o.parsed,
	}
}

func (o *Object) WithSource(source structure.Source) structure.ObjectParser {
	return &Object{
		base:   o.base.WithSource(source),
		object: o.object,
		parsed: o.parsed,
	}
}

func (o *Object) WithMeta(meta interface{}) structure.ObjectParser {
	return &Object{
		base:   o.base.WithMeta(meta),
		object: o.object,
		parsed: o.parsed,
	}
}

func (o *Object) WithReferenceObjectParser(reference string) structure.ObjectParser {
	return NewObjectParser(o.base.WithReference(reference), o.Object(reference))
}

func (o *Object) WithReferenceArrayParser(reference string) structure.ArrayParser {
	return NewArrayParser(o.base.WithReference(reference), o.Array(reference))
}

func (o *Object) WithReferenceErrorReporter(reference string) structure.ErrorReporter {
	return o.base.WithReference(reference)
}

func (o *Object) raw(reference string) (interface{}, bool) {
	if o.object == nil {
		return nil, false
	}

	o.parsed[reference] = true

	rawValue, ok := (*o.object)[reference]
	if !ok || rawValue == nil {
		return nil, false
	}

	return rawValue, true
}

type timeProvider interface {
	Time() time.Time
}
