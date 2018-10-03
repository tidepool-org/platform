package request

import (
	"net/http"

	"github.com/tidepool-org/platform/errors"
)

type ResponseInspector interface {
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
