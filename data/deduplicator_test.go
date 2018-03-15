package data_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"

	"github.com/tidepool-org/platform/data"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/id"
)

var _ = Describe("Deduplicator", func() {
	Context("DeduplicatorDescriptor", func() {
		Context("NewDeduplicatorDescriptor", func() {
			It("returns a new deduplicator descriptor", func() {
				Expect(data.NewDeduplicatorDescriptor()).ToNot(BeNil())
			})
		})

		Context("with a new deduplicator descriptor", func() {
			var testDeduplicatorDescriptor *data.DeduplicatorDescriptor
			var testName string
			var testVersion string

			BeforeEach(func() {
				testDeduplicatorDescriptor = data.NewDeduplicatorDescriptor()
				Expect(testDeduplicatorDescriptor).ToNot(BeNil())
				testName = id.New()
				testVersion = "1.2.3"
				testDeduplicatorDescriptor.Name = testName
				testDeduplicatorDescriptor.Version = testVersion
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
					Expect(testDeduplicatorDescriptor.IsRegisteredWithNamedDeduplicator(id.New())).To(BeFalse())
				})

				It("returns true if the deduplicator descriptor name is present and matches", func() {
					Expect(testDeduplicatorDescriptor.IsRegisteredWithNamedDeduplicator(testName)).To(BeTrue())
				})
			})

			Context("RegisterWithDeduplicator", func() {
				var testDeduplicator *testData.Deduplicator

				BeforeEach(func() {
					testDeduplicator = testData.NewDeduplicator()
				})

				AfterEach(func() {
					testDeduplicator.Expectations()
				})

				It("returns error if the deduplicator descriptor already has a name", func() {
					err := testDeduplicatorDescriptor.RegisterWithDeduplicator(testDeduplicator)
					Expect(err).To(MatchError(fmt.Sprintf("deduplicator descriptor already registered with %q", testName)))
				})

				It("returns successfully if the deduplicator descriptor does not already have a name", func() {
					testDeduplicatorDescriptor.Name = ""
					testDeduplicatorDescriptor.Version = ""
					testDeduplicator.NameOutputs = []string{testName}
					testDeduplicator.VersionOutputs = []string{testVersion}
					err := testDeduplicatorDescriptor.RegisterWithDeduplicator(testDeduplicator)
					Expect(err).ToNot(HaveOccurred())
					Expect(testDeduplicatorDescriptor.Name).To(Equal(testName))
					Expect(testDeduplicatorDescriptor.Version).To(Equal(testVersion))
				})
			})
		})
	})
})
