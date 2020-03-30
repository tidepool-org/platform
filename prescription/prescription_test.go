package prescription_test

import (
	"fmt"
	"time"

	"github.com/globalsign/mgo/bson"
	"syreclabs.com/go/faker"

	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/structure/validator"
	userTest "github.com/tidepool-org/platform/user/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/prescription/test"
	"github.com/tidepool-org/platform/user"

	"github.com/tidepool-org/platform/prescription"
)

var _ = Describe("Prescription", func() {
	Context("With a submitted revision", func() {
		var revisionCreate *prescription.RevisionCreate
		var userID string

		BeforeEach(func() {
			revisionCreate = test.RandomRevisionCreate()
			revisionCreate.State = prescription.StateSubmitted
			userID = user.NewID()
		})

		Context("Create new prescription", func() {
			var prescr *prescription.Prescription

			BeforeEach(func() {
				prescr = prescription.NewPrescription(userID, revisionCreate)
				Expect(prescr).ToNot(BeNil())
			})

			It("creates a non-empty id", func() {
				Expect(prescr.ID).ToNot(BeEmpty())
			})

			It("does not set the patientId", func() {
				Expect(prescr.PatientID).To(BeEmpty())
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
				Expect(prescr.CreatedUserID).To(Equal(userID))
			})

			It("sets the prescriber user id correctly", func() {
				Expect(prescr.PrescriberUserID).To(Equal(userID))
			})

			It("sets the created time correctly", func() {
				Expect(prescr.CreatedTime).To(BeTemporally("~", time.Now()))
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
		})
	})

	Describe("Filter", func() {

		Describe("Validate", func() {
			When("current user is NOT a clinician", func() {
				var usr *user.User
				var filter *prescription.Filter
				var validate structure.Validator

				BeforeEach(func() {
					usr = userTest.RandomUser()
					usr.Roles = &[]string{}

					var err error
					filter, err = prescription.NewFilter(usr)
					Expect(err).ToNot(HaveOccurred())

					validate = validator.New()
					Expect(validate.Validate(filter)).ToNot(HaveOccurred())
				})

				It("fails when patient id is not same as current user id", func() {
					patientID := userTest.RandomID()
					Expect(patientID).ToNot(Equal(filter.PatientID))
					filter.PatientID = patientID

					filter.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
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

				It("doesn't fail when the state is reviewed", func() {
					filter.State = prescription.StateReviewed

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
					filter.ID = bson.NewObjectId().Hex()

					filter.Validate(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())
				})

				It("fails when the id is 13 hex characters", func() {
					filter.ID = fmt.Sprintf("%sa", bson.NewObjectId().Hex())

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
				var usr *user.User
				var filter *prescription.Filter
				var validate structure.Validator

				BeforeEach(func() {
					usr = userTest.RandomUser()
					usr.Roles = &[]string{user.RoleClinic}

					var err error
					filter, err = prescription.NewFilter(usr)
					Expect(err).ToNot(HaveOccurred())

					validate = validator.New()
					Expect(validate.Validate(filter)).ToNot(HaveOccurred())
				})

				It("fails when patient id is invalid", func() {
					filter.PatientID = "invalid"

					filter.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("doesn't fail when patient id invalid", func() {
					filter.PatientID = userTest.RandomID()

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

				It("doesn't fail when the state is reviewed", func() {
					filter.State = prescription.StateReviewed

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
					filter.ID = bson.NewObjectId().Hex()

					filter.Validate(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())
				})

				It("fails when the id is 13 hex characters", func() {
					filter.ID = fmt.Sprintf("%sa", bson.NewObjectId().Hex())

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
