package bloodglucose

import (
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"
	"github.com/tidepool-org/platform/data/_fixtures"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/validate"
)

var _ = Describe("Selfmonitored", func() {

	var processing validate.ErrorProcessing
	var bgObj types.Datum

	Context("smbg from obj", func() {

		BeforeEach(func() {
			processing = validate.ErrorProcessing{BasePath: "0", ErrorsArray: validate.NewErrorsArray()}
			bgObj = fixtures.TestingDatumBase()
			bgObj["type"] = "smbg"
			bgObj["value"] = 5.5
			bgObj["units"] = "mmol/l"
		})

		It("returns a bolus if the obj is valid", func() {
			selfMonitored := BuildSelfMonitored(bgObj, processing)
			var bgType *SelfMonitored
			Expect(selfMonitored).To(BeAssignableToTypeOf(bgType))
			Expect(processing.HasErrors()).To(BeFalse())
		})

	})
	Context("validation", func() {

		BeforeEach(func() {
			processing = validate.ErrorProcessing{BasePath: "0", ErrorsArray: validate.NewErrorsArray()}
			bgObj = fixtures.TestingDatumBase()
			bgObj["type"] = "smbg"
			bgObj["value"] = 5.5
			bgObj["units"] = "mmol/l"
		})
		Context("units", func() {
			It("is required", func() {
				delete(bgObj, "units")
				smbg := BuildSelfMonitored(bgObj, processing)
				types.GetPlatformValidator().Struct(smbg, processing)
				Expect(processing.HasErrors()).To(BeTrue())
			})

			It("can be mmol/l", func() {
				bgObj["units"] = "mmol/l"
				smbg := BuildSelfMonitored(bgObj, processing)
				types.GetPlatformValidator().Struct(smbg, processing)
				Expect(processing.HasErrors()).To(BeFalse())
			})

			It("can be mg/dl", func() {
				bgObj["units"] = "mg/dl"
				smbg := BuildSelfMonitored(bgObj, processing)
				types.GetPlatformValidator().Struct(smbg, processing)
				Expect(processing.HasErrors()).To(BeFalse())
			})

			It("cannot be anything else", func() {
				bgObj["units"] = "grams"
				smbg := BuildSelfMonitored(bgObj, processing)
				types.GetPlatformValidator().Struct(smbg, processing)
				Expect(processing.HasErrors()).To(BeTrue())
			})

		})
		Context("value", func() {
			It("is required", func() {
				delete(bgObj, "value")
				smbg := BuildSelfMonitored(bgObj, processing)
				types.GetPlatformValidator().Struct(smbg, processing)
				Expect(processing.HasErrors()).To(BeTrue())
			})

		})
	})
})
