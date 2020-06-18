package mailer

import "github.com/prometheus/client_golang/prometheus"

var (
	errorCounter = createErrorCounter()
)

func createErrorCounter() *prometheus.CounterVec {
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "tidepool",
			Subsystem: "mailer",
			Name: "backend_errors",
		},
		[]string{"code", "backend"},
	)

	prometheus.MustRegister(counter)
	return counter
}

func ObserveError(code string, backend string) {
	errorCounter.WithLabelValues(code, backend).Inc()
}