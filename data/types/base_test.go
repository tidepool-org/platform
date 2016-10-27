package types_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
)

// TODO: Finish tests

var _ = Describe("Base", func() {
	Context("with new base", func() {
		var testBase *types.Base

		BeforeEach(func() {
			testBase = &types.Base{}
			testBase.Init()
		})

		Context("IdentityFields", func() {
			var userID string
			var deviceID string

			BeforeEach(func() {
				userID = app.NewID()
				deviceID = app.NewID()
				testBase.UserID = userID
				testBase.DeviceID = &deviceID
				testBase.Time = app.StringAsPointer("2016-09-06T13:45:58-07:00")
				testBase.Type = "testBase"
			})

			It("returns error if user id is empty", func() {
				testBase.UserID = ""
				identityFields, err := testBase.IdentityFields()
				Expect(err).To(MatchError("base: user id is empty"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns error if device id is missing", func() {
				testBase.DeviceID = nil
				identityFields, err := testBase.IdentityFields()
				Expect(err).To(MatchError("base: device id is missing"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns error if device id is empty", func() {
				testBase.DeviceID = app.StringAsPointer("")
				identityFields, err := testBase.IdentityFields()
				Expect(err).To(MatchError("base: device id is empty"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns error if time is missing", func() {
				testBase.Time = nil
				identityFields, err := testBase.IdentityFields()
				Expect(err).To(MatchError("base: time is missing"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns error if time is empty", func() {
				testBase.Time = app.StringAsPointer("")
				identityFields, err := testBase.IdentityFields()
				Expect(err).To(MatchError("base: time is empty"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns error if type is empty", func() {
				testBase.Type = ""
				identityFields, err := testBase.IdentityFields()
				Expect(err).To(MatchError("base: type is empty"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns the expected identity fields", func() {
				identityFields, err := testBase.IdentityFields()
				Expect(err).ToNot(HaveOccurred())
				Expect(identityFields).To(Equal([]string{userID, deviceID, "2016-09-06T13:45:58-07:00", "testBase"}))
			})
		})

		Context("with deduplicator descriptor", func() {
			var testDeduplicatorDescriptor *data.DeduplicatorDescriptor

			BeforeEach(func() {
				testDeduplicatorDescriptor = &data.DeduplicatorDescriptor{Name: app.NewID(), Hash: app.NewID()}
			})

			Context("DeduplicatorDescriptor", func() {
				It("gets the deduplicator descriptor", func() {
					testBase.Deduplicator = testDeduplicatorDescriptor
					Expect(testBase.DeduplicatorDescriptor()).To(Equal(testDeduplicatorDescriptor))
				})
			})

			Context("SetDeduplicatorDescriptor", func() {
				It("sets the deduplicator descriptor", func() {
					testBase.SetDeduplicatorDescriptor(testDeduplicatorDescriptor)
					Expect(testBase.Deduplicator).To(Equal(testDeduplicatorDescriptor))
				})
			})
		})
	})
})
