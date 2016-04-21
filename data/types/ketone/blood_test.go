package ketone

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fixtures "github.com/tidepool-org/platform/data/_fixtures"
	"github.com/tidepool-org/platform/data/types"
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
			Expect(helper.ValidDataType(Build(bloodKetoneObj, helper.ErrorProcessing))).To(BeNil())
		})

		Context("validation", func() {
			Context("value", func() {
				It("fails greater than zero", func() {
					bloodKetoneObj["value"] = 0.0

					Expect(
						helper.ErrorIsExpected(
							Build(bloodKetoneObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/value",
								Detail: "Must be greater than 0.0 given '0'",
							}),
					).To(BeNil())
				})

			})
			Context("iunits", func() {
				It("fails if not mmol/L", func() {
					bloodKetoneObj["units"] = "mg/dL"
					Expect(
						helper.ErrorIsExpected(
							Build(bloodKetoneObj, helper.ErrorProcessing),
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
