package validator

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

type Standard struct {
	context data.Context
}

func NewStandard(context data.Context) *Standard {
	return &Standard{
		context: context,
	}
}

func (s *Standard) Context() data.Context {
	return s.context
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

func (s *Standard) NewChildValidator(reference interface{}) data.Validator {
	return NewStandard(s.context.NewChildContext(reference))
}
