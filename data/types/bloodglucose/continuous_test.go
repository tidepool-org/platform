package bloodglucose

import (
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/_fixtures"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/validate"
)

var _ = Describe("Continuous", func() {
	var bgObj = fixtures.TestingDatumBase()
	var processing validate.ErrorProcessing

	Context("cbg from obj", func() {

		BeforeEach(func() {
			processing = validate.ErrorProcessing{BasePath: "0", ErrorsArray: validate.NewErrorsArray()}
			bgObj["type"] = "cbg"
			bgObj["value"] = 5.5
			bgObj["units"] = "mmol/l"
			bgObj["isig"] = 6.5
		})

		It("returns a bolus if the obj is valid", func() {
			continuous := BuildContinuous(bgObj, processing)
			var bgType *Continuous
			Expect(continuous).To(BeAssignableToTypeOf(bgType))
			Expect(processing.HasErrors()).To(BeFalse())
		})

	})
	Context("validation", func() {

		BeforeEach(func() {
			processing = validate.ErrorProcessing{BasePath: "0", ErrorsArray: validate.NewErrorsArray()}
			bgObj["type"] = "cbg"
			bgObj["value"] = 5.5
			bgObj["units"] = "mmol/l"
			bgObj["isig"] = 6.5
		})

		Context("units", func() {
			It("is required", func() {
				delete(bgObj, "units")
				cbg := BuildContinuous(bgObj, processing)
				types.GetPlatformValidator().Struct(cbg, processing)
				Expect(processing.HasErrors()).To(BeTrue())
			})

			It("can be mmol/l", func() {
				bgObj["units"] = "mmol/l"
				cbg := BuildContinuous(bgObj, processing)
				types.GetPlatformValidator().Struct(cbg, processing)
				Expect(processing.HasErrors()).To(BeFalse())
			})

			It("can be mg/dl", func() {
				bgObj["units"] = "mg/dl"
				cbg := BuildContinuous(bgObj, processing)
				types.GetPlatformValidator().Struct(cbg, processing)
				Expect(processing.HasErrors()).To(BeFalse())
			})

			It("cannot be anything else", func() {
				bgObj["units"] = "grams"
				cbg := BuildContinuous(bgObj, processing)
				types.GetPlatformValidator().Struct(cbg, processing)
				Expect(processing.HasErrors()).To(BeTrue())
			})

		})
		Context("value", func() {
			It("is required", func() {
				delete(bgObj, "value")
				cbg := BuildContinuous(bgObj, processing)
				types.GetPlatformValidator().Struct(cbg, processing)
				Expect(processing.HasErrors()).To(BeTrue())
			})
		})
		Context("isig", func() {

			It("is required", func() {
				delete(bgObj, "isig")
				cbg := BuildContinuous(bgObj, processing)
				types.GetPlatformValidator().Struct(cbg, processing)
				Expect(processing.HasErrors()).To(BeTrue())
			})
		})
	})
})
