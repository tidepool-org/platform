package cgm_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	testDataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose/test"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/settings/cgm"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewHighLevelAlert(units *string) *cgm.HighLevelAlert {
	datum := cgm.NewHighLevelAlert()
	datum.Enabled = pointer.FromBool(test.RandomBool())
	datum.Level = pointer.FromFloat64(test.RandomFloat64FromRange(datum.LevelRangeForUnits(units)))
	datum.Snooze = pointer.FromInt(test.RandomIntFromArray(cgm.LevelAlertSnoozes()))
	return datum
}

func CloneHighLevelAlert(datum *cgm.HighLevelAlert) *cgm.HighLevelAlert {
	if datum == nil {
		return nil
	}
	clone := cgm.NewHighLevelAlert()
	clone.Enabled = test.CloneBool(datum.Enabled)
	clone.Level = test.CloneFloat64(datum.Level)
	clone.Snooze = test.CloneInt(datum.Snooze)
	return clone
}

func NewLowLevelAlert(units *string) *cgm.LowLevelAlert {
	datum := cgm.NewLowLevelAlert()
	datum.Enabled = pointer.FromBool(test.RandomBool())
	datum.Level = pointer.FromFloat64(test.RandomFloat64FromRange(datum.LevelRangeForUnits(units)))
	datum.Snooze = pointer.FromInt(test.RandomIntFromArray(cgm.LevelAlertSnoozes()))
	return datum
}

func CloneLowLevelAlert(datum *cgm.LowLevelAlert) *cgm.LowLevelAlert {
	if datum == nil {
		return nil
	}
	clone := cgm.NewLowLevelAlert()
	clone.Enabled = test.CloneBool(datum.Enabled)
	clone.Level = test.CloneFloat64(datum.Level)
	clone.Snooze = test.CloneInt(datum.Snooze)
	return clone
}

