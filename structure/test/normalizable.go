package test

import (
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/test"
)

type Normalizable struct {
	*test.Mock
	NormalizeInvocations int
	NormalizeInputs      []structure.Normalizer
}

func NewNormalizable() *Normalizable {
	return &Normalizable{
		Mock: test.NewMock(),
	}
}

func (n *Normalizable) Normalize(normalizer structure.Normalizer) {
	n.NormalizeInvocations++

	n.NormalizeInputs = append(n.NormalizeInputs, normalizer)
}

func (n *Normalizable) Expectations() {
	n.Mock.Expectations()
}
