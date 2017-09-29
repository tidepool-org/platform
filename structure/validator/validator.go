package validator

import (
	"time"

	"github.com/tidepool-org/platform/structure"
	structureBase "github.com/tidepool-org/platform/structure/base"
)

type Validator struct {
	structure.Base
}

func New() *Validator {
	return NewValidator(structureBase.New())
}

func NewValidator(base structure.Base) *Validator {
	return &Validator{
		Base: base,
	}
}

func (v *Validator) Validate(validatable structure.Validatable) error {
	validatable.Validate(v)
	return v.Error()
}

func (v *Validator) Validating(reference string, value structure.Validatable) structure.Validating {
	return NewValidating(v.Base.WithReference(reference), value)
}

func (v *Validator) Bool(reference string, value *bool) structure.Bool {
	return NewBool(v.Base.WithReference(reference), value)
}

func (v *Validator) Float64(reference string, value *float64) structure.Float64 {
	return NewFloat64(v.Base.WithReference(reference), value)
}

func (v *Validator) Int(reference string, value *int) structure.Int {
	return NewInt(v.Base.WithReference(reference), value)
}

func (v *Validator) String(reference string, value *string) structure.String {
	return NewString(v.Base.WithReference(reference), value)
}

func (v *Validator) StringArray(reference string, value *[]string) structure.StringArray {
	return NewStringArray(v.Base.WithReference(reference), value)
}

func (v *Validator) Time(reference string, value *time.Time) structure.Time {
	return NewTime(v.Base.WithReference(reference), value)
}

func (v *Validator) WithSource(source structure.Source) *Validator {
	return &Validator{
		Base: v.Base.WithSource(source),
	}
}

func (v *Validator) WithMeta(meta interface{}) structure.Validator {
	return &Validator{
		Base: v.Base.WithMeta(meta),
	}
}

func (v *Validator) WithReference(reference string) structure.Validator {
	return &Validator{
		Base: v.Base.WithReference(reference),
	}
}
