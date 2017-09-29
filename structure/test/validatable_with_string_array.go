package test

import (
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/test"
)

type ValidatableWithStringArrayInput struct {
	Validator   structure.Validator
	StringArray *[]string
}

type ValidatableWithStringArray struct {
	*test.Mock
	ValidateInvocations int
	ValidateInputs      []ValidatableWithStringArrayInput
}

func NewValidatableWithStringArray() *ValidatableWithStringArray {
	return &ValidatableWithStringArray{
		Mock: test.NewMock(),
	}
}

func (v *ValidatableWithStringArray) Validate(validator structure.Validator, strArray *[]string) {
	v.ValidateInvocations++

	v.ValidateInputs = append(v.ValidateInputs, ValidatableWithStringArrayInput{Validator: validator, StringArray: strArray})
}

func (v *ValidatableWithStringArray) Expectations() {
	v.Mock.Expectations()
}
