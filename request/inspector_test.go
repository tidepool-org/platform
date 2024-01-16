package request_test

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	logtest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("Inspector", func() {
	Context("HeadersInspector", func() {
		Context("NewHeadersInspector", func() {
			It("returns successfully", func() {
				Expect(request.NewHeadersInspector(nil)).ToNot(BeNil())
			})
		})

		Context("with new headers inspector", func() {
			var inspector *request.HeadersInspector
			var testLogger *logtest.Logger

			BeforeEach(func() {
				testLogger = logtest.NewLogger()
				inspector = request.NewHeadersInspector(testLogger)
				Expect(inspector).ToNot(BeNil())
			})

			It("has no headers before inspection", func() {
				Expect(inspector.Headers).To(BeNil())
			})

			Context("InspectResponse", func() {
				var headers http.Header
				var res *http.Response

				BeforeEach(func() {
					headers = http.Header{}
					for _, key := range test.RandomStringArrayFromRangeAndGeneratorWithDuplicates(1, 3, testHttp.NewHeaderKey) {
						headers[key] = test.RandomStringArrayFromRangeAndGeneratorWithDuplicates(0, 2, testHttp.NewHeaderValue)
					}
					res = &http.Response{Header: headers}
				})

				It("logs an error if the response is missing", func() {
					inspector.InspectResponse(nil)
					testLogger.Serializer.AssertWarn("response is missing")
				})

				It("captures nil headers", func() {
					res.Header = nil
					inspector.InspectResponse(res)
					Expect(inspector.Headers).To(BeNil())
				})

				It("captures empty headers", func() {
					res.Header = http.Header{}
					inspector.InspectResponse(res)
					Expect(inspector.Headers).To(BeEmpty())
				})

				It("captures non-empty headers", func() {
					inspector.InspectResponse(res)
					Expect(inspector.Headers).To(Equal(headers))
				})
			})
		})
	})
})
