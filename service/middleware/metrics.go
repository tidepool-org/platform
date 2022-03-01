package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/mdblp/go-json-rest/rest"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	dflBuckets = []float64{300, 1200, 5000}
)

const (
	reqsName    = "dblp_data_http_requests_total"
	latencyName = "dblp_data_http_requests_duration_seconds"
)

// Middleware is a handler that exposes prometheus metrics for the number of requests,
// the latency and the response size, partitioned by status code, method and HTTP path.
type PromMiddleware struct {
	reqs    *prometheus.CounterVec
	latency *prometheus.HistogramVec
	name    string
}

// NewMiddleware returns a new prometheus Middleware handler.
func (mw *PromMiddleware) MiddlewareFunc(h rest.HandlerFunc) rest.HandlerFunc {
	start := time.Now()
	mw.reqs = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        reqsName,
			Help:        "How many HTTP requests processed, partitioned by status code, method and HTTP path.",
			ConstLabels: prometheus.Labels{"service": mw.name},
		},
		[]string{"code", "method", "path"},
	)
	prometheus.MustRegister(mw.reqs)

	mw.latency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        latencyName,
		Help:        "How long it took to process the request, partitioned by status code, method and HTTP path.",
		ConstLabels: prometheus.Labels{"service": mw.name},
		Buckets:     dflBuckets,
	},
		[]string{"code", "method", "path"},
	)
	prometheus.MustRegister(mw.latency)

	return func(w rest.ResponseWriter, r *rest.Request) {
		h(w, r)

		if r.Env["STATUS_CODE"] == nil {
			log.Fatal("StatusMiddleware: Env[\"STATUS_CODE\"] is nil, " +
				"RecorderMiddleware may not be in the wrapped Middlewares.")
		}
		statusCode := r.Env["STATUS_CODE"].(int)

		mw.reqs.WithLabelValues(http.StatusText(statusCode), r.Method, r.PathExp).Inc()
		mw.latency.WithLabelValues(http.StatusText(statusCode), r.Method, r.PathExp).Observe(float64(time.Since(start).Nanoseconds()) / 1000000000)
	}

}
