package data_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
)

var _ = Describe("Delegate", func() {
	Context("DeduplicatorDescriptor", func() {
		Context("NewDeduplicatorDescriptor", func() {
			It("returns a new deduplicator descriptor", func() {
				Expect(data.NewDeduplicatorDescriptor()).ToNot(BeNil())
			})
		})

		Context("with a new deduplicator descriptor", func() {
			var testDeduplicatorDescriptor *data.DeduplicatorDescriptor
			var testName string

			BeforeEach(func() {
				testDeduplicatorDescriptor = data.NewDeduplicatorDescriptor()
				Expect(testDeduplicatorDescriptor).ToNot(BeNil())
				testName = app.NewID()
				testDeduplicatorDescriptor.Name = testName
			})

			Context("IsRegisteredWithAnyDeduplicator", func() {
				It("returns false if the deduplicator descriptor name is missing", func() {
					testDeduplicatorDescriptor.Name = ""
					Expect(testDeduplicatorDescriptor.IsRegisteredWithAnyDeduplicator()).To(BeFalse())
				})

				It("returns true if the deduplicator descriptor name is present", func() {
					Expect(testDeduplicatorDescriptor.IsRegisteredWithAnyDeduplicator()).To(BeTrue())
				})
			})

			Context("IsRegisteredWithNamedDeduplicator", func() {
				It("returns false if the deduplicator descriptor name is missing", func() {
					testDeduplicatorDescriptor.Name = ""
					Expect(testDeduplicatorDescriptor.IsRegisteredWithNamedDeduplicator(testName)).To(BeFalse())
				})

				It("returns true if the deduplicator descriptor name is present, but does not match", func() {
					Expect(testDeduplicatorDescriptor.IsRegisteredWithNamedDeduplicator(app.NewID())).To(BeFalse())
				})

				It("returns true if the deduplicator descriptor name is present and matches", func() {
					Expect(testDeduplicatorDescriptor.IsRegisteredWithNamedDeduplicator(testName)).To(BeTrue())
				})
			})

			Context("RegisterWithNamedDeduplicator", func() {
				It("returns error if the deduplicator descriptor already has a name", func() {
					err := testDeduplicatorDescriptor.RegisterWithNamedDeduplicator(app.NewID())
					Expect(err).To(MatchError(fmt.Sprintf(`data: deduplicator descriptor already registered with "%s"`, testName)))
				})

				It("returns successfully if the deduplicator descriptor does not already have a name", func() {
					testDeduplicatorDescriptor.Name = ""
					err := testDeduplicatorDescriptor.RegisterWithNamedDeduplicator(testName)
					Expect(err).ToNot(HaveOccurred())
					Expect(testDeduplicatorDescriptor.Name).To(Equal(testName))
				})
			})
		})
	})
})
