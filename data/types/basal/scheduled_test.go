package basal

import (
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"
	"github.com/tidepool-org/platform/data/types"

	"github.com/tidepool-org/platform/data/_fixtures"
	"github.com/tidepool-org/platform/validate"
)

var _ = Describe("Scheduled", func() {

	var processing validate.ErrorProcessing

	var basalObj = fixtures.TestingDatumBase()
	basalObj["type"] = "basal"
	basalObj["deliveryType"] = "scheduled"
	basalObj["scheduleName"] = "DEFAULT"
	basalObj["rate"] = 1.75
	basalObj["duration"] = 7200000

	Context("from obj", func() {

		BeforeEach(func() {
			processing = validate.ErrorProcessing{BasePath: "0", ErrorsArray: validate.NewErrorsArray()}
		})

		It("should return a basal if the obj is valid", func() {
			basal := Build(basalObj, processing)
			var basalType *Scheduled
			Expect(basal).To(BeAssignableToTypeOf(basalType))
			Expect(processing.HasErrors()).To(BeFalse())
		})

		Context("validation", func() {

			Context("rate", func() {

				BeforeEach(func() {
					processing = validate.ErrorProcessing{BasePath: "0", ErrorsArray: validate.NewErrorsArray()}
				})

				It("is required", func() {
					delete(basalObj, "rate")
					basal := Build(basalObj, processing)
					types.GetPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeTrue())
				})

				It("invalid when zero", func() {
					basalObj["rate"] = 0.0
					basal := Build(basalObj, processing)
					types.GetPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeTrue())
					Expect(processing.Errors[0].Detail).To(Equal("'Rate' failed with 'required' when given '0'"))
				})

				It("valid when greater than zero", func() {
					basalObj["rate"] = 0.7
					basal := Build(basalObj, processing)
					types.GetPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})

			})
			Context("scheduleName", func() {

				BeforeEach(func() {
					processing = validate.ErrorProcessing{BasePath: "0", ErrorsArray: validate.NewErrorsArray()}
				})

				It("is not required", func() {
					delete(basalObj, "scheduleName")
					basal := Build(basalObj, processing)
					types.GetPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})

				It("is free text", func() {
					basalObj["scheduleName"] = "my schedule"
					basal := Build(basalObj, processing)
					types.GetPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})

			})
		})
	})
})
