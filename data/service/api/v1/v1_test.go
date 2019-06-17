package v1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	dataServiceApiV1 "github.com/tidepool-org/platform/data/service/api/v1"
)

var _ = Describe("V1", func() {
	Context("Routes", func() {
		It("returns the correct routes", func() {
			Expect(dataServiceApiV1.Routes()).ToNot(BeEmpty())
		})
	})
})
