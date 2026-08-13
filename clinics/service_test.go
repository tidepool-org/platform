package clinics_test

import (
	"context"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/clinics"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("Client", func() {
	Context("SyncEHRDataForPatient", func() {
		var server *httptest.Server
		var requestPath string
		var responseStatusCode int
		var client clinics.Client
		var patientID string

		BeforeEach(func() {
			patientID = userTest.RandomUserID()
			requestPath = ""
			responseStatusCode = http.StatusAccepted

			server = httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
				requestPath = req.URL.Path
				res.WriteHeader(responseStatusCode)
			}))
			GinkgoT().Setenv("TIDEPOOL_CLINIC_CLIENT_ADDRESS", server.URL)

			externalAccessor := authTest.NewExternalAccessor()
			externalAccessor.ServerSessionTokenOutputs = []authTest.ServerSessionTokenOutput{{Token: authTest.NewSessionToken()}}

			var err error
			client, err = clinics.NewClient(externalAccessor)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			server.Close()
		})

		It("requests a synchronization for the patient", func() {
			Expect(client.SyncEHRDataForPatient(context.Background(), patientID)).To(Succeed())
			Expect(requestPath).To(Equal("/v1/patients/" + patientID + "/ehr/sync"))
		})

		// The clinic service reports a patient with no active subscription to any clinic enabled for an
		// electronic health record as not found. Most users are not such a patient, so reporting that as
		// a failure would fail the work of nearly every user.
		It("returns no error when the patient has no active subscription", func() {
			responseStatusCode = http.StatusNotFound
			Expect(client.SyncEHRDataForPatient(context.Background(), patientID)).To(Succeed())
		})

		It("returns an error when the clinic service reports an unexpected status", func() {
			responseStatusCode = http.StatusInternalServerError
			err := client.SyncEHRDataForPatient(context.Background(), patientID)
			Expect(err).To(MatchError(ContainSubstring("unexpected response status code 500")))
		})
	})
})
