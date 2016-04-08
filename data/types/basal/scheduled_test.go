package basal

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tidepool-org/platform/data/types"

	"github.com/tidepool-org/platform/data/_fixtures"
)

var _ = Describe("Scheduled", func() {

	var helper *types.TestingHelper

	var basalObj = fixtures.TestingDatumBase()
	basalObj["type"] = "basal"
	basalObj["deliveryType"] = "scheduled"
	basalObj["scheduleName"] = "DEFAULT"
	basalObj["rate"] = 1.75
	basalObj["duration"] = 7200000

	BeforeEach(func() {
		helper = types.NewTestingHelper()
	})

	Context("from obj", func() {

		It("should return a basal if the obj is valid", func() {
			Expect(helper.ValidDataType(Build(basalObj, helper.ErrorProcessing))).To(BeNil())
		})

		Context("validation", func() {

			Context("rate", func() {

				It("is required", func() {
					delete(basalObj, "rate")
					Expect(
						helper.ErrorIsExpected(
							Build(basalObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/rate",
								Detail: "Must be greater than 0.0 given '<nil>'",
							}),
					).To(BeNil())
				})

				It("invalid when zero", func() {
					basalObj["rate"] = 0.0

					Expect(
						helper.ErrorIsExpected(
							Build(basalObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/rate",
								Detail: "Must be greater than 0.0 given '0'",
							}),
					).To(BeNil())
				})

				It("valid when greater than zero", func() {
					basalObj["rate"] = 0.7
					Expect(helper.ValidDataType(Build(basalObj, helper.ErrorProcessing))).To(BeNil())
				})

			})
			Context("scheduleName", func() {

				It("is not required", func() {
					delete(basalObj, "scheduleName")
					Expect(helper.ValidDataType(basalObj)).To(BeNil())
				})

				It("is free text", func() {
					basalObj["scheduleName"] = "my schedule"
					Expect(helper.ValidDataType(Build(basalObj, helper.ErrorProcessing))).To(BeNil())
				})

			})
		})
	})
})
