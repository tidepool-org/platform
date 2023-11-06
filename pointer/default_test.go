package pointer_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Default", func() {
	Context("DefaultBool", func() {
		It("returns a pointer to the default value if the pointer to the value is nil", func() {
			defaultValue := test.RandomBool()
			result := pointer.DefaultBool(nil, defaultValue)
			Expect(result).ToNot(BeNil())
			Expect(*result).To(Equal(defaultValue))
		})

		It("returns the pointer to the value if the pointer to the value is not nil", func() {
			value := test.RandomBool()
			defaultValue := test.RandomBool()
			result := pointer.DefaultBool(&value, defaultValue)
			Expect(result).To(Equal(&value))
		})
	})

	Context("DefaultDuration", func() {
		It("returns a pointer to the default value if the pointer to the value is nil", func() {
			defaultValue := test.RandomDuration()
			result := pointer.DefaultDuration(nil, defaultValue)
			Expect(result).ToNot(BeNil())
			Expect(*result).To(Equal(defaultValue))
		})

		It("returns the pointer to the value if the pointer to the value is not nil", func() {
			value := test.RandomDuration()
			defaultValue := test.RandomDuration()
			result := pointer.DefaultDuration(&value, defaultValue)
			Expect(result).To(Equal(&value))
		})
	})

	Context("DefaultFloat64", func() {
		It("returns a pointer to the default value if the pointer to the value is nil", func() {
			defaultValue := test.RandomFloat64()
			result := pointer.DefaultFloat64(nil, defaultValue)
			Expect(result).ToNot(BeNil())
			Expect(*result).To(Equal(defaultValue))
		})

		It("returns the pointer to the value if the pointer to the value is not nil", func() {
			value := test.RandomFloat64()
			defaultValue := test.RandomFloat64()
			result := pointer.DefaultFloat64(&value, defaultValue)
			Expect(result).To(Equal(&value))
		})
	})

	Context("DefaultInt", func() {
		It("returns a pointer to the default value if the pointer to the value is nil", func() {
			defaultValue := test.RandomInt()
			result := pointer.DefaultInt(nil, defaultValue)
			Expect(result).ToNot(BeNil())
			Expect(*result).To(Equal(defaultValue))
		})

		It("returns the pointer to the value if the pointer to the value is not nil", func() {
			value := test.RandomInt()
			defaultValue := test.RandomInt()
			result := pointer.DefaultInt(&value, defaultValue)
			Expect(result).To(Equal(&value))
		})
	})

	Context("DefaultString", func() {
		It("returns a pointer to the default value if the pointer to the value is nil", func() {
			defaultValue := test.RandomString()
			result := pointer.DefaultString(nil, defaultValue)
			Expect(result).ToNot(BeNil())
			Expect(*result).To(Equal(defaultValue))
		})

		It("returns the pointer to the value if the pointer to the value is not nil", func() {
			value := test.RandomString()
			defaultValue := test.RandomString()
			result := pointer.DefaultString(&value, defaultValue)
			Expect(result).To(Equal(&value))
		})
	})

	Context("DefaultStringArray", func() {
		It("returns a pointer to the default value if the pointer to the value is nil", func() {
			defaultValue := test.RandomStringArray()
			result := pointer.DefaultStringArray(nil, defaultValue)
			Expect(result).ToNot(BeNil())
			Expect(*result).To(Equal(defaultValue))
		})

		It("returns the pointer to the value if the pointer to the value is not nil with nil value", func() {
			var value []string
			defaultValue := test.RandomStringArray()
			result := pointer.DefaultStringArray(&value, defaultValue)
			Expect(result).To(Equal(&value))
		})

		It("returns the pointer to the value if the pointer to the value is not nil with empty value", func() {
			value := []string{}
			defaultValue := test.RandomStringArray()
			result := pointer.DefaultStringArray(&value, defaultValue)
			Expect(result).To(Equal(&value))
		})

		It("returns the pointer to the value if the pointer to the value is not nil with non-empty value", func() {
			value := test.RandomStringArray()
			defaultValue := test.RandomStringArray()
			result := pointer.DefaultStringArray(&value, defaultValue)
			Expect(result).To(Equal(&value))
		})
	})

	Context("DefaultTime", func() {
		It("returns a pointer to the default value if the pointer to the value is nil", func() {
			defaultValue := test.RandomTime()
			result := pointer.DefaultTime(nil, defaultValue)
			Expect(result).ToNot(BeNil())
			Expect(*result).To(Equal(defaultValue))
		})

		It("returns the pointer to the value if the pointer to the value is not nil", func() {
			value := test.RandomTime()
			defaultValue := test.RandomTime()
			result := pointer.DefaultTime(&value, defaultValue)
			Expect(result).To(Equal(&value))
		})
	})
})
