package context_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/client"
)

var _ = Describe("Source", func() {
	Describe("DefaultHTTPTimeout", func() {
		It("is 60 seconds", func() {
			Expect(client.DefaultHTTPTimeout).To(Equal(60 * time.Second))
		})
	})
})
