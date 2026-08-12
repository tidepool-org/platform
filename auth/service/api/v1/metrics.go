package v1

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (r *Router) MetricsRoutes() []*rest.Route {
	return []*rest.Route{
		rest.Get("/v1/metrics", r.PrometheusMetrics),
	}
}

func (r *Router) PrometheusMetrics(res rest.ResponseWriter, req *rest.Request) {
	// The default go-json-rest middleware gzips the content
	promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{DisableCompression: true}).
		ServeHTTP(res.(http.ResponseWriter), req.Request)
}
