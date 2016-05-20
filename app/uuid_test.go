package app_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/app"
)

var _ = Describe("UUID", func() {
	Context("NewUUID", func() {
		It("returns a string of 32 lowercase hexidecimal characters", func() {
			Expect(app.NewUUID()).To(MatchRegexp("[0-9a-f]{32}"))
		})

		It("returns different UUIDs for each invocation", func() {
			Expect(app.NewUUID()).ToNot(Equal(app.NewUUID()))
		})
	})
})
