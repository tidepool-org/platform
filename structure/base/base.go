package base

import (
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/structure"
)

type Base struct {
	logger       log.Logger
	origin       structure.Origin
	source       structure.Source
	meta         interface{}
	serializable *errors.Serializable
}

func New(logger log.Logger) *Base {
	return &Base{
		logger:       logger,
		origin:       structure.OriginExternal,
		serializable: &errors.Serializable{},
	}
}

func (b *Base) Logger() log.Logger {
	return b.logger
}

func (b *Base) Origin() structure.Origin {
	return b.origin
}

func (b *Base) HasSource() bool {
	return b.source != nil
}

func (b *Base) Source() structure.Source {
	return b.source
}

func (b *Base) HasMeta() bool {
	return b.meta != nil
}

func (b *Base) Meta() interface{} {
	return b.meta
}

func (b *Base) HasError() bool {
	return b.serializable.Error != nil
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

func (b *Base) WithOrigin(origin structure.Origin) *Base {
	return &Base{
		logger:       b.logger,
		origin:       origin,
		source:       b.source,
		meta:         b.meta,
		serializable: b.serializable,
	}
}

func (b *Base) WithSource(source structure.Source) *Base {
	return &Base{
		logger:       b.logger,
		origin:       b.origin,
		source:       source,
		meta:         b.meta,
		serializable: b.serializable,
	}
}

func (b *Base) WithMeta(meta interface{}) *Base {
	return &Base{
		logger:       b.logger.WithField("meta", meta),
		origin:       b.origin,
		source:       b.source,
		meta:         meta,
		serializable: b.serializable,
	}
}

func (b *Base) WithReference(reference string) *Base {
	base := &Base{
		logger:       b.logger,
		origin:       b.origin,
		source:       b.source,
		meta:         b.meta,
		serializable: b.serializable,
	}
	if base.source != nil {
		base.source = base.source.WithReference(reference)
	}
	return base
}
