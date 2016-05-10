package version_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/version"
)

var _ = Describe("Version", func() {

	Describe("Current", func() {
		var current version.Version

		BeforeEach(func() {
			current = version.Current()
		})

		It("exists", func() {
			Expect(current).ToNot(BeNil())
		})

		It("responds to Base", func() {
			Expect(current.Base).ToNot(BeNil())
		})

		It("responds to ShortCommit", func() {
			Expect(current.ShortCommit).ToNot(BeNil())
		})

		It("responds to FullCommit", func() {
			Expect(current.FullCommit).ToNot(BeNil())
		})

		It("responds to Short", func() {
			Expect(current.Short).ToNot(BeNil())
		})

		It("responds to Long", func() {
			Expect(current.Long).ToNot(BeNil())
		})
	})
})
