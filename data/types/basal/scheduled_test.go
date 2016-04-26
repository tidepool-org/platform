package basal_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fixtures "github.com/tidepool-org/platform/data/_fixtures"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/basal"
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
			Expect(helper.ValidDataType(basal.Build(basalObj, helper.ErrorProcessing))).To(BeNil())
		})

		Context("validation", func() {

			Context("rate", func() {

				It("is required", func() {
					delete(basalObj, "rate")
					Expect(
						helper.ErrorIsExpected(
							basal.Build(basalObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/rate",
								Detail: "Must be  >= 0.0 and <= 20.0 given '<nil>'",
							}),
					).To(BeNil())
				})

				It("invalid < 0", func() {
					basalObj["rate"] = -0.1

					Expect(
						helper.ErrorIsExpected(
							basal.Build(basalObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/rate",
								Detail: "Must be  >= 0.0 and <= 20.0 given '-0.1'",
							}),
					).To(BeNil())
				})

				It("invalid > 20", func() {
					basalObj["rate"] = 20.1

					Expect(
						helper.ErrorIsExpected(
							basal.Build(basalObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/rate",
								Detail: "Must be  >= 0.0 and <= 20.0 given '20.1'",
							}),
					).To(BeNil())
				})

				It("valid when greater than zero", func() {
					basalObj["rate"] = 0.7
					Expect(helper.ValidDataType(basal.Build(basalObj, helper.ErrorProcessing))).To(BeNil())
				})

			})
			Context("scheduleName", func() {

				It("is not required", func() {
					delete(basalObj, "scheduleName")
					Expect(helper.ValidDataType(basalObj)).To(BeNil())
				})

				It("is free text", func() {
					basalObj["scheduleName"] = "my schedule"
					Expect(helper.ValidDataType(basal.Build(basalObj, helper.ErrorProcessing))).To(BeNil())
				})

			})
		})
	})
})
