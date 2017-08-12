package auth_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/auth"
)

var _ = Describe("Client", func() {
	Context("TidepoolAuthTokenHeaderName", func() {
		It("is the correct header name", func() {
			Expect(auth.TidepoolAuthTokenHeaderName).To(Equal("X-Tidepool-Session-Token"))
		})
	})
})
