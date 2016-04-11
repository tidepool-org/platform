package basal

import (
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"
	"github.com/tidepool-org/platform/data/types"

	"github.com/tidepool-org/platform/data/_fixtures"
	"github.com/tidepool-org/platform/validate"
)

var _ = Describe("Temporary", func() {

	var processing validate.ErrorProcessing

	var basalObj = fixtures.TestingDatumBase()
	basalObj["type"] = "basal"
	basalObj["deliveryType"] = "temp"
	basalObj["rate"] = 1.75
	basalObj["percent"] = 0.5
	basalObj["duration"] = 1800000

	Context("Temporary from obj", func() {

		BeforeEach(func() {
			processing = validate.ErrorProcessing{BasePath: "0", ErrorsArray: validate.NewErrorsArray()}
		})

		It("should return a basal if the obj is valid", func() {
			basal := Build(basalObj, processing)
			var basalType *Temporary
			Expect(basal).To(BeAssignableToTypeOf(basalType))
		})

		It("should produce no error when valid", func() {
			Build(basalObj, processing)
			Expect(processing.HasErrors()).To(BeFalse())
		})

		Context("validation", func() {

			Context("rate", func() {

				BeforeEach(func() {
					processing = validate.ErrorProcessing{BasePath: "0", ErrorsArray: validate.NewErrorsArray()}
				})

				It("is not required", func() {
					delete(basalObj, "rate")
					basal := Build(basalObj, processing)
					types.GetPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})

				It("invalid less than zero", func() {
					basalObj["rate"] = -0.1
					basal := Build(basalObj, processing)
					types.GetPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeTrue())
					Expect(processing.Errors[0].Detail).To(ContainSubstring("'Rate' failed with 'Must be greater than 0.0' when given '-0.1'"))

				})

				It("valid when greater than zero", func() {
					basalObj["rate"] = 0.7
					basal := Build(basalObj, processing)
					types.GetPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})

			})

			Context("percent", func() {

				BeforeEach(func() {
					processing = validate.ErrorProcessing{BasePath: "0", ErrorsArray: validate.NewErrorsArray()}
				})

				It("is not required", func() {
					delete(basalObj, "percent")
					basal := Build(basalObj, processing)
					types.GetPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})

				It("invalid less than zero", func() {
					basalObj["percent"] = -0.1
					basal := Build(basalObj, processing)
					types.GetPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeTrue())
					Expect(processing.Errors[0].Detail).To(ContainSubstring("'Percent' failed with 'Must be greater than 0.0' when given '-0.1'"))

				})

				It("invalid when greater than 1.0", func() {
					basalObj["percent"] = 1.1
					basal := Build(basalObj, processing)
					types.GetPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeTrue())
					Expect(processing.Errors[0].Detail).To(ContainSubstring("'Percent' failed with 'Must be greater than 0.0' when given '1.1'"))

				})

				It("valid when between 0.0 and 1.0", func() {
					basalObj["percent"] = 0.7
					basal := Build(basalObj, processing)
					types.GetPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})

			})

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
