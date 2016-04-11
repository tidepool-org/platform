package bolus

import (
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/_fixtures"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/validate"
)

var _ = Describe("Bolus", func() {

	var bolusObj = fixtures.TestingDatumBase()
	bolusObj["type"] = "bolus"
	bolusObj["subType"] = "normal"
	bolusObj["normal"] = 1.0
	var processing validate.ErrorProcessing

	Context("type from obj", func() {

		BeforeEach(func() {
			processing = validate.ErrorProcessing{BasePath: "0", ErrorsArray: validate.NewErrorsArray()}
		})

		It("returns a valid bolus", func() {
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

				It("is required", func() {
					delete(bolusObj, "subType")
					bolus := Build(bolusObj, processing)
					types.GetPlatformValidator().Struct(bolus, processing)
					Expect(processing.HasErrors()).To(BeTrue())
				})

				It("invalid when no matching subType", func() {
					bolusObj["subType"] = "superfly"
					bolus := Build(bolusObj, processing)
					types.GetPlatformValidator().Struct(bolus, processing)
					Expect(processing.HasErrors()).To(BeTrue())
					Expect(processing.Errors[0].Detail).To(ContainSubstring("'SubType' failed with 'Must be one of normal, square, dual/square' when given 'superfly'"))
				})

				It("injected type is not supported", func() {
					bolusObj["subType"] = "injected"
					bolus := Build(bolusObj, processing)
					types.GetPlatformValidator().Struct(bolus, processing)
					Expect(processing.HasErrors()).To(BeTrue())
				})

				It("normal type is supported", func() {
					bolusObj["subType"] = "normal"
					bolus := Build(bolusObj, processing)
					types.GetPlatformValidator().Struct(bolus, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})

				It("square type is supported", func() {
					bolusObj["subType"] = "square"
					bolus := Build(bolusObj, processing)
					types.GetPlatformValidator().Struct(bolus, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})

				It("dual/square type is supported", func() {
					bolusObj["subType"] = "dual/square"
					bolus := Build(bolusObj, processing)
					types.GetPlatformValidator().Struct(bolus, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})

			})
		})
	})
})
