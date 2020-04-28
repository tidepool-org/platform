package status_test

import (
	"encoding/json"
	"net/http"

	"github.com/tidepool-org/platform/status"
	"github.com/tidepool-org/platform/status/test"
	"github.com/tidepool-org/platform/version"
	versionTest "github.com/tidepool-org/platform/version/test"

	"github.com/ant0ine/go-json-rest/rest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/log"
	logNull "github.com/tidepool-org/platform/log/null"
	testRest "github.com/tidepool-org/platform/test/rest"
)

var _ = Describe("StatusGet", func() {
	var response *testRest.ResponseWriter
	var request *rest.Request
	var rtr *status.Router
	var versionReporter version.Reporter
	var storeStatusReporter *test.StoreStatusReporter

	BeforeEach(func() {
		response = testRest.NewResponseWriter()
		request = testRest.NewRequest()
		request.Request = request.WithContext(log.NewContextWithLogger(request.Context(), logNull.NewLogger()))
		versionReporter = versionTest.NewReporter()
		storeStatusReporter = test.NewStoreStatusReporter()

		rtr = status.NewRouter(status.Params{
			VersionReporter:     versionReporter,
			StoreStatusReporter: storeStatusReporter,
		}).(*status.Router)

		Expect(rtr).ToNot(BeNil())
	})

	AfterEach(func() {
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
			BeforeEach(func() {
				response.HeaderOutput = &http.Header{}
				response.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
			})

			It("returns successfully", func() {
				rtr.StatusGet(response, request)
				Expect(response.WriteInputs).To(HaveLen(1))
			})

			It("returns the expected response", func() {
				storeStatusReporter.SetStatus(test.OkStoreStatus())

				sts := status.Status{
					Version: versionReporter.Long(),
					Store:   test.OkStoreStatus(),
				}

				rtr.StatusGet(response, request)
				Expect(json.Marshal(sts)).To(MatchJSON(response.WriteInputs[0]))
			})
		})
	})
})
