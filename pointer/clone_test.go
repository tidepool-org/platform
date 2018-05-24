package pointer_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Clone", func() {
	Context("CloneBool", func() {
		It("returns nil if the source is nil", func() {
			Expect(pointer.CloneBool(nil)).To(BeNil())
		})

		It("returns a clone of the specified source", func() {
			source := test.RandomBool()
			result := pointer.CloneBool(&source)
			Expect(result).ToNot(BeNil())
			Expect(result).ToNot(BeIdenticalTo(&source))
			Expect(*result).To(Equal(source))
		})
	})

	Context("CloneDuration", func() {
		It("returns nil if the source is nil", func() {
			Expect(pointer.CloneDuration(nil)).To(BeNil())
		})

		It("returns a clone of the specified source", func() {
			source := test.RandomDuration()
			result := pointer.CloneDuration(&source)
			Expect(result).ToNot(BeNil())
			Expect(result).ToNot(BeIdenticalTo(&source))
			Expect(*result).To(Equal(source))
		})
	})

	Context("CloneFloat64", func() {
		It("returns nil if the source is nil", func() {
			Expect(pointer.CloneFloat64(nil)).To(BeNil())
		})

		It("returns a clone of the specified source", func() {
			source := test.RandomFloat64()
			result := pointer.CloneFloat64(&source)
			Expect(result).ToNot(BeNil())
			Expect(result).ToNot(BeIdenticalTo(&source))
			Expect(*result).To(Equal(source))
		})
	})

	Context("CloneInt", func() {
		It("returns nil if the source is nil", func() {
			Expect(pointer.CloneInt(nil)).To(BeNil())
		})

		It("returns a clone of the specified source", func() {
			source := test.RandomInt()
			result := pointer.CloneInt(&source)
			Expect(result).ToNot(BeNil())
			Expect(result).ToNot(BeIdenticalTo(&source))
			Expect(*result).To(Equal(source))
		})
	})

	Context("CloneString", func() {
		It("returns nil if the source is nil", func() {
			Expect(pointer.CloneString(nil)).To(BeNil())
		})

		It("returns a clone of the specified source", func() {
			source := test.RandomString()
			result := pointer.CloneString(&source)
			Expect(result).ToNot(BeNil())
			Expect(result).ToNot(BeIdenticalTo(&source))
			Expect(*result).To(Equal(source))
		})
	})

	Context("CloneStringArray", func() {
		It("returns nil if the source is nil", func() {
			Expect(pointer.CloneStringArray(nil)).To(BeNil())
		})

		It("returns a clone of the specified nil source", func() {
			var source []string
			result := pointer.CloneStringArray(&source)
			Expect(result).ToNot(BeNil())
			Expect(result).ToNot(BeIdenticalTo(&source))
			Expect(*result).To(Equal(source))
		})

		It("returns a clone of the specified empty source", func() {
			source := []string{}
			result := pointer.CloneStringArray(&source)
			Expect(result).ToNot(BeNil())
			Expect(result).ToNot(BeIdenticalTo(&source))
			Expect(*result).To(Equal(source))
		})

		It("returns a clone of the specified source", func() {
			source := test.RandomStringArray()
			result := pointer.CloneStringArray(&source)
			Expect(result).ToNot(BeNil())
			Expect(result).ToNot(BeIdenticalTo(&source))
			Expect(*result).To(Equal(source))
		})
	})

	Context("CloneTime", func() {
		It("returns nil if the source is nil", func() {
			Expect(pointer.CloneTime(nil)).To(BeNil())
		})

		It("returns a clone of the specified source", func() {
			source := test.RandomTime()
			result := pointer.CloneTime(&source)
			Expect(result).ToNot(BeNil())
			Expect(result).ToNot(BeIdenticalTo(&source))
			Expect(*result).To(Equal(source))
		})
	})
})
