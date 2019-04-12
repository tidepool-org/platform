package pump_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	dataBloodGlucoseTest "github.com/tidepool-org/platform/data/blood/glucose/test"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTypesBasalTest "github.com/tidepool-org/platform/data/types/basal/test"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewInsulinSensitivityStart(units *string, startMinimum int) *pump.InsulinSensitivityStart {
	datum := pump.NewInsulinSensitivityStart()
	datum.Amount = pointer.FromFloat64(test.RandomFloat64FromRange(dataBloodGlucose.ValueRangeForUnits(units)))
	datum.Start = pointer.FromInt(test.RandomIntFromRange(pump.InsulinSensitivityStartStartMinimum, pump.InsulinSensitivityStartStartMaximum))
	if startMinimum == pump.InsulinSensitivityStartStartMinimum {
		datum.Start = pointer.FromInt(pump.InsulinSensitivityStartStartMinimum)
	} else {
		datum.Start = pointer.FromInt(test.RandomIntFromRange(startMinimum, pump.InsulinSensitivityStartStartMaximum))
	}
	return datum
}

func CloneInsulinSensitivityStart(datum *pump.InsulinSensitivityStart) *pump.InsulinSensitivityStart {
	if datum == nil {
		return nil
	}
	clone := pump.NewInsulinSensitivityStart()
	clone.Amount = pointer.CloneFloat64(datum.Amount)
	clone.Start = pointer.CloneInt(datum.Start)
	return clone
}

func NewInsulinSensitivityStartArray(units *string) *pump.InsulinSensitivityStartArray {
	datum := pump.NewInsulinSensitivityStartArray()
	*datum = append(*datum, NewInsulinSensitivityStart(units, pump.InsulinSensitivityStartStartMinimum))
	*datum = append(*datum, NewInsulinSensitivityStart(units, *datum.Last().Start+1))
	*datum = append(*datum, NewInsulinSensitivityStart(units, *datum.Last().Start+1))
	return datum
}

func CloneInsulinSensitivityStartArray(datumArray *pump.InsulinSensitivityStartArray) *pump.InsulinSensitivityStartArray {
	if datumArray == nil {
		return nil
	}
	clone := pump.NewInsulinSensitivityStartArray()
	for _, datum := range *datumArray {
		*clone = append(*clone, CloneInsulinSensitivityStart(datum))
	}
	return clone
}

func NewInsulinSensitivityStartArrayMap(units *string) *pump.InsulinSensitivityStartArrayMap {
	datum := pump.NewInsulinSensitivityStartArrayMap()
	datum.Set(dataTypesBasalTest.NewScheduleName(), NewInsulinSensitivityStartArray(units))
	return datum
}

func CloneInsulinSensitivityStartArrayMap(datumArrayMap *pump.InsulinSensitivityStartArrayMap) *pump.InsulinSensitivityStartArrayMap {
	if datumArrayMap == nil {
		return nil
	}
	clone := pump.NewInsulinSensitivityStartArrayMap()
	for datumName, datumArray := range *datumArrayMap {
		clone.Set(datumName, CloneInsulinSensitivityStartArray(datumArray))
	}
	return clone
}

