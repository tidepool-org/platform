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
	basalObj["deliveryType"] = "suspend"
	basalObj["duration"] = 1800000

	Context("suspend from obj", func() {

		BeforeEach(func() {
			processing = validate.ErrorProcessing{BasePath: "0/basal", ErrorsArray: validate.NewErrorsArray()}
		})

		It("should return a basal if the obj is valid", func() {
			basal := Build(basalObj, processing)
			var basalType *Suspend
			Expect(basal).To(BeAssignableToTypeOf(basalType))
			Expect(processing.HasErrors()).To(BeFalse())
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
		})
	})
})
