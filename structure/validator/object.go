package validator

import (
	"github.com/tidepool-org/platform/structure"
	structureBase "github.com/tidepool-org/platform/structure/base"
)

type Object struct {
	base  *structureBase.Base
	value *map[string]interface{}
}

func NewObject(base *structureBase.Base, value *map[string]interface{}) *Object {
	return &Object{
		base:  base,
		value: value,
	}
}

func (o *Object) Exists() structure.Object {
	if o.value == nil {
		o.base.ReportError(ErrorValueNotExists())
	}
	return o
}

func (o *Object) NotExists() structure.Object {
	if o.value != nil {
		o.base.ReportError(ErrorValueExists())
	}
	return o
}

func (o *Object) Empty() structure.Object {
	if o.value != nil {
		if len(*o.value) != 0 {
			o.base.ReportError(ErrorValueNotEmpty())
		}
	}
	return o
}

func (o *Object) NotEmpty() structure.Object {
	if o.value != nil {
		if len(*o.value) == 0 {
			o.base.ReportError(ErrorValueEmpty())
		}
	}
	return o
}

func (o *Object) Using(using func(value map[string]interface{}, errorReporter structure.ErrorReporter)) structure.Object {
	if o.value != nil {
		if using != nil {
			using(*o.value, o.base)
		}
	}
	return o
}
