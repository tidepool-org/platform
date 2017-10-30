package id_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/id"
)

var _ = Describe("ID", func() {
	Context("New", func() {
		It("returns a string of 32 lowercase hexidecimal characters", func() {
			Expect(id.New()).To(MatchRegexp("[0-9a-f]{32}"))
		})

		It("returns different IDs for each invocation", func() {
			Expect(id.New()).ToNot(Equal(id.New()))
		})
	})
})
