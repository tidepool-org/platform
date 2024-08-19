package pointer_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("From", func() {
	Context("FromBool", func() {
		It("returns a pointer to the specified value", func() {
			value := test.RandomBool()
			result := pointer.FromBool(value)
			Expect(result).ToNot(BeNil())
			Expect(*result).To(Equal(value))
		})
	})

	Context("FromDuration", func() {
		It("returns a pointer to the specified value", func() {
			value := test.RandomDuration()
			result := pointer.FromDuration(value)
			Expect(result).ToNot(BeNil())
			Expect(*result).To(Equal(value))
		})
	})

	Context("FromFloat64", func() {
		It("returns a pointer to the specified value", func() {
			value := test.RandomFloat64()
			result := pointer.FromFloat64(value)
			Expect(result).ToNot(BeNil())
			Expect(*result).To(Equal(value))
		})
	})

	Context("FromInt", func() {
		It("returns a pointer to the specified value", func() {
			value := test.RandomInt()
			result := pointer.FromInt(value)
			Expect(result).ToNot(BeNil())
			Expect(*result).To(Equal(value))
		})
	})

	Context("FromString", func() {
		It("returns a pointer to the specified value", func() {
			value := test.RandomString()
			result := pointer.FromString(value)
			Expect(result).ToNot(BeNil())
			Expect(*result).To(Equal(value))
		})
	})

	Context("FromStringArray", func() {
		It("returns a pointer to the specified nil value", func() {
			var value []string
			result := pointer.FromStringArray(value)
			Expect(result).ToNot(BeNil())
			Expect(*result).To(Equal(value))
		})

		It("returns a pointer to the specified empty value", func() {
			value := []string{}
			result := pointer.FromStringArray(value)
			Expect(result).ToNot(BeNil())
			Expect(*result).To(Equal(value))
		})

		It("returns a pointer to the specified non-empty value", func() {
			value := test.RandomStringArray()
			result := pointer.FromStringArray(value)
			Expect(result).ToNot(BeNil())
			Expect(*result).To(Equal(value))
		})
	})

	Context("FromTime", func() {
		It("returns a pointer to the specified value", func() {
			value := test.RandomTime()
			result := pointer.FromTime(value)
			Expect(result).ToNot(BeNil())
			Expect(*result).To(Equal(value))
		})
	})
})
