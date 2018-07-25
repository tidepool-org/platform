package test

import (
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/test"
)

type ValidatableWithIntInput struct {
	Validator structure.Validator
	Int       *int
}

type ValidatableWithInt struct {
	*test.Mock
	ValidateInvocations int
	ValidateInputs      []ValidatableWithIntInput
}

func NewValidatableWithInt() *ValidatableWithInt {
	return &ValidatableWithInt{
		Mock: test.NewMock(),
	}
}

func (v *ValidatableWithInt) Validate(validator structure.Validator, i *int) {
	v.ValidateInvocations++

	v.ValidateInputs = append(v.ValidateInputs, ValidatableWithIntInput{Validator: validator, Int: i})
}

func (v *ValidatableWithInt) Expectations() {
	v.Mock.Expectations()
}
