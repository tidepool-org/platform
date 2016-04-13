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

			/*Context("timeProcessing", func() {

				It("is required", func() {
					delete(uploadObj, "timeProcessing")

					Expect(
						helper.ErrorIsExpected(
							Build(uploadObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/timeProcessing",
								Detail: "Must be one of across-the-board-timezone, utc-bootstrapping, none given '<nil>'",
							}),
					).To(BeNil())
				})

				It("can be across-the-board-timezone", func() {
					uploadObj["timeProcessing"] = "across-the-board-timezone"
					Expect(helper.ValidDataType(Build(uploadObj, helper.ErrorProcessing))).To(BeNil())
				})
			})*/
		})

	})
})
