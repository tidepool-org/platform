package app_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/app"
)

var _ = Describe("App", func() {

	Describe("Error", func() {
		It("returns a formatted error", func() {
			Expect(app.Error("test", "one").Error()).To(Equal("test: one"))
		})
	})

	Describe("Errorf", func() {
		It("returns a formatted error", func() {
			Expect(app.Errorf("test", "%d %s", 2, "two").Error()).To(Equal("test: 2 two"))
		})
	})

	Describe("ExtError", func() {
		It("returns a formatted error", func() {
			err := fmt.Errorf("error: inner")
			Expect(app.ExtError(err, "test", "three").Error()).To(Equal("test: three; error: inner"))
		})
	})

	Describe("ExtErrorf", func() {
		It("returns a formatted error", func() {
			err := fmt.Errorf("error: inner")
			Expect(app.ExtErrorf(err, "test", "%d %s", 4, "four").Error()).To(Equal("test: 4 four; error: inner"))
		})
	})
})
