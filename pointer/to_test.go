package pointer_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("To", func() {
	Context("ToBool", func() {
		It("returns the zero value if the pointer is nil", func() {
			Expect(pointer.ToBool(nil)).To(BeFalse())
		})

		It("returns the dereferenced pointer", func() {
			value := test.RandomBool()
			result := pointer.ToBool(&value)
			Expect(result).To(Equal(value))
		})
	})

	Context("ToDuration", func() {
		It("returns the zero value if the pointer is nil", func() {
			Expect(pointer.ToDuration(nil)).To(Equal(time.Duration(0)))
		})

		It("returns the dereferenced pointer", func() {
			value := test.RandomDuration()
			result := pointer.ToDuration(&value)
			Expect(result).To(Equal(value))
		})
	})

	Context("ToFloat64", func() {
		It("returns the zero value if the pointer is nil", func() {
			Expect(pointer.ToFloat64(nil)).To(Equal(0.))
		})

		It("returns the dereferenced pointer", func() {
			value := test.RandomFloat64()
			result := pointer.ToFloat64(&value)
			Expect(result).To(Equal(value))
		})
	})

	Context("ToInt", func() {
		It("returns the zero value if the pointer is nil", func() {
			Expect(pointer.ToInt(nil)).To(Equal(0))
		})

		It("returns the dereferenced pointer", func() {
			value := test.RandomInt()
			result := pointer.ToInt(&value)
			Expect(result).To(Equal(value))
		})
	})

	Context("ToString", func() {
		It("returns the zero value if the pointer is nil", func() {
			Expect(pointer.ToString(nil)).To(BeEmpty())
		})

		It("returns the dereferenced pointer", func() {
			value := test.RandomString()
			result := pointer.ToString(&value)
			Expect(result).To(Equal(value))
		})
	})

	Context("ToStringArray", func() {
		It("returns the zero value if the pointer is nil", func() {
			Expect(pointer.ToStringArray(nil)).To(BeNil())
		})

		It("returns the dereferenced pointer with nil value", func() {
			var value []string
			result := pointer.ToStringArray(&value)
			Expect(result).To(Equal(value))
		})

		It("returns the dereferenced pointer with empty value", func() {
			value := []string{}
			result := pointer.ToStringArray(&value)
			Expect(result).To(Equal(value))
		})

		It("returns the dereferenced pointer with non-empty value", func() {
			value := test.RandomStringArray()
			result := pointer.ToStringArray(&value)
			Expect(result).To(Equal(value))
		})
	})

	Context("ToTime", func() {
		It("returns the zero value if the pointer is nil", func() {
			Expect(pointer.ToTime(nil)).To(Equal(time.Time{}))
		})

		It("returns the dereferenced pointer", func() {
			value := test.RandomTime()
			result := pointer.ToTime(&value)
			Expect(result).To(Equal(value))
		})
	})
})
