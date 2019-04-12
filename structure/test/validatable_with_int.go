package test

import "github.com/tidepool-org/platform/structure"

type ValidatableWithIntInput struct {
	Validator structure.Validator
	Int       *int
}

type ValidatableWithInt struct {
	ValidateInvocations int
	ValidateInputs      []ValidatableWithIntInput
}

func NewValidatableWithInt() *ValidatableWithInt {
	return &ValidatableWithInt{}
}

func (v *ValidatableWithInt) Validate(validator structure.Validator, i *int) {
	v.ValidateInvocations++

	v.ValidateInputs = append(v.ValidateInputs, ValidatableWithIntInput{Validator: validator, Int: i})
}

func (v *ValidatableWithInt) Expectations() {}
