package test

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/test"
)

type Normalizable struct {
	*test.Mock
	NormalizeInvocations int
	NormalizeInputs      []data.Normalizer
	NormalizeStub        func(normalizer data.Normalizer)
}

func NewNormalizable() *Normalizable {
	return &Normalizable{
		Mock: test.NewMock(),
	}
}

func (n *Normalizable) Normalize(normalizer data.Normalizer) {
	n.NormalizeInvocations++

	n.NormalizeInputs = append(n.NormalizeInputs, normalizer)

	if n.NormalizeStub != nil {
		n.NormalizeStub(normalizer)
	}
}

func (n *Normalizable) Expectations() {
	n.Mock.Expectations()
}
