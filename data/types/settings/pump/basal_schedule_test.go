package pump_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	testDataTypesBasal "github.com/tidepool-org/platform/data/types/basal/test"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewBasalSchedule() *pump.BasalSchedule {
	datum := pump.NewBasalSchedule()
	datum.Rate = pointer.Float64(test.RandomFloat64FromRange(pump.BasalScheduleRateMinimum, pump.BasalScheduleRateMaximum))
	datum.Start = pointer.Int(test.RandomIntFromRange(pump.BasalScheduleStartMinimum, pump.BasalScheduleStartMaximum))
	return datum
}

func CloneBasalSchedule(datum *pump.BasalSchedule) *pump.BasalSchedule {
	if datum == nil {
		return nil
	}
	clone := pump.NewBasalSchedule()
	clone.Rate = test.CloneFloat64(datum.Rate)
	clone.Start = test.CloneInt(datum.Start)
	return clone
}

func NewBasalScheduleArray() *pump.BasalScheduleArray {
	datum := pump.NewBasalScheduleArray()
	*datum = append(*datum, NewBasalSchedule())
	return datum
}

func CloneBasalScheduleArray(datumArray *pump.BasalScheduleArray) *pump.BasalScheduleArray {
	if datumArray == nil {
		return nil
	}
	clone := pump.NewBasalScheduleArray()
	for _, datum := range *datumArray {
		*clone = append(*clone, CloneBasalSchedule(datum))
	}
	return clone
}

func NewBasalScheduleArrayMap() *pump.BasalScheduleArrayMap {
	datum := pump.NewBasalScheduleArrayMap()
	datum.Set(testDataTypesBasal.NewScheduleName(), NewBasalScheduleArray())
	return datum
}

func CloneBasalScheduleArrayMap(datumMap *pump.BasalScheduleArrayMap) *pump.BasalScheduleArrayMap {
	if datumMap == nil {
		return nil
	}
	clone := pump.NewBasalScheduleArrayMap()
	for datumName, datumArray := range *datumMap {
		clone.Set(datumName, CloneBasalScheduleArray(datumArray))
	}
	return clone
}

