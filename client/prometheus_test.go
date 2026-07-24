package client_test

import (
	"net/http"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/tidepool-org/platform/client"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	prometheusTest "github.com/tidepool-org/platform/prometheus/test"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("Prometheus", func() {
	It("PathPatternAny is expected", func() {
		Expect(client.PathPatternAny).To(Equal("/"))
	})

	It("DurationBucketsDefault is expected", func() {
		Expect(client.DurationBucketsDefault).To(Equal([]float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 15, 20, 30, 60}))
	})

	It("PrometheusLabelNameMethod is expected", func() {
		Expect(client.PrometheusLabelNameMethod).To(Equal("method"))
	})

	It("PrometheusLabelNamePath is expected", func() {
		Expect(client.PrometheusLabelNamePath).To(Equal("path"))
	})

	It("PrometheusLabelNameStatus is expected", func() {
		Expect(client.PrometheusLabelNameStatus).To(Equal("status"))
	})

	It("PrometheusLabelValueError is expected", func() {
		Expect(client.PrometheusLabelValueError).To(Equal("ERROR"))
	})

	Context("PrometheusLabelNames", func() {
		It("returns the expected label names", func() {
			Expect(client.PrometheusLabelNames()).To(Equal([]string{
				client.PrometheusLabelNameMethod,
				client.PrometheusLabelNamePath,
				client.PrometheusLabelNameStatus,
			}))
		})
	})

	Context("PrometheusRequestURLPathPatternMatcher", func() {
		Context("NewPrometheusRequestURLPathPatternMatcher", func() {
			It("returns successfully with no path patterns", func() {
				matcher := client.NewPrometheusRequestURLPathPatternMatcher()
				Expect(matcher).ToNot(BeNil())
			})
			It("returns successfully with path patterns", func() {
				matcher := client.NewPrometheusRequestURLPathPatternMatcher("/one/{id}", client.PathPatternAny)
				Expect(matcher).ToNot(BeNil())
			})
		})

		Context("MatchPath", func() {
			It("returns the request path unchanged when there are no path patterns", func() {
				matcher := client.NewPrometheusRequestURLPathPatternMatcher()
				request := testHttp.NewRequest()
				Expect(matcher.MatchPath(request)).To(PointTo(Equal(request.URL.Path)))
			})

			It("returns nil when the path does not match any path pattern", func() {
				matcher := client.NewPrometheusRequestURLPathPatternMatcher("/one/{id}")
				request := test.Must(http.NewRequest(http.MethodGet, "http://example.com/two/456", nil))
				Expect(matcher.MatchPath(request)).To(BeNil())
			})

			It("returns the matched path pattern when the path matches a specific pattern", func() {
				matcher := client.NewPrometheusRequestURLPathPatternMatcher("/one/{id}", client.PathPatternAny)
				request := test.Must(http.NewRequest(http.MethodGet, "http://example.com/one/123", nil))
				Expect(matcher.MatchPath(request)).To(PointTo(Equal("/one/{id}")))
			})

			It("returns the request path unchanged when the path matches the any path pattern", func() {
				matcher := client.NewPrometheusRequestURLPathPatternMatcher("/one/{id}", client.PathPatternAny)
				request := test.Must(http.NewRequest(http.MethodGet, "http://example.com/two/456", nil))
				Expect(matcher.MatchPath(request)).To(PointTo(Equal("/two/456")))
			})
		})
	})

	Context("PrometheusRequestRoundTripper", func() {
		Context("NewPrometheusRequestRoundTripper", func() {
			It("returns successfully", func() {
				roundTripper := client.NewPrometheusRequestRoundTripper()
				Expect(roundTripper).ToNot(BeNil())
				Expect(roundTripper.RoundTripper).ToNot(BeNil())
				Expect(roundTripper.PrometheusRequestURLPathMatcher).ToNot(BeNil())
			})

			It("returns successfully with path patterns", func() {
				roundTripper := client.NewPrometheusRequestRoundTripper("/one/{id}", client.PathPatternAny)
				Expect(roundTripper).ToNot(BeNil())
				Expect(roundTripper.RoundTripper).ToNot(BeNil())
				Expect(roundTripper.PrometheusRequestURLPathMatcher).ToNot(BeNil())
			})
		})
	})

	Context("PrometheusRequestMetricsRoundTripper", func() {
		Context("NewPrometheusRequestMetricsRoundTripper", func() {
			It("returns successfully", func() {
				roundTripper := client.NewPrometheusRequestMetricsRoundTripper(prometheusTest.RandomMetricName(), prometheusTest.RandomMetricHelp())
				Expect(roundTripper).ToNot(BeNil())
				Expect(roundTripper.PrometheusRequestRoundTripper).ToNot(BeNil())
			})
		})

		Context("NewPrometheusRequestMetricsRoundTripperWithPathPatternsAndDurationBuckets", func() {
			It("returns successfully", func() {
				roundTripper := client.NewPrometheusRequestMetricsRoundTripperWithPathPatternsAndDurationBuckets(prometheusTest.RandomMetricName(), prometheusTest.RandomMetricHelp(), []string{"/one/{id}"}, []float64{1, 2, 3})
				Expect(roundTripper).ToNot(BeNil())
				Expect(roundTripper.PrometheusRequestRoundTripper).ToNot(BeNil())
			})
		})

		Context("with a metrics round tripper", func() {
			var testRoundTripper *testHttp.RoundTripper
			var name string
			var help string
			var request *http.Request

			BeforeEach(func() {
				testRoundTripper = testHttp.NewRoundTripper()
				name = prometheusTest.RandomMetricName()
				help = prometheusTest.RandomMetricHelp()
				request = testHttp.NewRequest()
			})

			Context("RoundTrip", func() {
				It("returns the response from the resolved round tripper", func() {
					roundTripper := client.NewPrometheusRequestMetricsRoundTripper(name, help)
					roundTripper.WithRoundTripper(testRoundTripper)
					testRoundTripper.Response = &http.Response{StatusCode: testHttp.NewStatusCode()}

					result := test.Must(roundTripper.RoundTrip(request))
					Expect(result).To(BeIdenticalTo(testRoundTripper.Response))
					Expect(testRoundTripper.Request).To(BeIdenticalTo(request))
				})

				It("returns the error from the resolved round tripper", func() {
					testErr := errorsTest.RandomError()
					roundTripper := client.NewPrometheusRequestMetricsRoundTripper(name, help)
					roundTripper.WithRoundTripper(testRoundTripper)
					testRoundTripper.Error = testErr

					result, err := roundTripper.RoundTrip(request)
					Expect(err).To(Equal(testErr))
					Expect(result).To(BeNil())
				})

				It("records request count and duration metrics for a successful request", func() {
					roundTripper := client.NewPrometheusRequestMetricsRoundTripper(name, help)
					roundTripper.WithRoundTripper(testRoundTripper)
					statusCode := testHttp.NewStatusCode()
					testRoundTripper.Response = &http.Response{StatusCode: statusCode}

					_ = test.Must(roundTripper.RoundTrip(request))

					expectedLabels := map[string]string{
						client.PrometheusLabelNameMethod: request.Method,
						client.PrometheusLabelNamePath:   request.URL.Path,
						client.PrometheusLabelNameStatus: strconv.Itoa(statusCode),
					}

					countFamily := prometheusTest.MetricFamilyFromName(name + "_request_count")
					Expect(countFamily).ToNot(BeNil())
					Expect(countFamily.GetMetric()).To(HaveLen(1))
					countMetric := countFamily.GetMetric()[0]
					Expect(countMetric.GetCounter().GetValue()).To(Equal(float64(1)))
					Expect(prometheusTest.LabelPairsToMap(countMetric.GetLabel())).To(Equal(expectedLabels))

					durationFamily := prometheusTest.MetricFamilyFromName(name + "_request_duration_seconds")
					Expect(durationFamily).ToNot(BeNil())
					Expect(durationFamily.GetMetric()).To(HaveLen(1))
					durationMetric := durationFamily.GetMetric()[0]
					Expect(durationMetric.GetHistogram().GetSampleCount()).To(Equal(uint64(1)))
					Expect(prometheusTest.LabelPairsToMap(durationMetric.GetLabel())).To(Equal(expectedLabels))
				})

				It("records request count and duration metrics with an ERROR status for a failed request", func() {
					testErr := errorsTest.RandomError()
					roundTripper := client.NewPrometheusRequestMetricsRoundTripper(name, help)
					roundTripper.WithRoundTripper(testRoundTripper)
					testRoundTripper.Error = testErr

					_, err := roundTripper.RoundTrip(request)
					Expect(err).To(Equal(testErr))

					expectedLabels := map[string]string{
						client.PrometheusLabelNameMethod: request.Method,
						client.PrometheusLabelNamePath:   request.URL.Path,
						client.PrometheusLabelNameStatus: client.PrometheusLabelValueError,
					}

					countFamily := prometheusTest.MetricFamilyFromName(name + "_request_count")
					Expect(countFamily).ToNot(BeNil())
					Expect(countFamily.GetMetric()).To(HaveLen(1))
					Expect(prometheusTest.LabelPairsToMap(countFamily.GetMetric()[0].GetLabel())).To(Equal(expectedLabels))

					durationFamily := prometheusTest.MetricFamilyFromName(name + "_request_duration_seconds")
					Expect(durationFamily).ToNot(BeNil())
					Expect(durationFamily.GetMetric()).To(HaveLen(1))
					Expect(prometheusTest.LabelPairsToMap(durationFamily.GetMetric()[0].GetLabel())).To(Equal(expectedLabels))
				})

				It("does not record metrics when the path does not match a path pattern", func() {
					roundTripper := client.NewPrometheusRequestMetricsRoundTripperWithPathPatternsAndDurationBuckets(name, help, []string{"/one/{id}"}, nil)
					roundTripper.WithRoundTripper(testRoundTripper)
					testRoundTripper.Response = &http.Response{StatusCode: testHttp.NewStatusCode()}
					request = test.Must(http.NewRequest(http.MethodGet, "http://example.com/two/456", nil))

					_ = test.Must(roundTripper.RoundTrip(request))

					Expect(prometheusTest.MetricFamilyFromName(name + "_request_count")).To(BeNil())
					Expect(prometheusTest.MetricFamilyFromName(name + "_request_duration_seconds")).To(BeNil())
				})

				It("uses the custom duration buckets when specified", func() {
					roundTripper := client.NewPrometheusRequestMetricsRoundTripperWithPathPatternsAndDurationBuckets(name, help, nil, []float64{1, 2, 3})
					roundTripper.WithRoundTripper(testRoundTripper)
					testRoundTripper.Response = &http.Response{StatusCode: testHttp.NewStatusCode()}

					_ = test.Must(roundTripper.RoundTrip(request))

					durationFamily := prometheusTest.MetricFamilyFromName(name + "_request_duration_seconds")
					Expect(durationFamily).ToNot(BeNil())
					Expect(durationFamily.GetMetric()).To(HaveLen(1))
					buckets := durationFamily.GetMetric()[0].GetHistogram().GetBucket()
					Expect(buckets).To(HaveLen(3))
					Expect(buckets[0].GetUpperBound()).To(Equal(1.0))
					Expect(buckets[1].GetUpperBound()).To(Equal(2.0))
					Expect(buckets[2].GetUpperBound()).To(Equal(3.0))
				})
			})

			Context("Labels", func() {
				It("returns labels with the response status code", func() {
					roundTripper := client.NewPrometheusRequestMetricsRoundTripper(name, help)
					statusCode := testHttp.NewStatusCode()
					response := &http.Response{StatusCode: statusCode}

					Expect(roundTripper.Labels(request, response)).To(PointTo(Equal(prometheus.Labels{
						client.PrometheusLabelNameMethod: request.Method,
						client.PrometheusLabelNamePath:   request.URL.Path,
						client.PrometheusLabelNameStatus: strconv.Itoa(statusCode),
					})))
				})

				It("returns labels with an ERROR status when the response is nil", func() {
					roundTripper := client.NewPrometheusRequestMetricsRoundTripper(name, help)

					Expect(roundTripper.Labels(request, nil)).To(PointTo(Equal(prometheus.Labels{
						client.PrometheusLabelNameMethod: request.Method,
						client.PrometheusLabelNamePath:   request.URL.Path,
						client.PrometheusLabelNameStatus: client.PrometheusLabelValueError,
					})))
				})

				It("returns nil when the path does not match a path pattern", func() {
					roundTripper := client.NewPrometheusRequestMetricsRoundTripperWithPathPatternsAndDurationBuckets(name, help, []string{"/one/{id}"}, nil)
					request := test.Must(http.NewRequest(http.MethodGet, "http://example.com/two/456", nil))

					Expect(roundTripper.Labels(request, nil)).To(BeNil())
				})
			})
		})
	})
})
