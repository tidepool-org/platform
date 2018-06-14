package request_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"net/http"

	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("Inspector", func() {
	Context("HeadersInspector", func() {
		Context("NewHeadersInspector", func() {
			It("returns successfully", func() {
				Expect(request.NewHeadersInspector()).ToNot(BeNil())
			})
		})

		Context("with new headers inspector", func() {
			var inspector *request.HeadersInspector

			BeforeEach(func() {
				inspector = request.NewHeadersInspector()
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

				It("returns an error if the response is missing", func() {
					Expect(inspector.InspectResponse(nil)).To(MatchError("response is missing"))
				})

				It("captures nil headers", func() {
					res.Header = nil
					Expect(inspector.InspectResponse(res)).To(Succeed())
					Expect(inspector.Headers).To(BeNil())
				})

				It("captures empty headers", func() {
					res.Header = http.Header{}
					Expect(inspector.InspectResponse(res)).To(Succeed())
					Expect(inspector.Headers).To(BeEmpty())
				})

				It("captures non-empty headers", func() {
					Expect(inspector.InspectResponse(res)).To(Succeed())
					Expect(inspector.Headers).To(Equal(headers))
				})
			})
		})
	})
})
