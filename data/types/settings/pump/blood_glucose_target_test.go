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

func NewBloodGlucoseTarget(units *string) *pump.BloodGlucoseTarget {
	datum := pump.NewBloodGlucoseTarget()
	datum.Target = *testDataBloodGlucose.NewTarget(units)
	datum.Start = pointer.Int(test.RandomIntFromRange(pump.BloodGlucoseTargetStartMinimum, pump.BloodGlucoseTargetStartMaximum))
	return datum
}

func CloneBloodGlucoseTarget(datum *pump.BloodGlucoseTarget) *pump.BloodGlucoseTarget {
	if datum == nil {
		return nil
	}
	clone := pump.NewBloodGlucoseTarget()
	clone.Target = *testDataBloodGlucose.CloneTarget(&datum.Target)
	clone.Start = test.CloneInt(datum.Start)
	return clone
}

func NewBloodGlucoseTargetArray(units *string) *pump.BloodGlucoseTargetArray {
	datum := pump.NewBloodGlucoseTargetArray()
	*datum = append(*datum, NewBloodGlucoseTarget(units))
	return datum
}

func CloneBloodGlucoseTargetArray(datumArray *pump.BloodGlucoseTargetArray) *pump.BloodGlucoseTargetArray {
	if datumArray == nil {
		return nil
	}
	clone := pump.NewBloodGlucoseTargetArray()
	for _, datum := range *datumArray {
		*clone = append(*clone, CloneBloodGlucoseTarget(datum))
	}
	return clone
}

