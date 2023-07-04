package v1

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	dataService "github.com/tidepool-org/platform/data/service"
)

func GetMetrics(ctx dataService.Context) {
	res := ctx.Response()
	req := ctx.Request()
	w, ok := res.(http.ResponseWriter)
	if !ok {
		rest.Error(res, "unexpected writer", http.StatusInternalServerError)
		return
	}
	promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{DisableCompression: true}).
		ServeHTTP(w, req.Request)
}
