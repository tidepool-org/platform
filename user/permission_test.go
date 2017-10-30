package user_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/user"
)

var _ = Describe("Permission", func() {
	Context("OwnerPermission", func() {
		It("has the expected permissions", func() {
			Expect(user.OwnerPermission).To(Equal("root"))
		})
	})

	Context("CustodianPermission", func() {
		It("has the expected permissions", func() {
			Expect(user.CustodianPermission).To(Equal("custodian"))
		})
	})

	Context("UploadPermission", func() {
		It("has the expected permissions", func() {
			Expect(user.UploadPermission).To(Equal("upload"))
		})
	})

	Context("ViewPermission", func() {
		It("has the expected permissions", func() {
			Expect(user.ViewPermission).To(Equal("view"))
		})
	})
})
