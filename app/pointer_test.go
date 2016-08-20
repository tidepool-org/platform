package app_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"time"

	"github.com/tidepool-org/platform/app"
)

var _ = Describe("Pointer", func() {
	Context("StringAsPointer", func() {
		It("returns a pointer to the specified source", func() {
			source := "abc"
			Expect(*app.StringAsPointer(source)).To(Equal(source))
		})
	})

	Context("StringArrayAsPointer", func() {
		It("returns a pointer to the specified nil source", func() {
			var source []string
			Expect(*app.StringArrayAsPointer(source)).To(Equal(source))
		})

		It("returns a pointer to the specified non-nil source", func() {
			source := []string{"abc", "def"}
			Expect(*app.StringArrayAsPointer(source)).To(Equal(source))
		})
	})

	Context("IntegerAsPointer", func() {
		It("returns a pointer to the specified source", func() {
			source := 123
			Expect(*app.IntegerAsPointer(source)).To(Equal(source))
		})
	})

	Context("DurationAsPointer", func() {
		It("returns a pointer to the specified source", func() {
			source := time.Hour
			Expect(*app.DurationAsPointer(source)).To(Equal(source))
		})
	})
})
