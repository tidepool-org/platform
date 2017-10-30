package test

import (
	"net/http"

	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/test"
)

type Mutator struct {
	*test.Mock
	MutateInvocations int
	MutateInputs      []*http.Request
	MutateOutputs     []error
}

func NewMutator() *Mutator {
	return &Mutator{
		Mock: test.NewMock(),
	}
}

func (m *Mutator) Mutate(request *http.Request) error {
	m.MutateInvocations++

	m.MutateInputs = append(m.MutateInputs, request)

	gomega.Expect(m.MutateOutputs).ToNot(gomega.BeEmpty())

	output := m.MutateOutputs[0]
	m.MutateOutputs = m.MutateOutputs[1:]
	return output
}

func (m *Mutator) Expectations() {
	m.Mock.Expectations()
	gomega.Expect(m.MutateOutputs).To(gomega.BeEmpty())
}