var _ = Describe("InsulinSensitivityStart", func() {
	It("InsulinSensitivityStartStartMaximum is expected", func() {
		Expect(pump.InsulinSensitivityStartStartMaximum).To(Equal(86400000))
	})

	It("InsulinSensitivityStartStartMinimum is expected", func() {
		Expect(pump.InsulinSensitivityStartStartMinimum).To(Equal(0))
	})

	Context("ParseInsulinSensitivityStart", func() {
		// TODO
	})

	Context("NewInsulinSensitivityStart", func() {
		It("is successful", func() {
			Expect(pump.NewInsulinSensitivityStart()).To(Equal(&pump.InsulinSensitivityStart{}))
		})
	})

	Context("InsulinSensitivityStart", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(units *string, mutator func(datum *pump.InsulinSensitivityStart, units *string), expectedErrors ...error) {
					datum := NewInsulinSensitivityStart(units, pump.InsulinSensitivityStartStartMinimum)
					mutator(datum, units)
					dataTypesTest.ValidateWithExpectedOrigins(NewValidatableWithUnitsAndStartMinimumAdapter(datum, units, pointer.FromInt(pump.InsulinSensitivityStartStartMinimum)), structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStart, units *string) {},
				),
				Entry("units missing; amount missing",
					nil,
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Amount = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
				),
				Entry("units missing; amount out of range (lower)",
					nil,
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Amount = pointer.FromFloat64(-0.1) },
				),
				Entry("units missing; amount in range (lower)",
					nil,
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Amount = pointer.FromFloat64(0.0) },
				),
				Entry("units missing; amount in range (upper)",
					nil,
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Amount = pointer.FromFloat64(1000.0) },
				),
				Entry("units missing; amount out of range (upper)",
					nil,
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Amount = pointer.FromFloat64(1000.1) },
				),
				Entry("units invalid; amount missing",
					pointer.FromString("invalid"),
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Amount = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
				),
				Entry("units invalid; amount out of range (lower)",
					pointer.FromString("invalid"),
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Amount = pointer.FromFloat64(-0.1) },
				),
				Entry("units invalid; amount in range (lower)",
					pointer.FromString("invalid"),
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Amount = pointer.FromFloat64(0.0) },
				),
				Entry("units invalid; amount in range (upper)",
					pointer.FromString("invalid"),
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Amount = pointer.FromFloat64(1000.0) },
				),
				Entry("units invalid; amount out of range (upper)",
					pointer.FromString("invalid"),
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Amount = pointer.FromFloat64(1000.1) },
				),
				Entry("units mmol/L; amount missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Amount = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
				),
				Entry("units mmol/L; amount out of range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Amount = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 55.0), "/amount"),
				),
				Entry("units mmol/L; amount in range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Amount = pointer.FromFloat64(0.0) },
				),
				Entry("units mmol/L; amount in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Amount = pointer.FromFloat64(55.0) },
				),
				Entry("units mmol/L; amount out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Amount = pointer.FromFloat64(55.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(55.1, 0.0, 55.0), "/amount"),
				),
				Entry("units mmol/l; amount missing",
					pointer.FromString("mmol/l"),
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Amount = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
				),
				Entry("units mmol/l; amount out of range (lower)",
					pointer.FromString("mmol/l"),
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Amount = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 55.0), "/amount"),
				),
				Entry("units mmol/l; amount in range (lower)",
					pointer.FromString("mmol/l"),
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Amount = pointer.FromFloat64(0.0) },
				),
				Entry("units mmol/l; amount in range (upper)",
					pointer.FromString("mmol/l"),
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Amount = pointer.FromFloat64(55.0) },
				),
				Entry("units mmol/l; amount out of range (upper)",
					pointer.FromString("mmol/l"),
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Amount = pointer.FromFloat64(55.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(55.1, 0.0, 55.0), "/amount"),
				),
				Entry("units mg/dL; amount missing",
					pointer.FromString("mg/dL"),
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Amount = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
				),
				Entry("units mg/dL; amount out of range (lower)",
					pointer.FromString("mg/dL"),
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Amount = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/amount"),
				),
				Entry("units mg/dL; amount in range (lower)",
					pointer.FromString("mg/dL"),
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Amount = pointer.FromFloat64(0.0) },
				),
				Entry("units mg/dL; amount in range (upper)",
					pointer.FromString("mg/dL"),
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Amount = pointer.FromFloat64(1000.0) },
				),
				Entry("units mg/dL; amount out of range (upper)",
					pointer.FromString("mg/dL"),
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Amount = pointer.FromFloat64(1000.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/amount"),
				),
				Entry("units mg/dl; amount missing",
					pointer.FromString("mg/dl"),
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Amount = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
				),
				Entry("units mg/dl; amount out of range (lower)",
					pointer.FromString("mg/dl"),
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Amount = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/amount"),
				),
				Entry("units mg/dl; amount in range (lower)",
					pointer.FromString("mg/dl"),
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Amount = pointer.FromFloat64(0.0) },
				),
				Entry("units mg/dl; amount in range (upper)",
					pointer.FromString("mg/dl"),
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Amount = pointer.FromFloat64(1000.0) },
				),
				Entry("units mg/dl; amount out of range (upper)",
					pointer.FromString("mg/dl"),
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Amount = pointer.FromFloat64(1000.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/amount"),
				),
				Entry("start missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Start = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/start"),
				),
				Entry("multiple errors",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStart, units *string) {
						datum.Amount = nil
						datum.Start = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/start"),
				),
			)

			DescribeTable("validates the datum with minimum start",
				func(units *string, mutator func(datum *pump.InsulinSensitivityStart, units *string), expectedErrors ...error) {
					datum := NewInsulinSensitivityStart(units, pump.InsulinSensitivityStartStartMinimum)
					mutator(datum, units)
					dataTypesTest.ValidateWithExpectedOrigins(NewValidatableWithUnitsAndStartMinimumAdapter(datum, units, pointer.FromInt(pump.InsulinSensitivityStartStartMinimum)), structure.Origins(), expectedErrors...)
				},
				Entry("start out of range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Start = pointer.FromInt(-1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo(-1, 0), "/start"),
				),
				Entry("start in range",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Start = pointer.FromInt(0) },
				),
				Entry("start out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Start = pointer.FromInt(1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo(1, 0), "/start"),
				),
			)

			DescribeTable("validates the datum with non-minimum start",
				func(units *string, mutator func(datum *pump.InsulinSensitivityStart, units *string), expectedErrors ...error) {
					datum := NewInsulinSensitivityStart(units, pump.InsulinSensitivityStartStartMinimum+1)
					mutator(datum, units)
					dataTypesTest.ValidateWithExpectedOrigins(NewValidatableWithUnitsAndStartMinimumAdapter(datum, units, pointer.FromInt(pump.InsulinSensitivityStartStartMinimum+1)), structure.Origins(), expectedErrors...)
				},
				Entry("start out of range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Start = pointer.FromInt(0) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(0, 1, 86400000), "/start"),
				),
				Entry("start in range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Start = pointer.FromInt(1) },
				),
				Entry("start in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Start = pointer.FromInt(86400000) },
				),
				Entry("start out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Start = pointer.FromInt(86400001) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(86400001, 1, 86400000), "/start"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(units *string, mutator func(datum *pump.InsulinSensitivityStart, units *string), expectator func(datum *pump.InsulinSensitivityStart, expectedDatum *pump.InsulinSensitivityStart, units *string)) {
					for _, origin := range structure.Origins() {
						datum := NewInsulinSensitivityStart(units, pump.InsulinSensitivityStartStartMinimum+1)
						mutator(datum, units)
						expectedDatum := CloneInsulinSensitivityStart(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin), units)
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						if expectator != nil {
							expectator(datum, expectedDatum, units)
						}
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStart, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units missing",
					nil,
					func(datum *pump.InsulinSensitivityStart, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units invalid",
					pointer.FromString("invalid"),
					func(datum *pump.InsulinSensitivityStart, units *string) {},
					nil,
				),
				Entry("does not modify the datum; start missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStart, units *string) { datum.Start = nil },
					nil,
				),
			)

			DescribeTable("normalizes the datum with origin external",
				func(units *string, mutator func(datum *pump.InsulinSensitivityStart, units *string), expectator func(datum *pump.InsulinSensitivityStart, expectedDatum *pump.InsulinSensitivityStart, units *string)) {
					datum := NewInsulinSensitivityStart(units, pump.InsulinSensitivityStartStartMinimum+1)
					mutator(datum, units)
					expectedDatum := CloneInsulinSensitivityStart(datum)
					normalizer := dataNormalizer.New()
					Expect(normalizer).ToNot(BeNil())
					datum.Normalize(normalizer.WithOrigin(structure.OriginExternal), units)
					Expect(normalizer.Error()).To(BeNil())
					Expect(normalizer.Data()).To(BeEmpty())
					if expectator != nil {
						expectator(datum, expectedDatum, units)
					}
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("does not modify the datum; units mmol/L",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStart, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *pump.InsulinSensitivityStart, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *pump.InsulinSensitivityStart, units *string) {},
					func(datum *pump.InsulinSensitivityStart, expectedDatum *pump.InsulinSensitivityStart, units *string) {
						dataBloodGlucoseTest.ExpectNormalizedValue(datum.Amount, expectedDatum.Amount, units)
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *pump.InsulinSensitivityStart, units *string) {},
					func(datum *pump.InsulinSensitivityStart, expectedDatum *pump.InsulinSensitivityStart, units *string) {
						dataBloodGlucoseTest.ExpectNormalizedValue(datum.Amount, expectedDatum.Amount, units)
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(units *string, mutator func(datum *pump.InsulinSensitivityStart, units *string), expectator func(datum *pump.InsulinSensitivityStart, expectedDatum *pump.InsulinSensitivityStart, units *string)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := NewInsulinSensitivityStart(units, pump.InsulinSensitivityStartStartMinimum+1)
						mutator(datum, units)
						expectedDatum := CloneInsulinSensitivityStart(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin), units)
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						if expectator != nil {
							expectator(datum, expectedDatum, units)
						}
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum; units mmol/L",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStart, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *pump.InsulinSensitivityStart, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *pump.InsulinSensitivityStart, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *pump.InsulinSensitivityStart, units *string) {},
					nil,
				),
			)
		})
	})

	Context("ParseInsulinSensitivityStartArray", func() {
		// TODO
	})

	Context("NewInsulinSensitivityStartArray", func() {
		It("is successful", func() {
			Expect(pump.NewInsulinSensitivityStartArray()).To(Equal(&pump.InsulinSensitivityStartArray{}))
		})
	})

	Context("InsulinSensitivityStartArray", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(units *string, mutator func(datum *pump.InsulinSensitivityStartArray, units *string), expectedErrors ...error) {
					datum := pump.NewInsulinSensitivityStartArray()
					mutator(datum, units)
					dataTypesTest.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithStringAdapter(datum, units), structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStartArray, units *string) {},
				),
				Entry("empty",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStartArray, units *string) {
						*datum = *pump.NewInsulinSensitivityStartArray()
					},
				),
				Entry("nil",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStartArray, units *string) { *datum = append(*datum, nil) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
				),
				Entry("single invalid",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStartArray, units *string) {
						invalid := NewInsulinSensitivityStart(pointer.FromString("mmol/L"), pump.InsulinSensitivityStartStartMinimum)
						invalid.Amount = nil
						*datum = append(*datum, invalid)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0/amount"),
				),
				Entry("single valid",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStartArray, units *string) {
						*datum = append(*datum, NewInsulinSensitivityStart(pointer.FromString("mmol/L"), pump.InsulinSensitivityStartStartMinimum))
					},
				),
				Entry("multiple invalid",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStartArray, units *string) {
						*datum = append(*datum, NewInsulinSensitivityStart(pointer.FromString("mmol/L"), pump.InsulinSensitivityStartStartMinimum))
						invalid := NewInsulinSensitivityStart(pointer.FromString("mmol/L"), *datum.Last().Start+1)
						invalid.Amount = nil
						*datum = append(*datum, invalid)
						*datum = append(*datum, NewInsulinSensitivityStart(pointer.FromString("mmol/L"), *datum.Last().Start+1))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/1/amount"),
				),
				Entry("multiple valid",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStartArray, units *string) {
						*datum = append(*datum, NewInsulinSensitivityStart(pointer.FromString("mmol/L"), pump.InsulinSensitivityStartStartMinimum))
						*datum = append(*datum, NewInsulinSensitivityStart(pointer.FromString("mmol/L"), *datum.Last().Start+1))
						*datum = append(*datum, NewInsulinSensitivityStart(pointer.FromString("mmol/L"), *datum.Last().Start+1))
					},
				),
				Entry("multiple errors",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStartArray, units *string) {
						invalid := NewInsulinSensitivityStart(pointer.FromString("mmol/L"), pump.InsulinSensitivityStartStartMinimum)
						invalid.Amount = nil
						*datum = append(*datum, nil, invalid)
						*datum = append(*datum, nil, NewInsulinSensitivityStart(pointer.FromString("mmol/L"), *datum.Last().Start+1))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/1/amount"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/2"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(units *string, mutator func(datum *pump.InsulinSensitivityStartArray, units *string), expectator func(datum *pump.InsulinSensitivityStartArray, expectedDatum *pump.InsulinSensitivityStartArray, units *string)) {
					for _, origin := range structure.Origins() {
						datum := NewInsulinSensitivityStartArray(units)
						mutator(datum, units)
						expectedDatum := CloneInsulinSensitivityStartArray(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin), units)
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						if expectator != nil {
							expectator(datum, expectedDatum, units)
						}
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStartArray, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units missing",
					nil,
					func(datum *pump.InsulinSensitivityStartArray, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units invalid",
					pointer.FromString("invalid"),
					func(datum *pump.InsulinSensitivityStartArray, units *string) {},
					nil,
				),
				Entry("does not modify the datum; start missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStartArray, units *string) { (*datum)[0].Start = nil },
					nil,
				),
			)

			DescribeTable("normalizes the datum with origin external",
				func(units *string, mutator func(datum *pump.InsulinSensitivityStartArray, units *string), expectator func(datum *pump.InsulinSensitivityStartArray, expectedDatum *pump.InsulinSensitivityStartArray, units *string)) {
					datum := NewInsulinSensitivityStartArray(units)
					mutator(datum, units)
					expectedDatum := CloneInsulinSensitivityStartArray(datum)
					normalizer := dataNormalizer.New()
					Expect(normalizer).ToNot(BeNil())
					datum.Normalize(normalizer.WithOrigin(structure.OriginExternal), units)
					Expect(normalizer.Error()).To(BeNil())
					Expect(normalizer.Data()).To(BeEmpty())
					if expectator != nil {
						expectator(datum, expectedDatum, units)
					}
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("does not modify the datum; units mmol/L",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStartArray, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *pump.InsulinSensitivityStartArray, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *pump.InsulinSensitivityStartArray, units *string) {},
					func(datum *pump.InsulinSensitivityStartArray, expectedDatum *pump.InsulinSensitivityStartArray, units *string) {
						for index := range *datum {
							dataBloodGlucoseTest.ExpectNormalizedValue((*datum)[index].Amount, (*expectedDatum)[index].Amount, units)
						}
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *pump.InsulinSensitivityStartArray, units *string) {},
					func(datum *pump.InsulinSensitivityStartArray, expectedDatum *pump.InsulinSensitivityStartArray, units *string) {
						for index := range *datum {
							dataBloodGlucoseTest.ExpectNormalizedValue((*datum)[index].Amount, (*expectedDatum)[index].Amount, units)
						}
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(units *string, mutator func(datum *pump.InsulinSensitivityStartArray, units *string), expectator func(datum *pump.InsulinSensitivityStartArray, expectedDatum *pump.InsulinSensitivityStartArray, units *string)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := NewInsulinSensitivityStartArray(units)
						mutator(datum, units)
						expectedDatum := CloneInsulinSensitivityStartArray(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin), units)
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						if expectator != nil {
							expectator(datum, expectedDatum, units)
						}
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum; units mmol/L",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStartArray, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *pump.InsulinSensitivityStartArray, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *pump.InsulinSensitivityStartArray, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *pump.InsulinSensitivityStartArray, units *string) {},
					nil,
				),
			)
		})
	})

	Context("ParseInsulinSensitivityStartArrayMap", func() {
		// TODO
	})

	Context("NewInsulinSensitivityStartArrayMap", func() {
		It("is successful", func() {
			Expect(pump.NewInsulinSensitivityStartArrayMap()).To(Equal(&pump.InsulinSensitivityStartArrayMap{}))
		})
	})

	Context("InsulinSensitivityStartArrayMap", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(units *string, mutator func(datum *pump.InsulinSensitivityStartArrayMap, units *string), expectedErrors ...error) {
					datum := pump.NewInsulinSensitivityStartArrayMap()
					mutator(datum, units)
					dataTypesTest.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithStringAdapter(datum, units), structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStartArrayMap, units *string) {},
				),
				Entry("empty",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStartArrayMap, units *string) {
						*datum = *pump.NewInsulinSensitivityStartArrayMap()
					},
				),
				Entry("empty name",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStartArrayMap, units *string) {
						datum.Set("", NewInsulinSensitivityStartArray(units))
					},
				),
				Entry("nil value",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStartArrayMap, units *string) { datum.Set("", nil) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/"),
				),
				Entry("single invalid",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStartArrayMap, units *string) {
						invalid := NewInsulinSensitivityStartArray(units)
						(*invalid)[0].Start = nil
						datum.Set("one", invalid)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/one/0/start"),
				),
				Entry("single valid",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStartArrayMap, units *string) {
						datum.Set("one", NewInsulinSensitivityStartArray(units))
					},
				),
				Entry("multiple invalid",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStartArrayMap, units *string) {
						invalid := NewInsulinSensitivityStartArray(units)
						(*invalid)[0].Start = nil
						datum.Set("one", NewInsulinSensitivityStartArray(units))
						datum.Set("two", invalid)
						datum.Set("three", NewInsulinSensitivityStartArray(units))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/two/0/start"),
				),
				Entry("multiple valid",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStartArrayMap, units *string) {
						datum.Set("one", NewInsulinSensitivityStartArray(units))
						datum.Set("two", NewInsulinSensitivityStartArray(units))
						datum.Set("three", NewInsulinSensitivityStartArray(units))
					},
				),
				Entry("multiple errors",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStartArrayMap, units *string) {
						invalid := NewInsulinSensitivityStartArray(units)
						(*invalid)[0].Start = nil
						datum.Set("one", nil)
						datum.Set("two", invalid)
						datum.Set("three", NewInsulinSensitivityStartArray(units))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/one"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/two/0/start"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(units *string, mutator func(datum *pump.InsulinSensitivityStartArrayMap, units *string), expectator func(datum *pump.InsulinSensitivityStartArrayMap, expectedDatum *pump.InsulinSensitivityStartArrayMap, units *string)) {
					for _, origin := range structure.Origins() {
						datum := NewInsulinSensitivityStartArrayMap(units)
						mutator(datum, units)
						expectedDatum := CloneInsulinSensitivityStartArrayMap(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin), units)
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						if expectator != nil {
							expectator(datum, expectedDatum, units)
						}
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStartArrayMap, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units missing",
					nil,
					func(datum *pump.InsulinSensitivityStartArrayMap, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units invalid",
					pointer.FromString("invalid"),
					func(datum *pump.InsulinSensitivityStartArrayMap, units *string) {},
					nil,
				),
				Entry("does not modify the datum; start missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStartArrayMap, units *string) {
						for name := range *datum {
							(*(*datum)[name])[0].Start = nil
						}
					},
					nil,
				),
			)

			DescribeTable("normalizes the datum with structure external",
				func(units *string, mutator func(datum *pump.InsulinSensitivityStartArrayMap, units *string), expectator func(datum *pump.InsulinSensitivityStartArrayMap, expectedDatum *pump.InsulinSensitivityStartArrayMap, units *string)) {
					datum := NewInsulinSensitivityStartArrayMap(units)
					mutator(datum, units)
					expectedDatum := CloneInsulinSensitivityStartArrayMap(datum)
					normalizer := dataNormalizer.New()
					Expect(normalizer).ToNot(BeNil())
					datum.Normalize(normalizer.WithOrigin(structure.OriginExternal), units)
					Expect(normalizer.Error()).To(BeNil())
					Expect(normalizer.Data()).To(BeEmpty())
					if expectator != nil {
						expectator(datum, expectedDatum, units)
					}
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("does not modify the datum; units mmol/L",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStartArrayMap, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *pump.InsulinSensitivityStartArrayMap, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *pump.InsulinSensitivityStartArrayMap, units *string) {},
					func(datum *pump.InsulinSensitivityStartArrayMap, expectedDatum *pump.InsulinSensitivityStartArrayMap, units *string) {
						for name := range *datum {
							for index := range *(*datum)[name] {
								dataBloodGlucoseTest.ExpectNormalizedValue((*(*datum)[name])[index].Amount, (*(*expectedDatum)[name])[index].Amount, units)
							}
						}
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *pump.InsulinSensitivityStartArrayMap, units *string) {},
					func(datum *pump.InsulinSensitivityStartArrayMap, expectedDatum *pump.InsulinSensitivityStartArrayMap, units *string) {
						for name := range *datum {
							for index := range *(*datum)[name] {
								dataBloodGlucoseTest.ExpectNormalizedValue((*(*datum)[name])[index].Amount, (*(*expectedDatum)[name])[index].Amount, units)
							}
						}
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(units *string, mutator func(datum *pump.InsulinSensitivityStartArrayMap, units *string), expectator func(datum *pump.InsulinSensitivityStartArrayMap, expectedDatum *pump.InsulinSensitivityStartArrayMap, units *string)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := NewInsulinSensitivityStartArrayMap(units)
						mutator(datum, units)
						expectedDatum := CloneInsulinSensitivityStartArrayMap(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin), units)
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						if expectator != nil {
							expectator(datum, expectedDatum, units)
						}
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum; units mmol/L",
					pointer.FromString("mmol/L"),
					func(datum *pump.InsulinSensitivityStartArrayMap, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *pump.InsulinSensitivityStartArrayMap, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *pump.InsulinSensitivityStartArrayMap, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *pump.InsulinSensitivityStartArrayMap, units *string) {},
					nil,
				),
			)
		})
	})
})
