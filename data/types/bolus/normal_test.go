package bolus

import (
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"
	"github.com/tidepool-org/platform/data/types"

	"github.com/tidepool-org/platform/data/_fixtures"
)

var _ = Describe("Normal", func() {

	var bolusObj = fixtures.TestingDatumBase()
	bolusObj["type"] = "bolus"
	bolusObj["subType"] = "normal"
	bolusObj["normal"] = 1.0

	var helper *types.TestingHelper

	BeforeEach(func() {
		helper = types.NewTestingHelper()
	})

	Context("from obj", func() {

		It("if the obj is valid", func() {
			Expect(helper.ValidDataType(Build(bolusObj, helper.ErrorProcessing))).To(BeNil())
		})

		Context("validation", func() {

			Context("normal", func() {

				It("is not required", func() {
					delete(bolusObj, "normal")
					Expect(helper.ValidDataType(Build(bolusObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("invalid when less than 0.0", func() {
					bolusObj["normal"] = -0.1

					Expect(
						helper.ErrorIsExpected(
							Build(bolusObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/normal",
								Detail: "Must be greater than 0.0 given '-0.1'",
							}),
					).To(BeNil())

				})

				It("valid when than 0.0", func() {
					bolusObj["normal"] = 0.7
					Expect(helper.ValidDataType(Build(bolusObj, helper.ErrorProcessing))).To(BeNil())
				})

			})
		})
	})
})
