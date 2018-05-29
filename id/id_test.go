package id_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("ID", func() {
	Context("New", func() {
		It("returns an error if length is negative", func() {
			value, err := id.New(-1)
			Expect(err).To(MatchError("length is invalid"))
			Expect(value).To(BeEmpty())
		})

		It("returns an error if length is zero", func() {
			value, err := id.New(0)
			Expect(err).To(MatchError("length is invalid"))
			Expect(value).To(BeEmpty())
		})

		It("returns a 2-character hexidecimal string if the length is one", func() {
			Expect(id.New(1)).To(MatchRegexp("^[0-9a-f]{2}$"))
		})

		It("returns a 10-character hexidecimal string if the length is five", func() {
			Expect(id.New(5)).To(MatchRegexp("^[0-9a-f]{10}$"))
		})

		It("returns a 32-character hexidecimal string if the length is sixteen", func() {
			Expect(id.New(16)).To(MatchRegexp("^[0-9a-f]{32}$"))
		})

		It("returns different IDs for each invocation", func() {
			valueOne, errOne := id.New(16)
			Expect(errOne).ToNot(HaveOccurred())
			Expect(valueOne).ToNot(BeEmpty())
			valueTwo, errTwo := id.New(16)
			Expect(errTwo).ToNot(HaveOccurred())
			Expect(valueTwo).ToNot(BeEmpty())
			Expect(valueOne).ToNot(Equal(valueTwo))
		})
	})

	Context("Must", func() {
		It("panics if the error is not nil", func() {
			Expect(func() { id.Must(test.NewString(32, test.CharsetHexidecimalLowercase), errorsTest.NewError()) }).To(Panic())
		})

		It("returns the value if the error is nil", func() {
			value := test.NewString(32, test.CharsetHexidecimalLowercase)
			Expect(id.Must(value, nil)).To(Equal(value))
		})
	})
})
