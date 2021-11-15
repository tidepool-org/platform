package api_test

import (
	"encoding/json"
	"net/http"

	"github.com/mdblp/go-json-rest/rest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/auth/service"
	"github.com/tidepool-org/platform/auth/service/api"
	serviceTest "github.com/tidepool-org/platform/auth/service/test"
	"github.com/tidepool-org/platform/log"
	logNull "github.com/tidepool-org/platform/log/null"
	testRest "github.com/tidepool-org/platform/test/rest"
)

var _ = Describe("StatusGet", func() {
	var response *testRest.ResponseWriter
	var request *rest.Request
	var svc *serviceTest.Service
	var rtr *api.Router

	BeforeEach(func() {
		response = testRest.NewResponseWriter()
		request = testRest.NewRequest()
		request.Request = request.WithContext(log.NewContextWithLogger(request.Context(), logNull.NewLogger()))
		svc = serviceTest.NewService()
		var err error
		rtr, err = api.NewRouter(svc)
		Expect(err).ToNot(HaveOccurred())
		Expect(rtr).ToNot(BeNil())
	})

	AfterEach(func() {
		svc.Expectations()
		response.AssertOutputsEmpty()
	})

	Context("StatusGet", func() {
		It("panics if response is missing", func() {
			Expect(func() { rtr.StatusGet(nil, request) }).To(Panic())
		})

		It("panics if request is missing", func() {
			Expect(func() { rtr.StatusGet(response, nil) }).To(Panic())
		})

		Context("with service status", func() {
			var sts *service.Status

			BeforeEach(func() {
				sts = &service.Status{}
				svc.StatusOutputs = []*service.Status{sts}
				response.HeaderOutput = &http.Header{}
				response.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
			})

			It("returns successfully", func() {
				rtr.StatusGet(response, request)
				Expect(response.WriteInputs).To(HaveLen(1))
				Expect(json.Marshal(sts)).To(MatchJSON(response.WriteInputs[0]))
			})
		})
	})
})
