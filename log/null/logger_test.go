package null_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	logNull "github.com/tidepool-org/platform/log/null"
)

var _ = Describe("Logger", func() {
	Context("NewLogger", func() {
		It("returns successfully", func() {
			Expect(logNull.NewLogger()).ToNot(BeNil())
		})
	})
})
