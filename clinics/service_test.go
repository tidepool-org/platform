package clinics_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"

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
	var requestQuery url.Values
	var requestBody []byte
	var responseStatusCode int
	var responseBodies [][]byte
	var client clinics.Client
	var patientID string
	var externalAccessor *authTest.ExternalAccessor

	BeforeEach(func() {
		patientID = userTest.RandomUserID()
		requestPath = ""
		requestBody = nil

		requestQuery = nil
		responseBodies = nil

		server = httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			requestPath = req.URL.Path
			requestQuery = req.URL.Query()
			var err error
			requestBody, err = io.ReadAll(req.Body)
			Expect(err).ToNot(HaveOccurred())
			if len(responseBodies) > 0 {
				res.Header().Set("Content-Type", "application/json")
			}
			res.WriteHeader(responseStatusCode)
			if len(responseBodies) > 0 {
				_, err = res.Write(responseBodies[0])
				Expect(err).ToNot(HaveOccurred())
				responseBodies = responseBodies[1:]
			}
		}))
		GinkgoT().Setenv("TIDEPOOL_CLINIC_CLIENT_ADDRESS", server.URL)

		externalAccessor = authTest.NewExternalAccessor()
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
			found, err := client.UpdatePatientSummary(context.Background(), patientID, nil, patientSummary)
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(requestPath).To(Equal("/v1/patients/" + patientID + "/summary"))
			Expect(requestQuery).ToNot(HaveKey("glycemicRangesKey"))

			decoded := &clinic.PatientSummaryV1{}
			Expect(json.Unmarshal(requestBody, decoded)).To(Succeed())
			Expect(decoded.CgmStats).ToNot(BeNil())
			Expect(decoded.CgmStats.Id).To(Equal(patientSummary.CgmStats.Id))
			Expect(decoded.BgmStats).To(BeNil())
		})

		It("tags the summary with the glycemic ranges key given", func() {
			key := "custom:mg/dL:54,70,180,250"
			_, err := client.UpdatePatientSummary(context.Background(), patientID, &key, patientSummary)
			Expect(err).ToNot(HaveOccurred())
			Expect(requestQuery.Get("glycemicRangesKey")).To(Equal(key))
		})

		// The clinic service reports a user who is not a patient of any clinic as no
		// change. Most users are not, so reporting that as a failure would fail the work of
		// nearly every user.
		It("reports the user is not a patient of any clinic, without error", func() {
			responseStatusCode = http.StatusNoContent
			found, err := client.UpdatePatientSummary(context.Background(), patientID, nil, patientSummary)
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeFalse())
		})

		It("reports the user as not found, without error", func() {
			responseStatusCode = http.StatusNotFound
			found, err := client.UpdatePatientSummary(context.Background(), patientID, nil, patientSummary)
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeFalse())
		})

		It("returns an error when the clinic service reports an unexpected status", func() {
			responseStatusCode = http.StatusInternalServerError
			_, err := client.UpdatePatientSummary(context.Background(), patientID, nil, patientSummary)
			Expect(err).To(MatchError(ContainSubstring("unexpected response status code 500")))
		})
	})

	Context("DeletePatientSummaryOfType", func() {
		BeforeEach(func() {
			responseStatusCode = http.StatusOK
		})

		It("deletes the stats of the type from every patient record of the user", func() {
			Expect(client.DeletePatientSummaryOfType(context.Background(), patientID, "cgm")).To(Succeed())
			Expect(requestPath).To(Equal("/v1/patients/" + patientID + "/summary/cgm"))
		})

		It("returns no error when the user is not a patient of any clinic", func() {
			responseStatusCode = http.StatusNoContent
			Expect(client.DeletePatientSummaryOfType(context.Background(), patientID, "bgm")).To(Succeed())
		})

		// The clinic service reports a user without patient records as no change, so not
		// found can only come from a clinic service deployed without the route
		It("returns an error when the route is not found", func() {
			responseStatusCode = http.StatusNotFound
			err := client.DeletePatientSummaryOfType(context.Background(), patientID, "bgm")
			Expect(err).To(MatchError(ContainSubstring("unexpected response status code 404")))
		})

		It("returns an error when the clinic service reports an unexpected status", func() {
			responseStatusCode = http.StatusInternalServerError
			err := client.DeletePatientSummaryOfType(context.Background(), patientID, "cgm")
			Expect(err).To(MatchError(ContainSubstring("unexpected response status code 500")))
		})
	})

	Context("ListClinicsForPatient", func() {
		newRelationships := func(count int) []byte {
			relationships := make([]clinic.PatientClinicRelationshipV1, count)
			for i := range relationships {
				relationships[i].Patient.Id = &patientID
				relationships[i].Patient.GlycemicRanges = &clinic.GlycemicRangesV1{Type: clinic.Preset, Preset: clinic.ADAHighRisk, Key: "preset:adaHighRisk"}
			}
			encoded, err := json.Marshal(relationships)
			Expect(err).ToNot(HaveOccurred())
			return encoded
		}

		BeforeEach(func() {
			responseStatusCode = http.StatusOK
		})

		It("lists the patient records of the user with their glycemic ranges", func() {
			responseBodies = [][]byte{newRelationships(2)}
			relationships, err := client.ListClinicsForPatient(context.Background(), patientID)
			Expect(err).ToNot(HaveOccurred())
			Expect(requestPath).To(Equal("/v1/patients/" + patientID + "/clinics"))
			Expect(requestQuery.Get("limit")).To(Equal("1000"))
			Expect(relationships).To(HaveLen(2))
			Expect(relationships[0].Patient.GlycemicRanges.Key).To(Equal("preset:adaHighRisk"))
		})

		It("lists every page", func() {
			// Each request takes a server session token
			externalAccessor.ServerSessionTokenOutputs = append(externalAccessor.ServerSessionTokenOutputs,
				authTest.ServerSessionTokenOutput{Token: authTest.NewSessionToken()})
			responseBodies = [][]byte{newRelationships(1000), newRelationships(1)}
			relationships, err := client.ListClinicsForPatient(context.Background(), patientID)
			Expect(err).ToNot(HaveOccurred())
			Expect(requestQuery.Get("offset")).To(Equal("1000"))
			Expect(relationships).To(HaveLen(1001))
		})

		It("lists nothing when the user is not found", func() {
			responseStatusCode = http.StatusNotFound
			relationships, err := client.ListClinicsForPatient(context.Background(), patientID)
			Expect(err).ToNot(HaveOccurred())
			Expect(relationships).To(BeEmpty())
		})

		It("returns an error when the clinic service reports an unexpected status", func() {
			responseStatusCode = http.StatusInternalServerError
			_, err := client.ListClinicsForPatient(context.Background(), patientID)
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
