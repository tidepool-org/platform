package test

import "net/http"

type ResponseInspector struct {
	InspectResponseInvocations int
	InspectResponseInputs      []*http.Response
	InspectResponseStub        func(res *http.Response) error
	InspectResponseOutputs     []error
	InspectResponseOutput      *error
}

func NewResponseInspector() *ResponseInspector {
	return &ResponseInspector{}
}

func (r *ResponseInspector) InspectResponse(res *http.Response) error {
	r.InspectResponseInvocations++
	r.InspectResponseInputs = append(r.InspectResponseInputs, res)
	if r.InspectResponseStub != nil {
		return r.InspectResponseStub(res)
	}
	if len(r.InspectResponseOutputs) > 0 {
		output := r.InspectResponseOutputs[0]
		r.InspectResponseOutputs = r.InspectResponseOutputs[1:]
		return output
	}
	if r.InspectResponseOutput != nil {
		return *r.InspectResponseOutput
	}
	panic("InspectResponse has no output")
}

func (r *ResponseInspector) AssertOutputsEmpty() {
	if len(r.InspectResponseOutputs) > 0 {
		panic("InspectResponseOutputs is not empty")
	}
}
