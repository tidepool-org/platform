package bloodglucose_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fixtures "github.com/tidepool-org/platform/data/_fixtures"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/bloodglucose"
)

var _ = Describe("Continuous", func() {
	var bgObj = fixtures.TestingDatumBase()
	var helper *types.TestingHelper

	BeforeEach(func() {
		helper = types.NewTestingHelper()
	})

	Context("cbg from obj", func() {

		BeforeEach(func() {
			bgObj["type"] = "cbg"
			bgObj["value"] = 5.5
			bgObj["units"] = "mmol/l"
			bgObj["isig"] = 6.5
		})

		It("returns a bolus if the obj is valid", func() {
			Expect(helper.ValidDataType(bloodglucose.BuildContinuous(bgObj, helper.ErrorProcessing))).To(BeNil())
		})

	})
	Context("validation", func() {

		BeforeEach(func() {
			bgObj["type"] = "cbg"
			bgObj["value"] = 5.5
			bgObj["units"] = "mmol/l"
			bgObj["isig"] = 6.5
		})

		Context("units", func() {
			It("is required", func() {
				delete(bgObj, "units")
				Expect(
					helper.ErrorIsExpected(
						bloodglucose.BuildContinuous(bgObj, helper.ErrorProcessing),
						types.ExpectedErrorDetails{
							Path:   "0/units",
							Detail: "Must be one of mmol/L, mg/dL given '<nil>'",
						}),
				).To(BeNil())
			})

			It("can be mmol/l", func() {
				bgObj["units"] = "mmol/l"
				Expect(helper.ValidDataType(bloodglucose.BuildContinuous(bgObj, helper.ErrorProcessing))).To(BeNil())
			})

			It("can be mg/dl", func() {
				bgObj["units"] = "mg/dl"
				Expect(helper.ValidDataType(bloodglucose.BuildContinuous(bgObj, helper.ErrorProcessing))).To(BeNil())
			})

			It("cannot be anything else", func() {
				bgObj["units"] = "grams"

				Expect(
					helper.ErrorIsExpected(
						bloodglucose.BuildContinuous(bgObj, helper.ErrorProcessing),
						types.ExpectedErrorDetails{
							Path:   "0/units",
							Detail: "Must be one of mmol/L, mg/dL given 'grams'",
						}),
				).To(BeNil())

			})

		})
		Context("value", func() {
			It("is required", func() {
				delete(bgObj, "value")
				Expect(
					helper.ErrorIsExpected(
						bloodglucose.BuildContinuous(bgObj, helper.ErrorProcessing),
						types.ExpectedErrorDetails{
							Path:   "0/value",
							Detail: "Must be greater than 0.0 given '<nil>'",
						}),
				).To(BeNil())
			})
		})
		// Context("isig", func() {

		// 	It("is required", func() {
		// 		delete(bgObj, "isig")
		// 		Expect(
		// 			helper.ErrorIsExpected(
		// 				bloodglucose.BuildContinuous(bgObj, helper.ErrorProcessing),
		// 				types.ExpectedErrorDetails{
		// 					Path:   "0/isig",
		// 					Detail: "Must be greater than 0.0 given '<nil>'",
		// 				}),
		// 		).To(BeNil())
		// 	})
		// })
	})
})
