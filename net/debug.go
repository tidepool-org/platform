package net

import (
	"net/http"
	"net/http/httputil"

	"github.com/tidepool-org/platform/log"
)

type DebugTransport struct {
	Logger log.Logger
}

func (d *DebugTransport) RoundTrip(request *http.Request) (*http.Response, error) {

	requestBytes, _ := httputil.DumpRequestOut(request, true)
	response, err := http.DefaultTransport.RoundTrip(request)
	responseBytes, _ := httputil.DumpResponse(response, true)

	d.Logger.Debug(string(append(requestBytes, responseBytes...)))
	return response, err
}
