package service_test

import (
	"context"

	"github.com/golang/mock/gomock"
	clinic "github.com/tidepool-org/clinic/client"

	"github.com/tidepool-org/platform/clinics"
	"github.com/tidepool-org/platform/prescription"
	"github.com/tidepool-org/platform/prescription/service"
	prescriptionStoreTest "github.com/tidepool-org/platform/prescription/store/test"
	prescriptionTest "github.com/tidepool-org/platform/prescription/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PrescriptionService", func() {
	var svc prescription.Service
	var str *prescriptionStoreTest.Store
	var ctrl *gomock.Controller
	var clinicsClient *clinics.MockClient

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		clinicsClient = clinics.NewMockClient(ctrl)
		str = prescriptionStoreTest.NewStore()

		var err error
		svc, err = service.NewService(str, clinicsClient)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		ctrl.Finish()
		str.Expectations()
	})

	Context("Claim Prescription", func() {
		It("uses the clinic service to share the patient account with the clinic", func() {
			prescr := prescriptionTest.RandomPrescription()
			patient := clinic.Patient{Id: clinic.TidepoolUserId(prescr.PatientUserID)}
			claim := &prescription.Claim{
				PatientID:  prescr.PatientUserID,
				AccessCode: prescr.AccessCode,
				Birthday:   prescr.LatestRevision.Attributes.Birthday,
			}
			str.GetPrescriptionRepositoryImpl.ClaimPrescriptionOutputs = []prescriptionTest.ClaimPrescriptionOutput{{
				Prescr: prescr,
				Err:    nil,
			}}

			clinicsClient.EXPECT().SharePatientAccount(gomock.Any(), prescr.ClinicID, prescr.PatientUserID).Return(&patient, nil)

			result, err := svc.ClaimPrescription(context.Background(), claim)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(prescr))
			Expect(str.GetPrescriptionRepositoryImpl.ClaimPrescriptionInputs[0]).ToNot(BeNil())
			Expect(str.GetPrescriptionRepositoryImpl.ClaimPrescriptionInputs[0].Claim).To(Equal(claim))
		})
	})
})
