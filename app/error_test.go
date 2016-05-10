package app_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/app"
)

var _ = Describe("Error", func() {

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
			err := errors.New("error: inner")
			Expect(app.ExtError(err, "test", "three").Error()).To(Equal("test: three; error: inner"))
		})

		It("does not fail when err is nil", func() {
			Expect(app.ExtError(nil, "test", "three").Error()).To(Equal("test: three; app: error is nil"))
		})
	})

	Describe("ExtErrorf", func() {
		It("returns a formatted error", func() {
			err := errors.New("error: inner")
			Expect(app.ExtErrorf(err, "test", "%d %s", 4, "four").Error()).To(Equal("test: 4 four; error: inner"))
		})

		It("does not fail when err is nil", func() {
			Expect(app.ExtErrorf(nil, "test", "%d %s", 4, "four").Error()).To(Equal("test: 4 four; app: error is nil"))
		})
	})
})
