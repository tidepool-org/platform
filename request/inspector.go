package request

import (
	"net/http"

	"github.com/tidepool-org/platform/log"
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
	logger  log.Logger
}

func NewHeadersInspector(logger log.Logger) *HeadersInspector {
	return &HeadersInspector{logger: logger}
}

func (h *HeadersInspector) InspectResponse(res *http.Response) {
	if res == nil {
		if h.logger != nil {
			h.logger.Warnf("response is missing")
		}
		return
	}

	h.Headers = res.Header
}
