package test

import "github.com/tidepool-org/platform/structure"

type ValidatableWithStringInput struct {
	Validator structure.Validator
	String    *string
}

type ValidatableWithString struct {
	ValidateInvocations int
	ValidateInputs      []ValidatableWithStringInput
}

func NewValidatableWithString() *ValidatableWithString {
	return &ValidatableWithString{}
}

func (v *ValidatableWithString) Validate(validator structure.Validator, str *string) {
	v.ValidateInvocations++

	v.ValidateInputs = append(v.ValidateInputs, ValidatableWithStringInput{Validator: validator, String: str})
}

func (v *ValidatableWithString) Expectations() {}
