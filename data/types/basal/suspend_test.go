package basal

import (
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"
	"github.com/tidepool-org/platform/data/types"

	"github.com/tidepool-org/platform/data/_fixtures"
	"github.com/tidepool-org/platform/validate"
)

var _ = Describe("Suspend", func() {

	var processing validate.ErrorProcessing

	var basalObj = fixtures.TestingDatumBase()
	basalObj["type"] = "basal"
	basalObj["deliveryType"] = "suspend"
	basalObj["duration"] = 1800000

	Context("from obj", func() {

		BeforeEach(func() {
			processing = validate.ErrorProcessing{BasePath: "0", ErrorsArray: validate.NewErrorsArray()}
		})

		It("should return a basal if the obj is valid", func() {
			basal := Build(basalObj, processing)
			var basalType *Suspend
			Expect(basal).To(BeAssignableToTypeOf(basalType))
			Expect(processing.HasErrors()).To(BeFalse())
		})

		Context("validation", func() {

			Context("suppressed", func() {
				suppressed := make(map[string]interface{})
				BeforeEach(func() {
					processing = validate.ErrorProcessing{BasePath: "0", ErrorsArray: validate.NewErrorsArray()}
					suppressed["deliveryType"] = "scheduled"
					suppressed["scheduleName"] = "DEFAULT"
					suppressed["rate"] = 1.75
					basalObj["suppressed"] = suppressed
				})

				It("is not required", func() {
					delete(basalObj, "suppressed")
					basal := Build(basalObj, processing)
					types.GetPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})

				It("when present is validated", func() {
					basal := Build(basalObj, processing)
					types.GetPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})

			})
		})
	})
})
