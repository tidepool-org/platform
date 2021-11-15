package api_test

import (
	"encoding/json"
	"net/http"

	"github.com/mdblp/go-json-rest/rest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/log"
	logNull "github.com/tidepool-org/platform/log/null"
	serviceApi "github.com/tidepool-org/platform/service/api"
	serviceApiTest "github.com/tidepool-org/platform/service/api/test"
	"github.com/tidepool-org/platform/test"
	testRest "github.com/tidepool-org/platform/test/rest"
)

var _ = Describe("StatusRouter", func() {
	var statusProvider *serviceApiTest.StatusProvider

	BeforeEach(func() {
		statusProvider = serviceApiTest.NewStatusProvider()
	})

	AfterEach(func() {
		statusProvider.AssertOutputsEmpty()
	})

	Context("NewStatusRouter", func() {
		It("returns an error if status provider is missing", func() {
			statusRouter, err := serviceApi.NewStatusRouter(nil)
			Expect(err).To(MatchError("status provider is missing"))
			Expect(statusRouter).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(serviceApi.NewStatusRouter(statusProvider)).ToNot(BeNil())
		})
	})

	Context("with new status router", func() {
		var statusRouter *serviceApi.StatusRouter

		BeforeEach(func() {
			var err error
			statusRouter, err = serviceApi.NewStatusRouter(statusProvider)
			Expect(err).ToNot(HaveOccurred())
			Expect(statusRouter).ToNot(BeNil())
		})

		Context("Routes", func() {
			It("returns the expected routes", func() {
				Expect(statusRouter.Routes()).ToNot(BeEmpty())
			})
		})

		Context("StatusGet", func() {
			var res *testRest.ResponseWriter
			var req *rest.Request

			BeforeEach(func() {
				res = testRest.NewResponseWriter()
				req = testRest.NewRequest()
				req.Request = req.WithContext(log.NewContextWithLogger(req.Context(), logNull.NewLogger()))
			})

			AfterEach(func() {
				res.AssertOutputsEmpty()
			})

			It("panics if response is missing", func() {
				Expect(func() { statusRouter.StatusGet(nil, req) }).To(Panic())
			})

			It("panics if request is missing", func() {
				Expect(func() { statusRouter.StatusGet(res, nil) }).To(Panic())
			})

			Context("with service status", func() {
				var status interface{}

				BeforeEach(func() {
					status = test.RandomStringFromRange(0, 32)
					statusProvider.StatusOutputs = []interface{}{status}
					res.HeaderOutput = &http.Header{}
					res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				})

				It("returns successfully", func() {
					statusRouter.StatusGet(res, req)
					Expect(res.WriteHeaderInputs).To(Equal([]int{200}))
					Expect(res.WriteInputs).To(HaveLen(1))
					Expect(json.Marshal(status)).To(MatchJSON(res.WriteInputs[0]))
				})
			})
		})
	})
})
