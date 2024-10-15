package food_test

import (
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataTypesFood "github.com/tidepool-org/platform/data/types/food"
	dataTypesFoodTest "github.com/tidepool-org/platform/data/types/food/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Energy", func() {
	It("EnergyKilojoulesPerKilocalorie is expected", func() {
		Expect(dataTypesFood.EnergyKilojoulesPerKilocalorie).To(Equal(4.1858))
	})

	It("EnergyUnitsCalories is expected", func() {
		Expect(dataTypesFood.EnergyUnitsCalories).To(Equal("calories"))
	})

	It("EnergyUnitsJoules is expected", func() {
		Expect(dataTypesFood.EnergyUnitsJoules).To(Equal("joules"))
	})

	It("EnergyUnitsKilocalories is expected", func() {
		Expect(dataTypesFood.EnergyUnitsKilocalories).To(Equal("kilocalories"))
	})

	It("EnergyUnitsKilojoules is expected", func() {
		Expect(dataTypesFood.EnergyUnitsKilojoules).To(Equal("kilojoules"))
	})

	It("EnergyValueCaloriesMaximum is expected", func() {
		Expect(dataTypesFood.EnergyValueCaloriesMaximum).To(Equal(10000000.0))
	})

	It("EnergyValueCaloriesMinimum is expected", func() {
		Expect(dataTypesFood.EnergyValueCaloriesMinimum).To(Equal(0.0))
	})

	It("EnergyValueJoulesMaximum is expected", func() {
		Expect(dataTypesFood.EnergyValueJoulesMaximum).To(Equal(41858000.0))
	})

	It("EnergyValueJoulesMinimum is expected", func() {
		Expect(dataTypesFood.EnergyValueJoulesMinimum).To(Equal(0.0))
	})

	It("EnergyValueKilocaloriesMaximum is expected", func() {
		Expect(dataTypesFood.EnergyValueKilocaloriesMaximum).To(Equal(10000.0))
	})

	It("EnergyValueKilocaloriesMinimum is expected", func() {
		Expect(dataTypesFood.EnergyValueKilocaloriesMinimum).To(Equal(0.0))
	})

	It("EnergyValueKilojoulesMaximum is expected", func() {
		Expect(dataTypesFood.EnergyValueKilojoulesMaximum).To(Equal(41858.0))
	})

	It("EnergyValueKilojoulesMinimum is expected", func() {
		Expect(dataTypesFood.EnergyValueKilojoulesMinimum).To(Equal(0.0))
	})

	It("EnergyUnits returns expected", func() {
		Expect(dataTypesFood.EnergyUnits()).To(Equal([]string{"calories", "joules", "kilocalories", "kilojoules"}))
	})

	Context("Energy", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesFood.Energy)) {
				datum := dataTypesFoodTest.RandomEnergy()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesFoodTest.NewObjectFromEnergy(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesFoodTest.NewObjectFromEnergy(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesFood.Energy) {},
			),
			Entry("empty",
				func(datum *dataTypesFood.Energy) {
					*datum = *dataTypesFood.NewEnergy()
				},
			),
			Entry("all",
				func(datum *dataTypesFood.Energy) {
					datum.Units = pointer.FromString(test.RandomStringFromArray(dataTypesFood.EnergyUnits()))
					datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesFood.EnergyValueCaloriesMinimum, dataTypesFood.EnergyValueCaloriesMaximum))
				},
			),
		)

		Context("ParseEnergy", func() {
			It("returns nil when the object is missing", func() {
				Expect(dataTypesFood.ParseEnergy(structureParser.NewObject(logTest.NewLogger(), nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := dataTypesFoodTest.RandomEnergy()
				object := dataTypesFoodTest.NewObjectFromEnergy(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(logTest.NewLogger(), &object)
				Expect(dataTypesFood.ParseEnergy(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewEnergy", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesFood.NewEnergy()
				Expect(datum).ToNot(BeNil())
				Expect(datum.Units).To(BeNil())
				Expect(datum.Value).To(BeNil())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dataTypesFood.Energy), expectedErrors ...error) {
					expectedDatum := dataTypesFoodTest.RandomEnergy()
					object := dataTypesFoodTest.NewObjectFromEnergy(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesFood.NewEnergy()
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataTypesFood.Energy) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *dataTypesFood.Energy) {
						object["units"] = true
						object["value"] = true
						expectedDatum.Units = nil
						expectedDatum.Value = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/units"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotFloat64(true), "/value"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesFood.Energy), expectedErrors ...error) {
					datum := dataTypesFoodTest.RandomEnergy()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesFood.Energy) {},
				),
				Entry("units missing",
					func(datum *dataTypesFood.Energy) { datum.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid",
					func(datum *dataTypesFood.Energy) { datum.Units = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"calories", "joules", "kilocalories", "kilojoules"}), "/units"),
				),
				Entry("units calories",
					func(datum *dataTypesFood.Energy) {
						datum.Units = pointer.FromString("calories")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units joules",
					func(datum *dataTypesFood.Energy) {
						datum.Units = pointer.FromString("joules")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units kilocalories",
					func(datum *dataTypesFood.Energy) {
						datum.Units = pointer.FromString("kilocalories")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units kilojoules",
					func(datum *dataTypesFood.Energy) {
						datum.Units = pointer.FromString("kilojoules")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units missing; value missing",
					func(datum *dataTypesFood.Energy) {
						datum.Units = nil
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units missing; value out of range (lower)",
					func(datum *dataTypesFood.Energy) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value in range (lower)",
					func(datum *dataTypesFood.Energy) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value in range (upper)",
					func(datum *dataTypesFood.Energy) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(41858000.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value out of range (upper)",
					func(datum *dataTypesFood.Energy) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(41858000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid; value missing",
					func(datum *dataTypesFood.Energy) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"calories", "joules", "kilocalories", "kilojoules"}), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units invalid; value out of range (lower)",
					func(datum *dataTypesFood.Energy) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"calories", "joules", "kilocalories", "kilojoules"}), "/units"),
				),
				Entry("units invalid; value in range (lower)",
					func(datum *dataTypesFood.Energy) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"calories", "joules", "kilocalories", "kilojoules"}), "/units"),
				),
				Entry("units invalid; value in range (upper)",
					func(datum *dataTypesFood.Energy) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(41858000.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"calories", "joules", "kilocalories", "kilojoules"}), "/units"),
				),
				Entry("units invalid; value out of range (upper)",
					func(datum *dataTypesFood.Energy) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(41858000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"calories", "joules", "kilocalories", "kilojoules"}), "/units"),
				),
				Entry("units calories; value missing",
					func(datum *dataTypesFood.Energy) {
						datum.Units = pointer.FromString("calories")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units calories; value out of range (lower)",
					func(datum *dataTypesFood.Energy) {
						datum.Units = pointer.FromString("calories")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 10000000.0), "/value"),
				),
				Entry("units calories; value in range (lower)",
					func(datum *dataTypesFood.Energy) {
						datum.Units = pointer.FromString("calories")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units calories; value in range (upper)",
					func(datum *dataTypesFood.Energy) {
						datum.Units = pointer.FromString("calories")
						datum.Value = pointer.FromFloat64(10000000.0)
					},
				),
				Entry("units calories; value out of range (upper)",
					func(datum *dataTypesFood.Energy) {
						datum.Units = pointer.FromString("calories")
						datum.Value = pointer.FromFloat64(10000000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(10000000.1, 0.0, 10000000.0), "/value"),
				),
				Entry("units joules; value missing",
					func(datum *dataTypesFood.Energy) {
						datum.Units = pointer.FromString("joules")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units joules; value out of range (lower)",
					func(datum *dataTypesFood.Energy) {
						datum.Units = pointer.FromString("joules")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 41858000.0), "/value"),
				),
				Entry("units joules; value in range (lower)",
					func(datum *dataTypesFood.Energy) {
						datum.Units = pointer.FromString("joules")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units joules; value in range (upper)",
					func(datum *dataTypesFood.Energy) {
						datum.Units = pointer.FromString("joules")
						datum.Value = pointer.FromFloat64(41858000.0)
					},
				),
				Entry("units joules; value out of range (upper)",
					func(datum *dataTypesFood.Energy) {
						datum.Units = pointer.FromString("joules")
						datum.Value = pointer.FromFloat64(41858000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(41858000.1, 0.0, 41858000.0), "/value"),
				),
				Entry("units kilocalories; value missing",
					func(datum *dataTypesFood.Energy) {
						datum.Units = pointer.FromString("kilocalories")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units kilocalories; value out of range (lower)",
					func(datum *dataTypesFood.Energy) {
						datum.Units = pointer.FromString("kilocalories")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 10000.0), "/value"),
				),
				Entry("units kilocalories; value in range (lower)",
					func(datum *dataTypesFood.Energy) {
						datum.Units = pointer.FromString("kilocalories")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units kilocalories; value in range (upper)",
					func(datum *dataTypesFood.Energy) {
						datum.Units = pointer.FromString("kilocalories")
						datum.Value = pointer.FromFloat64(10000.0)
					},
				),
				Entry("units kilocalories; value out of range (upper)",
					func(datum *dataTypesFood.Energy) {
						datum.Units = pointer.FromString("kilocalories")
						datum.Value = pointer.FromFloat64(10000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(10000.1, 0.0, 10000.0), "/value"),
				),
				Entry("units kilojoules; value missing",
					func(datum *dataTypesFood.Energy) {
						datum.Units = pointer.FromString("kilojoules")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units kilojoules; value out of range (lower)",
					func(datum *dataTypesFood.Energy) {
						datum.Units = pointer.FromString("kilojoules")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 41858.0), "/value"),
				),
				Entry("units kilojoules; value in range (lower)",
					func(datum *dataTypesFood.Energy) {
						datum.Units = pointer.FromString("kilojoules")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units kilojoules; value in range (upper)",
					func(datum *dataTypesFood.Energy) {
						datum.Units = pointer.FromString("kilojoules")
						datum.Value = pointer.FromFloat64(41858.0)
					},
				),
				Entry("units kilojoules; value out of range (upper)",
					func(datum *dataTypesFood.Energy) {
						datum.Units = pointer.FromString("kilojoules")
						datum.Value = pointer.FromFloat64(41858.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(41858.1, 0.0, 41858.0), "/value"),
				),
				Entry("multiple errors",
					func(datum *dataTypesFood.Energy) {
						datum.Units = nil
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
			)
		})
	})

	Context("EnergyValueRangeForUnits", func() {
		It("returns expected range for units missing", func() {
			minimum, maximum := dataTypesFood.EnergyValueRangeForUnits(nil)
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units invalid", func() {
			minimum, maximum := dataTypesFood.EnergyValueRangeForUnits(pointer.FromString("invalid"))
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units calories", func() {
			minimum, maximum := dataTypesFood.EnergyValueRangeForUnits(pointer.FromString("calories"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(10000000.0))
		})

		It("returns expected range for units joules", func() {
			minimum, maximum := dataTypesFood.EnergyValueRangeForUnits(pointer.FromString("joules"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(41858000.0))
		})

		It("returns expected range for units kilocalories", func() {
			minimum, maximum := dataTypesFood.EnergyValueRangeForUnits(pointer.FromString("kilocalories"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(10000.0))
		})

		It("returns expected range for units kilojoules", func() {
			minimum, maximum := dataTypesFood.EnergyValueRangeForUnits(pointer.FromString("kilojoules"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(41858.0))
		})
	})
})
