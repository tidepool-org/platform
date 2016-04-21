package bolus_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fixtures "github.com/tidepool-org/platform/data/_fixtures"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/bolus"
)

var _ = Describe("Bolus", func() {

	var bolusObj = fixtures.TestingDatumBase()
	bolusObj["type"] = "bolus"
	bolusObj["subType"] = "dual/square"
	bolusObj["normal"] = 2.0
	bolusObj["extended"] = 1.0
	bolusObj["duration"] = 3600000

	var helper *types.TestingHelper

	BeforeEach(func() {
		helper = types.NewTestingHelper()
	})

	Context("dual/square from obj", func() {

		It("if the obj is valid", func() {
			Expect(helper.ValidDataType(bolus.Build(bolusObj, helper.ErrorProcessing))).To(BeNil())
		})

		Context("validation", func() {

			Context("duration", func() {

				It("is not required", func() {
					delete(bolusObj, "duration")
					Expect(helper.ValidDataType(bolus.Build(bolusObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("invalid when less than zero", func() {
					bolusObj["duration"] = -1

					Expect(
						helper.ErrorIsExpected(
							bolus.Build(bolusObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/duration",
								Detail: "Must be greater than 0 given '-1'",
							}),
					).To(BeNil())

				})

				It("valid greater than zero", func() {
					bolusObj["duration"] = 4000
					Expect(helper.ValidDataType(bolus.Build(bolusObj, helper.ErrorProcessing))).To(BeNil())
				})

			})
			Context("extended", func() {

				It("is not required", func() {
					delete(bolusObj, "extended")
					Expect(helper.ValidDataType(bolus.Build(bolusObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("invalid when zero", func() {
					bolusObj["extended"] = -0.1

					Expect(
						helper.ErrorIsExpected(
							bolus.Build(bolusObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/extended",
								Detail: "Must be greater than 0.0 given '-0.1'",
							}),
					).To(BeNil())
				})

				It("valid when greater than zero", func() {
					bolusObj["extended"] = 0.7
					Expect(helper.ValidDataType(bolus.Build(bolusObj, helper.ErrorProcessing))).To(BeNil())
				})

			})
			Context("normal", func() {

				It("is not required", func() {
					delete(bolusObj, "normal")
					Expect(helper.ValidDataType(bolus.Build(bolusObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("invalid when zero", func() {
					bolusObj["normal"] = -0.1

					Expect(
						helper.ErrorIsExpected(
							bolus.Build(bolusObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/normal",
								Detail: "Must be greater than 0.0 given '-0.1'",
							}),
					).To(BeNil())

				})

				It("valid when greater than zero", func() {
					bolusObj["normal"] = 0.7
					Expect(helper.ValidDataType(bolus.Build(bolusObj, helper.ErrorProcessing))).To(BeNil())
				})

			})
		})
	})
})