var _ = Describe("BasalSchedule", func() {
	It("BasalScheduleRateMaximum is expected", func() {
		Expect(pump.BasalScheduleRateMaximum).To(Equal(100.0))
	})

	It("BasalScheduleRateMinimum is expected", func() {
		Expect(pump.BasalScheduleRateMinimum).To(Equal(0.0))
	})

	It("BasalScheduleStartMaximum is expected", func() {
		Expect(pump.BasalScheduleStartMaximum).To(Equal(86400000))
	})

	It("BasalScheduleStartMinimum is expected", func() {
		Expect(pump.BasalScheduleStartMinimum).To(Equal(0))
	})

	Context("ParseBasalSchedule", func() {
		// TODO
	})

	Context("NewBasalSchedule", func() {
		It("is successful", func() {
			Expect(pump.NewBasalSchedule()).To(Equal(&pump.BasalSchedule{}))
		})
	})

	Context("BasalSchedule", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *pump.BasalSchedule), expectedErrors ...error) {
					datum := NewBasalSchedule()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pump.BasalSchedule) {},
				),
				Entry("rate missing",
					func(datum *pump.BasalSchedule) { datum.Rate = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("rate out of range (lower)",
					func(datum *pump.BasalSchedule) { datum.Rate = pointer.Float64(-0.1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0, 100), "/rate"),
				),
				Entry("rate in range (lower)",
					func(datum *pump.BasalSchedule) { datum.Rate = pointer.Float64(0.0) },
				),
				Entry("rate in range (upper)",
					func(datum *pump.BasalSchedule) { datum.Rate = pointer.Float64(100.0) },
				),
				Entry("rate out of range (upper)",
					func(datum *pump.BasalSchedule) { datum.Rate = pointer.Float64(100.1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(100.1, 0, 100), "/rate"),
				),
				Entry("start missing",
					func(datum *pump.BasalSchedule) { datum.Start = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/start"),
				),
				Entry("start out of range (lower)",
					func(datum *pump.BasalSchedule) { datum.Start = pointer.Int(-1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/start"),
				),
				Entry("start in range (lower)",
					func(datum *pump.BasalSchedule) { datum.Start = pointer.Int(0) },
				),
				Entry("start in range (upper)",
					func(datum *pump.BasalSchedule) { datum.Start = pointer.Int(86400000) },
				),
				Entry("start out of range (upper)",
					func(datum *pump.BasalSchedule) { datum.Start = pointer.Int(86400001) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/start"),
				),
				Entry("multiple errors",
					func(datum *pump.BasalSchedule) {
						datum.Rate = nil
						datum.Start = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/start"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *pump.BasalSchedule)) {
					for _, origin := range structure.Origins() {
						datum := NewBasalSchedule()
						mutator(datum)
						expectedDatum := CloneBasalSchedule(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *pump.BasalSchedule) {},
				),
				Entry("does not modify the datum; rate missing",
					func(datum *pump.BasalSchedule) { datum.Rate = nil },
				),
				Entry("does not modify the datum; start missing",
					func(datum *pump.BasalSchedule) { datum.Start = nil },
				),
			)
		})
	})

	Context("ParseBasalScheduleArray", func() {
		// TODO
	})

	Context("NewBasalScheduleArray", func() {
		It("is successful", func() {
			Expect(pump.NewBasalScheduleArray()).To(Equal(&pump.BasalScheduleArray{}))
		})
	})

	Context("BasalScheduleArray", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *pump.BasalScheduleArray), expectedErrors ...error) {
					datum := pump.NewBasalScheduleArray()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pump.BasalScheduleArray) {},
				),
				Entry("empty",
					func(datum *pump.BasalScheduleArray) { *datum = *pump.NewBasalScheduleArray() },
				),
				Entry("nil",
					func(datum *pump.BasalScheduleArray) { *datum = append(*datum, nil) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
				),
				Entry("single invalid",
					func(datum *pump.BasalScheduleArray) {
						invalid := NewBasalSchedule()
						invalid.Rate = nil
						*datum = append(*datum, invalid)
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0/rate"),
				),
				Entry("single valid",
					func(datum *pump.BasalScheduleArray) {
						*datum = append(*datum, NewBasalSchedule())
					},
				),
				Entry("multiple invalid",
					func(datum *pump.BasalScheduleArray) {
						invalid := NewBasalSchedule()
						invalid.Rate = nil
						*datum = append(*datum, NewBasalSchedule(), invalid, NewBasalSchedule())
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/1/rate"),
				),
				Entry("multiple valid",
					func(datum *pump.BasalScheduleArray) {
						*datum = append(*datum, NewBasalSchedule(), NewBasalSchedule(), NewBasalSchedule())
					},
				),
				Entry("multiple errors",
					func(datum *pump.BasalScheduleArray) {
						invalid := NewBasalSchedule()
						invalid.Rate = nil
						*datum = append(*datum, nil, invalid, NewBasalSchedule())
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/1/rate"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *pump.BasalScheduleArray)) {
					for _, origin := range structure.Origins() {
						datum := NewBasalScheduleArray()
						mutator(datum)
						expectedDatum := CloneBasalScheduleArray(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *pump.BasalScheduleArray) {},
				),
				Entry("does not modify the datum; rate missing",
					func(datum *pump.BasalScheduleArray) { (*datum)[0].Rate = nil },
				),
				Entry("does not modify the datum; start missing",
					func(datum *pump.BasalScheduleArray) { (*datum)[0].Start = nil },
				),
			)
		})
	})

	Context("ParseBasalScheduleArrayMap", func() {
		// TODO
	})

	Context("NewBasalScheduleArrayMap", func() {
		It("is successful", func() {
			Expect(pump.NewBasalScheduleArrayMap()).To(Equal(&pump.BasalScheduleArrayMap{}))
		})
	})

	Context("BasalScheduleArrayMap", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *pump.BasalScheduleArrayMap), expectedErrors ...error) {
					datum := pump.NewBasalScheduleArrayMap()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pump.BasalScheduleArrayMap) {},
				),
				Entry("empty",
					func(datum *pump.BasalScheduleArrayMap) { *datum = *pump.NewBasalScheduleArrayMap() },
				),
				Entry("empty name",
					func(datum *pump.BasalScheduleArrayMap) { datum.Set("", NewBasalScheduleArray()) },
				),
				Entry("nil value",
					func(datum *pump.BasalScheduleArrayMap) { datum.Set("", nil) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/"),
				),
				Entry("single invalid",
					func(datum *pump.BasalScheduleArrayMap) {
						invalid := NewBasalScheduleArray()
						(*invalid)[0].Rate = nil
						datum.Set("one", invalid)
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/one/0/rate"),
				),
				Entry("single valid",
					func(datum *pump.BasalScheduleArrayMap) {
						datum.Set("one", NewBasalScheduleArray())
					},
				),
				Entry("multiple invalid",
					func(datum *pump.BasalScheduleArrayMap) {
						invalid := NewBasalScheduleArray()
						(*invalid)[0].Rate = nil
						datum.Set("one", NewBasalScheduleArray())
						datum.Set("two", invalid)
						datum.Set("three", NewBasalScheduleArray())
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/two/0/rate"),
				),
				Entry("multiple valid",
					func(datum *pump.BasalScheduleArrayMap) {
						datum.Set("one", NewBasalScheduleArray())
						datum.Set("two", NewBasalScheduleArray())
						datum.Set("three", NewBasalScheduleArray())
					},
				),
				Entry("multiple errors",
					func(datum *pump.BasalScheduleArrayMap) {
						invalid := NewBasalScheduleArray()
						(*invalid)[0].Rate = nil
						datum.Set("one", nil)
						datum.Set("two", invalid)
						datum.Set("three", NewBasalScheduleArray())
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/one"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/two/0/rate"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *pump.BasalScheduleArrayMap)) {
					for _, origin := range structure.Origins() {
						datum := NewBasalScheduleArrayMap()
						mutator(datum)
						expectedDatum := CloneBasalScheduleArrayMap(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *pump.BasalScheduleArrayMap) {},
				),
			)
		})
	})
})
