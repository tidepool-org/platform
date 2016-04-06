package basal

import (
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"
	"github.com/tidepool-org/platform/data/types"

	"github.com/tidepool-org/platform/validate"
)

var _ = Describe("Basal", func() {

	var processing validate.ErrorProcessing

	var basalObj = TestingDatumBase()
	basalObj["type"] = "basal"
	basalObj["deliveryType"] = "injected"
	basalObj["value"] = 3.0
	basalObj["insulin"] = "levemir"

	Context("injection from obj", func() {

		BeforeEach(func() {
			processing = validate.ErrorProcessing{BasePath: "0/basal", ErrorsArray: validate.NewErrorsArray()}
		})

		It("should return a basal if the obj is valid", func() {
			basal := Build(basalObj, processing)
			var basalType *Injected
			Expect(basal).To(BeAssignableToTypeOf(basalType))
			Expect(processing.HasErrors()).To(BeFalse())
		})

		Context("validation", func() {

			Context("insulin", func() {
				BeforeEach(func() {
					processing = validate.ErrorProcessing{BasePath: "0/basal", ErrorsArray: validate.NewErrorsArray()}
				})
				Context("is invalid when", func() {

					It("there is no matching type", func() {
						basalObj["insulin"] = "good"
						basal := Build(basalObj, processing)
						types.GetPlatformValidator().Struct(basal, processing)
						Expect(processing.HasErrors()).To(BeTrue())
					})

					It("gives detailed error", func() {
						basalObj["insulin"] = "good"
						basal := Build(basalObj, processing)
						types.GetPlatformValidator().Struct(basal, processing)
						Expect(processing.HasErrors()).To(BeTrue())
						Expect(processing.Errors[0].Detail).To(ContainSubstring("'Insulin' failed with 'Must be one of levemir, lantus' when given 'good'"))
					})

				})
				Context("is valid when", func() {

					It("levemir type", func() {
						basalObj["insulin"] = "levemir"
						basal := Build(basalObj, processing)
						types.GetPlatformValidator().Struct(basal, processing)
						Expect(processing.HasErrors()).To(BeFalse())
					})

					It("lantus type", func() {
						basalObj["insulin"] = "lantus"
						basal := Build(basalObj, processing)
						types.GetPlatformValidator().Struct(basal, processing)
						Expect(processing.HasErrors()).To(BeFalse())
					})
				})
			})
			Context("value", func() {
				BeforeEach(func() {
					processing = validate.ErrorProcessing{BasePath: "0/basal", ErrorsArray: validate.NewErrorsArray()}
				})
				Context("is invalid when", func() {

					/*It("zero", func() {
						basalObj["value"] = -1
						basal := BuildBasal(basalObj, processing)
						types.GetPlatformValidator().Struct(basal, processing)
						Expect(processing.HasErrors()).To(BeTrue())
					})

					It("gives detailed error", func() {
						basalObj["value"] = -1
						basal := BuildBasal(basalObj, processing)
						types.GetPlatformValidator().Struct(basal, processing)
						Expect(processing.HasErrors()).To(BeTrue())
						Expect(processing.Errors[0].Detail).To(ContainSubstring("'Insulin' failed with 'Must be one of levemir, lantus' when given 'good'"))
					})*/

				})
				Context("is valid when", func() {

					It("greater than zero", func() {
						basalObj["value"] = 1
						basal := Build(basalObj, processing)
						types.GetPlatformValidator().Struct(basal, processing)
						Expect(processing.HasErrors()).To(BeFalse())
					})

				})
			})
		})
	})
})
