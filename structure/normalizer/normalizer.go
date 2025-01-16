package normalizer

import (
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/structure"
	structureBase "github.com/tidepool-org/platform/structure/base"
)

type Normalizer struct {
	base *structureBase.Base
}

func New(logger log.Logger) *Normalizer {
	return NewNormalizer(structureBase.New(logger).WithSource(structure.NewPointerSource()))
}

func NewNormalizer(base *structureBase.Base) *Normalizer {
	return &Normalizer{
		base: base,
	}
}

func (n *Normalizer) Logger() log.Logger {
	return n.base.Logger()
}

func (n *Normalizer) Origin() structure.Origin {
	return n.base.Origin()
}

func (n *Normalizer) HasSource() bool {
	return n.base.HasSource()
}

func (n *Normalizer) Source() structure.Source {
	return n.base.Source()
}

func (n *Normalizer) HasMeta() bool {
	return n.base.HasMeta()
}

func (n *Normalizer) Meta() interface{} {
	return n.base.Meta()
}

func (n *Normalizer) HasError() bool {
	return n.base.HasError()
}

func (n *Normalizer) Error() error {
	return n.base.Error()
}

func (n *Normalizer) ReportError(err error) {
	n.base.ReportError(err)
}

func (n *Normalizer) Normalize(normalizable structure.Normalizable) error {
	normalizable.Normalize(n)
	return n.Error()
}

func (n *Normalizer) WithOrigin(origin structure.Origin) structure.Normalizer {
	return &Normalizer{
		base: n.base.WithOrigin(origin),
	}
}

func (n *Normalizer) WithSource(source structure.Source) structure.Normalizer {
	return &Normalizer{
		base: n.base.WithSource(source),
	}
}

func (n *Normalizer) WithMeta(meta interface{}) structure.Normalizer {
	return &Normalizer{
		base: n.base.WithMeta(meta),
	}
}

func (n *Normalizer) WithReference(reference string) structure.Normalizer {
	return &Normalizer{
		base: n.base.WithReference(reference),
	}
}
