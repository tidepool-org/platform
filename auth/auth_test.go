package auth_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/auth"
)

var _ = Describe("Client", func() {
	Context("TidepoolSessionTokenHeaderKey", func() {
		It("is the correct header name", func() {
			Expect(auth.TidepoolSessionTokenHeaderKey).To(Equal("X-Tidepool-Session-Token"))
		})
	})
})
