package auth_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/auth"
	authTest "github.com/tidepool-org/platform/auth/test"
)

var _ = Describe("DeviceAuthorization", func() {
	Context("NewDeviceAuthorizationID", func() {
		It("generates a string of 16 lowercase hexadecimal characters", func() {
			Expect(auth.NewDeviceAuthorizationID()).To(MatchRegexp("^[0-9a-f]{16}$"))
		})

		It("generates a different string on each invocation", func() {
			Expect(auth.NewDeviceAuthorizationID()).To(Not(Equal(auth.NewDeviceAuthorizationID())))
		})
	})

	Context("NewDeviceAuthorizationToken", func() {
		It("generates a string of 32 lowercase hexadecimal characters", func() {
			Expect(auth.NewDeviceAuthorizationToken()).To(MatchRegexp("^[0-9a-f]{32}$"))
		})

		It("generates a different string on each invocation", func() {
			Expect(auth.NewDeviceAuthorizationToken()).To(Not(Equal(auth.NewDeviceAuthorizationToken())))
		})
	})

	Context("NewDeviceAuthorization", func() {
		var userID string
		var create *auth.DeviceAuthorizationCreate
		var authz *auth.DeviceAuthorization

		BeforeEach(func() {
			create = auth.NewDeviceAuthorizationCreate()
			create.DevicePushToken = authTest.RandomDevicePushToken()

			userID = authTest.RandomUserID()
			authz, _ = auth.NewDeviceAuthorization(userID, create)
		})

		It("creates a valid id", func() {
			Expect(authz.ID).To(MatchRegexp("^[0-9a-f]{16}$"))
		})

		It("creates a valid token", func() {
			Expect(authz.Token).To(MatchRegexp("^[0-9a-f]{32}$"))
		})

		It("sets the correct user id", func() {
			Expect(authz.UserID).To(Equal(userID))
		})

		It("sets the correct device push token", func() {
			Expect(authz.DevicePushToken).To(Equal(create.DevicePushToken))
		})

		It("sets the status to pending", func() {
			Expect(authz.Status).To(Equal(auth.DeviceAuthorizationPending))
		})
	})

	Context("UpdateBundleID", func() {
		var authz *auth.DeviceAuthorization

		BeforeEach(func() {
			authz = &auth.DeviceAuthorization{}
		})

		It("doesn't return an error with loop bundle id", func() {
			Expect(authz.UpdateBundleID("org.tidepool.Loop")).To(Succeed())
		})

		It("doesn't return an error with loop bundle id", func() {
			Expect(authz.UpdateBundleID("75U4X84TEG.org.tidepool.Loop")).To(Succeed())
		})

		It("returns an error with invalid bundle", func() {
			Expect(authz.UpdateBundleID("com.todd.Loop")).To(MatchError("bundle id is not valid"))
		})

		It("return an error if bundle id is already set", func() {
			Expect(authz.UpdateBundleID("org.tidepool.Loop")).To(Succeed())
			Expect(authz.UpdateBundleID("75U4X84TEG.org.tidepool.Loop")).To(MatchError("bundle id is already set"))
		})
	})

	Context("UpdateStatus", func() {
		var authz *auth.DeviceAuthorization

		BeforeEach(func() {
			authz = &auth.DeviceAuthorization{}
		})

		It("doesn't return an error for status successful", func() {
			Expect(authz.UpdateStatus(auth.DeviceAuthorizationSuccessful)).To(Succeed())
		})

		It("doesn't return an error for status failed", func() {
			Expect(authz.UpdateStatus(auth.DeviceAuthorizationFailed)).To(Succeed())
		})

		It("doesn't return an error for status expired", func() {
			Expect(authz.UpdateStatus(auth.DeviceAuthorizationExpired)).To(Succeed())
		})

		It("returns an error for for invalid status", func() {
			Expect(authz.UpdateStatus("invalid-status")).To(MatchError("status is not valid"))
		})

		It("returns an error if it's already set to successful", func() {
			Expect(authz.UpdateStatus(auth.DeviceAuthorizationSuccessful)).To(Succeed())
			Expect(authz.UpdateStatus(auth.DeviceAuthorizationFailed)).To(MatchError("cannot update status of a completed device authorization"))
		})

		It("returns an error if it's already set to failed", func() {
			Expect(authz.UpdateStatus(auth.DeviceAuthorizationFailed)).To(Succeed())
			Expect(authz.UpdateStatus(auth.DeviceAuthorizationSuccessful)).To(MatchError("cannot update status of a completed device authorization"))
		})

		It("returns an error if it's already set to expired", func() {
			Expect(authz.UpdateStatus(auth.DeviceAuthorizationExpired)).To(Succeed())
			Expect(authz.UpdateStatus(auth.DeviceAuthorizationSuccessful)).To(MatchError("cannot update status of a completed device authorization"))
		})
	})

	Context("IsCompleted", func() {
		var authz *auth.DeviceAuthorization

		BeforeEach(func() {
			authz = &auth.DeviceAuthorization{}
		})

		It("returns true if it's successful", func() {
			_ = authz.UpdateStatus(auth.DeviceAuthorizationSuccessful)
			Expect(authz.IsCompleted()).To(BeTrue())
		})

		It("returns true if it's failed", func() {
			_ = authz.UpdateStatus(auth.DeviceAuthorizationFailed)
			Expect(authz.IsCompleted()).To(BeTrue())
		})

		It("returns true if it's expired", func() {
			_ = authz.UpdateStatus(auth.DeviceAuthorizationExpired)
			Expect(authz.IsCompleted()).To(BeTrue())
		})

		It("returns false if it's pending", func() {
			_ = authz.UpdateStatus(auth.DeviceAuthorizationPending)
			Expect(authz.IsCompleted()).To(BeFalse())
		})
	})
})
