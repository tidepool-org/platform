package normalizer

import (
	"github.com/tidepool-org/platform/structure"
	structureBase "github.com/tidepool-org/platform/structure/base"
)

type Normalizer struct {
	structure.Base
}

func New() *Normalizer {
	return NewNormalizer(structureBase.New())
}

func NewNormalizer(base structure.Base) *Normalizer {
	return &Normalizer{
		Base: base,
	}
}

func (n *Normalizer) Normalize(normalizable structure.Normalizable) error {
	normalizable.Normalize(n)
	return n.Error()
}

func (n *Normalizer) WithSource(source structure.Source) *Normalizer {
	return &Normalizer{
		Base: n.Base.WithSource(source),
	}
}

func (n *Normalizer) WithMeta(meta interface{}) structure.Normalizer {
	return &Normalizer{
		Base: n.Base.WithMeta(meta),
	}
}

func (n *Normalizer) WithReference(reference string) structure.Normalizer {
	return &Normalizer{
		Base: n.Base.WithReference(reference),
	}
}
