package validator

import "github.com/tidepool-org/platform/structure"

type ValidatableWithString interface {
	Validate(validator structure.Validator, str *string)
}

type ValidatableWithStringAdapter struct {
	validatableWithString ValidatableWithString
	str                   *string
}

func NewValidatableWithStringAdapter(validatableWithString ValidatableWithString, str *string) *ValidatableWithStringAdapter {
	return &ValidatableWithStringAdapter{
		validatableWithString: validatableWithString,
		str: str,
	}
}

func (v *ValidatableWithStringAdapter) Validate(validator structure.Validator) {
	v.validatableWithString.Validate(validator, v.str)
}

type ValidatableWithStringArray interface {
	Validate(validator structure.Validator, strArray *[]string)
}

type ValidatableWithStringArrayAdapter struct {
	validatableWithStringArray ValidatableWithStringArray
	strArray                   *[]string
}

func NewValidatableWithStringArrayAdapter(validatableWithStringArray ValidatableWithStringArray, strArray *[]string) *ValidatableWithStringArrayAdapter {
	return &ValidatableWithStringArrayAdapter{
		validatableWithStringArray: validatableWithStringArray,
		strArray:                   strArray,
	}
}

func (v *ValidatableWithStringArrayAdapter) Validate(validator structure.Validator) {
	v.validatableWithStringArray.Validate(validator, v.strArray)
}
