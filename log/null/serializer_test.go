package null_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/log"
	logNull "github.com/tidepool-org/platform/log/null"
)

var _ = Describe("Serializer", func() {
	Context("NewSerializer", func() {
		It("returns successfully", func() {
			Expect(logNull.NewSerializer()).ToNot(BeNil())
		})
	})

	Context("with new serializer", func() {
		var serializer log.Serializer

		BeforeEach(func() {
			serializer = logNull.NewSerializer()
			Expect(serializer).ToNot(BeNil())
		})

		Context("Serialize", func() {
			It("returns an error if fields are missing", func() {
				Expect(serializer.Serialize(nil)).To(MatchError("fields are missing"))
			})

			It("returns successfully after writing buffer with empty fields", func() {
				Expect(serializer.Serialize(log.Fields{})).To(Succeed())
			})

			It("returns successfully after writing buffer with non-empty fields", func() {
				Expect(serializer.Serialize(log.Fields{"b": "right", "a": "left"})).To(Succeed())
			})
		})
	})
})
