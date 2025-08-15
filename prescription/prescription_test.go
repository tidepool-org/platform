package prescription_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"syreclabs.com/go/faker"

	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/prescription"
	"github.com/tidepool-org/platform/prescription/test"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/user"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("Prescription", func() {
	Context("With a submitted revision", func() {
		var revisionCreate *prescription.RevisionCreate

		BeforeEach(func() {
			revisionCreate = test.RandomRevisionCreate()
			revisionCreate.State = prescription.StateSubmitted
		})

		Context("Create new prescription", func() {
			var prescr *prescription.Prescription

			BeforeEach(func() {
				prescr = prescription.NewPrescription(revisionCreate)
				Expect(prescr).ToNot(BeNil())
			})

			It("creates a non-empty id", func() {
				Expect(prescr.ID).ToNot(BeEmpty())
			})

			It("does not set the patientId", func() {
				Expect(prescr.PatientUserID).To(BeEmpty())
			})

			It("generates a non-empty access code", func() {
				Expect(prescr.AccessCode).ToNot(BeEmpty())
			})

			It("sets the state to the revision state", func() {
				Expect(prescr.State).To(Equal(revisionCreate.State))
			})

			It("sets the latest revision attribute to new revision", func() {
				Expect(prescr.LatestRevision).ToNot(BeNil())
			})

			It("populates the revision history with the newly created revision", func() {
				Expect(prescr.RevisionHistory).ToNot(BeEmpty())
				Expect(prescr.RevisionHistory[0]).To(Equal(prescr.LatestRevision))
			})

			It("sets the created user id correctly", func() {
				Expect(prescr.CreatedUserID).To(Equal(revisionCreate.ClinicianID))
			})

			It("sets the prescriber user id correctly", func() {
				Expect(prescr.PrescriberUserID).To(Equal(revisionCreate.ClinicianID))
			})

			It("sets the created time correctly", func() {
				Expect(prescr.CreatedTime).To(BeTemporally("~", time.Now(), 10*time.Millisecond))
			})

			It("does not set the deleted time", func() {
				Expect(prescr.DeletedTime).To(BeNil())
			})

			It("does not set the deleted user id", func() {
				Expect(prescr.DeletedUserID).To(Equal(""))
			})

			It("creates a revision with id 0", func() {
				Expect(prescr.LatestRevision.RevisionID).To(Equal(0))
			})

			It("sets the modified time", func() {
				Expect(prescr.ModifiedTime).To(BeTemporally("~", time.Now(), 10*time.Millisecond))
			})

			It("sets the modified user id", func() {
				Expect(prescr.ModifiedUserID).To(Equal(revisionCreate.ClinicianID))
			})

			It("sets the submitted time", func() {
				Expect(prescr.SubmittedTime).ToNot(BeNil())
				Expect(*prescr.SubmittedTime).To(BeTemporally("~", time.Now(), 10*time.Millisecond))
			})
		})
	})

	Describe("Update", func() {
		var revisionCreate *prescription.RevisionCreate

		BeforeEach(func() {
			revisionCreate = test.RandomRevisionCreate()
			revisionCreate.State = prescription.StatePending
		})

		Describe("AddRevision", func() {
			var prescr *prescription.Prescription
			var newRevision *prescription.RevisionCreate
			var update *prescription.Update

			BeforeEach(func() {
				prescr = prescription.NewPrescription(revisionCreate)
				newRevision = test.RandomRevisionCreate()
				newRevision.State = prescription.StateSubmitted
				update = prescription.NewPrescriptionAddRevisionUpdate(prescr, newRevision)
			})

			It("sets the revision correctly", func() {
				expectedRevision := prescription.NewRevision(prescr.LatestRevision.RevisionID+1, newRevision)
				expectedRevision.Attributes.CreatedTime = update.Revision.Attributes.CreatedTime
				Expect(*update.Revision).To(Equal(*expectedRevision))
			})

			It("sets the state correctly", func() {
				Expect(update.State).To(Equal(newRevision.State))
			})

			It("sets the prescriber id correctly", func() {
				Expect(update.PrescriberUserID).To(Equal(newRevision.ClinicianID))
			})

			It("doesn't set the patient id", func() {
				Expect(update.PatientUserID).To(BeEmpty())
			})

			It("sets the expiration time", func() {
				Expect(*update.ExpirationTime).To(BeTemporally(">", time.Now()))
			})

			It("sets the modified time", func() {
				Expect(update.ModifiedTime).To(BeTemporally("~", time.Now(), 10*time.Millisecond))
			})

			It("sets the modified user id", func() {
				Expect(update.ModifiedUserID).To(Equal(newRevision.ClinicianID))
			})

			It("sets the submitted time", func() {
				Expect(update.SubmittedTime).ToNot(BeNil())
				Expect(*update.SubmittedTime).To(BeTemporally("~", time.Now(), 10*time.Millisecond))
			})
		})

		Describe("ClaimUpdate", func() {
			var prescr *prescription.Prescription
			var update *prescription.Update
			var userID string

			BeforeEach(func() {
				userID = userTest.RandomUserID()
				revisionCreate.State = prescription.StateSubmitted
				prescr = prescription.NewPrescription(revisionCreate)
				update = prescription.NewPrescriptionClaimUpdate(userID, prescr)
			})

			It("sets the state to claimed", func() {
				Expect(update.State).To(Equal(prescription.StateClaimed))
			})

			It("sets the patient id correctly", func() {
				Expect(update.PatientUserID).To(Equal(userID))
			})

			It("doesn't set the prescriber id", func() {
				Expect(update.PrescriberUserID).To(BeEmpty())
			})

			It("resets the expiration time", func() {
				Expect(update.ExpirationTime).To(BeNil())
			})

			It("sets the modified time", func() {
				Expect(update.ModifiedTime).To(BeTemporally("~", time.Now(), time.Second))
			})

			It("sets the modified user id", func() {
				Expect(update.ModifiedUserID).To(Equal(userID))
			})
		})

		Describe("StateUpdate", func() {
			var prescr *prescription.Prescription
			var update *prescription.Update
			var userID string

			BeforeEach(func() {
				userID = userTest.RandomUserID()
				revisionCreate.State = prescription.StateClaimed
				prescr = prescription.NewPrescription(revisionCreate)
				stateUpdate := prescription.NewStateUpdate(userID)
				stateUpdate.State = prescription.StateActive
				update = prescription.NewPrescriptionStateUpdate(prescr, stateUpdate)
			})

			It("doesn't create a new revision", func() {
				Expect(update.Revision).To(BeNil())
			})

			It("sets the state to active", func() {
				Expect(update.State).To(Equal(prescription.StateActive))
			})

			It("sets the patient id", func() {
				Expect(update.PatientUserID).To(Equal(userID))
			})

			It("doesn't set the prescriber id", func() {
				Expect(update.PrescriberUserID).To(BeEmpty())
			})

			It("resets the expiration time", func() {
				Expect(update.ExpirationTime).To(BeNil())
			})

			It("sets the modified time", func() {
				Expect(update.ModifiedTime).To(BeTemporally("~", time.Now(), 10*time.Millisecond))
			})

			It("sets the modified user id", func() {
				Expect(update.ModifiedUserID).To(Equal(userID))
			})
		})
	})

	Describe("Filter", func() {
		Describe("Validate", func() {
			When("current user is a patient", func() {
				var usr *user.User
				var filter *prescription.Filter
				var validate structure.Validator

				BeforeEach(func() {
					usr = userTest.RandomUser()
					usr.Roles = &[]string{}

					var err error
					filter, err = prescription.NewPatientFilter(*usr.UserID)
					Expect(err).ToNot(HaveOccurred())

					validate = validator.New(logTest.NewLogger())
					Expect(validate.Validate(filter)).ToNot(HaveOccurred())
				})

				It("fails when the state is draft", func() {
					filter.State = prescription.StateDraft

					filter.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("fails when the state is pending", func() {
					filter.State = prescription.StatePending

					filter.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("fails when the state is expired", func() {
					filter.State = prescription.StateExpired

					filter.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("doesn't fail when the state is submitted", func() {
					filter.State = prescription.StateSubmitted

					filter.Validate(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())
				})

				It("doesn't fail when the state is active", func() {
					filter.State = prescription.StateActive

					filter.Validate(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())
				})

				It("doesn't fail when the state is inactive", func() {
					filter.State = prescription.StateInactive

					filter.Validate(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())
				})

				It("doesn't fail when the state is claimed", func() {
					filter.State = prescription.StateClaimed

					filter.Validate(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())
				})

				It("fails when the state is unrecognized", func() {
					filter.State = "invalid"

					filter.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("fails when patient email is set", func() {
					filter.PatientEmail = faker.Internet().Email()

					filter.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("doesn't fail with a valid id", func() {
					filter.ID = primitive.NewObjectID().Hex()

					filter.Validate(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())
				})

				It("fails when the id is 13 hex characters", func() {
					filter.ID = fmt.Sprintf("%sa", primitive.NewObjectID().Hex())

					filter.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("fails when the id contains non-hex character", func() {
					filter.ID = "507f1f77bcf86cd799439011z"

					filter.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})
			})

			When("current user is a clinician", func() {
				var clinicID string
				var filter *prescription.Filter
				var validate structure.Validator

				BeforeEach(func() {
					clinicID = faker.Number().Hexadecimal(24)
					var err error
					filter, err = prescription.NewClinicFilter(clinicID)
					Expect(err).ToNot(HaveOccurred())

					validate = validator.New(logTest.NewLogger())
					Expect(validate.Validate(filter)).ToNot(HaveOccurred())
				})

				It("fails when patient id is invalid", func() {
					filter.PatientUserID = "invalid"

					filter.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("doesn't fail when patient id invalid", func() {
					filter.PatientUserID = userTest.RandomUserID()

					filter.Validate(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())
				})

				It("doesn't fail when the state is draft", func() {
					filter.State = prescription.StateDraft

					filter.Validate(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())
				})

				It("doesn't fail when the state is pending", func() {
					filter.State = prescription.StatePending

					filter.Validate(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())
				})

				It("doesn't fail when the state is expired", func() {
					filter.State = prescription.StateExpired

					filter.Validate(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())
				})

				It("doesn't fail when the state is submitted", func() {
					filter.State = prescription.StateSubmitted

					filter.Validate(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())
				})

				It("doesn't fail when the state is active", func() {
					filter.State = prescription.StateActive

					filter.Validate(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())
				})

				It("doesn't fail when the state is inactive", func() {
					filter.State = prescription.StateInactive

					filter.Validate(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())
				})

				It("doesn't fail when the state is claimed", func() {
					filter.State = prescription.StateClaimed

					filter.Validate(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())
				})

				It("fails when the state is unrecognized", func() {
					filter.State = "invalid"

					filter.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("doesn't fail when patient email is valid", func() {
					filter.PatientEmail = faker.Internet().Email()

					filter.Validate(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())
				})

				It("fails when patient email is invalid", func() {
					filter.PatientEmail = "invalid-email.com"

					filter.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("doesn't fail when the patient email is valid", func() {
					filter.PatientEmail = faker.Internet().Email()

					filter.Validate(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())
				})

				It("doesn't fail with a valid id", func() {
					filter.ID = primitive.NewObjectID().Hex()

					filter.Validate(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())
				})

				It("fails when the id is 13 hex characters", func() {
					filter.ID = fmt.Sprintf("%sa", primitive.NewObjectID().Hex())

					filter.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("fails when the id contains non-hex character", func() {
					filter.ID = "507f1f77bcf86cd799439011z"

					filter.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})
			})
		})
	})
})
