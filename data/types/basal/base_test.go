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
	basalObj["rate"] = 1.0
	basalObj["duration"] = 28800000

	Context("type from obj", func() {

		BeforeEach(func() {
			processing = validate.ErrorProcessing{BasePath: "0", ErrorsArray: validate.NewErrorsArray()}
		})

		It("returns a valid basal type", func() {
			basal := Build(basalObj, processing)
			var basalType *Scheduled
			Expect(basal).To(BeAssignableToTypeOf(basalType))
			Expect(processing.HasErrors()).To(BeFalse())
		})

	})

	Context("validation", func() {

		Context("duration", func() {
			BeforeEach(func() {
				processing = validate.ErrorProcessing{BasePath: "0", ErrorsArray: validate.NewErrorsArray()}
			})

			It("is not required", func() {
				delete(basalObj, "duration")
				basal := Build(basalObj, processing)
				types.GetPlatformValidator().Struct(basal, processing)
				Expect(processing.HasErrors()).To(BeFalse())
			})

			It("fails if less than zero", func() {
				basalObj["duration"] = -1
				basal := Build(basalObj, processing)
				types.GetPlatformValidator().Struct(basal, processing)
				Expect(processing.HasErrors()).To(BeTrue())
				Expect(processing.Errors[0].Detail).To(ContainSubstring("'Duration' failed with 'Must be greater than 0' when given '-1'"))
			})

			It("valid when greater than zero", func() {
				basalObj["duration"] = 4000
				basal := Build(basalObj, processing)
				types.GetPlatformValidator().Struct(basal, processing)
				Expect(processing.HasErrors()).To(BeFalse())
			})

		})

		Context("deliveryType", func() {
			BeforeEach(func() {
				processing = validate.ErrorProcessing{BasePath: "0", ErrorsArray: validate.NewErrorsArray()}
			})

			It("is required", func() {
				delete(basalObj, "deliveryType")
				basal := Build(basalObj, processing)
				types.GetPlatformValidator().Struct(basal, processing)
				Expect(processing.HasErrors()).To(BeTrue())
				Expect(processing.Errors[0].Detail).To(Equal("'DeliveryType' failed with 'required' when given '<nil>'"))
			})

			It("invalid when no matching type", func() {
				basalObj["deliveryType"] = "superfly"
				basal := Build(basalObj, processing)
				types.GetPlatformValidator().Struct(basal, processing)
				Expect(processing.HasErrors()).To(BeTrue())
				Expect(processing.Errors[0].Detail).To(ContainSubstring("'DeliveryType' failed with 'Must be one of scheduled, suspend, temp' when given 'superfly'"))

			})

			It("valid if unsupported injected type", func() {
				basalObj["deliveryType"] = "injected"
				basal := Build(basalObj, processing)
				types.GetPlatformValidator().Struct(basal, processing)
				Expect(processing.HasErrors()).To(BeTrue())
			})

			It("valid if scheduled type", func() {
				basalObj["deliveryType"] = "scheduled"
				basal := Build(basalObj, processing)
				types.GetPlatformValidator().Struct(basal, processing)
				Expect(processing.HasErrors()).To(BeFalse())
			})

			It("valid if suspend type", func() {
				basalObj["deliveryType"] = "suspend"
				basal := Build(basalObj, processing)
				types.GetPlatformValidator().Struct(basal, processing)
				Expect(processing.HasErrors()).To(BeFalse())
			})

			It("valid if temp type", func() {
				basalObj["deliveryType"] = "temp"
				basal := Build(basalObj, processing)
				types.GetPlatformValidator().Struct(basal, processing)
				Expect(processing.HasErrors()).To(BeFalse())
			})

		})
	})
})
