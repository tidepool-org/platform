package test

import "github.com/tidepool-org/platform/structure"

type Normalizable struct {
	NormalizeInvocations int
	NormalizeInputs      []structure.Normalizer
	NormalizeStub        func(normalizer structure.Normalizer)
}

func NewNormalizable() *Normalizable {
	return &Normalizable{}
}

func (n *Normalizable) Normalize(normalizer structure.Normalizer) {
	n.NormalizeInvocations++

	n.NormalizeInputs = append(n.NormalizeInputs, normalizer)

	if n.NormalizeStub != nil {
		n.NormalizeStub(normalizer)
	}
}

func (n *Normalizable) Expectations() {}
