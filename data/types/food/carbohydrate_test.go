package food_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataTypesFood "github.com/tidepool-org/platform/data/types/food"
	dataTypesFoodTest "github.com/tidepool-org/platform/data/types/food/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Carbohydrate", func() {
	It("CarbohydrateDietaryFiberGramsMaximum is expected", func() {
		Expect(dataTypesFood.CarbohydrateDietaryFiberGramsMaximum).To(Equal(1000.0))
	})

	It("CarbohydrateDietaryFiberGramsMinimum is expected", func() {
		Expect(dataTypesFood.CarbohydrateDietaryFiberGramsMinimum).To(Equal(0.0))
	})

	It("CarbohydrateNetGramsMaximum is expected", func() {
		Expect(dataTypesFood.CarbohydrateNetGramsMaximum).To(Equal(1000.0))
	})

	It("CarbohydrateNetGramsMinimum is expected", func() {
		Expect(dataTypesFood.CarbohydrateNetGramsMinimum).To(Equal(0.0))
	})

	It("CarbohydrateSugarsGramsMaximum is expected", func() {
		Expect(dataTypesFood.CarbohydrateSugarsGramsMaximum).To(Equal(1000.0))
	})

	It("CarbohydrateSugarsGramsMinimum is expected", func() {
		Expect(dataTypesFood.CarbohydrateSugarsGramsMinimum).To(Equal(0.0))
	})

	It("CarbohydrateTotalGramsMaximum is expected", func() {
		Expect(dataTypesFood.CarbohydrateTotalGramsMaximum).To(Equal(1000.0))
	})

	It("CarbohydrateTotalGramsMinimum is expected", func() {
		Expect(dataTypesFood.CarbohydrateTotalGramsMinimum).To(Equal(0.0))
	})

	It("CarbohydrateUnitsGrams is expected", func() {
		Expect(dataTypesFood.CarbohydrateUnitsGrams).To(Equal("grams"))
	})

	It("CarbohydrateUnits returns expected", func() {
		Expect(dataTypesFood.CarbohydrateUnits()).To(Equal([]string{"grams"}))
	})

	Context("Carbohydrate", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesFood.Carbohydrate)) {
				datum := dataTypesFoodTest.RandomCarbohydrate()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesFoodTest.NewObjectFromCarbohydrate(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesFoodTest.NewObjectFromCarbohydrate(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesFood.Carbohydrate) {},
			),
			Entry("empty",
				func(datum *dataTypesFood.Carbohydrate) {
					*datum = *dataTypesFood.NewCarbohydrate()
				},
			),
			Entry("all",
				func(datum *dataTypesFood.Carbohydrate) {
					datum.DietaryFiber = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesFood.CarbohydrateDietaryFiberGramsMinimum, dataTypesFood.CarbohydrateDietaryFiberGramsMaximum))
					datum.Net = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesFood.CarbohydrateNetGramsMinimum, dataTypesFood.CarbohydrateNetGramsMaximum))
					datum.Sugars = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesFood.CarbohydrateSugarsGramsMinimum, dataTypesFood.CarbohydrateSugarsGramsMaximum))
					datum.Total = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesFood.CarbohydrateTotalGramsMinimum, dataTypesFood.CarbohydrateTotalGramsMaximum))
					datum.Units = pointer.FromString(test.RandomStringFromArray(dataTypesFood.CarbohydrateUnits()))
				},
			),
		)

		Context("ParseCarbohydrate", func() {
			It("returns nil when the object is missing", func() {
				Expect(dataTypesFood.ParseCarbohydrate(structureParser.NewObject(nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := dataTypesFoodTest.RandomCarbohydrate()
				object := dataTypesFoodTest.NewObjectFromCarbohydrate(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(&object)
				Expect(dataTypesFood.ParseCarbohydrate(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewCarbohydrate", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesFood.NewCarbohydrate()
				Expect(datum).ToNot(BeNil())
				Expect(datum.DietaryFiber).To(BeNil())
				Expect(datum.Net).To(BeNil())
				Expect(datum.Sugars).To(BeNil())
				Expect(datum.Total).To(BeNil())
				Expect(datum.Units).To(BeNil())
			})
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesFood.Carbohydrate), expectedErrors ...error) {
					datum := dataTypesFoodTest.RandomCarbohydrate()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesFood.Carbohydrate) {},
				),

				Entry("dietary fiber missing",
					func(datum *dataTypesFood.Carbohydrate) { datum.DietaryFiber = nil },
				),
				Entry("dietary fiber out of range (lower)",
					func(datum *dataTypesFood.Carbohydrate) { datum.DietaryFiber = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/dietaryFiber"),
				),
				Entry("dietary fiber in range (lower)",
					func(datum *dataTypesFood.Carbohydrate) { datum.DietaryFiber = pointer.FromFloat64(0.0) },
				),
				Entry("dietary fiber in range (upper)",
					func(datum *dataTypesFood.Carbohydrate) { datum.DietaryFiber = pointer.FromFloat64(1000.0) },
				),
				Entry("dietary fiber out of range (upper)",
					func(datum *dataTypesFood.Carbohydrate) { datum.DietaryFiber = pointer.FromFloat64(1000.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/dietaryFiber"),
				),
				Entry("net missing",
					func(datum *dataTypesFood.Carbohydrate) { datum.Net = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/net"),
				),
				Entry("net out of range (lower)",
					func(datum *dataTypesFood.Carbohydrate) { datum.Net = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/net"),
				),
				Entry("net in range (lower)",
					func(datum *dataTypesFood.Carbohydrate) { datum.Net = pointer.FromFloat64(0.0) },
				),
				Entry("net in range (upper)",
					func(datum *dataTypesFood.Carbohydrate) { datum.Net = pointer.FromFloat64(1000.0) },
				),
				Entry("net out of range (upper)",
					func(datum *dataTypesFood.Carbohydrate) { datum.Net = pointer.FromFloat64(1000.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/net"),
				),
				Entry("sugars missing",
					func(datum *dataTypesFood.Carbohydrate) { datum.Sugars = nil },
				),
				Entry("sugars out of range (lower)",
					func(datum *dataTypesFood.Carbohydrate) { datum.Sugars = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/sugars"),
				),
				Entry("sugars in range (lower)",
					func(datum *dataTypesFood.Carbohydrate) { datum.Sugars = pointer.FromFloat64(0.0) },
				),
				Entry("sugars in range (upper)",
					func(datum *dataTypesFood.Carbohydrate) { datum.Sugars = pointer.FromFloat64(1000.0) },
				),
				Entry("sugars out of range (upper)",
					func(datum *dataTypesFood.Carbohydrate) { datum.Sugars = pointer.FromFloat64(1000.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/sugars"),
				),
				Entry("total missing",
					func(datum *dataTypesFood.Carbohydrate) { datum.Total = nil },
				),
				Entry("total out of range (lower)",
					func(datum *dataTypesFood.Carbohydrate) { datum.Total = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/total"),
				),
				Entry("total in range (lower)",
					func(datum *dataTypesFood.Carbohydrate) { datum.Total = pointer.FromFloat64(0.0) },
				),
				Entry("total in range (upper)",
					func(datum *dataTypesFood.Carbohydrate) { datum.Total = pointer.FromFloat64(1000.0) },
				),
				Entry("total out of range (upper)",
					func(datum *dataTypesFood.Carbohydrate) { datum.Total = pointer.FromFloat64(1000.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/total"),
				),
				Entry("units missing",
					func(datum *dataTypesFood.Carbohydrate) { datum.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid",
					func(datum *dataTypesFood.Carbohydrate) { datum.Units = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"grams"}), "/units"),
				),
				Entry("units grams",
					func(datum *dataTypesFood.Carbohydrate) { datum.Units = pointer.FromString("grams") },
				),
				Entry("multiple errors",
					func(datum *dataTypesFood.Carbohydrate) {
						datum.DietaryFiber = pointer.FromFloat64(-0.1)
						datum.Net = pointer.FromFloat64(-0.1)
						datum.Sugars = pointer.FromFloat64(-0.1)
						datum.Total = pointer.FromFloat64(-0.1)
						datum.Units = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/dietaryFiber"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/net"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/sugars"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/total"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
			)
		})
	})
})
