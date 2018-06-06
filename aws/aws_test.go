package aws_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/aws"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("AWS", func() {
	Context("String", func() {
		It("returns a pointer to the specified value", func() {
			value := test.RandomString()
			result := aws.String(value)
			Expect(result).ToNot(BeNil())
			Expect(*result).To(Equal(value))
		})
	})

	Context("NewWriteAtBuffer", func() {
		It("returns successfully with nil bytes", func() {
			Expect(aws.NewWriteAtBuffer(nil)).ToNot(BeNil())
		})

		It("returns successfully with empty bytes", func() {
			Expect(aws.NewWriteAtBuffer([]byte{})).ToNot(BeNil())
		})

		It("returns successfully with non-empty bytes", func() {
			Expect(aws.NewWriteAtBuffer([]byte("test"))).ToNot(BeNil())
		})
	})
})
