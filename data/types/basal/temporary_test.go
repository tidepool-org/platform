package basal

import (
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"
	"github.com/tidepool-org/platform/data/types"

	"github.com/tidepool-org/platform/data/_fixtures"
	"github.com/tidepool-org/platform/validate"
)

var _ = Describe("Basal", func() {

	var processing validate.ErrorProcessing

	var basalObj = fixtures.TestingDatumBase()
	basalObj["type"] = "basal"
	basalObj["deliveryType"] = "temp"
	basalObj["rate"] = 1.75
	basalObj["percent"] = 0.5
	basalObj["duration"] = 1800000

	Context("Temporary from obj", func() {

		BeforeEach(func() {
			processing = validate.ErrorProcessing{BasePath: "0/basal", ErrorsArray: validate.NewErrorsArray()}
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

			Context("percent", func() {

				BeforeEach(func() {
					processing = validate.ErrorProcessing{BasePath: "0/basal", ErrorsArray: validate.NewErrorsArray()}
				})

				Context("is invalid when", func() {

					It("zero", func() {
						basalObj["rate"] = -1.0
						basal := Build(basalObj, processing)
						types.GetPlatformValidator().Struct(basal, processing)
						Expect(processing.HasErrors()).To(BeTrue())
					})

					It("zero and gives me context in error", func() {
						basalObj["rate"] = -0.1
						basal := Build(basalObj, processing)
						types.GetPlatformValidator().Struct(basal, processing)
						Expect(processing.HasErrors()).To(BeTrue())
						Expect(processing.Errors[0].Detail).To(ContainSubstring("'Rate' failed with 'Must be greater than 0.0' when given '-0.1'"))
					})

				})
				Context("is valid when", func() {

					It("greater than zero", func() {
						basalObj["rate"] = 0.7
						basal := Build(basalObj, processing)
						types.GetPlatformValidator().Struct(basal, processing)
						Expect(processing.HasErrors()).To(BeFalse())
					})

				})
			})

		})
	})
})
