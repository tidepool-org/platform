package basal_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fixtures "github.com/tidepool-org/platform/data/_fixtures"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/basal"
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
								Detail: "Must be >= 0.0 and <= 20.0 given '<nil>'",
							}),
					).To(BeNil())
				})

				It("invalid when < 0", func() {
					basalObj["rate"] = -0.1

					Expect(
						helper.ErrorIsExpected(
							basal.Build(basalObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/rate",
								Detail: "Must be >= 0.0 and <= 20.0 given '-0.1'",
							}),
					).To(BeNil())

				})

				It("invalid when > 20.0", func() {
					basalObj["rate"] = 20.1

					Expect(
						helper.ErrorIsExpected(
							basal.Build(basalObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/rate",
								Detail: "Must be >= 0.0 and <= 20.0 given '20.1'",
							}),
					).To(BeNil())

				})

				It("valid when >= 0.0", func() {
					basalObj["rate"] = 0.0
					Expect(helper.ValidDataType(basal.Build(basalObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("valid when <= 20.0", func() {
					basalObj["rate"] = 20.0
					Expect(helper.ValidDataType(basal.Build(basalObj, helper.ErrorProcessing))).To(BeNil())
				})

			})

			Context("duration", func() {

				It("is required", func() {
					delete(basalObj, "duration")

					Expect(
						helper.ErrorIsExpected(
							basal.Build(basalObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/duration",
								Detail: "Must be >= 0 and <= 86400000 given '<nil>'",
							}),
					).To(BeNil())

				})

				It("invalid when < 0", func() {
					basalObj["duration"] = -1

					Expect(
						helper.ErrorIsExpected(
							basal.Build(basalObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/duration",
								Detail: "Must be >= 0 and <= 86400000 given '-1'",
							}),
					).To(BeNil())

				})

				It("invalid when > 86400000", func() {
					basalObj["duration"] = 86400001

					Expect(
						helper.ErrorIsExpected(
							basal.Build(basalObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/duration",
								Detail: "Must be >= 0 and <= 86400000 given '86400001'",
							}),
					).To(BeNil())

				})

				It("valid when >= 0", func() {
					basalObj["duration"] = 0
					Expect(helper.ValidDataType(basal.Build(basalObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("valid when <= 86400000", func() {
					basalObj["duration"] = 86400000
					Expect(helper.ValidDataType(basal.Build(basalObj, helper.ErrorProcessing))).To(BeNil())
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
							basal.Build(basalObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/percent",
								Detail: "Must be >= 0.0 and <= 10.0 given '-0.1'",
							}),
					).To(BeNil())
				})

				It("invalid when greater than 10.0", func() {
					basalObj["percent"] = 10.1
					Expect(
						helper.ErrorIsExpected(
							basal.Build(basalObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/percent",
								Detail: "Must be >= 0.0 and <= 10.0 given '10.1'",
							}),
					).To(BeNil())
				})

				It("valid when >= 0.0", func() {
					basalObj["percent"] = 0.0
					Expect(helper.ValidDataType(basal.Build(basalObj, helper.ErrorProcessing))).To(BeNil())
				})
				It("valid when <= 10.0", func() {
					basalObj["percent"] = 10.0
					Expect(helper.ValidDataType(basal.Build(basalObj, helper.ErrorProcessing))).To(BeNil())
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
					Expect(helper.ValidDataType(basal.Build(basalObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("when present is validated", func() {
					Expect(helper.ValidDataType(basal.Build(basalObj, helper.ErrorProcessing))).To(BeNil())
				})

			})

		})
	})
})
