package bolus

import (
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"
	"github.com/tidepool-org/platform/data/types"

	"github.com/tidepool-org/platform/validate"
)

var _ = Describe("Bolus", func() {

	var bolusObj = TestingDatumBase()
	bolusObj["type"] = "bolus"
	bolusObj["subType"] = "normal"
	bolusObj["normal"] = 1.0

	var processing validate.ErrorProcessing

	Context("normal from obj", func() {

		BeforeEach(func() {
			processing = validate.ErrorProcessing{BasePath: "0", ErrorsArray: validate.NewErrorsArray()}
		})

		It("returns a bolus if the obj is valid", func() {
			bolus := Build(bolusObj, processing)
			var bolusType *Normal
			Expect(bolus).To(BeAssignableToTypeOf(bolusType))
		})

		It("produces no error when valid", func() {
			Build(bolusObj, processing)
			Expect(processing.HasErrors()).To(BeFalse())
		})

		Context("validation", func() {

			Context("normal", func() {
				BeforeEach(func() {
					processing = validate.ErrorProcessing{BasePath: "0", ErrorsArray: validate.NewErrorsArray()}
				})
				Context("is invalid when", func() {

					It("zero", func() {
						bolusObj["normal"] = -0.1
						bolus := Build(bolusObj, processing)
						types.GetPlatformValidator().Struct(bolus, processing)
						Expect(processing.HasErrors()).To(BeTrue())
						Expect(processing.Errors[0].Detail).To(ContainSubstring("'Normal' failed with 'Must be greater than 0.0' when given '-0.1'"))
					})

				})
				Context("is valid when", func() {

					It("greater than zero", func() {
						bolusObj["normal"] = 0.7
						bolus := Build(bolusObj, processing)
						types.GetPlatformValidator().Struct(bolus, processing)
						Expect(processing.HasErrors()).To(BeFalse())
					})

				})

			})
		})
	})
})
