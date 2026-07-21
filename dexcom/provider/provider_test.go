package provider_test

import (
	"net/http"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/client"
	dexcomProvider "github.com/tidepool-org/platform/dexcom/provider"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	prometheusTest "github.com/tidepool-org/platform/prometheus/test"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("Provider", func() {
	It("RequestTimeHeaderName is expected", func() {
		Expect(dexcomProvider.RequestTimeHeaderName).To(Equal("request-time"))
	})

	Context("PrometheusRequestMetricsRoundTripper", func() {
		Context("NewPrometheusRequestMetricsRoundTripper", func() {
			It("returns successfully", func() {
				roundTripper := dexcomProvider.NewPrometheusRequestMetricsRoundTripper(prometheusTest.RandomMetricName(), prometheusTest.RandomMetricHelp())
				Expect(roundTripper).ToNot(BeNil())
				Expect(roundTripper.PrometheusRequestMetricsRoundTripper).ToNot(BeNil())
			})
		})

		Context("RoundTrip", func() {
			var testRoundTripper *testHttp.RoundTripper
			var name string
			var roundTripper *dexcomProvider.PrometheusRequestMetricsRoundTripper
			var request *http.Request

			BeforeEach(func() {
				testRoundTripper = testHttp.NewRoundTripper()
				name = prometheusTest.RandomMetricName()
				roundTripper = dexcomProvider.NewPrometheusRequestMetricsRoundTripper(name, prometheusTest.RandomMetricHelp())
				roundTripper.WithRoundTripper(testRoundTripper)
				request = testHttp.NewRequest()
			})

			It("returns the response from the resolved round tripper", func() {
				testRoundTripper.Response = &http.Response{StatusCode: testHttp.NewStatusCode()}

				result := test.Must(roundTripper.RoundTrip(request))
				Expect(result).To(BeIdenticalTo(testRoundTripper.Response))
				Expect(testRoundTripper.Request).To(BeIdenticalTo(request))
			})

			It("returns the error from the resolved round tripper", func() {
				testErr := errorsTest.RandomError()
				testRoundTripper.Error = testErr

				result, err := roundTripper.RoundTrip(request)
				Expect(err).To(Equal(testErr))
				Expect(result).To(BeNil())
			})

			It("does not record a request time metric when the resolved round tripper returns an error", func() {
				testRoundTripper.Error = errorsTest.RandomError()

				_, _ = roundTripper.RoundTrip(request)

				Expect(prometheusTest.MetricFamilyFromName(name + "_request_time_seconds")).To(BeNil())
			})

			It("records a request time metric when the response has a valid request-time header", func() {
				statusCode := testHttp.NewStatusCode()
				requestTime := time.Duration(test.RandomIntFromRange(1, 60*1000)) * time.Millisecond
				header := http.Header{}
				header.Set(dexcomProvider.RequestTimeHeaderName, requestTime.String())
				testRoundTripper.Response = &http.Response{StatusCode: statusCode, Header: header}

				_ = test.Must(roundTripper.RoundTrip(request))

				family := prometheusTest.MetricFamilyFromName(name + "_request_time_seconds")
				Expect(family).ToNot(BeNil())
				Expect(family.GetMetric()).To(HaveLen(1))
				metric := family.GetMetric()[0]
				Expect(metric.GetHistogram().GetSampleCount()).To(Equal(uint64(1)))
				Expect(metric.GetHistogram().GetSampleSum()).To(Equal(requestTime.Seconds()))
				Expect(prometheusTest.LabelPairsToMap(metric.GetLabel())).To(Equal(map[string]string{
					client.PrometheusLabelNameMethod: request.Method,
					client.PrometheusLabelNamePath:   request.URL.Path,
					client.PrometheusLabelNameStatus: strconv.Itoa(statusCode),
				}))
			})

			It("does not record a request time metric when the response does not have a request-time header", func() {
				testRoundTripper.Response = &http.Response{StatusCode: testHttp.NewStatusCode(), Header: http.Header{}}

				_ = test.Must(roundTripper.RoundTrip(request))

				Expect(prometheusTest.MetricFamilyFromName(name + "_request_time_seconds")).To(BeNil())
			})

			It("does not record a request time metric when the request-time header is not a valid duration", func() {
				header := http.Header{}
				header.Set(dexcomProvider.RequestTimeHeaderName, test.RandomString())
				testRoundTripper.Response = &http.Response{StatusCode: testHttp.NewStatusCode(), Header: header}

				_ = test.Must(roundTripper.RoundTrip(request))

				Expect(prometheusTest.MetricFamilyFromName(name + "_request_time_seconds")).To(BeNil())
			})
		})
	})
})
