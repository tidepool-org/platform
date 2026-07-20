package v1

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	dataService "github.com/tidepool-org/platform/data/service"
)

func MetricsRoutes() []dataService.Route {
	return []dataService.Route{
		dataService.Get("/v1/metrics", PrometheusMetrics),
	}
}
func PrometheusMetrics(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

	// The default go-json-rest middleware gzips the content
	promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{DisableCompression: true}).
		ServeHTTP(res.(http.ResponseWriter), req.Request)
}
