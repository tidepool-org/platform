package null_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/log/null"
)

var _ = Describe("Logger", func() {
	Context("NewLogger", func() {
		It("returns successfully", func() {
			Expect(null.NewLogger()).ToNot(BeNil())
		})
	})
})
