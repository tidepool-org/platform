package twiist_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/twiist"
)

var _ = Describe("ServiceAccountAuthorizer", func() {
	var (
		authorizer twiist.ServiceAccountAuthorizer
		err        error
	)

	Context("When no service accounts are configured", func() {
		BeforeEach(func() {
			// Set an empty list of service accounts
			GinkgoT().Setenv("TIDEPOOL_TWIIST_SERVICE_ACCOUNT_IDS", "")

			authorizer, err = twiist.NewServiceAccountAuthorizer()
		})

		It("should create an authorizer without error", func() {
			Expect(err).To(BeNil())
			Expect(authorizer).To(Not(BeNil()))
		})

		It("should not authorize any user", func() {
			Expect(authorizer.IsAuthorized("any-user-id")).To(BeFalse())
		})
	})

	Context("When service accounts are configured", func() {
		BeforeEach(func() {
			// Set specific service account IDs
			GinkgoT().Setenv("TIDEPOOL_TWIIST_SERVICE_ACCOUNT_IDS", "service-account-1,service-account-2")

			authorizer, err = twiist.NewServiceAccountAuthorizer()
		})

		It("should create an authorizer without error", func() {
			Expect(err).To(BeNil())
			Expect(authorizer).To(Not(BeNil()))
		})

		It("should authorize configured service accounts", func() {
			Expect(authorizer.IsAuthorized("service-account-1")).To(BeTrue())
			Expect(authorizer.IsAuthorized("service-account-2")).To(BeTrue())
		})

		It("should not authorize unconfigured accounts", func() {
			Expect(authorizer.IsAuthorized("service-account-3")).To(BeFalse())
			Expect(authorizer.IsAuthorized("random-user")).To(BeFalse())
		})
	})
})
