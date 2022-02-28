package dosingdecision_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataTypesDosingDecision "github.com/tidepool-org/platform/data/types/dosingdecision"
	dataTypesDosingDecisionTest "github.com/tidepool-org/platform/data/types/dosingdecision/test"
	dataTypesFoodTest "github.com/tidepool-org/platform/data/types/food/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Food", func() {
	Context("Food", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesDosingDecision.Food)) {
				datum := dataTypesDosingDecisionTest.RandomFood()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesDosingDecisionTest.NewObjectFromFood(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesDosingDecisionTest.NewObjectFromFood(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesDosingDecision.Food) {},
			),
			Entry("empty",
				func(datum *dataTypesDosingDecision.Food) {
					*datum = *dataTypesDosingDecision.NewFood()
				},
			),
			Entry("all",
				func(datum *dataTypesDosingDecision.Food) {
					datum.Time = pointer.FromTime(test.RandomTime())
					datum.Nutrition = dataTypesFoodTest.RandomNutrition()
				},
			),
		)

		Context("ParseFood", func() {
			It("returns nil when the object is missing", func() {
				Expect(dataTypesDosingDecision.ParseFood(structureParser.NewObject(nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := dataTypesDosingDecisionTest.RandomFood()
				object := dataTypesDosingDecisionTest.NewObjectFromFood(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(&object)
				Expect(dataTypesDosingDecision.ParseFood(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewFood", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesDosingDecision.NewFood()
				Expect(datum).ToNot(BeNil())
				Expect(datum.Time).To(BeNil())
				Expect(datum.Nutrition).To(BeNil())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dataTypesDosingDecision.Food), expectedErrors ...error) {
					expectedDatum := dataTypesDosingDecisionTest.RandomFood()
					object := dataTypesDosingDecisionTest.NewObjectFromFood(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesDosingDecision.NewFood()
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataTypesDosingDecision.Food) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *dataTypesDosingDecision.Food) {
						object["time"] = true
						object["nutrition"] = true
						expectedDatum.Time = nil
						expectedDatum.Nutrition = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/time"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/nutrition"),
				),
			)
		})
		Context("Validate", func() {
			DescribeTable("return the expected results when the input",
				func(mutator func(datum *dataTypesDosingDecision.Food), expectedErrors ...error) {
					datum := dataTypesDosingDecisionTest.RandomFood()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesDosingDecision.Food) {},
				),
				Entry("time missing",
					func(datum *dataTypesDosingDecision.Food) {
						datum.Time = nil
					},
				),
				Entry("time exists",
					func(datum *dataTypesDosingDecision.Food) {
						datum.Time = pointer.FromTime(test.RandomTime())
					},
				),
				Entry("nutrition missing",
					func(datum *dataTypesDosingDecision.Food) { datum.Nutrition = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/nutrition"),
				),
				Entry("nutrition invalid",
					func(datum *dataTypesDosingDecision.Food) {
						datum.Nutrition.EstimatedAbsorptionDuration = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-1, 0, 86400), "/nutrition/estimatedAbsorptionDuration"),
				),
				Entry("nutrition valid",
					func(datum *dataTypesDosingDecision.Food) { datum.Nutrition = dataTypesFoodTest.RandomNutrition() },
				),
				Entry("multiple errors",
					func(datum *dataTypesDosingDecision.Food) {
						datum.Nutrition = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/nutrition"),
				),
			)
		})
	})
})
