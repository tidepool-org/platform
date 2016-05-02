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
	bolusObj["subType"] = "normal"
	bolusObj["normal"] = 1.0

	var helper *types.TestingHelper

	BeforeEach(func() {
		helper = types.NewTestingHelper()
	})

	Context("type from obj", func() {

		It("returns a valid bolus", func() {
			Expect(helper.ValidDataType(bolus.Build(bolusObj, helper.ErrorProcessing))).To(BeNil())
		})

		Context("validation", func() {

			Context("subType", func() {

				It("is required", func() {
					delete(bolusObj, "subType")

					Expect(
						helper.ErrorIsExpected(
							bolus.Build(bolusObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/subType",
								Detail: "Must be one of normal, square, dual/square given '<nil>'",
							}),
					).To(BeNil())
				})

				It("invalid when no matching subType", func() {
					bolusObj["subType"] = "superfly"
					Expect(
						helper.ErrorIsExpected(
							bolus.Build(bolusObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/subType",
								Detail: "Must be one of normal, square, dual/square given 'superfly'",
							}),
					).To(BeNil())
				})

				It("injected type is not supported", func() {
					bolusObj["subType"] = "injected"
					Expect(
						helper.ErrorIsExpected(
							bolus.Build(bolusObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/subType",
								Detail: "Must be one of normal, square, dual/square given 'injected'",
							}),
					).To(BeNil())
				})

				It("normal type is supported", func() {
					bolusObj["subType"] = "normal"
					Expect(helper.ValidDataType(bolus.Build(bolusObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("square type is supported", func() {
					bolusObj["subType"] = "square"
					bolusObj["extended"] = 1.0
					bolusObj["duration"] = 3600000
					Expect(helper.ValidDataType(bolus.Build(bolusObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("dual/square type is supported", func() {
					bolusObj["subType"] = "dual/square"
					bolusObj["extended"] = 1.0
					bolusObj["duration"] = 3600000
					Expect(helper.ValidDataType(bolus.Build(bolusObj, helper.ErrorProcessing))).To(BeNil())
				})

			})
		})
	})
})
