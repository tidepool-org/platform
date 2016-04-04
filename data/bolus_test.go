package data

import (
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"

	"github.com/tidepool-org/platform/validate"
)

var _ = Describe("Bolus", func() {

	var bolusObj = testingDatumBase()
	bolusObj["type"] = "bolus"
	bolusObj["subType"] = "injected"
	bolusObj["value"] = 3.0
	bolusObj["insulin"] = "novolog"

	var processing validate.ErrorProcessing

	Context("datum from obj", func() {

		BeforeEach(func() {
			processing = validate.ErrorProcessing{BasePath: "0/bolus", ErrorsArray: validate.NewErrorsArray()}
		})

		It("returns a bolus if the obj is valid", func() {
			bolus := BuildBolus(bolusObj, processing)
			var bolusType *Bolus
			Expect(bolus).To(BeAssignableToTypeOf(bolusType))
		})

		It("produces no error when valid", func() {
			BuildBolus(bolusObj, processing)
			Expect(processing.HasErrors()).To(BeFalse())
		})

	})
	Context("injection", func() {
		var bolusObj = testingDatumBase()
		bolusObj["type"] = "bolus"
		bolusObj["subType"] = "injected"
		bolusObj["value"] = 3.0
		bolusObj["insulin"] = "novolog"
	})
	Context("normal", func() {
		var bolusObj = testingDatumBase()
		bolusObj["type"] = "bolus"
		bolusObj["subType"] = "normal"
		bolusObj["normal"] = 1.0
	})
	Context("square", func() {
		var bolusObj = testingDatumBase()
		bolusObj["type"] = "bolus"
		bolusObj["subType"] = "square"
		bolusObj["extended"] = 1.0
		bolusObj["duration"] = 3600000

	})
	Context("dual/square", func() {
		var bolusObj = testingDatumBase()
		bolusObj["type"] = "bolus"
		bolusObj["subType"] = "dual/square"
		bolusObj["normal"] = 2.0
		bolusObj["extended"] = 1.0
		bolusObj["duration"] = 3600000
	})
	Context("validation", func() {

		Context("duration", func() {
			BeforeEach(func() {
				processing = validate.ErrorProcessing{BasePath: "0/Bolus", ErrorsArray: validate.NewErrorsArray()}
			})
			Context("is invalid when", func() {

				It("zero", func() {
					bolusObj["duration"] = -1
					bolus := BuildBolus(bolusObj, processing)
					getPlatformValidator().Struct(bolus, processing)
					Expect(processing.HasErrors()).To(BeTrue())
					Expect(processing.Errors[0].Detail).To(ContainSubstring("'Duration' failed with 'Must be greater than 0' when given '-1'"))
				})

			})
			Context("is valid when", func() {

				It("greater than zero", func() {
					bolusObj["duration"] = 4000
					bolus := BuildBolus(bolusObj, processing)
					getPlatformValidator().Struct(bolus, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})

			})
		})
		Context("extended", func() {
			BeforeEach(func() {
				processing = validate.ErrorProcessing{BasePath: "0/Bolus", ErrorsArray: validate.NewErrorsArray()}
			})
			Context("is invalid when", func() {

				It("zero", func() {
					bolusObj["extended"] = -0.1
					bolus := BuildBolus(bolusObj, processing)
					getPlatformValidator().Struct(bolus, processing)
					Expect(processing.HasErrors()).To(BeTrue())
					Expect(processing.Errors[0].Detail).To(ContainSubstring("'Extended' failed with 'Must be greater than 0.0' when given '-0.1'"))
				})

			})
			Context("is valid when", func() {

				It("greater than zero", func() {
					bolusObj["extended"] = 0.7
					bolus := BuildBolus(bolusObj, processing)
					getPlatformValidator().Struct(bolus, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})

			})
		})
		Context("normal", func() {
			BeforeEach(func() {
				processing = validate.ErrorProcessing{BasePath: "0/Bolus", ErrorsArray: validate.NewErrorsArray()}
			})
			Context("is invalid when", func() {

				It("zero", func() {
					bolusObj["normal"] = -0.1
					bolus := BuildBolus(bolusObj, processing)
					getPlatformValidator().Struct(bolus, processing)
					Expect(processing.HasErrors()).To(BeTrue())
					Expect(processing.Errors[0].Detail).To(ContainSubstring("'Normal' failed with 'Must be greater than 0.0' when given '-0.1'"))
				})

			})
			Context("is valid when", func() {

				It("greater than zero", func() {
					bolusObj["normal"] = 0.7
					bolus := BuildBolus(bolusObj, processing)
					getPlatformValidator().Struct(bolus, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})

			})
		})
		Context("subType", func() {
			BeforeEach(func() {
				processing = validate.ErrorProcessing{BasePath: "0/Bolus", ErrorsArray: validate.NewErrorsArray()}
			})
			Context("is invalid when", func() {
				It("there is no matching type", func() {
					bolusObj["subType"] = "superfly"
					bolus := BuildBolus(bolusObj, processing)
					getPlatformValidator().Struct(bolus, processing)
					Expect(processing.HasErrors()).To(BeTrue())
					Expect(processing.Errors[0].Detail).To(ContainSubstring("'SubType' failed with 'Must be one of injected, normal, square, dual/square' when given 'superfly'"))
				})
			})
			Context("is valid when", func() {
				It("injected type", func() {
					bolusObj["subType"] = "injected"
					bolus := BuildBolus(bolusObj, processing)
					getPlatformValidator().Struct(bolus, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})
				It("normal type", func() {
					bolusObj["subType"] = "normal"
					bolus := BuildBolus(bolusObj, processing)
					getPlatformValidator().Struct(bolus, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})
				It("square type", func() {
					bolusObj["subType"] = "square"
					bolus := BuildBolus(bolusObj, processing)
					getPlatformValidator().Struct(bolus, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})
				It("dual/square type", func() {
					bolusObj["subType"] = "dual/square"
					bolus := BuildBolus(bolusObj, processing)
					getPlatformValidator().Struct(bolus, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})
			})
		})
		Context("insulin", func() {
			BeforeEach(func() {
				processing = validate.ErrorProcessing{BasePath: "0/bolus", ErrorsArray: validate.NewErrorsArray()}
			})
			Context("is invalid when", func() {

				It("there is no matching type", func() {
					bolusObj["insulin"] = "good"
					bolus := BuildBolus(bolusObj, processing)
					getPlatformValidator().Struct(bolus, processing)
					Expect(processing.HasErrors()).To(BeTrue())
				})

				It("gives detailed error", func() {
					bolusObj["insulin"] = "good"
					bolus := BuildBolus(bolusObj, processing)
					getPlatformValidator().Struct(bolus, processing)
					Expect(processing.HasErrors()).To(BeTrue())
					Expect(processing.Errors[0].Detail).To(ContainSubstring("'Insulin' failed with 'Must be one of novolog, humalog' when given 'good'"))
				})

			})
			Context("is valid when", func() {

				It("novolog type", func() {
					bolusObj["insulin"] = "novolog"
					bolus := BuildBolus(bolusObj, processing)
					getPlatformValidator().Struct(bolus, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})

				It("humalog type", func() {
					bolusObj["insulin"] = "humalog"
					bolus := BuildBolus(bolusObj, processing)
					getPlatformValidator().Struct(bolus, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})
			})
		})
		Context("value", func() {
			BeforeEach(func() {
				processing = validate.ErrorProcessing{BasePath: "0/Bolus", ErrorsArray: validate.NewErrorsArray()}
			})
			Context("is invalid when", func() {

				/*It("zero", func() {
					bolusObj["value"] = -1
					Bolus := BuildBolus(bolusObj, processing)
					getPlatformValidator().Struct(Bolus, processing)
					Expect(processing.HasErrors()).To(BeTrue())
				})

				It("gives detailed error", func() {
					bolusObj["value"] = -1
					Bolus := BuildBolus(bolusObj, processing)
					getPlatformValidator().Struct(Bolus, processing)
					Expect(processing.HasErrors()).To(BeTrue())
					Expect(processing.Errors[0].Detail).To(ContainSubstring("'Insulin' failed with 'Must be one of levemir,lantus' when given 'good'"))
				})*/

			})
			Context("is valid when", func() {

				It("greater than zero", func() {
					bolusObj["value"] = 1
					bolus := BuildBolus(bolusObj, processing)
					getPlatformValidator().Struct(bolus, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})

			})
		})
	})
})
