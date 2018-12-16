package pump_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataBloodGlucoseTest "github.com/tidepool-org/platform/data/blood/glucose/test"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewUnits(unitsBloodGlucose *string) *pump.Units {
	datum := pump.NewUnits()
	datum.BloodGlucose = unitsBloodGlucose
	datum.Carbohydrate = pointer.FromString(test.RandomStringFromArray(pump.Carbohydrates()))
	return datum
}

func CloneUnits(datum *pump.Units) *pump.Units {
	if datum == nil {
		return nil
	}
	clone := pump.NewUnits()
	clone.BloodGlucose = test.CloneString(datum.BloodGlucose)
	clone.Carbohydrate = test.CloneString(datum.Carbohydrate)
	return clone
}

var _ = Describe("Units", func() {
	It("CarbohydrateExchanges is expected", func() {
		Expect(pump.CarbohydrateExchanges).To(Equal("exchanges"))
	})

	It("CarbohydrateGrams is expected", func() {
		Expect(pump.CarbohydrateGrams).To(Equal("grams"))
	})

	It("Carbohydrates returns expected", func() {
		Expect(pump.Carbohydrates()).To(Equal([]string{"exchanges", "grams"}))
	})

	Context("ParseUnits", func() {
		// TODO
	})

	Context("NewUnits", func() {
		It("is successful", func() {
			Expect(pump.NewUnits()).To(Equal(&pump.Units{}))
		})
	})

	Context("Units", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *pump.Units), expectedErrors ...error) {
					datum := NewUnits(pointer.FromString("mmol/L"))
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pump.Units) {},
				),
				Entry("blood glucose missing",
					func(datum *pump.Units) { datum.BloodGlucose = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/bg"),
				),
				Entry("blood glucose invalid",
					func(datum *pump.Units) { datum.BloodGlucose = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/bg"),
				),
				Entry("blood glucose mmol/L",
					func(datum *pump.Units) { datum.BloodGlucose = pointer.FromString("mmol/L") },
				),
				Entry("blood glucose mmol/l",
					func(datum *pump.Units) { datum.BloodGlucose = pointer.FromString("mmol/l") },
				),
				Entry("blood glucose mg/dL",
					func(datum *pump.Units) { datum.BloodGlucose = pointer.FromString("mg/dL") },
				),
				Entry("blood glucose mg/dl",
					func(datum *pump.Units) { datum.BloodGlucose = pointer.FromString("mg/dl") },
				),
				Entry("carbohydrate missing",
					func(datum *pump.Units) { datum.Carbohydrate = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/carb"),
				),
				Entry("carbohydrate invalid",
					func(datum *pump.Units) { datum.Carbohydrate = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"exchanges", "grams"}), "/carb"),
				),
				Entry("carbohydrate exchanges",
					func(datum *pump.Units) { datum.Carbohydrate = pointer.FromString("exchanges") },
				),
				Entry("carbohydrate grams",
					func(datum *pump.Units) { datum.Carbohydrate = pointer.FromString("grams") },
				),
				Entry("multiple errors",
					func(datum *pump.Units) {
						datum.BloodGlucose = nil
						datum.Carbohydrate = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/bg"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/carb"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *pump.Units), expectator func(datum *pump.Units, expectedDatum *pump.Units)) {
					for _, origin := range structure.Origins() {
						datum := NewUnits(pointer.FromString("mmol/L"))
						mutator(datum)
						expectedDatum := CloneUnits(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						if expectator != nil {
							expectator(datum, expectedDatum)
						}
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *pump.Units) {},
					nil,
				),
				Entry("does not modify the datum; blood glucose missing",
					func(datum *pump.Units) { datum.BloodGlucose = nil },
					nil,
				),
				Entry("does not modify the datum; blood glucose invalid",
					func(datum *pump.Units) { datum.BloodGlucose = pointer.FromString("invalid") },
					nil,
				),
				Entry("does not modify the datum; carbohydrate missing",
					func(datum *pump.Units) { datum.Carbohydrate = nil },
					nil,
				),
				Entry("does not modify the datum; carbohydrate invalid",
					func(datum *pump.Units) { datum.Carbohydrate = pointer.FromString("invalid") },
					nil,
				),
				Entry("does not modify the datum; carbohydrate exchanges",
					func(datum *pump.Units) { datum.Carbohydrate = pointer.FromString("exchanges") },
					nil,
				),
				Entry("does not modify the datum; carbohydrate grams",
					func(datum *pump.Units) { datum.Carbohydrate = pointer.FromString("grams") },
					nil,
				),
			)

			DescribeTable("normalizes the datum with origin external",
				func(mutator func(datum *pump.Units), expectator func(datum *pump.Units, expectedDatum *pump.Units)) {
					datum := NewUnits(pointer.FromString("mmol/L"))
					mutator(datum)
					expectedDatum := CloneUnits(datum)
					normalizer := dataNormalizer.New()
					Expect(normalizer).ToNot(BeNil())
					datum.Normalize(normalizer.WithOrigin(structure.OriginExternal))
					Expect(normalizer.Error()).To(BeNil())
					Expect(normalizer.Data()).To(BeEmpty())
					if expectator != nil {
						expectator(datum, expectedDatum)
					}
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("does not modify the datum; blood glucose mmol/L",
					func(datum *pump.Units) { datum.BloodGlucose = pointer.FromString("mmol/L") },
					nil,
				),
				Entry("modifies the datum; blood glucose mmol/l",
					func(datum *pump.Units) { datum.BloodGlucose = pointer.FromString("mmol/l") },
					func(datum *pump.Units, expectedDatum *pump.Units) {
						dataBloodGlucoseTest.ExpectNormalizedUnits(datum.BloodGlucose, expectedDatum.BloodGlucose)
					},
				),
				Entry("modifies the datum; blood glucose mg/dL",
					func(datum *pump.Units) { datum.BloodGlucose = pointer.FromString("mg/dL") },
					func(datum *pump.Units, expectedDatum *pump.Units) {
						dataBloodGlucoseTest.ExpectNormalizedUnits(datum.BloodGlucose, expectedDatum.BloodGlucose)
					},
				),
				Entry("modifies the datum; blood glucose mg/dl",
					func(datum *pump.Units) { datum.BloodGlucose = pointer.FromString("mg/dl") },
					func(datum *pump.Units, expectedDatum *pump.Units) {
						dataBloodGlucoseTest.ExpectNormalizedUnits(datum.BloodGlucose, expectedDatum.BloodGlucose)
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(mutator func(datum *pump.Units), expectator func(datum *pump.Units, expectedDatum *pump.Units)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := NewUnits(pointer.FromString("mmol/L"))
						mutator(datum)
						expectedDatum := CloneUnits(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						if expectator != nil {
							expectator(datum, expectedDatum)
						}
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum; blood glucose mmol/L",
					func(datum *pump.Units) { datum.BloodGlucose = pointer.FromString("mmol/L") },
					nil,
				),
				Entry("does not modify the datum; blood glucose mmol/l",
					func(datum *pump.Units) { datum.BloodGlucose = pointer.FromString("mmol/l") },
					nil,
				),
				Entry("does not modify the datum; blood glucose mg/dL",
					func(datum *pump.Units) { datum.BloodGlucose = pointer.FromString("mg/dL") },
					nil,
				),
				Entry("does not modify the datum; blood glucose mg/dl",
					func(datum *pump.Units) { datum.BloodGlucose = pointer.FromString("mg/dl") },
					nil,
				),
			)
		})
	})
})
