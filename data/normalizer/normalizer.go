package normalizer

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
	structureNormalizer "github.com/tidepool-org/platform/structure/normalizer"
)

type Normalizer struct {
	normalizer structure.Normalizer
	data       *[]data.Datum
}

func New() *Normalizer {
	return &Normalizer{
		normalizer: structureNormalizer.New(),
		data:       &[]data.Datum{},
	}
}

func (n *Normalizer) Error() error {
	return n.normalizer.Error()
}

func (n *Normalizer) ReportError(err error) {
	n.normalizer.ReportError(err)
}

func (n *Normalizer) Normalize(normalizable data.Normalizable) error {
	normalizable.Normalize(n)
	return n.Error()
}

func (n *Normalizer) Data() []data.Datum {
	return *n.data
}

func (n *Normalizer) AddData(data ...data.Datum) {
	for _, datum := range data {
		if datum != nil {
			*n.data = append(*n.data, datum)
		}
	}
}
func (n *Normalizer) WithSource(source structure.Source) data.Normalizer {
	return &Normalizer{
		normalizer: n.normalizer.WithSource(source),
		data:       n.data,
	}
}

func (n *Normalizer) WithMeta(meta interface{}) data.Normalizer {
	return &Normalizer{
		normalizer: n.normalizer.WithMeta(meta),
		data:       n.data,
	}
}

func (n *Normalizer) WithReference(reference string) data.Normalizer {
	return &Normalizer{
		normalizer: n.normalizer.WithReference(reference),
		data:       n.data,
	}
}
