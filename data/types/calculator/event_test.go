package calculator

import (
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/_fixtures"
	"github.com/tidepool-org/platform/data/types"
)

var _ = Describe("Event", func() {
	var calculatorObj = fixtures.TestingDatumBase()
	var recommendedObj = make(map[string]interface{})
	var bolusObj = make(map[string]interface{})
	var bgTargetObj = make(map[string]interface{})
	var helper *types.TestingHelper

	BeforeEach(func() {
		helper = types.NewTestingHelper()
		calculatorObj["type"] = "wizard"
		calculatorObj["carbInput"] = 45
		calculatorObj["bgInput"] = 99.0
		calculatorObj["insulinOnBoard"] = 1.3
		calculatorObj["insulinSensitivity"] = 75
		calculatorObj["units"] = "mg/dL"

		recommendedObj["carb"] = 4.0
		recommendedObj["correction"] = 1.0
		recommendedObj["net"] = 4.0

		calculatorObj["recommended"] = recommendedObj

		bolusObj["type"] = "bolus"
		bolusObj["subType"] = "normal"
		bolusObj["deviceId"] = "test"
		bolusObj["time"] = "2014-01-01T01:00:00.000Z"

		calculatorObj["bolus"] = bolusObj

		bgTargetObj["high"] = 120
		bgTargetObj["low"] = 80

		calculatorObj["bgTarget"] = bgTargetObj

	})

	Context("calculator record from obj", func() {

		It("when valid", func() {
			Expect(helper.ValidDataType(Build(calculatorObj, helper.ErrorProcessing))).To(BeNil())
		})

		Context("validation", func() {

			Context("carbInput", func() {

				It("is not required", func() {
					delete(calculatorObj, "carbInput")
					Expect(helper.ValidDataType(Build(calculatorObj, helper.ErrorProcessing))).To(BeNil())
				})

			})

			Context("bgInput", func() {

				It("is not required", func() {
					delete(calculatorObj, "bgInput")
					Expect(helper.ValidDataType(Build(calculatorObj, helper.ErrorProcessing))).To(BeNil())
				})

			})

			Context("insulinOnBoard", func() {

				It("is not required", func() {
					delete(calculatorObj, "insulinOnBoard")
					Expect(helper.ValidDataType(Build(calculatorObj, helper.ErrorProcessing))).To(BeNil())
				})

			})

			Context("units", func() {

				It("is required", func() {
					delete(calculatorObj, "units")

					Expect(
						helper.ErrorIsExpected(
							Build(calculatorObj, helper.ErrorProcessing),
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
					Expect(helper.ValidDataType(Build(calculatorObj, helper.ErrorProcessing))).To(BeNil())
				})

			})

			Context("bolus", func() {

				It("is not required", func() {
					delete(calculatorObj, "bolus")
					Expect(helper.ValidDataType(Build(calculatorObj, helper.ErrorProcessing))).To(BeNil())
				})

			})

			Context("bgTarget", func() {

				It("is not required", func() {
					delete(calculatorObj, "bgTarget")
					Expect(helper.ValidDataType(Build(calculatorObj, helper.ErrorProcessing))).To(BeNil())
				})

			})
		})

	})
})
