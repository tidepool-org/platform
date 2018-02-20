package pump_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	testDataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose/test"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewInsulinSensitivity(units *string) *pump.InsulinSensitivity {
	datum := pump.NewInsulinSensitivity()
	datum.Amount = pointer.Float64(test.RandomFloat64FromRange(dataBloodGlucose.ValueRangeForUnits(units)))
	datum.Start = pointer.Int(test.RandomIntFromRange(pump.InsulinSensitivityStartMinimum, pump.InsulinSensitivityStartMaximum))
	return datum
}

func CloneInsulinSensitivity(datum *pump.InsulinSensitivity) *pump.InsulinSensitivity {
	if datum == nil {
		return nil
	}
	clone := pump.NewInsulinSensitivity()
	clone.Amount = test.CloneFloat64(datum.Amount)
	clone.Start = test.CloneInt(datum.Start)
	return clone
}

func NewInsulinSensitivityArray(units *string) *pump.InsulinSensitivityArray {
	datum := pump.NewInsulinSensitivityArray()
	*datum = append(*datum, NewInsulinSensitivity(units))
	return datum
}

func CloneInsulinSensitivityArray(datumArray *pump.InsulinSensitivityArray) *pump.InsulinSensitivityArray {
	if datumArray == nil {
		return nil
	}
	clone := pump.NewInsulinSensitivityArray()
	for _, datum := range *datumArray {
		*clone = append(*clone, CloneInsulinSensitivity(datum))
	}
	return clone
}

