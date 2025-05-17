package request

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	prometheusPromauto "github.com/prometheus/client_golang/prometheus/promauto"
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

type PrometheusCodePathResponseInspector struct {
	*prometheus.CounterVec
	patternMux *http.ServeMux
}

// When there are only a few discrete paths possible, then no need to simplify via patterns.
//
// For example: /one, /two
func NewPrometheusCodePathResponseInspector(name string, help string) *PrometheusCodePathResponseInspector {
	return NewPrometheusCodePathResponseInspectorWithPatterns(name, help)
}

// Where there are numerous discrete paths possible, then must simplify via patterns to prevent overwhelming Prometheus.
// If no patterns are specified, then all paths are recorded as-is. If one or more patterns are specified, but a path
// does not match one of the patterns, then the path is NOT captured by Prometheus. If you wish to match "all other" paths
// and records those paths as-is, then add the pattern PatternAny at the end of your patterns.
//
// Uses standard Go HTTP pattern matching. See https://go.dev/src/net/http/pattern.go.
//
// For example: /one/{id}, /two/{id}
func NewPrometheusCodePathResponseInspectorWithPatterns(name string, help string, patterns ...string) *PrometheusCodePathResponseInspector {
	var patternMux *http.ServeMux

	if len(patterns) > 0 {
		patternMux = http.NewServeMux()
		for _, pattern := range patterns {
			patternMux.HandleFunc(pattern, func(http.ResponseWriter, *http.Request) {})
		}
	}

	return &PrometheusCodePathResponseInspector{
		CounterVec: prometheusPromauto.NewCounterVec(prometheus.CounterOpts{Name: name, Help: help}, []string{"code", "path"}),
		patternMux: patternMux,
	}
}

func (p *PrometheusCodePathResponseInspector) InspectResponse(res *http.Response) {
	path := res.Request.URL.Path
	if p.patternMux != nil {
		if _, pattern := p.patternMux.Handler(res.Request); pattern == "" {
			return
		} else if pattern != PatternAny {
			path = pattern
		}
	}
	p.With(prometheus.Labels{"code": strconv.Itoa(res.StatusCode), "path": path}).Inc()
}

const PatternAny = "/"
