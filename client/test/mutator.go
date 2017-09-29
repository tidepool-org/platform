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
	r.MutateInvocations++

	r.MutateInputs = append(r.MutateInputs, request)

	gomega.Expect(r.MutateOutputs).ToNot(gomega.BeEmpty())

	output := r.MutateOutputs[0]
	r.MutateOutputs = r.MutateOutputs[1:]
	return output
}

func (m *Mutator) Expectations() {
	m.Mock.Expectations()
	gomega.Expect(r.MutateOutputs).To(gomega.BeEmpty())
}
