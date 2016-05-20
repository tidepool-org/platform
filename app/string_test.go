package app_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/app"
)

var _ = Describe("String", func() {
	Context("FirstStringNotEmpty", func() {
		It("returns an empty string with no arguments", func() {
			Expect(app.FirstStringNotEmpty()).To(Equal(""))
		})

		It("returns an empty string with all empty arguments", func() {
			Expect(app.FirstStringNotEmpty("", "", "")).To(Equal(""))
		})

		It("returns the first empty string with only one argument", func() {
			Expect(app.FirstStringNotEmpty("pixie")).To(Equal("pixie"))
		})

		It("returns the first empty string with multiple arguments", func() {
			Expect(app.FirstStringNotEmpty("", "", "goblin", "")).To(Equal("goblin"))
		})
	})
})
