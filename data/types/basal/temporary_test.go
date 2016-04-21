package basal

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fixtures "github.com/tidepool-org/platform/data/_fixtures"
	"github.com/tidepool-org/platform/data/types"
)

var _ = Describe("Temporary", func() {

	var helper *types.TestingHelper

	var basalObj = fixtures.TestingDatumBase()
	basalObj["type"] = "basal"
	basalObj["deliveryType"] = "temp"
	basalObj["rate"] = 1.75
	basalObj["percent"] = 0.5
	basalObj["duration"] = 1800000

	BeforeEach(func() {
		helper = types.NewTestingHelper()
	})

	Context("Temporary from obj", func() {

		It("should return a basal if the obj is valid", func() {
			Expect(helper.ValidDataType(Build(basalObj, helper.ErrorProcessing))).To(BeNil())
		})

		Context("validation", func() {

			Context("rate", func() {

				It("is not required", func() {
					delete(basalObj, "rate")
					Expect(helper.ValidDataType(Build(basalObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("invalid less than zero", func() {
					basalObj["rate"] = -0.1

					Expect(
						helper.ErrorIsExpected(
							Build(basalObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/rate",
								Detail: "Must be greater than 0.0 given '-0.1'",
							}),
					).To(BeNil())

				})

				It("valid when greater than zero", func() {
					basalObj["rate"] = 0.7
					Expect(helper.ValidDataType(Build(basalObj, helper.ErrorProcessing))).To(BeNil())
				})

			})

			Context("percent", func() {

				It("is not required", func() {
					delete(basalObj, "percent")
					Expect(helper.ValidDataType(basalObj)).To(BeNil())
				})

				It("invalid less than zero", func() {
					basalObj["percent"] = -0.1

					Expect(
						helper.ErrorIsExpected(
							Build(basalObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/percent",
								Detail: "Must be greater than 0.0 given '-0.1'",
							}),
					).To(BeNil())
				})

				It("invalid when greater than 1.0", func() {
					basalObj["percent"] = 1.1
					Expect(
						helper.ErrorIsExpected(
							Build(basalObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/percent",
								Detail: "Must be greater than 0.0 given '1.1'",
							}),
					).To(BeNil())
				})

				It("valid when between 0.0 and 1.0", func() {
					basalObj["percent"] = 0.7
					Expect(helper.ValidDataType(Build(basalObj, helper.ErrorProcessing))).To(BeNil())
				})

			})

			Context("suppressed", func() {

				suppressed := make(map[string]interface{})

				BeforeEach(func() {
					suppressed["deliveryType"] = "scheduled"
					suppressed["scheduleName"] = "DEFAULT"
					suppressed["rate"] = 1.75
					basalObj["suppressed"] = suppressed
				})

				It("is not required", func() {
					delete(basalObj, "suppressed")
					Expect(helper.ValidDataType(Build(basalObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("when present is validated", func() {
					Expect(helper.ValidDataType(Build(basalObj, helper.ErrorProcessing))).To(BeNil())
				})

			})

		})
	})
})
