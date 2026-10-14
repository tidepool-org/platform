package clinics_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	clinic "github.com/tidepool-org/clinic/client"
	"go.mongodb.org/mongo-driver/bson/primitive"

	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/clinics"
	summaryTest "github.com/tidepool-org/platform/summary/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("Client", func() {
	var server *httptest.Server
	var requestPath string
	var requestBody []byte
	var responseStatusCode int
	var client clinics.Client
	var patientID string

	BeforeEach(func() {
		patientID = userTest.RandomUserID()
		requestPath = ""
		requestBody = nil

		server = httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			requestPath = req.URL.Path
			var err error
			requestBody, err = io.ReadAll(req.Body)
			Expect(err).ToNot(HaveOccurred())
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

	Context("SyncEHRDataForPatient", func() {
		BeforeEach(func() {
			responseStatusCode = http.StatusAccepted
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

	Context("UpdatePatientSummary", func() {
		var patientSummary *clinic.PatientSummaryV1

		BeforeEach(func() {
			responseStatusCode = http.StatusOK
			cgm := summaryTest.RandomCGMSummary(patientID)
			cgm.ID = primitive.NewObjectID()
			patientSummary = clinics.NewPatientSummary(cgm, nil)
		})

		It("updates the summary of the patient", func() {
			Expect(client.UpdatePatientSummary(context.Background(), patientID, patientSummary)).To(Succeed())
			Expect(requestPath).To(Equal("/v1/patients/" + patientID + "/summary"))

			decoded := &clinic.PatientSummaryV1{}
			Expect(json.Unmarshal(requestBody, decoded)).To(Succeed())
			Expect(decoded.CgmStats).ToNot(BeNil())
			Expect(decoded.CgmStats.Id).To(Equal(patientSummary.CgmStats.Id))
			Expect(decoded.BgmStats).To(BeNil())
		})

		// The clinic service reports a user who is not a patient of any clinic as no change. Most
		// users are not, so reporting that as a failure would fail the work of nearly every user.
		It("returns no error when the user is not a patient of any clinic", func() {
			responseStatusCode = http.StatusNoContent
			Expect(client.UpdatePatientSummary(context.Background(), patientID, patientSummary)).To(Succeed())
		})

		It("returns no error when the clinic service reports not found", func() {
			responseStatusCode = http.StatusNotFound
			Expect(client.UpdatePatientSummary(context.Background(), patientID, patientSummary)).To(Succeed())
		})

		It("returns an error when the clinic service reports an unexpected status", func() {
			responseStatusCode = http.StatusInternalServerError
			err := client.UpdatePatientSummary(context.Background(), patientID, patientSummary)
			Expect(err).To(MatchError(ContainSubstring("unexpected response status code 500")))
		})
	})

	Context("DeletePatientSummary", func() {
		var summaryID string

		BeforeEach(func() {
			responseStatusCode = http.StatusOK
			summaryID = primitive.NewObjectID().Hex()
		})

		It("deletes the summary from every patient record holding it", func() {
			Expect(client.DeletePatientSummary(context.Background(), summaryID)).To(Succeed())
			Expect(requestPath).To(Equal("/v1/summaries/" + summaryID + "/clinics"))
		})

		It("returns no error when no patient record holds the summary", func() {
			responseStatusCode = http.StatusNoContent
			Expect(client.DeletePatientSummary(context.Background(), summaryID)).To(Succeed())
		})

		// A delete resent by retried work may target a summary the clinic service already deleted
		It("returns no error when the summary is not found", func() {
			responseStatusCode = http.StatusNotFound
			Expect(client.DeletePatientSummary(context.Background(), summaryID)).To(Succeed())
		})

		It("returns an error when the clinic service reports an unexpected status", func() {
			responseStatusCode = http.StatusInternalServerError
			err := client.DeletePatientSummary(context.Background(), summaryID)
			Expect(err).To(MatchError(ContainSubstring("unexpected response status code 500")))
		})
	})
})
