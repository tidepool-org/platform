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
	bolusObj["subType"] = "injected"
	bolusObj["value"] = 3.0
	bolusObj["insulin"] = "novolog"

	var processing validate.ErrorProcessing

	Context("injected from obj", func() {

		BeforeEach(func() {
			processing = validate.ErrorProcessing{BasePath: "0", ErrorsArray: validate.NewErrorsArray()}
		})

		It("returns a bolus if the obj is valid", func() {
			bolus := Build(bolusObj, processing)
			var bolusType *Injected
			Expect(bolus).To(BeAssignableToTypeOf(bolusType))
		})

		It("produces no error when valid", func() {
			Build(bolusObj, processing)
			Expect(processing.HasErrors()).To(BeFalse())
		})

		Context("validation", func() {

			Context("insulin", func() {
				BeforeEach(func() {
					processing = validate.ErrorProcessing{BasePath: "0", ErrorsArray: validate.NewErrorsArray()}
				})
				Context("is invalid when", func() {

					It("there is no matching type", func() {
						bolusObj["insulin"] = "good"
						bolus := Build(bolusObj, processing)
						types.GetPlatformValidator().Struct(bolus, processing)
						Expect(processing.HasErrors()).To(BeTrue())
					})

					It("gives detailed error", func() {
						bolusObj["insulin"] = "good"
						bolus := Build(bolusObj, processing)
						types.GetPlatformValidator().Struct(bolus, processing)
						Expect(processing.HasErrors()).To(BeTrue())
						Expect(processing.Errors[0].Detail).To(ContainSubstring("'Insulin' failed with 'Must be one of novolog, humalog' when given 'good'"))
					})

				})
				Context("is valid when", func() {

					It("novolog type", func() {
						bolusObj["insulin"] = "novolog"
						bolus := Build(bolusObj, processing)
						types.GetPlatformValidator().Struct(bolus, processing)
						Expect(processing.HasErrors()).To(BeFalse())
					})

					It("humalog type", func() {
						bolusObj["insulin"] = "humalog"
						bolus := Build(bolusObj, processing)
						types.GetPlatformValidator().Struct(bolus, processing)
						Expect(processing.HasErrors()).To(BeFalse())
					})
				})
			})
			Context("value", func() {
				BeforeEach(func() {
					processing = validate.ErrorProcessing{BasePath: "0", ErrorsArray: validate.NewErrorsArray()}
				})
				Context("is invalid when", func() {

					/*It("zero", func() {
						bolusObj["value"] = -1
						Bolus := Build(bolusObj, processing)
						types.GetPlatformValidator().Struct(Bolus, processing)
						Expect(processing.HasErrors()).To(BeTrue())
					})

					It("gives detailed error", func() {
						bolusObj["value"] = -1
						Bolus := Build(bolusObj, processing)
						types.GetPlatformValidator().Struct(Bolus, processing)
						Expect(processing.HasErrors()).To(BeTrue())
						Expect(processing.Errors[0].Detail).To(ContainSubstring("'Insulin' failed with 'Must be one of levemir,lantus' when given 'good'"))
					})*/

				})
				Context("is valid when", func() {

					It("greater than zero", func() {
						bolusObj["value"] = 1
						bolus := Build(bolusObj, processing)
						types.GetPlatformValidator().Struct(bolus, processing)
						Expect(processing.HasErrors()).To(BeFalse())
					})

				})
			})
		})
	})
})
