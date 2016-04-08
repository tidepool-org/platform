package basal

import (
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/_fixtures"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/validate"
)

var _ = Describe("Basal", func() {

	var processing validate.ErrorProcessing

	var basalObj = fixtures.TestingDatumBase()
	basalObj["type"] = "basal"
	basalObj["deliveryType"] = "scheduled"
	basalObj["duration"] = 28800000

	Context("datum from obj", func() {

		BeforeEach(func() {
			processing = validate.ErrorProcessing{BasePath: "0/basal", ErrorsArray: validate.NewErrorsArray()}
		})

		It("should return a basal if the obj is valid", func() {
			basal := Build(basalObj, processing)
			var basalType *Base
			Expect(basal).To(BeAssignableToTypeOf(basalType))
			Expect(processing.HasErrors()).To(BeFalse())
		})

	})

	Context("validation", func() {

		Context("duration", func() {
			BeforeEach(func() {
				processing = validate.ErrorProcessing{BasePath: "0/basal", ErrorsArray: validate.NewErrorsArray()}
			})
			Context("is invalid when", func() {

				It("zero", func() {
					basalObj["duration"] = -1
					basal := Build(basalObj, processing)
					types.GetPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeTrue())
				})

				It("gives detailed error", func() {
					basalObj["duration"] = -1
					basal := Build(basalObj, processing)
					types.GetPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeTrue())
					Expect(processing.Errors[0].Detail).To(ContainSubstring("'Duration' failed with 'Must be greater than 0' when given '-1'"))
				})

			})
			Context("is valid when", func() {

				It("greater than zero", func() {
					basalObj["duration"] = 4000
					basal := Build(basalObj, processing)
					types.GetPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})

			})
		})

		Context("deliveryType", func() {
			BeforeEach(func() {
				processing = validate.ErrorProcessing{BasePath: "0/basal", ErrorsArray: validate.NewErrorsArray()}
			})
			Context("is invalid when", func() {

				It("there is no matching type", func() {
					basalObj["deliveryType"] = "superfly"
					basal := Build(basalObj, processing)
					types.GetPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeTrue())
				})

				It("gives detailed error", func() {
					basalObj["deliveryType"] = "superfly"
					basal := Build(basalObj, processing)
					types.GetPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeTrue())
					Expect(processing.Errors[0].Detail).To(ContainSubstring("'DeliveryType' failed with 'Must be one of scheduled, suspend, temp' when given 'superfly'"))
				})

				It("injected type is unsupported", func() {
					basalObj["deliveryType"] = "injected"
					basal := Build(basalObj, processing)
					types.GetPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeTrue())
				})
			})
			Context("is valid when", func() {

				It("scheduled type", func() {
					basalObj["deliveryType"] = "scheduled"
					basal := Build(basalObj, processing)
					types.GetPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})

				It("suspend type", func() {
					basalObj["deliveryType"] = "suspend"
					basal := Build(basalObj, processing)
					types.GetPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})

				It("temp type", func() {
					basalObj["deliveryType"] = "temp"
					basal := Build(basalObj, processing)
					types.GetPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})
			})
		})
	})
})