var _ = Describe("BloodGlucoseTarget", func() {
	It("BloodGlucoseTargetStartMaximum is expected", func() {
		Expect(pump.BloodGlucoseTargetStartMaximum).To(Equal(86400000))
	})

	It("BloodGlucoseTargetStartMinimum is expected", func() {
		Expect(pump.BloodGlucoseTargetStartMinimum).To(Equal(0))
	})

	Context("ParseBloodGlucoseTarget", func() {
		// TODO
	})

	Context("NewBloodGlucoseTarget", func() {
		It("is successful", func() {
			Expect(pump.NewBloodGlucoseTarget()).To(Equal(&pump.BloodGlucoseTarget{}))
		})
	})

	Context("BloodGlucoseTarget", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(units *string, mutator func(datum *pump.BloodGlucoseTarget, units *string), expectedErrors ...error) {
					datum := NewBloodGlucoseTarget(units)
					mutator(datum, units)
					testDataTypes.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithStringAdapter(datum, units), structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.String("mmol/L"),
					func(datum *pump.BloodGlucoseTarget, units *string) {},
				),
				Entry("target missing",
					pointer.String("mmol/L"),
					func(datum *pump.BloodGlucoseTarget, units *string) { datum.Target = *dataBloodGlucose.NewTarget() },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/target"),
				),
				Entry("start missing",
					pointer.String("mmol/L"),
					func(datum *pump.BloodGlucoseTarget, units *string) { datum.Start = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/start"),
				),
				Entry("start out of range (lower)",
					pointer.String("mmol/L"),
					func(datum *pump.BloodGlucoseTarget, units *string) { datum.Start = pointer.Int(-1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/start"),
				),
				Entry("start in range (lower)",
					pointer.String("mmol/L"),
					func(datum *pump.BloodGlucoseTarget, units *string) { datum.Start = pointer.Int(0) },
				),
				Entry("start in range (upper)",
					pointer.String("mmol/L"),
					func(datum *pump.BloodGlucoseTarget, units *string) { datum.Start = pointer.Int(86400000) },
				),
				Entry("start out of range (upper)",
					pointer.String("mmol/L"),
					func(datum *pump.BloodGlucoseTarget, units *string) { datum.Start = pointer.Int(86400001) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/start"),
				),
				Entry("multiple errors",
					pointer.String("mmol/L"),
					func(datum *pump.BloodGlucoseTarget, units *string) {
						datum.Target = *dataBloodGlucose.NewTarget()
						datum.Start = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/target"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/start"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(units *string, mutator func(datum *pump.BloodGlucoseTarget, units *string), expectator func(datum *pump.BloodGlucoseTarget, expectedDatum *pump.BloodGlucoseTarget, units *string)) {
					for _, origin := range structure.Origins() {
						datum := NewBloodGlucoseTarget(units)
						mutator(datum, units)
						expectedDatum := CloneBloodGlucoseTarget(datum)
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
					func(datum *pump.BloodGlucoseTarget, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units missing",
					nil,
					func(datum *pump.BloodGlucoseTarget, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units invalid",
					pointer.String("invalid"),
					func(datum *pump.BloodGlucoseTarget, units *string) {},
					nil,
				),
				Entry("does not modify the datum; start missing",
					pointer.String("mmol/L"),
					func(datum *pump.BloodGlucoseTarget, units *string) { datum.Start = nil },
					nil,
				),
			)

			DescribeTable("normalizes the datum with origin external",
				func(units *string, mutator func(datum *pump.BloodGlucoseTarget, units *string), expectator func(datum *pump.BloodGlucoseTarget, expectedDatum *pump.BloodGlucoseTarget, units *string)) {
					datum := NewBloodGlucoseTarget(units)
					mutator(datum, units)
					expectedDatum := CloneBloodGlucoseTarget(datum)
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
					func(datum *pump.BloodGlucoseTarget, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.String("mmol/l"),
					func(datum *pump.BloodGlucoseTarget, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dL",
					pointer.String("mg/dL"),
					func(datum *pump.BloodGlucoseTarget, units *string) {},
					func(datum *pump.BloodGlucoseTarget, expectedDatum *pump.BloodGlucoseTarget, units *string) {
						testDataBloodGlucose.ExpectNormalizedTarget(&datum.Target, &expectedDatum.Target, units)
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.String("mg/dl"),
					func(datum *pump.BloodGlucoseTarget, units *string) {},
					func(datum *pump.BloodGlucoseTarget, expectedDatum *pump.BloodGlucoseTarget, units *string) {
						testDataBloodGlucose.ExpectNormalizedTarget(&datum.Target, &expectedDatum.Target, units)
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(units *string, mutator func(datum *pump.BloodGlucoseTarget, units *string), expectator func(datum *pump.BloodGlucoseTarget, expectedDatum *pump.BloodGlucoseTarget, units *string)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := NewBloodGlucoseTarget(units)
						mutator(datum, units)
						expectedDatum := CloneBloodGlucoseTarget(datum)
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
					func(datum *pump.BloodGlucoseTarget, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.String("mmol/l"),
					func(datum *pump.BloodGlucoseTarget, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dL",
					pointer.String("mg/dL"),
					func(datum *pump.BloodGlucoseTarget, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dl",
					pointer.String("mg/dl"),
					func(datum *pump.BloodGlucoseTarget, units *string) {},
					nil,
				),
			)
		})
	})

	Context("ParseBloodGlucoseTargetArray", func() {
		// TODO
	})

	Context("NewBloodGlucoseTargetArray", func() {
		It("is successful", func() {
			Expect(pump.NewBloodGlucoseTargetArray()).To(Equal(&pump.BloodGlucoseTargetArray{}))
		})
	})

	Context("BloodGlucoseTargetArray", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(units *string, mutator func(datum *pump.BloodGlucoseTargetArray, units *string), expectedErrors ...error) {
					datum := pump.NewBloodGlucoseTargetArray()
					mutator(datum, units)
					testDataTypes.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithStringAdapter(datum, units), structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.String("mmol/L"),
					func(datum *pump.BloodGlucoseTargetArray, units *string) {},
				),
				Entry("empty",
					pointer.String("mmol/L"),
					func(datum *pump.BloodGlucoseTargetArray, units *string) { *datum = *pump.NewBloodGlucoseTargetArray() },
				),
				Entry("nil",
					pointer.String("mmol/L"),
					func(datum *pump.BloodGlucoseTargetArray, units *string) { *datum = append(*datum, nil) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
				),
				Entry("single invalid",
					pointer.String("mmol/L"),
					func(datum *pump.BloodGlucoseTargetArray, units *string) {
						invalid := NewBloodGlucoseTarget(units)
						invalid.Target = *dataBloodGlucose.NewTarget()
						*datum = append(*datum, invalid)
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0/target"),
				),
				Entry("single valid",
					pointer.String("mmol/L"),
					func(datum *pump.BloodGlucoseTargetArray, units *string) {
						*datum = append(*datum, NewBloodGlucoseTarget(units))
					},
				),
				Entry("multiple invalid",
					pointer.String("mmol/L"),
					func(datum *pump.BloodGlucoseTargetArray, units *string) {
						invalid := NewBloodGlucoseTarget(units)
						invalid.Target = *dataBloodGlucose.NewTarget()
						*datum = append(*datum, NewBloodGlucoseTarget(units), invalid, NewBloodGlucoseTarget(units))
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/1/target"),
				),
				Entry("multiple valid",
					pointer.String("mmol/L"),
					func(datum *pump.BloodGlucoseTargetArray, units *string) {
						*datum = append(*datum, NewBloodGlucoseTarget(units), NewBloodGlucoseTarget(units), NewBloodGlucoseTarget(units))
					},
				),
				Entry("multiple errors",
					pointer.String("mmol/L"),
					func(datum *pump.BloodGlucoseTargetArray, units *string) {
						invalid := NewBloodGlucoseTarget(units)
						invalid.Target = *dataBloodGlucose.NewTarget()
						*datum = append(*datum, nil, invalid, NewBloodGlucoseTarget(units))
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/1/target"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(units *string, mutator func(datum *pump.BloodGlucoseTargetArray, units *string), expectator func(datum *pump.BloodGlucoseTargetArray, expectedDatum *pump.BloodGlucoseTargetArray, units *string)) {
					for _, origin := range structure.Origins() {
						datum := NewBloodGlucoseTargetArray(units)
						mutator(datum, units)
						expectedDatum := CloneBloodGlucoseTargetArray(datum)
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
					func(datum *pump.BloodGlucoseTargetArray, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units missing",
					nil,
					func(datum *pump.BloodGlucoseTargetArray, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units invalid",
					pointer.String("invalid"),
					func(datum *pump.BloodGlucoseTargetArray, units *string) {},
					nil,
				),
				Entry("does not modify the datum; start missing",
					pointer.String("mmol/L"),
					func(datum *pump.BloodGlucoseTargetArray, units *string) { (*datum)[0].Start = nil },
					nil,
				),
			)

			DescribeTable("normalizes the datum with structure external",
				func(units *string, mutator func(datum *pump.BloodGlucoseTargetArray, units *string), expectator func(datum *pump.BloodGlucoseTargetArray, expectedDatum *pump.BloodGlucoseTargetArray, units *string)) {
					datum := NewBloodGlucoseTargetArray(units)
					mutator(datum, units)
					expectedDatum := CloneBloodGlucoseTargetArray(datum)
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
					func(datum *pump.BloodGlucoseTargetArray, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.String("mmol/l"),
					func(datum *pump.BloodGlucoseTargetArray, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dL",
					pointer.String("mg/dL"),
					func(datum *pump.BloodGlucoseTargetArray, units *string) {},
					func(datum *pump.BloodGlucoseTargetArray, expectedDatum *pump.BloodGlucoseTargetArray, units *string) {
						testDataBloodGlucose.ExpectNormalizedTarget(&(*datum)[0].Target, &(*expectedDatum)[0].Target, units)
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.String("mg/dl"),
					func(datum *pump.BloodGlucoseTargetArray, units *string) {},
					func(datum *pump.BloodGlucoseTargetArray, expectedDatum *pump.BloodGlucoseTargetArray, units *string) {
						testDataBloodGlucose.ExpectNormalizedTarget(&(*datum)[0].Target, &(*expectedDatum)[0].Target, units)
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(units *string, mutator func(datum *pump.BloodGlucoseTargetArray, units *string), expectator func(datum *pump.BloodGlucoseTargetArray, expectedDatum *pump.BloodGlucoseTargetArray, units *string)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := NewBloodGlucoseTargetArray(units)
						mutator(datum, units)
						expectedDatum := CloneBloodGlucoseTargetArray(datum)
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
					func(datum *pump.BloodGlucoseTargetArray, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.String("mmol/l"),
					func(datum *pump.BloodGlucoseTargetArray, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dL",
					pointer.String("mg/dL"),
					func(datum *pump.BloodGlucoseTargetArray, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dl",
					pointer.String("mg/dl"),
					func(datum *pump.BloodGlucoseTargetArray, units *string) {},
					nil,
				),
			)
		})
	})
})
