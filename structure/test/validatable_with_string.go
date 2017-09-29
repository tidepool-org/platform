package test

import (
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/test"
)

type ValidatableWithStringInput struct {
	Validator structure.Validator
	String    *string
}

type ValidatableWithString struct {
	*test.Mock
	ValidateInvocations int
	ValidateInputs      []ValidatableWithStringInput
}

func NewValidatableWithString() *ValidatableWithString {
	return &ValidatableWithString{
		Mock: test.NewMock(),
	}
}

func (v *ValidatableWithString) Validate(validator structure.Validator, str *string) {
	v.ValidateInvocations++

	v.ValidateInputs = append(v.ValidateInputs, ValidatableWithStringInput{Validator: validator, String: str})
}

func (v *ValidatableWithString) Expectations() {
	v.Mock.Expectations()
}
