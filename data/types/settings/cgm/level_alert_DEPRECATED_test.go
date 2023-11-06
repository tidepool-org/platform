package cgm_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataBloodGlucoseTest "github.com/tidepool-org/platform/data/blood/glucose/test"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTypesSettingsCgm "github.com/tidepool-org/platform/data/types/settings/cgm"
	dataTypesSettingsCgmTest "github.com/tidepool-org/platform/data/types/settings/cgm/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("LevelAlertDEPRECATED", func() {
	It("HighLevelAlertDEPRECATEDLevelMgdLMaximum returns expected", func() {
		Expect(dataTypesSettingsCgm.HighLevelAlertDEPRECATEDLevelMgdLMaximum).To(Equal(float64(400)))
	})

	It("HighLevelAlertDEPRECATEDLevelMgdLMinimum returns expected", func() {
		Expect(dataTypesSettingsCgm.HighLevelAlertDEPRECATEDLevelMgdLMinimum).To(Equal(float64(120)))
	})

	It("HighLevelAlertDEPRECATEDLevelMmolLMaximum returns expected", func() {
		Expect(dataTypesSettingsCgm.HighLevelAlertDEPRECATEDLevelMmolLMaximum).To(Equal(22.20299))
	})

	It("HighLevelAlertDEPRECATEDLevelMmolLMinimum returns expected", func() {
		Expect(dataTypesSettingsCgm.HighLevelAlertDEPRECATEDLevelMmolLMinimum).To(Equal(6.66090))
	})

	It("LowLevelAlertDEPRECATEDLevelMgdLMaximum returns expected", func() {
		Expect(dataTypesSettingsCgm.LowLevelAlertDEPRECATEDLevelMgdLMaximum).To(Equal(float64(100)))
	})

	It("LowLevelAlertDEPRECATEDLevelMgdLMinimum returns expected", func() {
		Expect(dataTypesSettingsCgm.LowLevelAlertDEPRECATEDLevelMgdLMinimum).To(Equal(float64(59)))
	})

	It("LowLevelAlertDEPRECATEDLevelMmolLMaximum returns expected", func() {
		Expect(dataTypesSettingsCgm.LowLevelAlertDEPRECATEDLevelMmolLMaximum).To(Equal(5.55075))
	})

	It("LowLevelAlertDEPRECATEDLevelMmolLMinimum returns expected", func() {
		Expect(dataTypesSettingsCgm.LowLevelAlertDEPRECATEDLevelMmolLMinimum).To(Equal(3.27494))
	})

	It("LevelAlertDEPRECATEDSnoozes returns expected", func() {
		Expect(dataTypesSettingsCgm.LevelAlertDEPRECATEDSnoozes()).To(Equal([]int{
			0, 900000, 1200000, 1500000, 1800000, 2100000, 2400000, 2700000,
			3000000, 3300000, 3600000, 3900000, 4200000, 4500000, 4800000, 5100000,
			5400000, 5700000, 6000000, 6300000, 6600000, 6900000, 7200000, 7500000,
			7800000, 8100000, 8400000, 8700000, 9000000, 9300000, 9600000, 9900000,
			10200000, 10500000, 10800000, 11100000, 11400000, 11700000, 12000000,
			12300000, 12600000, 12900000, 13200000, 13500000, 13800000, 14100000,
			14400000, 15300000, 16200000, 17100000, 18000000}))
	})

	Context("ParseHighLevelAlertDEPRECATED", func() {
		// TODO
	})

	Context("NewHighLevelAlertDEPRECATED", func() {
		It("is successful", func() {
			Expect(dataTypesSettingsCgm.NewHighLevelAlertDEPRECATED()).To(Equal(&dataTypesSettingsCgm.HighLevelAlertDEPRECATED{}))
		})
	})

	Context("HighLevelAlertDEPRECATED", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(units *string, mutator func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string), expectedErrors ...error) {
					datum := dataTypesSettingsCgmTest.RandomHighLevelAlertDEPRECATED(units)
					mutator(datum, units)
					dataTypesTest.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithStringAdapter(datum, units), structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {},
				),
				Entry("enabled missing",
					nil,
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) { datum.Enabled = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
				),
				Entry("enabled false",
					nil,
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {
						datum.Enabled = pointer.FromBool(false)
					},
				),
				Entry("enabled true",
					nil,
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {
						datum.Enabled = pointer.FromBool(true)
					},
				),
				Entry("units missing; level missing",
					nil,
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) { datum.Level = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/level"),
				),
				Entry("units missing; level out of range (lower)",
					nil,
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(6.66089)
					},
				),
				Entry("units missing; level in range (lower)",
					nil,
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(6.66090)
					},
				),
				Entry("units missing; level in range (upper)",
					nil,
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(400)
					},
				),
				Entry("units missing; level out of range (upper)",
					nil,
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(401)
					},
				),
				Entry("units invalid; level missing",
					pointer.FromString("invalid"),
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) { datum.Level = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/level"),
				),
				Entry("units invalid; level out of range (lower)",
					pointer.FromString("invalid"),
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(6.66089)
					},
				),
				Entry("units invalid; level in range (lower)",
					pointer.FromString("invalid"),
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(6.66090)
					},
				),
				Entry("units invalid; level in range (upper)",
					pointer.FromString("invalid"),
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(400)
					},
				),
				Entry("units invalid; level out of range (upper)",
					pointer.FromString("invalid"),
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(401)
					},
				),
				Entry("units mmol/L; level missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) { datum.Level = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/level"),
				),
				Entry("units mmol/L; level out of range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(6.66089)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(6.66089, 6.66090, 22.20299), "/level"),
				),
				Entry("units mmol/L; level in range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(6.66090)
					},
				),
				Entry("units mmol/L; level in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(22.20299)
					},
				),
				Entry("units mmol/L; level out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(22.20300)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(22.20300, 6.66090, 22.20299), "/level"),
				),
				Entry("units mmol/l; level missing",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) { datum.Level = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/level"),
				),
				Entry("units mmol/l; level out of range (lower)",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(6.66089)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(6.66089, 6.66090, 22.20299), "/level"),
				),
				Entry("units mmol/l; level in range (lower)",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(6.66090)
					},
				),
				Entry("units mmol/l; level in range (upper)",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(22.20299)
					},
				),
				Entry("units mmol/l; level out of range (upper)",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(22.20300)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(22.20300, 6.66090, 22.20299), "/level"),
				),
				Entry("units mg/dL; level missing",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) { datum.Level = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/level"),
				),
				Entry("units mg/dL; level out of range (lower)",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(119)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(119, 120, 400), "/level"),
				),
				Entry("units mg/dL; level in range (lower)",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(120)
					},
				),
				Entry("units mg/dL; level in range (upper)",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(400)
					},
				),
				Entry("units mg/dL; level out of range (upper)",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(401)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(401, 120, 400), "/level"),
				),
				Entry("units mg/dl; level missing",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) { datum.Level = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/level"),
				),
				Entry("units mg/dl; level out of range (lower)",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(119)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(119, 120, 400), "/level"),
				),
				Entry("units mg/dl; level in range (lower)",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(120)
					},
				),
				Entry("units mg/dl; level in range (upper)",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(400)
					},
				),
				Entry("units mg/dl; level out of range (upper)",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(401)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(401, 120, 400), "/level"),
				),
				Entry("snooze missing",
					nil,
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) { datum.Snooze = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/snooze"),
				),
				Entry("snooze invalid",
					nil,
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {
						datum.Snooze = pointer.FromInt(1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueIntNotOneOf(1, dataTypesSettingsCgm.LevelAlertDEPRECATEDSnoozes()), "/snooze"),
				),
				Entry("snooze valid",
					nil,
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {
						datum.Snooze = pointer.FromInt(0)
					},
				),
				Entry("multiple errors",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {
						datum.Enabled = nil
						datum.Level = nil
						datum.Snooze = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/level"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/snooze"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(units *string, mutator func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string), expectator func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, expectedDatum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesSettingsCgmTest.RandomHighLevelAlertDEPRECATED(units)
						mutator(datum, units)
						expectedDatum := dataTypesSettingsCgmTest.CloneHighLevelAlertDEPRECATED(datum)
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
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units missing",
					nil,
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units invalid",
					pointer.FromString("invalid"),
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; enabled missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) { datum.Enabled = nil },
					nil,
				),
				Entry("does not modify the datum; snooze missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) { datum.Snooze = nil },
					nil,
				),
			)

			DescribeTable("normalizes the datum with origin external",
				func(units *string, mutator func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string), expectator func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, expectedDatum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string)) {
					datum := dataTypesSettingsCgmTest.RandomHighLevelAlertDEPRECATED(units)
					mutator(datum, units)
					expectedDatum := dataTypesSettingsCgmTest.CloneHighLevelAlertDEPRECATED(datum)
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
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {},
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, expectedDatum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {
						dataBloodGlucoseTest.ExpectNormalizedValue(datum.Level, expectedDatum.Level, units)
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {},
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, expectedDatum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {
						dataBloodGlucoseTest.ExpectNormalizedValue(datum.Level, expectedDatum.Level, units)
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(units *string, mutator func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string), expectator func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, expectedDatum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := dataTypesSettingsCgmTest.RandomHighLevelAlertDEPRECATED(units)
						mutator(datum, units)
						expectedDatum := dataTypesSettingsCgmTest.CloneHighLevelAlertDEPRECATED(datum)
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
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED, units *string) {},
					nil,
				),
			)
		})
	})

	Context("ParseLowLevelAlertDEPRECATED", func() {
		// TODO
	})

	Context("NewLowLevelAlertDEPRECATED", func() {
		It("is successful", func() {
			Expect(dataTypesSettingsCgm.NewLowLevelAlertDEPRECATED()).To(Equal(&dataTypesSettingsCgm.LowLevelAlertDEPRECATED{}))
		})
	})

	Context("LowLevelAlertDEPRECATED", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(units *string, mutator func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string), expectedErrors ...error) {
					datum := dataTypesSettingsCgmTest.RandomLowLevelAlertDEPRECATED(units)
					mutator(datum, units)
					dataTypesTest.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithStringAdapter(datum, units), structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {},
				),
				Entry("enabled missing",
					nil,
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) { datum.Enabled = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
				),
				Entry("enabled false",
					nil,
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {
						datum.Enabled = pointer.FromBool(false)
					},
				),
				Entry("enabled true",
					nil,
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {
						datum.Enabled = pointer.FromBool(true)
					},
				),
				Entry("units missing; level missing",
					nil,
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) { datum.Level = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/level"),
				),
				Entry("units missing; level out of range (lower)",
					nil,
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(3.27493)
					},
				),
				Entry("units missing; level in range (lower)",
					nil,
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(3.27494)
					},
				),
				Entry("units missing; level in range (upper)",
					nil,
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(100)
					},
				),
				Entry("units missing; level out of range (upper)",
					nil,
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(101)
					},
				),
				Entry("units invalid; level missing",
					pointer.FromString("invalid"),
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) { datum.Level = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/level"),
				),
				Entry("units invalid; level out of range (lower)",
					pointer.FromString("invalid"),
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(3.27493)
					},
				),
				Entry("units invalid; level in range (lower)",
					pointer.FromString("invalid"),
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(3.27494)
					},
				),
				Entry("units invalid; level in range (upper)",
					pointer.FromString("invalid"),
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(100)
					},
				),
				Entry("units invalid; level out of range (upper)",
					pointer.FromString("invalid"),
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(101)
					},
				),
				Entry("units mmol/L; level missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) { datum.Level = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/level"),
				),
				Entry("units mmol/L; level out of range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(3.27493)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(3.27493, 3.27494, 5.55075), "/level"),
				),
				Entry("units mmol/L; level in range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(3.27494)
					},
				),
				Entry("units mmol/L; level in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(5.55075)
					},
				),
				Entry("units mmol/L; level out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(5.55076)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(5.55076, 3.27494, 5.55075), "/level"),
				),
				Entry("units mmol/l; level missing",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) { datum.Level = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/level"),
				),
				Entry("units mmol/l; level out of range (lower)",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(3.27493)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(3.27493, 3.27494, 5.55075), "/level"),
				),
				Entry("units mmol/l; level in range (lower)",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(3.27494)
					},
				),
				Entry("units mmol/l; level in range (upper)",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(5.55075)
					},
				),
				Entry("units mmol/l; level out of range (upper)",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(5.55076)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(5.55076, 3.27494, 5.55075), "/level"),
				),
				Entry("units mg/dL; level missing",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) { datum.Level = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/level"),
				),
				Entry("units mg/dL; level out of range (lower)",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(58)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(58, 59, 100), "/level"),
				),
				Entry("units mg/dL; level in range (lower)",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(59)
					},
				),
				Entry("units mg/dL; level in range (upper)",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(100)
					},
				),
				Entry("units mg/dL; level out of range (upper)",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(101)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(101, 59, 100), "/level"),
				),
				Entry("units mg/dl; level missing",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) { datum.Level = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/level"),
				),
				Entry("units mg/dl; level out of range (lower)",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(58)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(58, 59, 100), "/level"),
				),
				Entry("units mg/dl; level in range (lower)",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(59)
					},
				),
				Entry("units mg/dl; level in range (upper)",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(100)
					},
				),
				Entry("units mg/dl; level out of range (upper)",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {
						datum.Level = pointer.FromFloat64(101)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(101, 59, 100), "/level"),
				),
				Entry("snooze missing",
					nil,
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) { datum.Snooze = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/snooze"),
				),
				Entry("snooze invalid",
					nil,
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {
						datum.Snooze = pointer.FromInt(1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueIntNotOneOf(1, dataTypesSettingsCgm.LevelAlertDEPRECATEDSnoozes()), "/snooze"),
				),
				Entry("snooze valid",
					nil,
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {
						datum.Snooze = pointer.FromInt(0)
					},
				),
				Entry("multiple errors",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {
						datum.Enabled = nil
						datum.Level = nil
						datum.Snooze = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/level"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/snooze"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(units *string, mutator func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string), expectator func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, expectedDatum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesSettingsCgmTest.RandomLowLevelAlertDEPRECATED(units)
						mutator(datum, units)
						expectedDatum := dataTypesSettingsCgmTest.CloneLowLevelAlertDEPRECATED(datum)
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
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units missing",
					nil,
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units invalid",
					pointer.FromString("invalid"),
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; enabled missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) { datum.Enabled = nil },
					nil,
				),
				Entry("does not modify the datum; snooze missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) { datum.Snooze = nil },
					nil,
				),
			)

			DescribeTable("normalizes the datum with origin external",
				func(units *string, mutator func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string), expectator func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, expectedDatum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string)) {
					datum := dataTypesSettingsCgmTest.RandomLowLevelAlertDEPRECATED(units)
					mutator(datum, units)
					expectedDatum := dataTypesSettingsCgmTest.CloneLowLevelAlertDEPRECATED(datum)
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
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {},
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, expectedDatum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {
						dataBloodGlucoseTest.ExpectNormalizedValue(datum.Level, expectedDatum.Level, units)
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {},
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, expectedDatum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {
						dataBloodGlucoseTest.ExpectNormalizedValue(datum.Level, expectedDatum.Level, units)
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(units *string, mutator func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string), expectator func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, expectedDatum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := dataTypesSettingsCgmTest.RandomLowLevelAlertDEPRECATED(units)
						mutator(datum, units)
						expectedDatum := dataTypesSettingsCgmTest.CloneLowLevelAlertDEPRECATED(datum)
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
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED, units *string) {},
					nil,
				),
			)
		})
	})
})
