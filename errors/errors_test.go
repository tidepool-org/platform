package errors_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"

	"github.com/tidepool-org/platform/errors"
)

var _ = Describe("Errors", func() {
	Context("New", func() {
		It("returns a formatted error", func() {
			Expect(errors.New("errors", "one").Error()).To(Equal("errors: one"))
		})
	})

	Context("Newf", func() {
		It("returns a formatted error", func() {
			Expect(errors.Newf("errors", "%d %s", 2, "two").Error()).To(Equal("errors: 2 two"))
		})
	})

	Context("Wrap", func() {
		It("returns a formatted error", func() {
			err := fmt.Errorf("inner")
			Expect(errors.Wrap(err, "errors", "three").Error()).To(Equal("errors: three; inner"))
		})

		It("does not fail when err is nil", func() {
			Expect(errors.Wrap(nil, "errors", "three").Error()).To(Equal("errors: three; errors: error is nil"))
		})
	})

	Context("Wrapf", func() {
		It("returns a formatted error", func() {
			err := fmt.Errorf("inner")
			Expect(errors.Wrapf(err, "errors", "%d %s", 4, "four").Error()).To(Equal("errors: 4 four; inner"))
		})

		It("does not fail when err is nil", func() {
			Expect(errors.Wrapf(nil, "errors", "%d %s", 4, "four").Error()).To(Equal("errors: 4 four; errors: error is nil"))
		})
	})
})
