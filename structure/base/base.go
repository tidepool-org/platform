package base

import (
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
)

type Base struct {
	source       structure.Source
	meta         interface{}
	serializable *errors.Serializable
}

func New() *Base {
	return &Base{
		serializable: &errors.Serializable{},
	}
}

func (b *Base) Error() error {
	return b.serializable.Error
}

func (b *Base) ReportError(err error) {
	if err != nil {
		err = errors.WithSource(err, b.source)
		err = errors.WithMeta(err, b.meta)
		b.serializable.Error = errors.Append(b.serializable.Error, err)
	}
}

func (b *Base) WithSource(source structure.Source) *Base {
	return &Base{
		source:       source,
		meta:         b.meta,
		serializable: b.serializable,
	}
}

func (b *Base) WithMeta(meta interface{}) *Base {
	return &Base{
		source:       b.source,
		meta:         meta,
		serializable: b.serializable,
	}
}

func (b *Base) WithReference(reference string) *Base {
	base := &Base{
		source:       b.source,
		meta:         b.meta,
		serializable: b.serializable,
	}
	if base.source != nil {
		base.source = base.source.WithReference(reference)
	}
	return base
}
