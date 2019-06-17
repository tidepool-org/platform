package aws_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/aws"
)

var _ = Describe("AWS", func() {
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
