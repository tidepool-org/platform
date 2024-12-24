package v1

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/tidepool-org/platform/log"
)

type DebuggingTransport struct {
	Logger log.Logger
}

func (s *DebuggingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	bytes, _ := httputil.DumpRequestOut(r, true)

	resp, err := http.DefaultTransport.RoundTrip(r)
	// err is returned after dumping the response

	respBytes, _ := httputil.DumpResponse(resp, true)
	bytes = append(bytes, respBytes...)

	s.Logger.Debug(fmt.Sprintf("%s\n", string(bytes)))

	return resp, err
}
