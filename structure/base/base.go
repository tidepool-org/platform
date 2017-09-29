package base

import (
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
)

type Base struct {
	source structure.Source
	meta   interface{}
	errors error
}

func New() *Base {
	return &Base{}
}

func (b *Base) Source() structure.Source {
	return b.source
}

func (b *Base) Meta() interface{} {
	return b.meta
}

func (b *Base) Error() error {
	return b.errors
}

func (b *Base) ReportError(err error) {
	if err != nil {
		err = errors.WithSource(err, b.source)
		err = errors.WithMeta(err, b.meta)
		b.errors = errors.Append(b.errors, err)
	}
}

func (b *Base) WithSource(source structure.Source) structure.Base {
	return &Base{
		source: source,
		meta:   b.meta,
		errors: b.errors,
	}
}

func (b *Base) WithMeta(meta interface{}) structure.Base {
	return &Base{
		source: b.source,
		meta:   meta,
		errors: b.errors,
	}
}

func (b *Base) WithReference(reference string) structure.Base {
	base := &Base{
		source: b.source,
		meta:   b.meta,
		errors: b.errors,
	}
	if base.source != nil {
		base.source = base.source.WithReference(reference)
	}
	return base
}