var _ = Describe("LevelAlert", func() {
	It("HighLevelAlertLevelMgdLMaximum returns expected", func() {
		Expect(cgm.HighLevelAlertLevelMgdLMaximum).To(Equal(float64(400)))
	})

	It("HighLevelAlertLevelMgdLMinimum returns expected", func() {
		Expect(cgm.HighLevelAlertLevelMgdLMinimum).To(Equal(float64(120)))
	})

	It("HighLevelAlertLevelMmolLMaximum returns expected", func() {
		Expect(cgm.HighLevelAlertLevelMmolLMaximum).To(Equal(22.20299))
	})

	It("HighLevelAlertLevelMmolLMinimum returns expected", func() {
		Expect(cgm.HighLevelAlertLevelMmolLMinimum).To(Equal(6.66090))
	})

	It("LowLevelAlertLevelMgdLMaximum returns expected", func() {
		Expect(cgm.LowLevelAlertLevelMgdLMaximum).To(Equal(float64(100)))
	})

	It("LowLevelAlertLevelMgdLMinimum returns expected", func() {
		Expect(cgm.LowLevelAlertLevelMgdLMinimum).To(Equal(float64(59)))
	})

	It("LowLevelAlertLevelMmolLMaximum returns expected", func() {
		Expect(cgm.LowLevelAlertLevelMmolLMaximum).To(Equal(5.55075))
	})

	It("LowLevelAlertLevelMmolLMinimum returns expected", func() {
		Expect(cgm.LowLevelAlertLevelMmolLMinimum).To(Equal(3.27494))
	})

	It("LevelAlertSnoozes returns expected", func() {
		Expect(cgm.LevelAlertSnoozes()).To(Equal([]int{
			0, 900000, 1800000, 2700000, 3600000, 4500000, 5400000, 6300000,
			7200000, 8100000, 9000000, 9900000, 10800000, 11700000, 12600000,
			13500000, 14400000, 15300000, 16200000, 17100000, 18000000}))
	})

	Context("ParseHighLevelAlert", func() {
		// TODO
	})

	Context("NewHighLevelAlert", func() {
		It("is successful", func() {
			Expect(cgm.NewHighLevelAlert()).To(Equal(&cgm.HighLevelAlert{}))
		})
	})

	Context("HighLevelAlert", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(units *string, mutator func(datum *cgm.HighLevelAlert, units *string), expectedErrors ...error) {
					datum := NewHighLevelAlert(units)
					mutator(datum, units)
					testDataTypes.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithStringAdapter(datum, units), structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.FromString("mmol/L"),
					func(datum *cgm.HighLevelAlert, units *string) {},
				),
				Entry("enabled missing",
					nil,
					func(datum *cgm.HighLevelAlert, units *string) { datum.Enabled = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
				),
				Entry("enabled true",
					nil,
					func(datum *cgm.HighLevelAlert, units *string) { datum.Enabled = pointer.FromBool(true) },
				),
				Entry("enabled false",
					nil,
					func(datum *cgm.HighLevelAlert, units *string) { datum.Enabled = pointer.FromBool(false) },
				),
				Entry("units missing; level missing",
					nil,
					func(datum *cgm.HighLevelAlert, units *string) { datum.Level = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/level"),
				),
				Entry("units missing; level out of range (lower)",
					nil,
					func(datum *cgm.HighLevelAlert, units *string) { datum.Level = pointer.FromFloat64(6.66089) },
				),
				Entry("units missing; level in range (lower)",
					nil,
					func(datum *cgm.HighLevelAlert, units *string) { datum.Level = pointer.FromFloat64(6.66090) },
				),
				Entry("units missing; level in range (upper)",
					nil,
					func(datum *cgm.HighLevelAlert, units *string) { datum.Level = pointer.FromFloat64(400) },
				),
				Entry("units missing; level out of range (upper)",
					nil,
					func(datum *cgm.HighLevelAlert, units *string) { datum.Level = pointer.FromFloat64(401) },
				),
				Entry("units invalid; level missing",
					pointer.FromString("invalid"),
					func(datum *cgm.HighLevelAlert, units *string) { datum.Level = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/level"),
				),
				Entry("units invalid; level out of range (lower)",
					pointer.FromString("invalid"),
					func(datum *cgm.HighLevelAlert, units *string) { datum.Level = pointer.FromFloat64(6.66089) },
				),
				Entry("units invalid; level in range (lower)",
					pointer.FromString("invalid"),
					func(datum *cgm.HighLevelAlert, units *string) { datum.Level = pointer.FromFloat64(6.66090) },
				),
				Entry("units invalid; level in range (upper)",
					pointer.FromString("invalid"),
					func(datum *cgm.HighLevelAlert, units *string) { datum.Level = pointer.FromFloat64(400) },
				),
				Entry("units invalid; level out of range (upper)",
					pointer.FromString("invalid"),
					func(datum *cgm.HighLevelAlert, units *string) { datum.Level = pointer.FromFloat64(401) },
				),
				Entry("units mmol/L; level missing",
					pointer.FromString("mmol/L"),
					func(datum *cgm.HighLevelAlert, units *string) { datum.Level = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/level"),
				),
				Entry("units mmol/L; level out of range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *cgm.HighLevelAlert, units *string) { datum.Level = pointer.FromFloat64(6.66089) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(6.66089, 6.66090, 22.20299), "/level"),
				),
				Entry("units mmol/L; level in range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *cgm.HighLevelAlert, units *string) { datum.Level = pointer.FromFloat64(6.66090) },
				),
				Entry("units mmol/L; level in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *cgm.HighLevelAlert, units *string) { datum.Level = pointer.FromFloat64(22.20299) },
				),
				Entry("units mmol/L; level out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *cgm.HighLevelAlert, units *string) { datum.Level = pointer.FromFloat64(22.20300) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(22.20300, 6.66090, 22.20299), "/level"),
				),
				Entry("units mmol/l; level missing",
					pointer.FromString("mmol/l"),
					func(datum *cgm.HighLevelAlert, units *string) { datum.Level = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/level"),
				),
				Entry("units mmol/l; level out of range (lower)",
					pointer.FromString("mmol/l"),
					func(datum *cgm.HighLevelAlert, units *string) { datum.Level = pointer.FromFloat64(6.66089) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(6.66089, 6.66090, 22.20299), "/level"),
				),
				Entry("units mmol/l; level in range (lower)",
					pointer.FromString("mmol/l"),
					func(datum *cgm.HighLevelAlert, units *string) { datum.Level = pointer.FromFloat64(6.66090) },
				),
				Entry("units mmol/l; level in range (upper)",
					pointer.FromString("mmol/l"),
					func(datum *cgm.HighLevelAlert, units *string) { datum.Level = pointer.FromFloat64(22.20299) },
				),
				Entry("units mmol/l; level out of range (upper)",
					pointer.FromString("mmol/l"),
					func(datum *cgm.HighLevelAlert, units *string) { datum.Level = pointer.FromFloat64(22.20300) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(22.20300, 6.66090, 22.20299), "/level"),
				),
				Entry("units mg/dL; level missing",
					pointer.FromString("mg/dL"),
					func(datum *cgm.HighLevelAlert, units *string) { datum.Level = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/level"),
				),
				Entry("units mg/dL; level out of range (lower)",
					pointer.FromString("mg/dL"),
					func(datum *cgm.HighLevelAlert, units *string) { datum.Level = pointer.FromFloat64(119) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(119, 120, 400), "/level"),
				),
				Entry("units mg/dL; level in range (lower)",
					pointer.FromString("mg/dL"),
					func(datum *cgm.HighLevelAlert, units *string) { datum.Level = pointer.FromFloat64(120) },
				),
				Entry("units mg/dL; level in range (upper)",
					pointer.FromString("mg/dL"),
					func(datum *cgm.HighLevelAlert, units *string) { datum.Level = pointer.FromFloat64(400) },
				),
				Entry("units mg/dL; level out of range (upper)",
					pointer.FromString("mg/dL"),
					func(datum *cgm.HighLevelAlert, units *string) { datum.Level = pointer.FromFloat64(401) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(401, 120, 400), "/level"),
				),
				Entry("units mg/dl; level missing",
					pointer.FromString("mg/dl"),
					func(datum *cgm.HighLevelAlert, units *string) { datum.Level = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/level"),
				),
				Entry("units mg/dl; level out of range (lower)",
					pointer.FromString("mg/dl"),
					func(datum *cgm.HighLevelAlert, units *string) { datum.Level = pointer.FromFloat64(119) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(119, 120, 400), "/level"),
				),
				Entry("units mg/dl; level in range (lower)",
					pointer.FromString("mg/dl"),
					func(datum *cgm.HighLevelAlert, units *string) { datum.Level = pointer.FromFloat64(120) },
				),
				Entry("units mg/dl; level in range (upper)",
					pointer.FromString("mg/dl"),
					func(datum *cgm.HighLevelAlert, units *string) { datum.Level = pointer.FromFloat64(400) },
				),
				Entry("units mg/dl; level out of range (upper)",
					pointer.FromString("mg/dl"),
					func(datum *cgm.HighLevelAlert, units *string) { datum.Level = pointer.FromFloat64(401) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(401, 120, 400), "/level"),
				),
				Entry("snooze missing",
					nil,
					func(datum *cgm.HighLevelAlert, units *string) { datum.Snooze = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/snooze"),
				),
				Entry("snooze invalid",
					nil,
					func(datum *cgm.HighLevelAlert, units *string) { datum.Snooze = pointer.FromInt(1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueIntNotOneOf(1, cgm.LevelAlertSnoozes()), "/snooze"),
				),
				Entry("snooze valid",
					nil,
					func(datum *cgm.HighLevelAlert, units *string) { datum.Snooze = pointer.FromInt(0) },
				),
				Entry("multiple errors",
					pointer.FromString("mmol/L"),
					func(datum *cgm.HighLevelAlert, units *string) {
						datum.Enabled = nil
						datum.Level = nil
						datum.Snooze = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/level"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/snooze"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(units *string, mutator func(datum *cgm.HighLevelAlert, units *string), expectator func(datum *cgm.HighLevelAlert, expectedDatum *cgm.HighLevelAlert, units *string)) {
					for _, origin := range structure.Origins() {
						datum := NewHighLevelAlert(units)
						mutator(datum, units)
						expectedDatum := CloneHighLevelAlert(datum)
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
					func(datum *cgm.HighLevelAlert, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units missing",
					nil,
					func(datum *cgm.HighLevelAlert, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units invalid",
					pointer.FromString("invalid"),
					func(datum *cgm.HighLevelAlert, units *string) {},
					nil,
				),
				Entry("does not modify the datum; enabled missing",
					pointer.FromString("mmol/L"),
					func(datum *cgm.HighLevelAlert, units *string) { datum.Enabled = nil },
					nil,
				),
				Entry("does not modify the datum; snooze missing",
					pointer.FromString("mmol/L"),
					func(datum *cgm.HighLevelAlert, units *string) { datum.Snooze = nil },
					nil,
				),
			)

			DescribeTable("normalizes the datum with origin external",
				func(units *string, mutator func(datum *cgm.HighLevelAlert, units *string), expectator func(datum *cgm.HighLevelAlert, expectedDatum *cgm.HighLevelAlert, units *string)) {
					datum := NewHighLevelAlert(units)
					mutator(datum, units)
					expectedDatum := CloneHighLevelAlert(datum)
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
					func(datum *cgm.HighLevelAlert, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *cgm.HighLevelAlert, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *cgm.HighLevelAlert, units *string) {},
					func(datum *cgm.HighLevelAlert, expectedDatum *cgm.HighLevelAlert, units *string) {
						testDataBloodGlucose.ExpectNormalizedValue(datum.Level, expectedDatum.Level, units)
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *cgm.HighLevelAlert, units *string) {},
					func(datum *cgm.HighLevelAlert, expectedDatum *cgm.HighLevelAlert, units *string) {
						testDataBloodGlucose.ExpectNormalizedValue(datum.Level, expectedDatum.Level, units)
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(units *string, mutator func(datum *cgm.HighLevelAlert, units *string), expectator func(datum *cgm.HighLevelAlert, expectedDatum *cgm.HighLevelAlert, units *string)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := NewHighLevelAlert(units)
						mutator(datum, units)
						expectedDatum := CloneHighLevelAlert(datum)
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
					func(datum *cgm.HighLevelAlert, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *cgm.HighLevelAlert, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *cgm.HighLevelAlert, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *cgm.HighLevelAlert, units *string) {},
					nil,
				),
			)
		})
	})

	Context("ParseLowLevelAlert", func() {
		// TODO
	})

	Context("NewLowLevelAlert", func() {
		It("is successful", func() {
			Expect(cgm.NewLowLevelAlert()).To(Equal(&cgm.LowLevelAlert{}))
		})
	})

	Context("LowLevelAlert", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(units *string, mutator func(datum *cgm.LowLevelAlert, units *string), expectedErrors ...error) {
					datum := NewLowLevelAlert(units)
					mutator(datum, units)
					testDataTypes.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithStringAdapter(datum, units), structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.FromString("mmol/L"),
					func(datum *cgm.LowLevelAlert, units *string) {},
				),
				Entry("enabled missing",
					nil,
					func(datum *cgm.LowLevelAlert, units *string) { datum.Enabled = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
				),
				Entry("enabled true",
					nil,
					func(datum *cgm.LowLevelAlert, units *string) { datum.Enabled = pointer.FromBool(true) },
				),
				Entry("enabled false",
					nil,
					func(datum *cgm.LowLevelAlert, units *string) { datum.Enabled = pointer.FromBool(false) },
				),
				Entry("units missing; level missing",
					nil,
					func(datum *cgm.LowLevelAlert, units *string) { datum.Level = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/level"),
				),
				Entry("units missing; level out of range (lower)",
					nil,
					func(datum *cgm.LowLevelAlert, units *string) { datum.Level = pointer.FromFloat64(3.27493) },
				),
				Entry("units missing; level in range (lower)",
					nil,
					func(datum *cgm.LowLevelAlert, units *string) { datum.Level = pointer.FromFloat64(3.27494) },
				),
				Entry("units missing; level in range (upper)",
					nil,
					func(datum *cgm.LowLevelAlert, units *string) { datum.Level = pointer.FromFloat64(100) },
				),
				Entry("units missing; level out of range (upper)",
					nil,
					func(datum *cgm.LowLevelAlert, units *string) { datum.Level = pointer.FromFloat64(101) },
				),
				Entry("units invalid; level missing",
					pointer.FromString("invalid"),
					func(datum *cgm.LowLevelAlert, units *string) { datum.Level = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/level"),
				),
				Entry("units invalid; level out of range (lower)",
					pointer.FromString("invalid"),
					func(datum *cgm.LowLevelAlert, units *string) { datum.Level = pointer.FromFloat64(3.27493) },
				),
				Entry("units invalid; level in range (lower)",
					pointer.FromString("invalid"),
					func(datum *cgm.LowLevelAlert, units *string) { datum.Level = pointer.FromFloat64(3.27494) },
				),
				Entry("units invalid; level in range (upper)",
					pointer.FromString("invalid"),
					func(datum *cgm.LowLevelAlert, units *string) { datum.Level = pointer.FromFloat64(100) },
				),
				Entry("units invalid; level out of range (upper)",
					pointer.FromString("invalid"),
					func(datum *cgm.LowLevelAlert, units *string) { datum.Level = pointer.FromFloat64(101) },
				),
				Entry("units mmol/L; level missing",
					pointer.FromString("mmol/L"),
					func(datum *cgm.LowLevelAlert, units *string) { datum.Level = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/level"),
				),
				Entry("units mmol/L; level out of range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *cgm.LowLevelAlert, units *string) { datum.Level = pointer.FromFloat64(3.27493) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(3.27493, 3.27494, 5.55075), "/level"),
				),
				Entry("units mmol/L; level in range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *cgm.LowLevelAlert, units *string) { datum.Level = pointer.FromFloat64(3.27494) },
				),
				Entry("units mmol/L; level in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *cgm.LowLevelAlert, units *string) { datum.Level = pointer.FromFloat64(5.55075) },
				),
				Entry("units mmol/L; level out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *cgm.LowLevelAlert, units *string) { datum.Level = pointer.FromFloat64(5.55076) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(5.55076, 3.27494, 5.55075), "/level"),
				),
				Entry("units mmol/l; level missing",
					pointer.FromString("mmol/l"),
					func(datum *cgm.LowLevelAlert, units *string) { datum.Level = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/level"),
				),
				Entry("units mmol/l; level out of range (lower)",
					pointer.FromString("mmol/l"),
					func(datum *cgm.LowLevelAlert, units *string) { datum.Level = pointer.FromFloat64(3.27493) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(3.27493, 3.27494, 5.55075), "/level"),
				),
				Entry("units mmol/l; level in range (lower)",
					pointer.FromString("mmol/l"),
					func(datum *cgm.LowLevelAlert, units *string) { datum.Level = pointer.FromFloat64(3.27494) },
				),
				Entry("units mmol/l; level in range (upper)",
					pointer.FromString("mmol/l"),
					func(datum *cgm.LowLevelAlert, units *string) { datum.Level = pointer.FromFloat64(5.55075) },
				),
				Entry("units mmol/l; level out of range (upper)",
					pointer.FromString("mmol/l"),
					func(datum *cgm.LowLevelAlert, units *string) { datum.Level = pointer.FromFloat64(5.55076) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(5.55076, 3.27494, 5.55075), "/level"),
				),
				Entry("units mg/dL; level missing",
					pointer.FromString("mg/dL"),
					func(datum *cgm.LowLevelAlert, units *string) { datum.Level = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/level"),
				),
				Entry("units mg/dL; level out of range (lower)",
					pointer.FromString("mg/dL"),
					func(datum *cgm.LowLevelAlert, units *string) { datum.Level = pointer.FromFloat64(58) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(58, 59, 100), "/level"),
				),
				Entry("units mg/dL; level in range (lower)",
					pointer.FromString("mg/dL"),
					func(datum *cgm.LowLevelAlert, units *string) { datum.Level = pointer.FromFloat64(59) },
				),
				Entry("units mg/dL; level in range (upper)",
					pointer.FromString("mg/dL"),
					func(datum *cgm.LowLevelAlert, units *string) { datum.Level = pointer.FromFloat64(100) },
				),
				Entry("units mg/dL; level out of range (upper)",
					pointer.FromString("mg/dL"),
					func(datum *cgm.LowLevelAlert, units *string) { datum.Level = pointer.FromFloat64(101) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(101, 59, 100), "/level"),
				),
				Entry("units mg/dl; level missing",
					pointer.FromString("mg/dl"),
					func(datum *cgm.LowLevelAlert, units *string) { datum.Level = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/level"),
				),
				Entry("units mg/dl; level out of range (lower)",
					pointer.FromString("mg/dl"),
					func(datum *cgm.LowLevelAlert, units *string) { datum.Level = pointer.FromFloat64(58) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(58, 59, 100), "/level"),
				),
				Entry("units mg/dl; level in range (lower)",
					pointer.FromString("mg/dl"),
					func(datum *cgm.LowLevelAlert, units *string) { datum.Level = pointer.FromFloat64(59) },
				),
				Entry("units mg/dl; level in range (upper)",
					pointer.FromString("mg/dl"),
					func(datum *cgm.LowLevelAlert, units *string) { datum.Level = pointer.FromFloat64(100) },
				),
				Entry("units mg/dl; level out of range (upper)",
					pointer.FromString("mg/dl"),
					func(datum *cgm.LowLevelAlert, units *string) { datum.Level = pointer.FromFloat64(101) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(101, 59, 100), "/level"),
				),
				Entry("snooze missing",
					nil,
					func(datum *cgm.LowLevelAlert, units *string) { datum.Snooze = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/snooze"),
				),
				Entry("snooze invalid",
					nil,
					func(datum *cgm.LowLevelAlert, units *string) { datum.Snooze = pointer.FromInt(1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueIntNotOneOf(1, cgm.LevelAlertSnoozes()), "/snooze"),
				),
				Entry("snooze valid",
					nil,
					func(datum *cgm.LowLevelAlert, units *string) { datum.Snooze = pointer.FromInt(0) },
				),
				Entry("multiple errors",
					pointer.FromString("mmol/L"),
					func(datum *cgm.LowLevelAlert, units *string) {
						datum.Enabled = nil
						datum.Level = nil
						datum.Snooze = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/level"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/snooze"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(units *string, mutator func(datum *cgm.LowLevelAlert, units *string), expectator func(datum *cgm.LowLevelAlert, expectedDatum *cgm.LowLevelAlert, units *string)) {
					for _, origin := range structure.Origins() {
						datum := NewLowLevelAlert(units)
						mutator(datum, units)
						expectedDatum := CloneLowLevelAlert(datum)
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
					func(datum *cgm.LowLevelAlert, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units missing",
					nil,
					func(datum *cgm.LowLevelAlert, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units invalid",
					pointer.FromString("invalid"),
					func(datum *cgm.LowLevelAlert, units *string) {},
					nil,
				),
				Entry("does not modify the datum; enabled missing",
					pointer.FromString("mmol/L"),
					func(datum *cgm.LowLevelAlert, units *string) { datum.Enabled = nil },
					nil,
				),
				Entry("does not modify the datum; snooze missing",
					pointer.FromString("mmol/L"),
					func(datum *cgm.LowLevelAlert, units *string) { datum.Snooze = nil },
					nil,
				),
			)

			DescribeTable("normalizes the datum with origin external",
				func(units *string, mutator func(datum *cgm.LowLevelAlert, units *string), expectator func(datum *cgm.LowLevelAlert, expectedDatum *cgm.LowLevelAlert, units *string)) {
					datum := NewLowLevelAlert(units)
					mutator(datum, units)
					expectedDatum := CloneLowLevelAlert(datum)
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
					func(datum *cgm.LowLevelAlert, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *cgm.LowLevelAlert, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *cgm.LowLevelAlert, units *string) {},
					func(datum *cgm.LowLevelAlert, expectedDatum *cgm.LowLevelAlert, units *string) {
						testDataBloodGlucose.ExpectNormalizedValue(datum.Level, expectedDatum.Level, units)
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *cgm.LowLevelAlert, units *string) {},
					func(datum *cgm.LowLevelAlert, expectedDatum *cgm.LowLevelAlert, units *string) {
						testDataBloodGlucose.ExpectNormalizedValue(datum.Level, expectedDatum.Level, units)
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(units *string, mutator func(datum *cgm.LowLevelAlert, units *string), expectator func(datum *cgm.LowLevelAlert, expectedDatum *cgm.LowLevelAlert, units *string)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := NewLowLevelAlert(units)
						mutator(datum, units)
						expectedDatum := CloneLowLevelAlert(datum)
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
					func(datum *cgm.LowLevelAlert, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *cgm.LowLevelAlert, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *cgm.LowLevelAlert, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *cgm.LowLevelAlert, units *string) {},
					nil,
				),
			)
		})
	})
})
