package service_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/service"
)

var _ = Describe("Errors", func() {
	Context("InternalServerError", func() {
		It("matches the expected error", func() {
			Expect(service.ErrorInternalServerFailure()).To(Equal(
				&service.Error{
					Code:   "internal-server-failure",
					Status: 500,
					Title:  "internal server failure",
					Detail: "Internal server failure",
				}))
		})
	})
})
