package request

import (
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/tidepool-org/platform/structure"
	structureBase "github.com/tidepool-org/platform/structure/base"
	structureParser "github.com/tidepool-org/platform/structure/parser"
)

type Values struct {
	base   *structureBase.Base
	values *url.Values
	parsed map[string]int
}

func NewValues(values *url.Values) *Values {
	return NewValuesParser(structureBase.New(), values)
}

func NewValuesParser(base *structureBase.Base, values *url.Values) *Values {
	var parsed map[string]int
	if values != nil {
		parsed = make(map[string]int, len(*values))
	}

	return &Values{
		base:   base,
		values: values,
		parsed: parsed,
	}
}

func (v *Values) Error() error {
	return v.base.Error()
}

func (v *Values) Exists() bool {
	return v.values != nil
}

func (v *Values) Parse(objectParsable structure.ObjectParsable) error {
	objectParsable.Parse(v)
	return v.Error()
}

func (v *Values) References() []string {
	if v.values == nil {
		return nil
	}

	references := []string{}
	for reference := range *v.values {
		references = append(references, reference)
	}

	return references
}

func (v *Values) ReferenceExists(reference string) bool {
	if v.values == nil {
		return false
	}

	_, ok := (*v.values)[reference]
	return ok
}

func (v *Values) Bool(reference string) *bool {
	rawValue, ok := v.raw(reference)
	if !ok {
		return nil
	}

	boolValue, err := strconv.ParseBool(rawValue)
	if err != nil {
		v.base.WithReference(reference).ReportError(structureParser.ErrorTypeNotBool(rawValue))
		return nil
	}

	return &boolValue
}

func (v *Values) Float64(reference string) *float64 {
	rawValue, ok := v.raw(reference)
	if !ok {
		return nil
	}

	float64Value, err := strconv.ParseFloat(rawValue, 64)
	if err != nil {
		v.base.WithReference(reference).ReportError(structureParser.ErrorTypeNotFloat64(rawValue))
		return nil
	}

	return &float64Value
}

func (v *Values) Int(reference string) *int {
	rawValue, ok := v.raw(reference)
	if !ok {
		return nil
	}

	int64Value, err := strconv.ParseInt(rawValue, 10, 0)
	if err != nil {
		v.base.WithReference(reference).ReportError(structureParser.ErrorTypeNotInt(rawValue))
		return nil
	}

	intValue := int(int64Value)

	return &intValue
}

func (v *Values) String(reference string) *string {
	rawValue, ok := v.raw(reference)
	if !ok {
		return nil
	}

	return &rawValue
}

func (v *Values) StringArray(reference string) *[]string {
	if v.values == nil {
		return nil
	}

	values, ok := (*v.values)[reference]
	if !ok {
		return nil
	}

	index, _ := v.parsed[reference]
	if index >= len(values) {
		return nil
	}

	v.parsed[reference] = len(values)

	stringArrayValue := []string{}
	for _, value := range values[index:] {
		stringArrayValue = append(stringArrayValue, strings.Split(value, ",")...)
	}

	return &stringArrayValue
}

func (v *Values) Time(reference string, layout string) *time.Time {
	rawValue, ok := v.raw(reference)
	if !ok {
		return nil
	}

	timeValue, err := time.Parse(layout, rawValue)
	if err != nil {
		v.base.WithReference(reference).ReportError(structureParser.ErrorTimeNotParsable(rawValue, layout))
		return nil
	}

	return &timeValue
}

func (v *Values) Object(reference string) *map[string]interface{} {
	rawValue, ok := v.raw(reference)
	if !ok {
		return nil
	}

	v.base.WithReference(reference).ReportError(structureParser.ErrorTypeNotObject(rawValue))
	return nil
}

func (v *Values) Array(reference string) *[]interface{} {
	rawValue, ok := v.raw(reference)
	if !ok {
		return nil
	}

	v.base.WithReference(reference).ReportError(structureParser.ErrorTypeNotArray(rawValue))
	return nil
}

func (v *Values) Interface(reference string) *interface{} {
	rawValue, ok := v.raw(reference)
	if !ok {
		return nil
	}

	var rawInterface interface{} = rawValue
	return &rawInterface
}

func (v *Values) NotParsed() error {
	if v.values == nil {
		return v.Error()
	}

	for reference := range *v.values {
		if v.parsed[reference] < len((*v.values)[reference]) {
			v.base.WithReference(reference).ReportError(structureParser.ErrorNotParsed())
		}
	}

	return v.Error()
}

func (v *Values) WithSource(source structure.Source) structure.ObjectParser {
	return &Values{
		base: v.base.WithSource(source),
	}
}

func (v *Values) WithMeta(meta interface{}) structure.ObjectParser {
	return &Values{
		base: v.base.WithMeta(meta),
	}
}

func (v *Values) WithReferenceObjectParser(reference string) structure.ObjectParser {
	return structureParser.NewObjectParser(v.base.WithReference(reference), v.Object(reference))
}

func (v *Values) WithReferenceArrayParser(reference string) structure.ArrayParser {
	return structureParser.NewArrayParser(v.base.WithReference(reference), v.Array(reference))
}

func (v *Values) raw(reference string) (string, bool) {
	if v.values == nil {
		return "", false
	}

	values, ok := (*v.values)[reference]
	if !ok {
		return "", false
	}

	index, _ := v.parsed[reference]
	if index >= len(values) {
		return "", false
	}

	v.parsed[reference] = index + 1

	return values[index], true
}
