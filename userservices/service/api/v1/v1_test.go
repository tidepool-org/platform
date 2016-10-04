package v1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/userservices/service/api/v1"
)

var _ = Describe("V1", func() {
	Context("Routes", func() {
		It("returns the correct routes", func() {
			Expect(v1.Routes()).ToNot(BeEmpty())
		})
	})
})
