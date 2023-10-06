package validator

import (
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/structure/base"
)

// Bytes validates a byte slice.
type Bytes struct {
	base  *base.Base
	value []byte
}

func NewBytes(base *base.Base, value []byte) *Bytes {
	return &Bytes{
		base:  base,
		value: value,
	}
}

func (b *Bytes) NotEmpty() structure.Bytes {
	if len(b.value) == 0 {
		b.base.ReportError(ErrorValueEmpty())
	}
	return b
}
