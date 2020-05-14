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

func (n *Normalizer) Origin() structure.Origin {
	return n.normalizer.Origin()
}

func (n *Normalizer) HasSource() bool {
	return n.normalizer.HasSource()
}

func (n *Normalizer) Source() structure.Source {
	return n.normalizer.Source()
}

func (n *Normalizer) HasMeta() bool {
	return n.normalizer.HasMeta()
}

func (n *Normalizer) Meta() interface{} {
	return n.normalizer.Meta()
}

func (n *Normalizer) HasError() bool {
	return n.normalizer.HasError()
}

func (n *Normalizer) Error() error {
	return n.normalizer.Error()
}

func (n *Normalizer) HasWarning() bool {
	return n.normalizer.HasWarning()
}

func (n *Normalizer) Warning() error {
	return n.normalizer.Warning()
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
func (n *Normalizer) WithOrigin(origin structure.Origin) data.Normalizer {
	return &Normalizer{
		normalizer: n.normalizer.WithOrigin(origin),
		data:       n.data,
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
