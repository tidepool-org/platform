package api_test

import (
	"encoding/json"
	"net/http"

	"github.com/tidepool-org/platform/prescription/status"
	"github.com/tidepool-org/platform/status/test"

	"github.com/ant0ine/go-json-rest/rest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/prescription/api"

	"github.com/tidepool-org/platform/log"
	logNull "github.com/tidepool-org/platform/log/null"
	serviceTest "github.com/tidepool-org/platform/prescription/container/test"
	testRest "github.com/tidepool-org/platform/test/rest"
)

var _ = Describe("StatusGet", func() {
	var response *testRest.ResponseWriter
	var request *rest.Request
	var container *serviceTest.Container
	var rtr *api.Router

	BeforeEach(func() {
		response = testRest.NewResponseWriter()
		request = testRest.NewRequest()
		request.Request = request.WithContext(log.NewContextWithLogger(request.Context(), logNull.NewLogger()))

		container = serviceTest.NewContainer()
		var err error
		rtr, err = api.NewRouter(container)
		Expect(err).ToNot(HaveOccurred())
		Expect(rtr).ToNot(BeNil())
	})

	AfterEach(func() {
		container.Expectations()
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
			var sts *status.Status

			BeforeEach(func() {
				reporter := test.NewReporter()
				container.StatusReporterOutputs = []status.Reporter{reporter}
				sts = &status.Status{}
				reporter.StatusOutputs = []*status.Status{sts}
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
