package request

import (
	"net/http"

)

type ResponseInspector interface {
	// InspectResponse is passed a http.Response to inspect.
	//
	// An inspector must not modify the response. Doing so could impact later
	// inspectors.
	//
	// The state of the response's body is undefined. There could be multiple
	// inspectors before or after any given inspector, so when reading the
	// body, it's probably a good idea to restore it when done.
	InspectResponse(res *http.Response)
}

type HeadersInspector struct {
	Headers http.Header
}

func NewHeadersInspector() *HeadersInspector {
	return &HeadersInspector{}
}

func (h *HeadersInspector) InspectResponse(res *http.Response) {
	h.Headers = res.Header
}
