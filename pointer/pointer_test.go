package pointer_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"time"

	"github.com/tidepool-org/platform/pointer"
)

var _ = Describe("Pointer", func() {
	Context("Boolean", func() {
		It("returns a pointer to the specified source", func() {
			source := true
			Expect(*pointer.Boolean(source)).To(Equal(source))
		})
	})

	Context("Duration", func() {
		It("returns a pointer to the specified source", func() {
			source := 24 * time.Hour
			Expect(*pointer.Duration(source)).To(Equal(source))
		})
	})

	Context("Float", func() {
		It("returns a pointer to the specified source", func() {
			source := 123.45
			Expect(*pointer.Float(source)).To(Equal(source))
		})
	})

	Context("Integer", func() {
		It("returns a pointer to the specified source", func() {
			source := 123
			Expect(*pointer.Integer(source)).To(Equal(source))
		})
	})

	Context("String", func() {
		It("returns a pointer to the specified source", func() {
			source := "abc"
			Expect(*pointer.String(source)).To(Equal(source))
		})
	})

	Context("StringArray", func() {
		It("returns a pointer to the specified nil source", func() {
			var source []string
			Expect(*pointer.StringArray(source)).To(Equal(source))
		})

		It("returns a pointer to the specified non-nil, empty source", func() {
			source := []string{"abc", "def"}
			Expect(*pointer.StringArray(source)).To(Equal(source))
		})

		It("returns a pointer to the specified non-nil, non-empty source", func() {
			source := []string{"abc", "def"}
			Expect(*pointer.StringArray(source)).To(Equal(source))
		})
	})
})
