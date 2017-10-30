package validator

import (
	"github.com/tidepool-org/platform/structure"
	structureBase "github.com/tidepool-org/platform/structure/base"
)

type Validating struct {
	base  *structureBase.Base
	value structure.Validatable
}

func NewValidating(base *structureBase.Base, value structure.Validatable) *Validating {
	return &Validating{
		base:  base,
		value: value,
	}
}

func (v *Validating) Exists() structure.Validating {
	if v.value == nil {
		v.base.ReportError(ErrorValueNotExists())
	}
	return v
}

func (v *Validating) NotExists() structure.Validating {
	if v.value != nil {
		v.base.ReportError(ErrorValueExists())
	}
	return v
}

func (v *Validating) Validate() structure.Validating {
	if v.value != nil {
		v.value.Validate(NewValidator(v.base))
	}
	return v
}
