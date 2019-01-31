package multipart_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	imageMultipart "github.com/tidepool-org/platform/image/multipart"
)

var _ = Describe("FormEncoder", func() {
	Context("NewFormEncoder", func() {
		It("returns successfully", func() {
			Expect(imageMultipart.NewFormEncoder()).ToNot(BeNil())
		})
	})
})
