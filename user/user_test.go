package user_test

import (
	. "github.com/tidepool-org/platform/user"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("User", func() {
	Context("with no parameters", func() {
		It("should return user", func() {
			Expect(GetUser()).To(Equal("user"))
		})
	})
})
