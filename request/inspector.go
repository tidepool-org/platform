package request

import (
	"net/http"

	"github.com/tidepool-org/platform/errors"
)

type ResponseInspector interface {
	// InspectResponse is passed a response to inspect.
	//
	// An inspector must not modify the response. Doing so could impact later
	// inspectors.
	//
	// The state of the response's body is undefined. There could be multiple
	// inspectors before or after any given inspector, so when reading the
	// body, it's probably a good idea to restore it when done.
	//
	// Any error returned will simply be logged. This might be removed in the
	// future, so implementors are encouraged to use their own logging method.
	InspectResponse(res *http.Response) error
}

type HeadersInspector struct {
	Headers http.Header
}

func NewHeadersInspector() *HeadersInspector {
	return &HeadersInspector{}
}

func (h *HeadersInspector) InspectResponse(res *http.Response) error {
	if res == nil {
		return errors.New("response is missing")
	}

	h.Headers = res.Header
	return nil
}