var _ = Describe("InsulinSensitivity", func() {
	It("InsulinSensitivityStartMaximum is expected", func() {
		Expect(pump.InsulinSensitivityStartMaximum).To(Equal(86400000))
	})

	It("InsulinSensitivityStartMinimum is expected", func() {
		Expect(pump.InsulinSensitivityStartMinimum).To(Equal(0))
	})

	Context("ParseInsulinSensitivity", func() {
		// TODO
	})

	Context("NewInsulinSensitivity", func() {
		It("is successful", func() {
			Expect(pump.NewInsulinSensitivity()).To(Equal(&pump.InsulinSensitivity{}))
		})
	})

	Context("InsulinSensitivity", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(units *string, mutator func(datum *pump.InsulinSensitivity, units *string), expectedErrors ...error) {
					datum := NewInsulinSensitivity(units)
					mutator(datum, units)
					testDataTypes.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithStringAdapter(datum, units), structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.String("mmol/L"),
					func(datum *pump.InsulinSensitivity, units *string) {},
				),
				Entry("units missing; amount missing",
					nil,
					func(datum *pump.InsulinSensitivity, units *string) { datum.Amount = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
				),
				Entry("units missing; amount out of range (lower)",
					nil,
					func(datum *pump.InsulinSensitivity, units *string) { datum.Amount = pointer.Float64(-0.1) },
				),
				Entry("units missing; amount in range (lower)",
					nil,
					func(datum *pump.InsulinSensitivity, units *string) { datum.Amount = pointer.Float64(0.0) },
				),
				Entry("units missing; amount in range (upper)",
					nil,
					func(datum *pump.InsulinSensitivity, units *string) { datum.Amount = pointer.Float64(1000.0) },
				),
				Entry("units missing; amount out of range (upper)",
					nil,
					func(datum *pump.InsulinSensitivity, units *string) { datum.Amount = pointer.Float64(1000.1) },
				),
				Entry("units invalid; amount missing",
					pointer.String("invalid"),
					func(datum *pump.InsulinSensitivity, units *string) { datum.Amount = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
				),
				Entry("units invalid; amount out of range (lower)",
					pointer.String("invalid"),
					func(datum *pump.InsulinSensitivity, units *string) { datum.Amount = pointer.Float64(-0.1) },
				),
				Entry("units invalid; amount in range (lower)",
					pointer.String("invalid"),
					func(datum *pump.InsulinSensitivity, units *string) { datum.Amount = pointer.Float64(0.0) },
				),
				Entry("units invalid; amount in range (upper)",
					pointer.String("invalid"),
					func(datum *pump.InsulinSensitivity, units *string) { datum.Amount = pointer.Float64(1000.0) },
				),
				Entry("units invalid; amount out of range (upper)",
					pointer.String("invalid"),
					func(datum *pump.InsulinSensitivity, units *string) { datum.Amount = pointer.Float64(1000.1) },
				),
				Entry("units mmol/L; amount missing",
					pointer.String("mmol/L"),
					func(datum *pump.InsulinSensitivity, units *string) { datum.Amount = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
				),
				Entry("units mmol/L; amount out of range (lower)",
					pointer.String("mmol/L"),
					func(datum *pump.InsulinSensitivity, units *string) { datum.Amount = pointer.Float64(-0.1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 55.0), "/amount"),
				),
				Entry("units mmol/L; amount in range (lower)",
					pointer.String("mmol/L"),
					func(datum *pump.InsulinSensitivity, units *string) { datum.Amount = pointer.Float64(0.0) },
				),
				Entry("units mmol/L; amount in range (upper)",
					pointer.String("mmol/L"),
					func(datum *pump.InsulinSensitivity, units *string) { datum.Amount = pointer.Float64(55.0) },
				),
				Entry("units mmol/L; amount out of range (upper)",
					pointer.String("mmol/L"),
					func(datum *pump.InsulinSensitivity, units *string) { datum.Amount = pointer.Float64(55.1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(55.1, 0.0, 55.0), "/amount"),
				),
				Entry("units mmol/l; amount missing",
					pointer.String("mmol/l"),
					func(datum *pump.InsulinSensitivity, units *string) { datum.Amount = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
				),
				Entry("units mmol/l; amount out of range (lower)",
					pointer.String("mmol/l"),
					func(datum *pump.InsulinSensitivity, units *string) { datum.Amount = pointer.Float64(-0.1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 55.0), "/amount"),
				),
				Entry("units mmol/l; amount in range (lower)",
					pointer.String("mmol/l"),
					func(datum *pump.InsulinSensitivity, units *string) { datum.Amount = pointer.Float64(0.0) },
				),
				Entry("units mmol/l; amount in range (upper)",
					pointer.String("mmol/l"),
					func(datum *pump.InsulinSensitivity, units *string) { datum.Amount = pointer.Float64(55.0) },
				),
				Entry("units mmol/l; amount out of range (upper)",
					pointer.String("mmol/l"),
					func(datum *pump.InsulinSensitivity, units *string) { datum.Amount = pointer.Float64(55.1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(55.1, 0.0, 55.0), "/amount"),
				),
				Entry("units mg/dL; amount missing",
					pointer.String("mg/dL"),
					func(datum *pump.InsulinSensitivity, units *string) { datum.Amount = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
				),
				Entry("units mg/dL; amount out of range (lower)",
					pointer.String("mg/dL"),
					func(datum *pump.InsulinSensitivity, units *string) { datum.Amount = pointer.Float64(-0.1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/amount"),
				),
				Entry("units mg/dL; amount in range (lower)",
					pointer.String("mg/dL"),
					func(datum *pump.InsulinSensitivity, units *string) { datum.Amount = pointer.Float64(0.0) },
				),
				Entry("units mg/dL; amount in range (upper)",
					pointer.String("mg/dL"),
					func(datum *pump.InsulinSensitivity, units *string) { datum.Amount = pointer.Float64(1000.0) },
				),
				Entry("units mg/dL; amount out of range (upper)",
					pointer.String("mg/dL"),
					func(datum *pump.InsulinSensitivity, units *string) { datum.Amount = pointer.Float64(1000.1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/amount"),
				),
				Entry("units mg/dl; amount missing",
					pointer.String("mg/dl"),
					func(datum *pump.InsulinSensitivity, units *string) { datum.Amount = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
				),
				Entry("units mg/dl; amount out of range (lower)",
					pointer.String("mg/dl"),
					func(datum *pump.InsulinSensitivity, units *string) { datum.Amount = pointer.Float64(-0.1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/amount"),
				),
				Entry("units mg/dl; amount in range (lower)",
					pointer.String("mg/dl"),
					func(datum *pump.InsulinSensitivity, units *string) { datum.Amount = pointer.Float64(0.0) },
				),
				Entry("units mg/dl; amount in range (upper)",
					pointer.String("mg/dl"),
					func(datum *pump.InsulinSensitivity, units *string) { datum.Amount = pointer.Float64(1000.0) },
				),
				Entry("units mg/dl; amount out of range (upper)",
					pointer.String("mg/dl"),
					func(datum *pump.InsulinSensitivity, units *string) { datum.Amount = pointer.Float64(1000.1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/amount"),
				),
				Entry("start missing",
					pointer.String("mmol/L"),
					func(datum *pump.InsulinSensitivity, units *string) { datum.Start = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/start"),
				),
				Entry("start out of range (lower)",
					pointer.String("mmol/L"),
					func(datum *pump.InsulinSensitivity, units *string) { datum.Start = pointer.Int(-1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/start"),
				),
				Entry("start in range (lower)",
					pointer.String("mmol/L"),
					func(datum *pump.InsulinSensitivity, units *string) { datum.Start = pointer.Int(0) },
				),
				Entry("start in range (upper)",
					pointer.String("mmol/L"),
					func(datum *pump.InsulinSensitivity, units *string) { datum.Start = pointer.Int(86400000) },
				),
				Entry("start out of range (upper)",
					pointer.String("mmol/L"),
					func(datum *pump.InsulinSensitivity, units *string) { datum.Start = pointer.Int(86400001) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/start"),
				),
				Entry("multiple errors",
					pointer.String("mmol/L"),
					func(datum *pump.InsulinSensitivity, units *string) {
						datum.Amount = nil
						datum.Start = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/start"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(units *string, mutator func(datum *pump.InsulinSensitivity, units *string), expectator func(datum *pump.InsulinSensitivity, expectedDatum *pump.InsulinSensitivity, units *string)) {
					for _, origin := range structure.Origins() {
						datum := NewInsulinSensitivity(units)
						mutator(datum, units)
						expectedDatum := CloneInsulinSensitivity(datum)
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
					pointer.String("mmol/L"),
					func(datum *pump.InsulinSensitivity, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units missing",
					nil,
					func(datum *pump.InsulinSensitivity, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units invalid",
					pointer.String("invalid"),
					func(datum *pump.InsulinSensitivity, units *string) {},
					nil,
				),
				Entry("does not modify the datum; start missing",
					pointer.String("mmol/L"),
					func(datum *pump.InsulinSensitivity, units *string) { datum.Start = nil },
					nil,
				),
			)

			DescribeTable("normalizes the datum with origin external",
				func(units *string, mutator func(datum *pump.InsulinSensitivity, units *string), expectator func(datum *pump.InsulinSensitivity, expectedDatum *pump.InsulinSensitivity, units *string)) {
					datum := NewInsulinSensitivity(units)
					mutator(datum, units)
					expectedDatum := CloneInsulinSensitivity(datum)
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
					pointer.String("mmol/L"),
					func(datum *pump.InsulinSensitivity, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.String("mmol/l"),
					func(datum *pump.InsulinSensitivity, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dL",
					pointer.String("mg/dL"),
					func(datum *pump.InsulinSensitivity, units *string) {},
					func(datum *pump.InsulinSensitivity, expectedDatum *pump.InsulinSensitivity, units *string) {
						Expect(datum.Amount).ToNot(BeNil())
						Expect(*datum.Amount).To(Equal(*dataBloodGlucose.NormalizeValueForUnits(expectedDatum.Amount, units)))
						expectedDatum.Amount = datum.Amount
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.String("mg/dl"),
					func(datum *pump.InsulinSensitivity, units *string) {},
					func(datum *pump.InsulinSensitivity, expectedDatum *pump.InsulinSensitivity, units *string) {
						Expect(datum.Amount).ToNot(BeNil())
						Expect(*datum.Amount).To(Equal(*dataBloodGlucose.NormalizeValueForUnits(expectedDatum.Amount, units)))
						expectedDatum.Amount = datum.Amount
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(units *string, mutator func(datum *pump.InsulinSensitivity, units *string), expectator func(datum *pump.InsulinSensitivity, expectedDatum *pump.InsulinSensitivity, units *string)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := NewInsulinSensitivity(units)
						mutator(datum, units)
						expectedDatum := CloneInsulinSensitivity(datum)
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
					pointer.String("mmol/L"),
					func(datum *pump.InsulinSensitivity, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.String("mmol/l"),
					func(datum *pump.InsulinSensitivity, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dL",
					pointer.String("mg/dL"),
					func(datum *pump.InsulinSensitivity, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dl",
					pointer.String("mg/dl"),
					func(datum *pump.InsulinSensitivity, units *string) {},
					nil,
				),
			)
		})
	})

	Context("ParseInsulinSensitivityArray", func() {
		// TODO
	})

	Context("NewInsulinSensitivityArray", func() {
		It("is successful", func() {
			Expect(pump.NewInsulinSensitivityArray()).To(Equal(&pump.InsulinSensitivityArray{}))
		})
	})

	Context("InsulinSensitivityArray", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(units *string, mutator func(datum *pump.InsulinSensitivityArray, units *string), expectedErrors ...error) {
					datum := pump.NewInsulinSensitivityArray()
					mutator(datum, units)
					testDataTypes.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithStringAdapter(datum, units), structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.String("mmol/L"),
					func(datum *pump.InsulinSensitivityArray, units *string) {},
				),
				Entry("empty",
					pointer.String("mmol/L"),
					func(datum *pump.InsulinSensitivityArray, units *string) { *datum = *pump.NewInsulinSensitivityArray() },
				),
				Entry("nil",
					pointer.String("mmol/L"),
					func(datum *pump.InsulinSensitivityArray, units *string) { *datum = append(*datum, nil) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
				),
				Entry("single invalid",
					pointer.String("mmol/L"),
					func(datum *pump.InsulinSensitivityArray, units *string) {
						invalid := NewInsulinSensitivity(pointer.String("mmol/L"))
						invalid.Amount = nil
						*datum = append(*datum, invalid)
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0/amount"),
				),
				Entry("single valid",
					pointer.String("mmol/L"),
					func(datum *pump.InsulinSensitivityArray, units *string) {
						*datum = append(*datum, NewInsulinSensitivity(pointer.String("mmol/L")))
					},
				),
				Entry("multiple invalid",
					pointer.String("mmol/L"),
					func(datum *pump.InsulinSensitivityArray, units *string) {
						invalid := NewInsulinSensitivity(pointer.String("mmol/L"))
						invalid.Amount = nil
						*datum = append(*datum,
							NewInsulinSensitivity(pointer.String("mmol/L")),
							invalid,
							NewInsulinSensitivity(pointer.String("mmol/L")),
						)
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/1/amount"),
				),
				Entry("multiple valid",
					pointer.String("mmol/L"),
					func(datum *pump.InsulinSensitivityArray, units *string) {
						*datum = append(*datum,
							NewInsulinSensitivity(pointer.String("mmol/L")),
							NewInsulinSensitivity(pointer.String("mmol/L")),
							NewInsulinSensitivity(pointer.String("mmol/L")),
						)
					},
				),
				Entry("multiple errors",
					pointer.String("mmol/L"),
					func(datum *pump.InsulinSensitivityArray, units *string) {
						invalid := NewInsulinSensitivity(pointer.String("mmol/L"))
						invalid.Amount = nil
						*datum = append(*datum,
							nil,
							invalid,
							NewInsulinSensitivity(pointer.String("mmol/L")),
						)
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/1/amount"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(units *string, mutator func(datum *pump.InsulinSensitivityArray, units *string), expectator func(datum *pump.InsulinSensitivityArray, expectedDatum *pump.InsulinSensitivityArray, units *string)) {
					for _, origin := range structure.Origins() {
						datum := NewInsulinSensitivityArray(units)
						mutator(datum, units)
						expectedDatum := CloneInsulinSensitivityArray(datum)
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
					pointer.String("mmol/L"),
					func(datum *pump.InsulinSensitivityArray, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units missing",
					nil,
					func(datum *pump.InsulinSensitivityArray, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units invalid",
					pointer.String("invalid"),
					func(datum *pump.InsulinSensitivityArray, units *string) {},
					nil,
				),
				Entry("does not modify the datum; start missing",
					pointer.String("mmol/L"),
					func(datum *pump.InsulinSensitivityArray, units *string) { (*datum)[0].Start = nil },
					nil,
				),
			)

			DescribeTable("normalizes the datum with origin external",
				func(units *string, mutator func(datum *pump.InsulinSensitivityArray, units *string), expectator func(datum *pump.InsulinSensitivityArray, expectedDatum *pump.InsulinSensitivityArray, units *string)) {
					datum := NewInsulinSensitivityArray(units)
					mutator(datum, units)
					expectedDatum := CloneInsulinSensitivityArray(datum)
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
					pointer.String("mmol/L"),
					func(datum *pump.InsulinSensitivityArray, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.String("mmol/l"),
					func(datum *pump.InsulinSensitivityArray, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dL",
					pointer.String("mg/dL"),
					func(datum *pump.InsulinSensitivityArray, units *string) {},
					func(datum *pump.InsulinSensitivityArray, expectedDatum *pump.InsulinSensitivityArray, units *string) {
						testDataBloodGlucose.ExpectNormalizedValue((*datum)[0].Amount, (*expectedDatum)[0].Amount, units)
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.String("mg/dl"),
					func(datum *pump.InsulinSensitivityArray, units *string) {},
					func(datum *pump.InsulinSensitivityArray, expectedDatum *pump.InsulinSensitivityArray, units *string) {
						testDataBloodGlucose.ExpectNormalizedValue((*datum)[0].Amount, (*expectedDatum)[0].Amount, units)
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(units *string, mutator func(datum *pump.InsulinSensitivityArray, units *string), expectator func(datum *pump.InsulinSensitivityArray, expectedDatum *pump.InsulinSensitivityArray, units *string)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := NewInsulinSensitivityArray(units)
						mutator(datum, units)
						expectedDatum := CloneInsulinSensitivityArray(datum)
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
					pointer.String("mmol/L"),
					func(datum *pump.InsulinSensitivityArray, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.String("mmol/l"),
					func(datum *pump.InsulinSensitivityArray, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dL",
					pointer.String("mg/dL"),
					func(datum *pump.InsulinSensitivityArray, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dl",
					pointer.String("mg/dl"),
					func(datum *pump.InsulinSensitivityArray, units *string) {},
					nil,
				),
			)
		})
	})
})
