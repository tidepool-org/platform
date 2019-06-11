package test

import (
	"net/http"

	"github.com/onsi/gomega"
)

type RequestMutator struct {
	MutateRequestInvocations int
	MutateRequestInputs      []*http.Request
	MutateRequestOutputs     []error
}

func NewRequestMutator() *RequestMutator {
	return &RequestMutator{}
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
	gomega.Expect(r.MutateRequestOutputs).To(gomega.BeEmpty())
}
