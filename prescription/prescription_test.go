package prescription_test

import (
	"time"

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
				var err error
				prescr, err = prescription.NewPrescription(userID, revisionCreate)
				Expect(err).ToNot(HaveOccurred())
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
})
