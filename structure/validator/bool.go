package validator

import (
	"github.com/tidepool-org/platform/structure"
	structureBase "github.com/tidepool-org/platform/structure/base"
)

type Bool struct {
	base  *structureBase.Base
	value *bool
}

func NewBool(base *structureBase.Base, value *bool) *Bool {
	return &Bool{
		base:  base,
		value: value,
	}
}

func (b *Bool) Exists() structure.Bool {
	if b.value == nil {
		b.base.ReportError(ErrorValueNotExists())
	}
	return b
}

func (b *Bool) NotExists() structure.Bool {
	if b.value != nil {
		b.base.ReportError(ErrorValueExists())
	}
	return b
}

func (b *Bool) True() structure.Bool {
	if b.value != nil {
		if !*b.value {
			b.base.ReportError(ErrorValueNotTrue())
		}
	}
	return b
}

func (b *Bool) False() structure.Bool {
	if b.value != nil {
		if *b.value {
			b.base.ReportError(ErrorValueNotFalse())
		}
	}
	return b
}
