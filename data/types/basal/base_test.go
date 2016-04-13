package basal

import (
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/_fixtures"
	"github.com/tidepool-org/platform/data/types"
)

var _ = Describe("Basal", func() {

	var helper *types.TestingHelper

	var basalObj = fixtures.TestingDatumBase()
	basalObj["type"] = "basal"
	basalObj["deliveryType"] = "scheduled"
	basalObj["rate"] = 1.0
	basalObj["duration"] = 28800000

	BeforeEach(func() {
		helper = types.NewTestingHelper()
	})

	Context("type from obj", func() {

		It("returns a valid basal type", func() {
			Expect(helper.ValidDataType(Build(basalObj, helper.ErrorProcessing))).To(BeNil())
		})

	})

	Context("validation", func() {

		Context("duration", func() {

			It("is not required", func() {
				delete(basalObj, "duration")
				Expect(helper.ValidDataType(Build(basalObj, helper.ErrorProcessing))).To(BeNil())
			})

			It("fails if less than zero", func() {
				basalObj["duration"] = -1

				Expect(
					helper.ErrorIsExpected(
						Build(basalObj, helper.ErrorProcessing),
						types.ExpectedErrorDetails{
							Path:   "0/duration",
							Detail: "Must be greater than 0 given '-1'",
						}),
				).To(BeNil())

			})

			It("valid when greater than zero", func() {
				basalObj["duration"] = 4000
				Expect(helper.ValidDataType(Build(basalObj, helper.ErrorProcessing))).To(BeNil())
			})

		})

		Context("deliveryType", func() {

			It("is required", func() {
				delete(basalObj, "deliveryType")

				Expect(
					helper.ErrorIsExpected(
						Build(basalObj, helper.ErrorProcessing),
						types.ExpectedErrorDetails{
							Path:   "0/deliveryType",
							Detail: "Must be one of scheduled, suspend, temp given '<nil>'",
						}),
				).To(BeNil())

			})

			It("invalid when no matching type", func() {
				basalObj["deliveryType"] = "superfly"
				Expect(
					helper.ErrorIsExpected(
						Build(basalObj, helper.ErrorProcessing),
						types.ExpectedErrorDetails{
							Path:   "0/deliveryType",
							Detail: "Must be one of scheduled, suspend, temp given 'superfly'",
						}),
				).To(BeNil())

			})

			It("invalid if unsupported injected type", func() {
				basalObj["deliveryType"] = "injected"

				Expect(
					helper.ErrorIsExpected(
						Build(basalObj, helper.ErrorProcessing),
						types.ExpectedErrorDetails{
							Path:   "0/deliveryType",
							Detail: "Must be one of scheduled, suspend, temp given 'injected'",
						}),
				).To(BeNil())
			})

			It("valid if scheduled type", func() {
				basalObj["deliveryType"] = "scheduled"
				Expect(helper.ValidDataType(Build(basalObj, helper.ErrorProcessing))).To(BeNil())
			})

			It("valid if suspend type", func() {
				basalObj["deliveryType"] = "suspend"
				Expect(helper.ValidDataType(Build(basalObj, helper.ErrorProcessing))).To(BeNil())
			})

			It("valid if temp type", func() {
				basalObj["deliveryType"] = "temp"
				Expect(helper.ValidDataType(Build(basalObj, helper.ErrorProcessing))).To(BeNil())
			})

		})
	})
})
