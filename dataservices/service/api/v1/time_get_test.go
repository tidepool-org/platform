package v1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/dataservices/service/api/v1"
)

var _ = Describe("TimeGet", func() {
	Context("Unit Tests", func() {
		var context *TestContext

		BeforeEach(func() {
			context = NewTestContext()
		})

		It("succeeds", func() {
			v1.TimeGet(context)
			Expect(context.RespondWithStatusAndDataInputs).To(HaveLen(1))
			Expect(context.ValidateTest()).To(BeTrue())
		})
	})
})
