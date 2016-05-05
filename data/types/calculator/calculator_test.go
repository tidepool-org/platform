package calculator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fixtures "github.com/tidepool-org/platform/data/_fixtures"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/calculator"
)

var _ = Describe("Event", func() {
	var calculatorObj = fixtures.TestingDatumBase()
	var recommendedObj = make(map[string]interface{}, 0)
	var bgTargetObj = make(map[string]interface{}, 0)
	var helper *types.TestingHelper

	BeforeEach(func() {
		helper = types.NewTestingHelper()
		calculatorObj["type"] = "wizard"
		calculatorObj["carbInput"] = 45
		calculatorObj["bgInput"] = 99.0
		calculatorObj["insulinOnBoard"] = 1.3
		calculatorObj["insulinSensitivity"] = 75
		calculatorObj["units"] = "mg/dL"
		calculatorObj["bolus"] = "linked-bolus-id"

		recommendedObj["carb"] = 4.0
		recommendedObj["correction"] = 1.0
		recommendedObj["net"] = 4.0

		calculatorObj["recommended"] = recommendedObj

		bgTargetObj["high"] = 120.0
		bgTargetObj["low"] = 80.0

		calculatorObj["bgTarget"] = bgTargetObj

	})

	Context("calculator record from obj", func() {

		It("when valid", func() {
			Expect(helper.ValidDataType(calculator.Build(calculatorObj, helper.ErrorProcessing))).To(BeNil())
		})

		Context("validation", func() {

			Context("carbInput", func() {

				It("is not required", func() {
					delete(calculatorObj, "carbInput")
					Expect(helper.ValidDataType(calculator.Build(calculatorObj, helper.ErrorProcessing))).To(BeNil())
				})

			})

			Context("bgInput", func() {

				It("is not required", func() {
					delete(calculatorObj, "bgInput")
					Expect(helper.ValidDataType(calculator.Build(calculatorObj, helper.ErrorProcessing))).To(BeNil())
				})

			})

			Context("insulinOnBoard", func() {

				It("is not required", func() {
					delete(calculatorObj, "insulinOnBoard")
					Expect(helper.ValidDataType(calculator.Build(calculatorObj, helper.ErrorProcessing))).To(BeNil())
				})

			})

			Context("units", func() {

				It("is required", func() {
					delete(calculatorObj, "units")

					Expect(
						helper.ErrorIsExpected(
							calculator.Build(calculatorObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/units",
								Detail: "Must be one of mmol/L, mg/dL given '<nil>'",
							}),
					).To(BeNil())
				})

			})

			Context("recommended", func() {

				It("is not required", func() {
					delete(calculatorObj, "recommended")
					Expect(helper.ValidDataType(calculator.Build(calculatorObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("net is not required", func() {

					delete(recommendedObj, "net")
					calculatorObj["recommended"] = recommendedObj

					Expect(helper.ValidDataType(calculator.Build(calculatorObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("correction is not required", func() {

					delete(recommendedObj, "correction")
					calculatorObj["recommended"] = recommendedObj

					Expect(helper.ValidDataType(calculator.Build(calculatorObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("carb required", func() {

					delete(recommendedObj, "carb")
					calculatorObj["recommended"] = recommendedObj

					Expect(
						helper.ErrorIsExpected(
							calculator.Build(calculatorObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/recommended/carb",
								Detail: "This is a required field given '<nil>'",
							}),
					).To(BeNil())
				})

			})

			Context("bolus", func() {

				It("is not required", func() {
					delete(calculatorObj, "bolus")
					Expect(helper.ValidDataType(calculator.Build(calculatorObj, helper.ErrorProcessing))).To(BeNil())
				})

			})

			Context("bgTarget", func() {

				It("is not required", func() {
					delete(calculatorObj, "bgTarget")
					Expect(helper.ValidDataType(calculator.Build(calculatorObj, helper.ErrorProcessing))).To(BeNil())
				})

				// TODO_DATA: Commented out due to changes in calculator.go adding .SetValueAllowedToBeEmpty(true)
				// It("if present requires high", func() {

				// 	delete(bgTargetObj, "high")
				// 	calculatorObj["bgTarget"] = bgTargetObj

				// 	Expect(
				// 		helper.ErrorIsExpected(
				// 			calculator.Build(calculatorObj, helper.ErrorProcessing),
				// 			types.ExpectedErrorDetails{
				// 				Path:   "0/bgTarget/high",
				// 				Detail: "Must be between 0.0 and 1000.0 given '<nil>'",
				// 			}),
				// 	).To(BeNil())
				// })

				// TODO_DATA: Commented out due to changes in calculator.go adding .SetValueAllowedToBeEmpty(true)
				// It("if present requires low", func() {

				// 	delete(bgTargetObj, "low")
				// 	calculatorObj["bgTarget"] = bgTargetObj

				// 	Expect(
				// 		helper.ErrorIsExpected(
				// 			calculator.Build(calculatorObj, helper.ErrorProcessing),
				// 			types.ExpectedErrorDetails{
				// 				Path:   "0/bgTarget/low",
				// 				Detail: "Must be between 0.0 and 1000.0 given '<nil>'",
				// 			}),
				// 	).To(BeNil())
				// })

			})
		})

	})
})
