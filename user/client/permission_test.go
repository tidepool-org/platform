package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/user/client"
)

var _ = Describe("Permission", func() {
	Context("OwnerPermission", func() {
		It("has the expected permissions", func() {
			Expect(client.OwnerPermission).To(Equal("root"))
		})
	})

	Context("CustodianPermission", func() {
		It("has the expected permissions", func() {
			Expect(client.CustodianPermission).To(Equal("custodian"))
		})
	})

	Context("UploadPermission", func() {
		It("has the expected permissions", func() {
			Expect(client.UploadPermission).To(Equal("upload"))
		})
	})

	Context("ViewPermission", func() {
		It("has the expected permissions", func() {
			Expect(client.ViewPermission).To(Equal("view"))
		})
	})
})
