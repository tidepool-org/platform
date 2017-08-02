package null_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/null"
)

var _ = Describe("Serializer", func() {
	Context("NewSerializer", func() {
		It("returns successfully", func() {
			Expect(null.NewSerializer()).ToNot(BeNil())
		})
	})

	Context("with new serializer", func() {
		var serializer log.Serializer

		BeforeEach(func() {
			serializer = null.NewSerializer()
			Expect(serializer).ToNot(BeNil())
		})

		Context("Serialize", func() {
			It("returns an error if fields are missing", func() {
				Expect(serializer.Serialize(nil)).To(MatchError("null: fields are missing"))
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
