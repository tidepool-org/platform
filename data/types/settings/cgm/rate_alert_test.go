package cgm_test

import (
	"math"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	dataTest "github.com/tidepool-org/platform/data/test"
	dataTypesSettingsCgm "github.com/tidepool-org/platform/data/types/settings/cgm"
	dataTypesSettingsCgmTest "github.com/tidepool-org/platform/data/types/settings/cgm/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("RateAlert", func() {
	It("RateAlertUnitsMgdLMinute is expected", func() {
		Expect(dataTypesSettingsCgm.RateAlertUnitsMgdLMinute).To(Equal("mg/dL/minute"))
	})

	It("RateAlertUnitsMmolLMinute is expected", func() {
		Expect(dataTypesSettingsCgm.RateAlertUnitsMmolLMinute).To(Equal("mmol/L/minute"))
	})

	It("FallAlertRateMgdLMinuteMaximum is expected", func() {
		Expect(dataTypesSettingsCgm.FallAlertRateMgdLMinuteMaximum).To(Equal(10.0))
	})

	It("FallAlertRateMgdLMinuteMinimum is expected", func() {
		Expect(dataTypesSettingsCgm.FallAlertRateMgdLMinuteMinimum).To(Equal(1.0))
	})

	It("FallAlertRateMmolLMinuteMaximum is expected", func() {
		Expect(dataTypesSettingsCgm.FallAlertRateMmolLMinuteMaximum).To(Equal(0.55507))
	})

	It("FallAlertRateMmolLMinuteMinimum is expected", func() {
		Expect(dataTypesSettingsCgm.FallAlertRateMmolLMinuteMinimum).To(Equal(0.05551))
	})

	It("RiseAlertRateMgdLMinuteMaximum is expected", func() {
		Expect(dataTypesSettingsCgm.RiseAlertRateMgdLMinuteMaximum).To(Equal(10.0))
	})

	It("RiseAlertRateMgdLMinuteMinimum is expected", func() {
		Expect(dataTypesSettingsCgm.RiseAlertRateMgdLMinuteMinimum).To(Equal(1.0))
	})

	It("RiseAlertRateMmolLMinuteMaximum is expected", func() {
		Expect(dataTypesSettingsCgm.RiseAlertRateMmolLMinuteMaximum).To(Equal(0.55507))
	})

	It("RiseAlertRateMmolLMinuteMinimum is expected", func() {
		Expect(dataTypesSettingsCgm.RiseAlertRateMmolLMinuteMinimum).To(Equal(0.05551))
	})

	It("RateAlertUnits returns expected", func() {
		Expect(dataTypesSettingsCgm.RateAlertUnits()).To(Equal([]string{"mg/dL/minute", "mmol/L/minute"}))
	})

	Context("RateAlert", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesSettingsCgm.RateAlert)) {
				datum := dataTypesSettingsCgmTest.RandomRateAlert()
				mutator(datum)
				test.ExpectSerializedObjectBSON(datum, dataTypesSettingsCgmTest.NewObjectFromRateAlert(datum, test.ObjectFormatBSON))
				test.ExpectSerializedObjectJSON(datum, dataTypesSettingsCgmTest.NewObjectFromRateAlert(datum, test.ObjectFormatJSON))
			},
			Entry("succeeds",
				func(datum *dataTypesSettingsCgm.RateAlert) {},
			),
			Entry("empty",
				func(datum *dataTypesSettingsCgm.RateAlert) { *datum = dataTypesSettingsCgm.RateAlert{} },
			),
		)

		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesSettingsCgm.RateAlert), expectedErrors ...error) {
					datum := dataTypesSettingsCgmTest.RandomRateAlert()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesSettingsCgm.RateAlert) {},
				),
				Entry("enabled missing",
					func(datum *dataTypesSettingsCgm.RateAlert) { datum.Enabled = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
				),
				Entry("enabled false",
					func(datum *dataTypesSettingsCgm.RateAlert) { datum.Enabled = pointer.FromBool(false) },
				),
				Entry("enabled true",
					func(datum *dataTypesSettingsCgm.RateAlert) { datum.Enabled = pointer.FromBool(true) },
				),
				Entry("snooze missing",
					func(datum *dataTypesSettingsCgm.RateAlert) { datum.Snooze = nil },
				),
				Entry("snooze invalid",
					func(datum *dataTypesSettingsCgm.RateAlert) { datum.Snooze.Duration = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/snooze/duration"),
				),
				Entry("snooze valid",
					func(datum *dataTypesSettingsCgm.RateAlert) { datum.Snooze = dataTypesSettingsCgmTest.RandomSnooze() },
				),
				Entry("units missing; rate missing",
					func(datum *dataTypesSettingsCgm.RateAlert) {
						datum.Units = nil
						datum.Rate = nil
					},
				),
				Entry("units missing; rate exists",
					func(datum *dataTypesSettingsCgm.RateAlert) {
						datum.Units = nil
						datum.Rate = pointer.FromFloat64(test.RandomFloat64())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid; rate missing",
					func(datum *dataTypesSettingsCgm.RateAlert) {
						datum.Units = pointer.FromString("invalid")
						datum.Rate = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/units"),
				),
				Entry("units invalid; rate exists",
					func(datum *dataTypesSettingsCgm.RateAlert) {
						datum.Units = pointer.FromString("invalid")
						datum.Rate = pointer.FromFloat64(test.RandomFloat64())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mg/dL/minute", "mmol/L/minute"}), "/units"),
				),
				Entry("units mg/dL/minute; rate missing",
					func(datum *dataTypesSettingsCgm.RateAlert) {
						datum.Units = pointer.FromString("mg/dL/minute")
						datum.Rate = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/units"),
				),
				Entry("units mg/dL/minute; rate exists",
					func(datum *dataTypesSettingsCgm.RateAlert) {
						datum.Units = pointer.FromString("mg/dL/minute")
						datum.Rate = pointer.FromFloat64(test.RandomFloat64())
					},
				),
				Entry("units mmol/L/minute; rate missing",
					func(datum *dataTypesSettingsCgm.RateAlert) {
						datum.Units = pointer.FromString("mmol/L/minute")
						datum.Rate = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/units"),
				),
				Entry("units mmol/L/minute; rate exists",
					func(datum *dataTypesSettingsCgm.RateAlert) {
						datum.Units = pointer.FromString("mmol/L/minute")
						datum.Rate = pointer.FromFloat64(test.RandomFloat64())
					},
				),
				Entry("multiple errors",
					func(datum *dataTypesSettingsCgm.RateAlert) {
						datum.Enabled = nil
						datum.Snooze.Duration = nil
						datum.Units = nil
						datum.Rate = pointer.FromFloat64(test.RandomFloat64())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/snooze/duration"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
			)
		})
	})

	Context("ParseFallAlert", func() {
		// TODO
	})

	Context("NewFallAlert", func() {
		It("returns successfully with default values", func() {
			datum := dataTypesSettingsCgm.NewFallAlert()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Enabled).To(BeNil())
			Expect(datum.Snooze).To(BeNil())
			Expect(datum.Rate).To(BeNil())
			Expect(datum.Units).To(BeNil())
		})
	})

	Context("FallAlert", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesSettingsCgm.FallAlert)) {
				datum := dataTypesSettingsCgmTest.RandomFallAlert()
				mutator(datum)
				test.ExpectSerializedObjectBSON(datum, dataTypesSettingsCgmTest.NewObjectFromFallAlert(datum, test.ObjectFormatBSON))
				test.ExpectSerializedObjectJSON(datum, dataTypesSettingsCgmTest.NewObjectFromFallAlert(datum, test.ObjectFormatJSON))
				dataTest.ExpectSerializedObject(datum, dataTypesSettingsCgmTest.NewObjectFromFallAlert(datum, test.ObjectFormatJSON),
					func(parser data.ObjectParser) interface{} { return dataTypesSettingsCgm.ParseFallAlert(parser) })
			},
			Entry("succeeds",
				func(datum *dataTypesSettingsCgm.FallAlert) {},
			),
			Entry("empty",
				func(datum *dataTypesSettingsCgm.FallAlert) { *datum = dataTypesSettingsCgm.FallAlert{} },
			),
		)

		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesSettingsCgm.FallAlert), expectedErrors ...error) {
					datum := dataTypesSettingsCgmTest.RandomFallAlert()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesSettingsCgm.FallAlert) {},
				),
				Entry("enabled missing",
					func(datum *dataTypesSettingsCgm.FallAlert) { datum.Enabled = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
				),
				Entry("enabled false",
					func(datum *dataTypesSettingsCgm.FallAlert) { datum.Enabled = pointer.FromBool(false) },
				),
				Entry("enabled true",
					func(datum *dataTypesSettingsCgm.FallAlert) { datum.Enabled = pointer.FromBool(true) },
				),
				Entry("snooze missing",
					func(datum *dataTypesSettingsCgm.FallAlert) { datum.Snooze = nil },
				),
				Entry("snooze invalid",
					func(datum *dataTypesSettingsCgm.FallAlert) { datum.Snooze.Duration = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/snooze/duration"),
				),
				Entry("snooze valid",
					func(datum *dataTypesSettingsCgm.FallAlert) { datum.Snooze = dataTypesSettingsCgmTest.RandomSnooze() },
				),
				Entry("units missing; rate missing",
					func(datum *dataTypesSettingsCgm.FallAlert) {
						datum.Units = nil
						datum.Rate = nil
					},
				),
				Entry("units missing; rate exists",
					func(datum *dataTypesSettingsCgm.FallAlert) {
						datum.Units = nil
						datum.Rate = pointer.FromFloat64(test.RandomFloat64())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid; rate missing",
					func(datum *dataTypesSettingsCgm.FallAlert) {
						datum.Units = pointer.FromString("invalid")
						datum.Rate = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/units"),
				),
				Entry("units invalid; rate exists",
					func(datum *dataTypesSettingsCgm.FallAlert) {
						datum.Units = pointer.FromString("invalid")
						datum.Rate = pointer.FromFloat64(test.RandomFloat64())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mg/dL/minute", "mmol/L/minute"}), "/units"),
				),
				Entry("units mg/dL/minute; rate missing",
					func(datum *dataTypesSettingsCgm.FallAlert) {
						datum.Units = pointer.FromString("mg/dL/minute")
						datum.Rate = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/units"),
				),
				Entry("units mg/dL/minute; rate out of range (lower)",
					func(datum *dataTypesSettingsCgm.FallAlert) {
						datum.Units = pointer.FromString("mg/dL/minute")
						datum.Rate = pointer.FromFloat64(0.9)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(0.9, 1.0, 10.0), "/rate"),
				),
				Entry("units mg/dL/minute; rate in range (lower)",
					func(datum *dataTypesSettingsCgm.FallAlert) {
						datum.Units = pointer.FromString("mg/dL/minute")
						datum.Rate = pointer.FromFloat64(1.0)
					},
				),
				Entry("units mg/dL/minute; rate in range (upper)",
					func(datum *dataTypesSettingsCgm.FallAlert) {
						datum.Units = pointer.FromString("mg/dL/minute")
						datum.Rate = pointer.FromFloat64(10.0)
					},
				),
				Entry("units mg/dL/minute; rate out of range (upper)",
					func(datum *dataTypesSettingsCgm.FallAlert) {
						datum.Units = pointer.FromString("mg/dL/minute")
						datum.Rate = pointer.FromFloat64(10.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(10.1, 1.0, 10.0), "/rate"),
				),
				Entry("units mmol/L/minute; rate missing",
					func(datum *dataTypesSettingsCgm.FallAlert) {
						datum.Units = pointer.FromString("mmol/L/minute")
						datum.Rate = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/units"),
				),
				Entry("units mmol/L/minute; rate out of range (lower)",
					func(datum *dataTypesSettingsCgm.FallAlert) {
						datum.Units = pointer.FromString("mmol/L/minute")
						datum.Rate = pointer.FromFloat64(0.05550)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(0.05550, 0.05551, 0.55507), "/rate"),
				),
				Entry("units mmol/L/minute; rate in range (lower)",
					func(datum *dataTypesSettingsCgm.FallAlert) {
						datum.Units = pointer.FromString("mmol/L/minute")
						datum.Rate = pointer.FromFloat64(0.05551)
					},
				),
				Entry("units mmol/L/minute; rate in range (upper)",
					func(datum *dataTypesSettingsCgm.FallAlert) {
						datum.Units = pointer.FromString("mmol/L/minute")
						datum.Rate = pointer.FromFloat64(0.55507)
					},
				),
				Entry("units mmol/L/minute; rate out of range (upper)",
					func(datum *dataTypesSettingsCgm.FallAlert) {
						datum.Units = pointer.FromString("mmol/L/minute")
						datum.Rate = pointer.FromFloat64(0.55508)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(0.55508, 0.05551, 0.55507), "/rate"),
				),
				Entry("multiple errors",
					func(datum *dataTypesSettingsCgm.FallAlert) {
						datum.Enabled = nil
						datum.Snooze.Duration = nil
						datum.Units = nil
						datum.Rate = pointer.FromFloat64(test.RandomFloat64())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/snooze/duration"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
			)
		})
	})

	Context("FallAlertRateRangeForUnits", func() {
		It("returns expected range for units missing", func() {
			minimum, maximum := dataTypesSettingsCgm.FallAlertRateRangeForUnits(nil)
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units invalid", func() {
			minimum, maximum := dataTypesSettingsCgm.FallAlertRateRangeForUnits(pointer.FromString("invalid"))
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units mg/dL/minute", func() {
			minimum, maximum := dataTypesSettingsCgm.FallAlertRateRangeForUnits(pointer.FromString("mg/dL/minute"))
			Expect(minimum).To(Equal(1.0))
			Expect(maximum).To(Equal(10.0))
		})

		It("returns expected range for units mmol/L/minute", func() {
			minimum, maximum := dataTypesSettingsCgm.FallAlertRateRangeForUnits(pointer.FromString("mmol/L/minute"))
			Expect(minimum).To(Equal(0.05551))
			Expect(maximum).To(Equal(0.55507))
		})
	})

	Context("ParseRiseAlert", func() {
		// TODO
	})

	Context("NewRiseAlert", func() {
		It("returns successfully with default values", func() {
			datum := dataTypesSettingsCgm.NewRiseAlert()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Enabled).To(BeNil())
			Expect(datum.Snooze).To(BeNil())
			Expect(datum.Rate).To(BeNil())
			Expect(datum.Units).To(BeNil())
		})
	})

	Context("RiseAlert", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesSettingsCgm.RiseAlert)) {
				datum := dataTypesSettingsCgmTest.RandomRiseAlert()
				mutator(datum)
				test.ExpectSerializedObjectBSON(datum, dataTypesSettingsCgmTest.NewObjectFromRiseAlert(datum, test.ObjectFormatBSON))
				test.ExpectSerializedObjectJSON(datum, dataTypesSettingsCgmTest.NewObjectFromRiseAlert(datum, test.ObjectFormatJSON))
				dataTest.ExpectSerializedObject(datum, dataTypesSettingsCgmTest.NewObjectFromRiseAlert(datum, test.ObjectFormatJSON),
					func(parser data.ObjectParser) interface{} { return dataTypesSettingsCgm.ParseRiseAlert(parser) })
			},
			Entry("succeeds",
				func(datum *dataTypesSettingsCgm.RiseAlert) {},
			),
			Entry("empty",
				func(datum *dataTypesSettingsCgm.RiseAlert) { *datum = dataTypesSettingsCgm.RiseAlert{} },
			),
		)

		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesSettingsCgm.RiseAlert), expectedErrors ...error) {
					datum := dataTypesSettingsCgmTest.RandomRiseAlert()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesSettingsCgm.RiseAlert) {},
				),
				Entry("enabled missing",
					func(datum *dataTypesSettingsCgm.RiseAlert) { datum.Enabled = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
				),
				Entry("enabled false",
					func(datum *dataTypesSettingsCgm.RiseAlert) { datum.Enabled = pointer.FromBool(false) },
				),
				Entry("enabled true",
					func(datum *dataTypesSettingsCgm.RiseAlert) { datum.Enabled = pointer.FromBool(true) },
				),
				Entry("snooze missing",
					func(datum *dataTypesSettingsCgm.RiseAlert) { datum.Snooze = nil },
				),
				Entry("snooze invalid",
					func(datum *dataTypesSettingsCgm.RiseAlert) { datum.Snooze.Duration = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/snooze/duration"),
				),
				Entry("snooze valid",
					func(datum *dataTypesSettingsCgm.RiseAlert) { datum.Snooze = dataTypesSettingsCgmTest.RandomSnooze() },
				),
				Entry("units missing; rate missing",
					func(datum *dataTypesSettingsCgm.RiseAlert) {
						datum.Units = nil
						datum.Rate = nil
					},
				),
				Entry("units missing; rate exists",
					func(datum *dataTypesSettingsCgm.RiseAlert) {
						datum.Units = nil
						datum.Rate = pointer.FromFloat64(test.RandomFloat64())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid; rate missing",
					func(datum *dataTypesSettingsCgm.RiseAlert) {
						datum.Units = pointer.FromString("invalid")
						datum.Rate = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/units"),
				),
				Entry("units invalid; rate exists",
					func(datum *dataTypesSettingsCgm.RiseAlert) {
						datum.Units = pointer.FromString("invalid")
						datum.Rate = pointer.FromFloat64(test.RandomFloat64())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mg/dL/minute", "mmol/L/minute"}), "/units"),
				),
				Entry("units mg/dL/minute; rate missing",
					func(datum *dataTypesSettingsCgm.RiseAlert) {
						datum.Units = pointer.FromString("mg/dL/minute")
						datum.Rate = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/units"),
				),
				Entry("units mg/dL/minute; rate out of range (lower)",
					func(datum *dataTypesSettingsCgm.RiseAlert) {
						datum.Units = pointer.FromString("mg/dL/minute")
						datum.Rate = pointer.FromFloat64(0.9)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(0.9, 1.0, 10.0), "/rate"),
				),
				Entry("units mg/dL/minute; rate in range (lower)",
					func(datum *dataTypesSettingsCgm.RiseAlert) {
						datum.Units = pointer.FromString("mg/dL/minute")
						datum.Rate = pointer.FromFloat64(1.0)
					},
				),
				Entry("units mg/dL/minute; rate in range (upper)",
					func(datum *dataTypesSettingsCgm.RiseAlert) {
						datum.Units = pointer.FromString("mg/dL/minute")
						datum.Rate = pointer.FromFloat64(10.0)
					},
				),
				Entry("units mg/dL/minute; rate out of range (upper)",
					func(datum *dataTypesSettingsCgm.RiseAlert) {
						datum.Units = pointer.FromString("mg/dL/minute")
						datum.Rate = pointer.FromFloat64(10.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(10.1, 1.0, 10.0), "/rate"),
				),
				Entry("units mmol/L/minute; rate missing",
					func(datum *dataTypesSettingsCgm.RiseAlert) {
						datum.Units = pointer.FromString("mmol/L/minute")
						datum.Rate = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/units"),
				),
				Entry("units mmol/L/minute; rate out of range (lower)",
					func(datum *dataTypesSettingsCgm.RiseAlert) {
						datum.Units = pointer.FromString("mmol/L/minute")
						datum.Rate = pointer.FromFloat64(0.05550)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(0.05550, 0.05551, 0.55507), "/rate"),
				),
				Entry("units mmol/L/minute; rate in range (lower)",
					func(datum *dataTypesSettingsCgm.RiseAlert) {
						datum.Units = pointer.FromString("mmol/L/minute")
						datum.Rate = pointer.FromFloat64(0.05551)
					},
				),
				Entry("units mmol/L/minute; rate in range (upper)",
					func(datum *dataTypesSettingsCgm.RiseAlert) {
						datum.Units = pointer.FromString("mmol/L/minute")
						datum.Rate = pointer.FromFloat64(0.55507)
					},
				),
				Entry("units mmol/L/minute; rate out of range (upper)",
					func(datum *dataTypesSettingsCgm.RiseAlert) {
						datum.Units = pointer.FromString("mmol/L/minute")
						datum.Rate = pointer.FromFloat64(0.55508)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(0.55508, 0.05551, 0.55507), "/rate"),
				),
				Entry("multiple errors",
					func(datum *dataTypesSettingsCgm.RiseAlert) {
						datum.Enabled = nil
						datum.Snooze.Duration = nil
						datum.Units = nil
						datum.Rate = pointer.FromFloat64(test.RandomFloat64())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/snooze/duration"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
			)
		})
	})

	Context("RiseAlertRateRangeForUnits", func() {
		It("returns expected range for units missing", func() {
			minimum, maximum := dataTypesSettingsCgm.RiseAlertRateRangeForUnits(nil)
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units invalid", func() {
			minimum, maximum := dataTypesSettingsCgm.RiseAlertRateRangeForUnits(pointer.FromString("invalid"))
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units mg/dL/minute", func() {
			minimum, maximum := dataTypesSettingsCgm.RiseAlertRateRangeForUnits(pointer.FromString("mg/dL/minute"))
			Expect(minimum).To(Equal(1.0))
			Expect(maximum).To(Equal(10.0))
		})

		It("returns expected range for units mmol/L/minute", func() {
			minimum, maximum := dataTypesSettingsCgm.RiseAlertRateRangeForUnits(pointer.FromString("mmol/L/minute"))
			Expect(minimum).To(Equal(0.05551))
			Expect(maximum).To(Equal(0.55507))
		})
	})
})
