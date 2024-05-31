package validator

import (
	"time"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/structure"
	structureBase "github.com/tidepool-org/platform/structure/base"
)

type Validator struct {
	base *structureBase.Base
}

func New(logger log.Logger) *Validator {
	return NewValidator(structureBase.New(logger).WithSource(structure.NewPointerSource()))
}

func NewValidator(base *structureBase.Base) *Validator {
	return &Validator{
		base: base,
	}
}

func (v *Validator) Logger() log.Logger {
	return v.base.Logger()
}

func (v *Validator) Origin() structure.Origin {
	return v.base.Origin()
}

func (v *Validator) HasSource() bool {
	return v.base.HasSource()
}

func (v *Validator) Source() structure.Source {
	return v.base.Source()
}

func (v *Validator) HasMeta() bool {
	return v.base.HasMeta()
}

func (v *Validator) Meta() interface{} {
	return v.base.Meta()
}

func (v *Validator) HasError() bool {
	return v.base.HasError()
}

func (v *Validator) Error() error {
	return v.base.Error()
}

func (v *Validator) ReportError(err error) {
	v.base.ReportError(err)
}

func (v *Validator) Validate(validatable structure.Validatable) error {
	validatable.Validate(v)
	return v.Error()
}

func (v *Validator) Bool(reference string, value *bool) structure.Bool {
	return NewBool(v.base.WithReference(reference), value)
}

func (v *Validator) Float64(reference string, value *float64) structure.Float64 {
	return NewFloat64(v.base.WithReference(reference), value)
}

func (v *Validator) Int(reference string, value *int) structure.Int {
	return NewInt(v.base.WithReference(reference), value)
}

func (v *Validator) String(reference string, value *string) structure.String {
	return NewString(v.base.WithReference(reference), value)
}

func (v *Validator) StringArray(reference string, value *[]string) structure.StringArray {
	return NewStringArray(v.base.WithReference(reference), value)
}

func (v *Validator) Time(reference string, value *time.Time) structure.Time {
	return NewTime(v.base.WithReference(reference), value)
}

func (v *Validator) Duration(reference string, value *time.Duration) structure.Duration {
	return NewDuration(v.base.WithReference(reference), value)
}

func (v *Validator) Object(reference string, value *map[string]interface{}) structure.Object {
	return NewObject(v.base.WithReference(reference), value)
}

func (v *Validator) Array(reference string, value *[]interface{}) structure.Array {
	return NewArray(v.base.WithReference(reference), value)
}

func (v *Validator) Bytes(reference string, value []byte) structure.Bytes {
	return NewBytes(v.base.WithReference(reference), value)
}

func (v *Validator) WithOrigin(origin structure.Origin) structure.Validator {
	return &Validator{
		base: v.base.WithOrigin(origin),
	}
}

func (v *Validator) WithSource(source structure.Source) structure.Validator {
	return &Validator{
		base: v.base.WithSource(source),
	}
}

func (v *Validator) WithMeta(meta interface{}) structure.Validator {
	return &Validator{
		base: v.base.WithMeta(meta),
	}
}

func (v *Validator) WithReference(reference string) structure.Validator {
	return &Validator{
		base: v.base.WithReference(reference),
	}
}
