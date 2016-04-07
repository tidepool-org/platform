package bolus

import (
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/validate"
)

func TestingDatumBase() map[string]interface{} {
	return map[string]interface{}{
		"userId":           "b676436f60",
		"groupId":          "43099shgs55",
		"uploadId":         "upid_b856b0e6e519",
		"deviceTime":       "2014-06-11T06:00:00.000Z",
		"time":             "2014-06-11T06:00:00.000Z",
		"timezoneOffset":   0,
		"conversionOffset": 0,
		"clockDriftOffset": 0,
		"deviceId":         "InsOmn-111111111",
	}
}

var _ = Describe("Bolus", func() {

	var bolusObj = TestingDatumBase()
	bolusObj["type"] = "bolus"
	bolusObj["subType"] = "normal"
	bolusObj["normal"] = 1.0
	var processing validate.ErrorProcessing

	Context("bolus type from obj", func() {

		BeforeEach(func() {
			processing = validate.ErrorProcessing{BasePath: "0", ErrorsArray: validate.NewErrorsArray()}
		})

		It("returns a bolus if the obj is valid", func() {
			bolus := Build(bolusObj, processing)
			var bolusType *Normal
			Expect(bolus).To(BeAssignableToTypeOf(bolusType))
			Expect(processing.HasErrors()).To(BeFalse())
		})

		Context("validation", func() {

			Context("subType", func() {
				BeforeEach(func() {
					processing = validate.ErrorProcessing{BasePath: "0", ErrorsArray: validate.NewErrorsArray()}
				})
				Context("is invalid when", func() {
					It("there is no matching type", func() {
						bolusObj["subType"] = "superfly"
						bolus := Build(bolusObj, processing)
						types.GetPlatformValidator().Struct(bolus, processing)
						Expect(processing.HasErrors()).To(BeTrue())
						Expect(processing.Errors[0].Detail).To(ContainSubstring("'SubType' failed with 'Must be one of normal, square, dual/square' when given 'superfly'"))
					})
					It("injected type is unsupported", func() {
						bolusObj["subType"] = "injected"
						bolus := Build(bolusObj, processing)
						types.GetPlatformValidator().Struct(bolus, processing)
						Expect(processing.HasErrors()).To(BeTrue())
					})
				})
				Context("is valid when", func() {
					It("normal type", func() {
						bolusObj["subType"] = "normal"
						bolus := Build(bolusObj, processing)
						types.GetPlatformValidator().Struct(bolus, processing)
						Expect(processing.HasErrors()).To(BeFalse())
					})
					It("square type", func() {
						bolusObj["subType"] = "square"
						bolus := Build(bolusObj, processing)
						types.GetPlatformValidator().Struct(bolus, processing)
						Expect(processing.HasErrors()).To(BeFalse())
					})
					It("dual/square type", func() {
						bolusObj["subType"] = "dual/square"
						bolus := Build(bolusObj, processing)
						types.GetPlatformValidator().Struct(bolus, processing)
						Expect(processing.HasErrors()).To(BeFalse())
					})
				})

			})
		})
	})
})
