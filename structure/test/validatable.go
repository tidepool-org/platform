package test

import "github.com/tidepool-org/platform/structure"

type Validatable struct {
	ValidateInvocations int
	ValidateInputs      []structure.Validator
}

func NewValidatable() *Validatable {
	return &Validatable{}
}

func (v *Validatable) Validate(validator structure.Validator) {
	v.ValidateInvocations++

	v.ValidateInputs = append(v.ValidateInputs, validator)
}

func (v *Validatable) Expectations() {}
