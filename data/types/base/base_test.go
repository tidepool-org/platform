package base_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/base"
)

// TODO: Finish tests

var _ = Describe("Base", func() {
	Context("with new base", func() {
		var testBase *base.Base

		BeforeEach(func() {
			testBase = &base.Base{}
			testBase.Init()
		})

		Context("with deduplicator descriptor", func() {
			var testDeduplicatorDescriptor *data.DeduplicatorDescriptor

			BeforeEach(func() {
				testDeduplicatorDescriptor = &data.DeduplicatorDescriptor{Name: app.NewID(), Hash: app.NewID()}
			})

			Context("DeduplicatorDescriptor", func() {
				It("gets the deduplicator descriptor", func() {
					testBase.Deduplicator = testDeduplicatorDescriptor
					Expect(testBase.DeduplicatorDescriptor()).To(Equal(testDeduplicatorDescriptor))
				})
			})

			Context("SetDeduplicatorDescriptor", func() {
				It("sets the deduplicator descriptor", func() {
					testBase.SetDeduplicatorDescriptor(testDeduplicatorDescriptor)
					Expect(testBase.Deduplicator).To(Equal(testDeduplicatorDescriptor))
				})
			})
		})
	})
})
