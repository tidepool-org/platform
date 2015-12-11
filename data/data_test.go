package data_test

import (
	. "github.com/tidepool-org/platform/data"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Data", func() {
	Context("with no parameters", func() {
		It("should return data", func() {
			Expect(GetData()).To(Equal("data"))
		})
	})
})
