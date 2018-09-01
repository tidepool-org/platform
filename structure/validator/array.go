package validator

import (
	"github.com/tidepool-org/platform/structure"
	structureBase "github.com/tidepool-org/platform/structure/base"
)

type Array struct {
	base  *structureBase.Base
	value *[]interface{}
}

func NewArray(base *structureBase.Base, value *[]interface{}) *Array {
	return &Array{
		base:  base,
		value: value,
	}
}

func (a *Array) Exists() structure.Array {
	if a.value == nil {
		a.base.ReportError(ErrorValueNotExists())
	}
	return a
}

func (a *Array) NotExists() structure.Array {
	if a.value != nil {
		a.base.ReportError(ErrorValueExists())
	}
	return a
}

func (a *Array) Empty() structure.Array {
	if a.value != nil {
		if len(*a.value) > 0 {
			a.base.ReportError(ErrorValueNotEmpty())
		}
	}
	return a
}

func (a *Array) NotEmpty() structure.Array {
	if a.value != nil {
		if len(*a.value) == 0 {
			a.base.ReportError(ErrorValueEmpty())
		}
	}
	return a
}

func (a *Array) Using(usingFunc structure.ArrayUsingFunc) structure.Array {
	if a.value != nil {
		if usingFunc != nil {
			usingFunc(*a.value, a.base)
		}
	}
	return a
}
