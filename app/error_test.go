package app_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"

	"github.com/tidepool-org/platform/app"
)

var _ = Describe("Error", func() {
	Context("Error", func() {
		It("returns a formatted error", func() {
			Expect(app.Error("app", "one").Error()).To(Equal("app: one"))
		})
	})

	Context("Errorf", func() {
		It("returns a formatted error", func() {
			Expect(app.Errorf("app", "%d %s", 2, "two").Error()).To(Equal("app: 2 two"))
		})
	})

	Context("ExtError", func() {
		It("returns a formatted error", func() {
			err := errors.New("error: inner")
			Expect(app.ExtError(err, "app", "three").Error()).To(Equal("app: three; error: inner"))
		})

		It("does not fail when err is nil", func() {
			Expect(app.ExtError(nil, "app", "three").Error()).To(Equal("app: three; app: error is nil"))
		})
	})

	Context("ExtErrorf", func() {
		It("returns a formatted error", func() {
			err := errors.New("error: inner")
			Expect(app.ExtErrorf(err, "app", "%d %s", 4, "four").Error()).To(Equal("app: 4 four; error: inner"))
		})

		It("does not fail when err is nil", func() {
			Expect(app.ExtErrorf(nil, "app", "%d %s", 4, "four").Error()).To(Equal("app: 4 four; app: error is nil"))
		})
	})
})
