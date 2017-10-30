package test

import (
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/test"
)

type Validatable struct {
	*test.Mock
	ValidateInvocations int
	ValidateInputs      []structure.Validator
}

func NewValidatable() *Validatable {
	return &Validatable{
		Mock: test.NewMock(),
	}
}

func (v *Validatable) Validate(validator structure.Validator) {
	v.ValidateInvocations++

	v.ValidateInputs = append(v.ValidateInputs, validator)
}

func (v *Validatable) Expectations() {
	v.Mock.Expectations()
}
