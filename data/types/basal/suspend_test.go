package basal

import (
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"
	"github.com/tidepool-org/platform/data/types"

	"github.com/tidepool-org/platform/data/_fixtures"
)

var _ = Describe("Suspend", func() {

	var helper *types.TestingHelper

	var basalObj = fixtures.TestingDatumBase()
	basalObj["type"] = "basal"
	basalObj["deliveryType"] = "suspend"
	basalObj["duration"] = 1800000

	BeforeEach(func() {
		helper = types.NewTestingHelper()
	})

	Context("from obj", func() {

		It("should return a basal if the obj is valid", func() {
			Expect(helper.ValidDataType(Build(basalObj, helper.ErrorProcessing))).To(BeNil())
		})

		Context("validation", func() {

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
					Expect(helper.ValidDataType(Build(basalObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("when present is validated", func() {
					Expect(helper.ValidDataType(Build(basalObj, helper.ErrorProcessing))).To(BeNil())
				})

			})
		})
	})
})
