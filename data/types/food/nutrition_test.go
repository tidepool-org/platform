package food_test

import (
	. "github.com/onsi/ginkgo/v2"
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

var _ = Describe("Nutrition", func() {
	It("EstimatedAbsorptionDuration is expected", func() {
		Expect(dataTypesFood.EstimatedAbsorptionDurationSecondsMaximum).To(Equal(86400))
	})

	It("AbsorptionDurationSecondsMinimum is expected", func() {
		Expect(dataTypesFood.EstimatedAbsorptionDurationSecondsMinimum).To(Equal(0))
	})

	Context("Nutrition", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesFood.Nutrition)) {
				datum := dataTypesFoodTest.RandomNutrition()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesFoodTest.NewObjectFromNutrition(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesFoodTest.NewObjectFromNutrition(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesFood.Nutrition) {},
			),
			Entry("empty",
				func(datum *dataTypesFood.Nutrition) {
					*datum = *dataTypesFood.NewNutrition()
				},
			),
			Entry("all",
				func(datum *dataTypesFood.Nutrition) {
					datum.EstimatedAbsorptionDuration = pointer.FromInt(test.RandomIntFromRange(dataTypesFood.EstimatedAbsorptionDurationSecondsMinimum, dataTypesFood.EstimatedAbsorptionDurationSecondsMaximum))
					datum.Carbohydrate = dataTypesFoodTest.RandomCarbohydrate()
					datum.Energy = dataTypesFoodTest.RandomEnergy()
					datum.Fat = dataTypesFoodTest.RandomFat()
					datum.Protein = dataTypesFoodTest.RandomProtein()
				},
			),
		)

		Context("ParseNutrition", func() {
			It("returns nil when the object is missing", func() {
				Expect(dataTypesFood.ParseNutrition(structureParser.NewObject(nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := dataTypesFoodTest.RandomNutrition()
				object := dataTypesFoodTest.NewObjectFromNutrition(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(&object)
				Expect(dataTypesFood.ParseNutrition(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewNutrition", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesFood.NewNutrition()
				Expect(datum).ToNot(BeNil())
				Expect(datum.EstimatedAbsorptionDuration).To(BeNil())
				Expect(datum.Carbohydrate).To(BeNil())
				Expect(datum.Energy).To(BeNil())
				Expect(datum.Fat).To(BeNil())
				Expect(datum.Protein).To(BeNil())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dataTypesFood.Nutrition), expectedErrors ...error) {
					expectedDatum := dataTypesFoodTest.RandomNutrition()
					object := dataTypesFoodTest.NewObjectFromNutrition(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesFood.NewNutrition()
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataTypesFood.Nutrition) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *dataTypesFood.Nutrition) {
						object["estimatedAbsorptionDuration"] = true
						object["carbohydrate"] = true
						object["energy"] = true
						object["fat"] = true
						object["protein"] = true
						expectedDatum.EstimatedAbsorptionDuration = nil
						expectedDatum.Carbohydrate = nil
						expectedDatum.Energy = nil
						expectedDatum.Fat = nil
						expectedDatum.Protein = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/estimatedAbsorptionDuration"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/carbohydrate"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/energy"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/fat"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/protein"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesFood.Nutrition), expectedErrors ...error) {
					datum := dataTypesFoodTest.RandomNutrition()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesFood.Nutrition) {},
				),
				Entry("absorption duration missing",
					func(datum *dataTypesFood.Nutrition) { datum.EstimatedAbsorptionDuration = nil },
				),
				Entry("absorption duration out of range (lower)",
					func(datum *dataTypesFood.Nutrition) { datum.EstimatedAbsorptionDuration = pointer.FromInt(-1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-1, 0, 86400), "/estimatedAbsorptionDuration"),
				),
				Entry("absorption duration in range (lower)",
					func(datum *dataTypesFood.Nutrition) { datum.EstimatedAbsorptionDuration = pointer.FromInt(0) },
				),
				Entry("absorption duration in range (upper)",
					func(datum *dataTypesFood.Nutrition) { datum.EstimatedAbsorptionDuration = pointer.FromInt(86400) },
				),
				Entry("absorption duration out of range (upper)",
					func(datum *dataTypesFood.Nutrition) { datum.EstimatedAbsorptionDuration = pointer.FromInt(86401) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(86401, 0, 86400), "/estimatedAbsorptionDuration"),
				),
				Entry("carbohydrate missing",
					func(datum *dataTypesFood.Nutrition) { datum.Carbohydrate = nil },
				),
				Entry("carbohydrate invalid",
					func(datum *dataTypesFood.Nutrition) { datum.Carbohydrate.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/carbohydrate/units"),
				),
				Entry("carbohydrate valid",
					func(datum *dataTypesFood.Nutrition) { datum.Carbohydrate = dataTypesFoodTest.RandomCarbohydrate() },
				),
				Entry("energy missing",
					func(datum *dataTypesFood.Nutrition) { datum.Energy = nil },
				),
				Entry("energy invalid",
					func(datum *dataTypesFood.Nutrition) { datum.Energy.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/energy/units"),
				),
				Entry("energy valid",
					func(datum *dataTypesFood.Nutrition) { datum.Energy = dataTypesFoodTest.RandomEnergy() },
				),
				Entry("fat missing",
					func(datum *dataTypesFood.Nutrition) { datum.Fat = nil },
				),
				Entry("fat invalid",
					func(datum *dataTypesFood.Nutrition) { datum.Fat.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/fat/units"),
				),
				Entry("fat valid",
					func(datum *dataTypesFood.Nutrition) { datum.Fat = dataTypesFoodTest.RandomFat() },
				),
				Entry("protein missing",
					func(datum *dataTypesFood.Nutrition) { datum.Protein = nil },
				),
				Entry("protein invalid",
					func(datum *dataTypesFood.Nutrition) { datum.Protein.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/protein/units"),
				),
				Entry("protein valid",
					func(datum *dataTypesFood.Nutrition) { datum.Protein = dataTypesFoodTest.RandomProtein() },
				),
				Entry("multiple errors",
					func(datum *dataTypesFood.Nutrition) {
						datum.EstimatedAbsorptionDuration = pointer.FromInt(-1)
						datum.Carbohydrate.Units = nil
						datum.Energy.Units = nil
						datum.Fat.Units = nil
						datum.Protein.Units = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-1, 0, 86400), "/estimatedAbsorptionDuration"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/carbohydrate/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/energy/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/fat/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/protein/units"),
				),
			)
		})
	})
})
