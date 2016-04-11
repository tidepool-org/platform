package bolus

import (
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"
	"github.com/tidepool-org/platform/data/types"

	"github.com/tidepool-org/platform/data/_fixtures"
	"github.com/tidepool-org/platform/validate"
)

var _ = Describe("Normal", func() {

	var bolusObj = fixtures.TestingDatumBase()
	bolusObj["type"] = "bolus"
	bolusObj["subType"] = "normal"
	bolusObj["normal"] = 1.0

	var processing validate.ErrorProcessing

	Context("from obj", func() {

		BeforeEach(func() {
			processing = validate.ErrorProcessing{BasePath: "0", ErrorsArray: validate.NewErrorsArray()}
		})

		It("if the obj is valid", func() {
			bolus := Build(bolusObj, processing)
			var bolusType *Normal
			Expect(bolus).To(BeAssignableToTypeOf(bolusType))
			Expect(processing.HasErrors()).To(BeFalse())
		})

		Context("validation", func() {

			Context("normal", func() {
				BeforeEach(func() {
					processing = validate.ErrorProcessing{BasePath: "0", ErrorsArray: validate.NewErrorsArray()}
				})

				It("is not required", func() {
					delete(bolusObj, "normal")
					bolus := Build(bolusObj, processing)
					types.GetPlatformValidator().Struct(bolus, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})

				It("invalid when less than 0.0", func() {
					bolusObj["normal"] = -0.1
					bolus := Build(bolusObj, processing)
					types.GetPlatformValidator().Struct(bolus, processing)
					Expect(processing.HasErrors()).To(BeTrue())
					Expect(processing.Errors[0].Detail).To(ContainSubstring("'Normal' failed with 'Must be greater than 0.0' when given '-0.1'"))
				})

				It("valid when than 0.0", func() {
					bolusObj["normal"] = 0.7
					bolus := Build(bolusObj, processing)
					types.GetPlatformValidator().Struct(bolus, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})

			})
		})
	})
})
