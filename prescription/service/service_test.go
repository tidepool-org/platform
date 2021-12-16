package service_test

import (
	"context"

	"github.com/tidepool-org/go-common/events"

	prescriptionApplicationTest "github.com/tidepool-org/platform/prescription/application/test"

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
	var clinicsCtrl *gomock.Controller
	var mailerCtrl *gomock.Controller
	var clinicsClient *clinics.MockClient
	var mailerClient *prescriptionApplicationTest.MockMockMailer

	BeforeEach(func() {
		mailerCtrl = gomock.NewController(GinkgoT())
		mailerClient = prescriptionApplicationTest.NewMockMockMailer(mailerCtrl)

		clinicsCtrl = gomock.NewController(GinkgoT())
		clinicsClient = clinics.NewMockClient(clinicsCtrl)

		str = prescriptionStoreTest.NewStore()

		var err error
		svc, err = service.NewService(str, clinicsClient, mailerClient)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		clinicsCtrl.Finish()
		mailerCtrl.Finish()
		str.Expectations()
	})

	Context("Create Prescription", func() {
		It("sends an email when the prescription is in submitted state", func() {
			create := prescriptionTest.RandomRevisionCreate()
			create.State = prescription.StateSubmitted
			prescr := prescription.NewPrescription(create)

			str.GetPrescriptionRepositoryImpl.CreatePrescriptionInputs = []prescriptionTest.CreatePrescriptionInput{{
				RevisionCreate: create,
			}}
			str.GetPrescriptionRepositoryImpl.CreatePrescriptionOutputs = []prescriptionTest.CreatePrescriptionOutput{{
				Prescription: prescr,
				Error:        nil,
			}}

			expectedEmail := events.SendEmailTemplateEvent{
				Recipient: *create.Email,
				Template:  "prescription_access_code",
				Variables: map[string]string{
					"AccessCode": prescr.AccessCode,
				},
			}

			mailerClient.EXPECT().SendEmailTemplate(gomock.Any(), expectedEmail).Return(nil)

			result, err := svc.CreatePrescription(context.Background(), create)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(prescr))
		})
	})

	Context("Add Revision", func() {
		It("sends an email when the prescription is in submitted state", func() {
			create := prescriptionTest.RandomRevisionCreate()
			create.State = prescription.StateSubmitted
			prescr := prescription.NewPrescription(create)

			str.GetPrescriptionRepositoryImpl.AddRevisionInputs = []prescriptionTest.AddRevisionInput{{
				Create: create,
				ID:     prescr.ID.Hex(),
			}}
			str.GetPrescriptionRepositoryImpl.AddRevisionOutputs = []prescriptionTest.AddRevisionOutput{{
				Prescr: prescr,
				Err:    nil,
			}}

			expectedEmail := events.SendEmailTemplateEvent{
				Recipient: *create.Email,
				Template:  "prescription_access_code",
				Variables: map[string]string{
					"AccessCode": prescr.AccessCode,
				},
			}

			mailerClient.EXPECT().SendEmailTemplate(gomock.Any(), expectedEmail).Return(nil)

			result, err := svc.AddRevision(context.Background(), prescr.ID.Hex(), create)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(prescr))
		})
	})

	Context("Claim Prescription", func() {
		It("uses the clinic service to share the patient account with the clinic", func() {
			prescr := prescriptionTest.RandomPrescription()
			patient := clinic.Patient{Id: clinic.TidepoolUserId(prescr.PatientUserID)}
			claim := &prescription.Claim{
				PatientID:  prescr.PatientUserID,
				AccessCode: prescr.AccessCode,
				Birthday:   *prescr.LatestRevision.Attributes.Birthday,
			}
			str.GetPrescriptionRepositoryImpl.GetClaimablePrescriptionOutputs = []prescriptionTest.GetClaimablePrescriptionOutput{{
				Prescr: prescr,
				Err:    nil,
			}}
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
