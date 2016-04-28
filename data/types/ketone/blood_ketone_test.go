package ketone_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fixtures "github.com/tidepool-org/platform/data/_fixtures"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/ketone"
)

var _ = Describe("Blood", func() {
	var bloodKetoneObj = fixtures.TestingDatumBase()
	var helper *types.TestingHelper

	BeforeEach(func() {
		helper = types.NewTestingHelper()
		bloodKetoneObj["type"] = "bloodKetone"
		bloodKetoneObj["value"] = 2.2
		bloodKetoneObj["units"] = "mmol/L"
	})

	Context("ketone from obj", func() {

		It("when valid", func() {
			Expect(helper.ValidDataType(ketone.Build(bloodKetoneObj, helper.ErrorProcessing))).To(BeNil())
		})

		Context("validation", func() {
			Context("value", func() {
				It("is required", func() {

					delete(bloodKetoneObj, "value")

					Expect(
						helper.ErrorIsExpected(
							ketone.Build(bloodKetoneObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/value",
								Detail: "Needs to be in the range of >= 0.0 and <= 10.0 given '<nil>'",
							}),
					).To(BeNil())
				})
				It("fails < 0.0", func() {
					bloodKetoneObj["value"] = -0.1

					Expect(
						helper.ErrorIsExpected(
							ketone.Build(bloodKetoneObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/value",
								Detail: "Needs to be in the range of >= 0.0 and <= 10.0 given '-0.1'",
							}),
					).To(BeNil())
				})

				It("fails > 10.0", func() {
					bloodKetoneObj["value"] = 10.1

					Expect(
						helper.ErrorIsExpected(
							ketone.Build(bloodKetoneObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/value",
								Detail: "Needs to be in the range of >= 0.0 and <= 10.0 given '10.1'",
							}),
					).To(BeNil())
				})

				It("passes if  >= 0.0 and <= 10.0", func() {
					bloodKetoneObj["value"] = 4.1
					Expect(helper.ValidDataType(ketone.Build(bloodKetoneObj, helper.ErrorProcessing))).To(BeNil())
				})

			})
			Context("units", func() {
				It("is required", func() {

					delete(bloodKetoneObj, "units")

					Expect(
						helper.ErrorIsExpected(
							ketone.Build(bloodKetoneObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/units",
								Detail: "Must be mmol/L given '<nil>'",
							}),
					).To(BeNil())
				})
				It("fails if not mmol/L", func() {
					bloodKetoneObj["units"] = "mg/dL"
					Expect(
						helper.ErrorIsExpected(
							ketone.Build(bloodKetoneObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/units",
								Detail: "Must be mmol/L given 'mg/dL'",
							}),
					).To(BeNil())
				})

			})

		})
	})
})
