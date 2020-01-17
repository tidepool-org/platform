package pump_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func NewBloodGlucosePreMealTarget(units *string) *pump.BloodGlucosePreMealTarget {
	datum := pump.NewBloodGlucosePreMealTarget()
	min, max := glucose.ValueRangeForUnits(units)
	datum.Low = pointer.FromFloat64(test.RandomFloat64FromRange(min, max-5))
	datum.High = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Low, max))
	return datum
}

func CloneBloodGlucosePreMealTarget(datum *pump.BloodGlucosePreMealTarget, units *string) *pump.BloodGlucosePreMealTarget {
	if datum == nil {
		return nil
	}
	clone := NewBloodGlucosePreMealTarget(units)
	clone.Low = pointer.CloneFloat64(datum.Low)
	clone.High = pointer.CloneFloat64(datum.High)
	return clone
}

type ValidatableWithUnits interface {
	Validate(validator structure.Validator, units *string)
}

type ValidatableWithUnitsAdapter struct {
	validatableWithUnits ValidatableWithUnits
	units                *string
}

func NewValidatableWithUnitsAdapter(validatableWithUnits ValidatableWithUnits, units *string) *ValidatableWithUnitsAdapter {
	return &ValidatableWithUnitsAdapter{
		validatableWithUnits: validatableWithUnits,
		units:                units,
	}
}

func (v *ValidatableWithUnitsAdapter) Validate(validator structure.Validator) {
	v.validatableWithUnits.Validate(validator, v.units)
}

func low(a float64, b float64) float64  { return a }
func high(a float64, b float64) float64 { return b }

var _ = Describe("BloodGlucosePreMealTarget", func() {
	Context("ParseBloodGlucosePreMealTarget", func() {
		// TODO
	})

	Context("NewBloodGlucosePreMealTarget", func() {
		It("is successful", func() {
			Expect(pump.NewBloodGlucosePreMealTarget()).To(Equal(&pump.BloodGlucosePreMealTarget{}))
		})
	})

	Context("BloodGlucosePreMealTarget", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *pump.BloodGlucosePreMealTarget, units *string), units *string, expectedErrors ...error) {
					datum := NewBloodGlucosePreMealTarget(units)
					mutator(datum, units)
					dataTypesTest.ValidateWithExpectedOrigins(NewValidatableWithUnitsAdapter(datum, units), structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pump.BloodGlucosePreMealTarget, units *string) {},
					pointer.FromString("mmol/L"),
				),
				Entry("Low missing",
					func(datum *pump.BloodGlucosePreMealTarget, units *string) { datum.Low = nil },
					pointer.FromString("mmol/L"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/low"),
				),
				Entry("High missing",
					func(datum *pump.BloodGlucosePreMealTarget, units *string) { datum.High = nil },
					pointer.FromString("mmol/L"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/high"),
				),
				Entry("Low valid",
					func(datum *pump.BloodGlucosePreMealTarget, units *string) {
						min, _ := glucose.ValueRangeForUnits(units)
						datum.Low = pointer.FromFloat64(min)
					},
					pointer.FromString("mmol/L"),
				),
				Entry("Low too low",
					func(datum *pump.BloodGlucosePreMealTarget, units *string) {
						min, _ := glucose.ValueRangeForUnits(units)
						datum.Low = pointer.FromFloat64(min - 1)
					},
					pointer.FromString("mmol/L"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(low(glucose.ValueRangeForUnits(pointer.FromString("mmol/L")))-1,
						low(glucose.ValueRangeForUnits(pointer.FromString("mmol/L"))),
						high(glucose.ValueRangeForUnits(pointer.FromString("mmol/L")))), "/low"),
				),
				Entry("Low too High",
					func(datum *pump.BloodGlucosePreMealTarget, units *string) {
						_, max := glucose.ValueRangeForUnits(units)
						datum.Low = pointer.FromFloat64(max + 1)
					},
					pointer.FromString("mmol/L"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(high(glucose.ValueRangeForUnits(pointer.FromString("mmol/L")))+1,
						low(glucose.ValueRangeForUnits(pointer.FromString("mmol/L"))),
						high(glucose.ValueRangeForUnits(pointer.FromString("mmol/L")))), "/low"),
				),
				Entry("High valid",
					func(datum *pump.BloodGlucosePreMealTarget, units *string) {
						min, _ := glucose.ValueRangeForUnits(units)
						datum.High = pointer.FromFloat64(min)
					},
					pointer.FromString("mmol/L"),
				),
				Entry("High too low",
					func(datum *pump.BloodGlucosePreMealTarget, units *string) {
						min, _ := glucose.ValueRangeForUnits(units)
						datum.High = pointer.FromFloat64(min - 1)
					},
					pointer.FromString("mmol/L"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(low(glucose.ValueRangeForUnits(pointer.FromString("mmol/L")))-1,
						low(glucose.ValueRangeForUnits(pointer.FromString("mmol/L"))),
						high(glucose.ValueRangeForUnits(pointer.FromString("mmol/L")))), "/high"),
				),
				Entry("High too High",
					func(datum *pump.BloodGlucosePreMealTarget, units *string) {
						_, max := glucose.ValueRangeForUnits(units)
						datum.High = pointer.FromFloat64(max + 1)
					},
					pointer.FromString("mmol/L"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(high(glucose.ValueRangeForUnits(pointer.FromString("mmol/L")))+1,
						low(glucose.ValueRangeForUnits(pointer.FromString("mmol/L"))),
						high(glucose.ValueRangeForUnits(pointer.FromString("mmol/L")))), "/high"),
				),
				Entry("multiple errors",
					func(datum *pump.BloodGlucosePreMealTarget, units *string) {
						min, max := glucose.ValueRangeForUnits(units)
						datum.Low = pointer.FromFloat64(min - 1)
						datum.High = pointer.FromFloat64(max + 1)
					},
					pointer.FromString("mmol/L"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(low(glucose.ValueRangeForUnits(pointer.FromString("mmol/L")))-1,
						low(glucose.ValueRangeForUnits(pointer.FromString("mmol/L"))),
						high(glucose.ValueRangeForUnits(pointer.FromString("mmol/L")))), "/low"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(high(glucose.ValueRangeForUnits(pointer.FromString("mmol/L")))+1,
						low(glucose.ValueRangeForUnits(pointer.FromString("mmol/L"))),
						high(glucose.ValueRangeForUnits(pointer.FromString("mmol/L")))), "/high"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(units *string, mutator func(datum *pump.BloodGlucosePreMealTarget, units *string)) {
					for _, origin := range structure.Origins() {
						datum := NewBloodGlucosePreMealTarget(units)
						mutator(datum, units)
						expectedDatum := CloneBloodGlucosePreMealTarget(datum, units)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin), units)
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucosePreMealTarget, units *string) {},
				),
				Entry("does not modify the datum; Low missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucosePreMealTarget, units *string) { datum.Low = nil },
				),
				Entry("does not modify the datum; High missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucosePreMealTarget, units *string) { datum.High = nil },
				),
			)
		})
	})
})
