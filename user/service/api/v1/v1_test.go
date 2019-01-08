package v1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	userServiceApiV1 "github.com/tidepool-org/platform/user/service/api/v1"
)

var _ = Describe("V1", func() {
	Context("Routes", func() {
		It("returns the correct routes", func() {
			Expect(userServiceApiV1.Routes()).ToNot(BeEmpty())
		})
	})
})
