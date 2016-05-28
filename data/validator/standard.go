package validator

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
	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/service"
)

type Standard struct {
	context data.Context
}

func NewStandard(context data.Context) (*Standard, error) {
	if context == nil {
		return nil, app.Error("validator", "context is missing")
	}

	return &Standard{
		context: context,
	}, nil
}

func (s *Standard) SetMeta(meta interface{}) {
	s.context.SetMeta(meta)
}

func (s *Standard) AppendError(reference interface{}, err *service.Error) {
	s.context.AppendError(reference, err)
}

func (s *Standard) ValidateBoolean(reference interface{}, value *bool) data.Boolean {
	return NewStandardBoolean(s.context, reference, value)
}

func (s *Standard) ValidateInteger(reference interface{}, value *int) data.Integer {
	return NewStandardInteger(s.context, reference, value)
}

func (s *Standard) ValidateFloat(reference interface{}, value *float64) data.Float {
	return NewStandardFloat(s.context, reference, value)
}

func (s *Standard) ValidateString(reference interface{}, value *string) data.String {
	return NewStandardString(s.context, reference, value)
}

func (s *Standard) ValidateStringArray(reference interface{}, value *[]string) data.StringArray {
	return NewStandardStringArray(s.context, reference, value)
}

func (s *Standard) ValidateObject(reference interface{}, value *map[string]interface{}) data.Object {
	return NewStandardObject(s.context, reference, value)
}

func (s *Standard) ValidateObjectArray(reference interface{}, value *[]map[string]interface{}) data.ObjectArray {
	return NewStandardObjectArray(s.context, reference, value)
}

func (s *Standard) ValidateInterface(reference interface{}, value *interface{}) data.Interface {
	return NewStandardInterface(s.context, reference, value)
}

func (s *Standard) ValidateInterfaceArray(reference interface{}, value *[]interface{}) data.InterfaceArray {
	return NewStandardInterfaceArray(s.context, reference, value)
}

func (s *Standard) ValidateStringAsTime(reference interface{}, stringValue *string, timeLayout string) data.Time {
	return NewStandardStringAsTime(s.context, reference, stringValue, timeLayout)
}

func (s *Standard) ValidateStringAsBloodGlucoseUnits(reference interface{}, stringValue *string) data.BloodGlucoseUnits {
	return NewStandardStringAsBloodGlucoseUnits(s.context, reference, stringValue)
}

func (s *Standard) ValidateFloatAsBloodGlucoseValue(reference interface{}, floatValue *float64) data.BloodGlucoseValue {
	return NewStandardFloatAsBloodGlucoseValue(s.context, reference, floatValue)
}

func (s *Standard) NewChildValidator(reference interface{}) data.Validator {
	standard, _ := NewStandard(s.context.NewChildContext(reference))
	return standard
}
