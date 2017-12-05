package normalizer

import (
	"github.com/tidepool-org/platform/structure"
	structureBase "github.com/tidepool-org/platform/structure/base"
)

type Normalizer struct {
	base *structureBase.Base
}

func New() *Normalizer {
	return NewNormalizer(structureBase.New().WithSource(structure.NewPointerSource()))
}

func NewNormalizer(base *structureBase.Base) *Normalizer {
	return &Normalizer{
		base: base,
	}
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
