package test

import (
	"net/http"

	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/test"
)

type RequestMutator struct {
	*test.Mock
	MutateRequestInvocations int
	MutateRequestInputs      []*http.Request
	MutateRequestOutputs     []error
}

func NewRequestMutator() *RequestMutator {
	return &RequestMutator{
		Mock: test.NewMock(),
	}
}

func (r *RequestMutator) MutateRequest(request *http.Request) error {
	r.MutateRequestInvocations++

	r.MutateRequestInputs = append(r.MutateRequestInputs, request)

	gomega.Expect(r.MutateRequestOutputs).ToNot(gomega.BeEmpty())

	output := r.MutateRequestOutputs[0]
	r.MutateRequestOutputs = r.MutateRequestOutputs[1:]
	return output
}

func (r *RequestMutator) Expectations() {
	r.Mock.Expectations()
	gomega.Expect(r.MutateRequestOutputs).To(gomega.BeEmpty())
}
