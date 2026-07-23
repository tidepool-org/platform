package client

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/tidepool-org/platform/pointer"
)

const PathPatternAny = "/"

var (
	DurationBucketsDefault = []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 15, 20, 30, 60}

	PrometheusLabelNameMethod = "method"
	PrometheusLabelNamePath   = "path"
	PrometheusLabelNameStatus = "status"

	PrometheusLabelValueError = "ERROR"
)

func PrometheusLabelNames() []string {
	return []string{
		PrometheusLabelNameMethod,
		PrometheusLabelNamePath,
		PrometheusLabelNameStatus,
	}
}

// Where there are numerous discrete paths possible, then must simplify via patterns to prevent overwhelming Prometheus.
// If no patterns are specified, then all paths are recorded as-is. If one or more patterns are specified, but a path
// does not match one of the patterns, then the path is NOT captured by Prometheus. If you wish to match "all other" paths
// and records those paths as-is, then add the pattern PathPatternAny at the end of your patterns.
//
// Uses standard Go HTTP pattern matching. See https://go.dev/src/net/http/pattern.go.
//
// For example: /one/{id}, /two/{id}
type PrometheusRequestURLPathMatcher struct {
	pathPatternMux *http.ServeMux
}

func NewPrometheusRequestURLPathPatternMatcher(pathPatterns ...string) *PrometheusRequestURLPathMatcher {
	var pathPatternMux *http.ServeMux

	if len(pathPatterns) > 0 {
		pathPatternMux = http.NewServeMux()
		for _, pattern := range pathPatterns {
			pathPatternMux.HandleFunc(pattern, func(http.ResponseWriter, *http.Request) {})
		}
	}

	return &PrometheusRequestURLPathMatcher{
		pathPatternMux: pathPatternMux,
	}
}

func (p *PrometheusRequestURLPathMatcher) MatchPath(req *http.Request) *string {
	path := req.URL.Path
	if p.pathPatternMux != nil {
		if _, pattern := p.pathPatternMux.Handler(req); pattern == "" {
			return nil
		} else if pattern != PathPatternAny {
			path = pattern
		}
	}
	return &path
}

type PrometheusRequestRoundTripper struct {
	*RoundTripper
	*PrometheusRequestURLPathMatcher
}

func NewPrometheusRequestRoundTripper(pathPatterns ...string) *PrometheusRequestRoundTripper {
	return &PrometheusRequestRoundTripper{
		RoundTripper:                    NewRoundTripper(nil),
		PrometheusRequestURLPathMatcher: NewPrometheusRequestURLPathPatternMatcher(pathPatterns...),
	}
}

type PrometheusRequestMetricsRoundTripper struct {
	*PrometheusRequestRoundTripper
	requestCountCounterVec      *prometheus.CounterVec
	requestDurationHistogramVec *prometheus.HistogramVec
}

func NewPrometheusRequestMetricsRoundTripper(name string, help string) *PrometheusRequestMetricsRoundTripper {
	return NewPrometheusRequestMetricsRoundTripperWithPathPatternsAndDurationBuckets(name, help, nil, nil)
}

func NewPrometheusRequestMetricsRoundTripperWithPathPatternsAndDurationBuckets(name string, help string, pathPatterns []string, durationBuckets []float64) *PrometheusRequestMetricsRoundTripper {
	return &PrometheusRequestMetricsRoundTripper{
		PrometheusRequestRoundTripper: NewPrometheusRequestRoundTripper(pathPatterns...),
		requestCountCounterVec: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: fmt.Sprintf("%s_request_count", name),
				Help: fmt.Sprintf("%s request count, sorted by method, path, and status", help),
			},
			PrometheusLabelNames(),
		),
		requestDurationHistogramVec: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    fmt.Sprintf("%s_request_duration_seconds", name),
				Help:    fmt.Sprintf("%s request duration, in seconds, sorted by method, path, and status", help),
				Buckets: pointer.DefaultArray(durationBuckets, DurationBucketsDefault),
			},
			PrometheusLabelNames(),
		),
	}
}

func (p *PrometheusRequestMetricsRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	res, err := p.PrometheusRequestRoundTripper.RoundTrip(req)
	duration := time.Since(start)

	if labels := p.Labels(req, res); labels != nil {
		p.requestCountCounterVec.With(*labels).Inc()
		p.requestDurationHistogramVec.With(*labels).Observe(duration.Seconds())
	}

	return res, err
}

func (p *PrometheusRequestMetricsRoundTripper) Labels(req *http.Request, res *http.Response) *prometheus.Labels {
	path := p.MatchPath(req)
	if path == nil {
		return nil
	}

	labels := prometheus.Labels{
		PrometheusLabelNameMethod: req.Method,
		PrometheusLabelNamePath:   *path,
	}
	if res != nil {
		labels[PrometheusLabelNameStatus] = strconv.Itoa(res.StatusCode)
	} else {
		labels[PrometheusLabelNameStatus] = PrometheusLabelValueError
	}

	return &labels
}
