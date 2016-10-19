package user_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/user"
)

var _ = Describe("Role", func() {
	Context("ClinicRole", func() {
		It("exists", func() {
			Expect(user.ClinicRole).To(Equal("clinic"))
		})
	})
})
