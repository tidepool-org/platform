package permission_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/permission"
)

var _ = Describe("permission", func() {
	Context("Owner", func() {
		It("has the expected permissions", func() {
			Expect(permission.Owner).To(Equal("root"))
		})
	})

	Context("Custodian", func() {
		It("has the expected permissions", func() {
			Expect(permission.Custodian).To(Equal("custodian"))
		})
	})

	Context("Write", func() {
		It("has the expected permissions", func() {
			Expect(permission.Write).To(Equal("upload"))
		})
	})

	Context("Read", func() {
		It("has the expected permissions", func() {
			Expect(permission.Read).To(Equal("view"))
		})
	})

	Context("Follow", func() {
		It("has the expected permissions", func() {
			Expect(permission.Follow).To(Equal("follow"))
		})
	})
})
