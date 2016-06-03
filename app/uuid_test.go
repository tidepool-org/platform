package app_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/app"
)

var _ = Describe("UUID", func() {
	Context("NewUUID", func() {
		It("returns a string of 36 lowercase hexidecimal characters with dashes", func() {
			Expect(app.NewUUID()).To(MatchRegexp("[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}"))
		})

		It("returns different UUIDs for each invocation", func() {
			Expect(app.NewUUID()).ToNot(Equal(app.NewUUID()))
		})
	})

	Context("NewID", func() {
		It("returns a string of 32 lowercase hexidecimal characters", func() {
			Expect(app.NewID()).To(MatchRegexp("[0-9a-f]{32}"))
		})

		It("returns different IDs for each invocation", func() {
			Expect(app.NewID()).ToNot(Equal(app.NewID()))
		})
	})
})
