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

var _ = Describe("LevelAlert", func() {
	It("LevelAlertUnitsMgdL is expected", func() {
		Expect(dataTypesSettingsCgm.LevelAlertUnitsMgdL).To(Equal("mg/dL"))
	})

	It("LevelAlertUnitsMmolL is expected", func() {
		Expect(dataTypesSettingsCgm.LevelAlertUnitsMmolL).To(Equal("mmol/L"))
	})

	It("HighAlertLevelMgdLMaximum is expected", func() {
		Expect(dataTypesSettingsCgm.HighAlertLevelMgdLMaximum).To(Equal(400.0))
	})

	It("HighAlertLevelMgdLMinimum is expected", func() {
		Expect(dataTypesSettingsCgm.HighAlertLevelMgdLMinimum).To(Equal(100.0))
	})

	It("HighAlertLevelMmolLMaximum is expected", func() {
		Expect(dataTypesSettingsCgm.HighAlertLevelMmolLMaximum).To(Equal(22.20299))
	})

	It("HighAlertLevelMmolLMinimum is expected", func() {
		Expect(dataTypesSettingsCgm.HighAlertLevelMmolLMinimum).To(Equal(5.55075))
	})

	It("LowAlertLevelMgdLMaximum is expected", func() {
		Expect(dataTypesSettingsCgm.LowAlertLevelMgdLMaximum).To(Equal(150.0))
	})

	It("LowAlertLevelMgdLMinimum is expected", func() {
		Expect(dataTypesSettingsCgm.LowAlertLevelMgdLMinimum).To(Equal(50.0))
	})

	It("LowAlertLevelMmolLMaximum is expected", func() {
		Expect(dataTypesSettingsCgm.LowAlertLevelMmolLMaximum).To(Equal(8.32612))
	})

	It("LowAlertLevelMmolLMinimum is expected", func() {
		Expect(dataTypesSettingsCgm.LowAlertLevelMmolLMinimum).To(Equal(2.77537))
	})

	It("UrgentLowAlertLevelMgdLMaximum is expected", func() {
		Expect(dataTypesSettingsCgm.UrgentLowAlertLevelMgdLMaximum).To(Equal(80.0))
	})

	It("UrgentLowAlertLevelMgdLMinimum is expected", func() {
		Expect(dataTypesSettingsCgm.UrgentLowAlertLevelMgdLMinimum).To(Equal(40.0))
	})

	It("UrgentLowAlertLevelMmolLMaximum is expected", func() {
		Expect(dataTypesSettingsCgm.UrgentLowAlertLevelMmolLMaximum).To(Equal(4.44060))
	})

	It("UrgentLowAlertLevelMmolLMinimum is expected", func() {
		Expect(dataTypesSettingsCgm.UrgentLowAlertLevelMmolLMinimum).To(Equal(2.22030))
	})

	It("LevelAlertUnits returns expected", func() {
		Expect(dataTypesSettingsCgm.LevelAlertUnits()).To(Equal([]string{"mg/dL", "mmol/L"}))
	})

	Context("LevelAlert", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesSettingsCgm.LevelAlert)) {
				datum := dataTypesSettingsCgmTest.RandomLevelAlert()
				mutator(datum)
				test.ExpectSerializedBSON(datum, dataTypesSettingsCgmTest.NewObjectFromLevelAlert(datum, test.ObjectFormatBSON))
				test.ExpectSerializedJSON(datum, dataTypesSettingsCgmTest.NewObjectFromLevelAlert(datum, test.ObjectFormatJSON))
			},
			Entry("succeeds",
				func(datum *dataTypesSettingsCgm.LevelAlert) {},
			),
			Entry("empty",
				func(datum *dataTypesSettingsCgm.LevelAlert) { *datum = dataTypesSettingsCgm.LevelAlert{} },
			),
		)

		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesSettingsCgm.LevelAlert), expectedErrors ...error) {
					datum := dataTypesSettingsCgmTest.RandomLevelAlert()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesSettingsCgm.LevelAlert) {},
				),
				Entry("enabled missing",
					func(datum *dataTypesSettingsCgm.LevelAlert) { datum.Enabled = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
				),
				Entry("enabled false",
					func(datum *dataTypesSettingsCgm.LevelAlert) { datum.Enabled = pointer.FromBool(false) },
				),
				Entry("enabled true",
					func(datum *dataTypesSettingsCgm.LevelAlert) { datum.Enabled = pointer.FromBool(true) },
				),
				Entry("snooze missing",
					func(datum *dataTypesSettingsCgm.LevelAlert) { datum.Snooze = nil },
				),
				Entry("snooze invalid",
					func(datum *dataTypesSettingsCgm.LevelAlert) { datum.Snooze.Duration = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/snooze/duration"),
				),
				Entry("snooze valid",
					func(datum *dataTypesSettingsCgm.LevelAlert) { datum.Snooze = dataTypesSettingsCgmTest.RandomSnooze() },
				),
				Entry("units missing; level missing",
					func(datum *dataTypesSettingsCgm.LevelAlert) {
						datum.Units = nil
						datum.Level = nil
					},
				),
				Entry("units missing; level exists",
					func(datum *dataTypesSettingsCgm.LevelAlert) {
						datum.Units = nil
						datum.Level = pointer.FromFloat64(test.RandomFloat64())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid; level missing",
					func(datum *dataTypesSettingsCgm.LevelAlert) {
						datum.Units = pointer.FromString("invalid")
						datum.Level = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/units"),
				),
				Entry("units invalid; level exists",
					func(datum *dataTypesSettingsCgm.LevelAlert) {
						datum.Units = pointer.FromString("invalid")
						datum.Level = pointer.FromFloat64(test.RandomFloat64())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mg/dL", "mmol/L"}), "/units"),
				),
				Entry("units mg/dL; level missing",
					func(datum *dataTypesSettingsCgm.LevelAlert) {
						datum.Units = pointer.FromString("mg/dL")
						datum.Level = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/units"),
				),
				Entry("units mg/dL; level exists",
					func(datum *dataTypesSettingsCgm.LevelAlert) {
						datum.Units = pointer.FromString("mg/dL")
						datum.Level = pointer.FromFloat64(test.RandomFloat64())
					},
				),
				Entry("units mmol/L; level missing",
					func(datum *dataTypesSettingsCgm.LevelAlert) {
						datum.Units = pointer.FromString("mmol/L")
						datum.Level = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/units"),
				),
				Entry("units mmol/L; level exists",
					func(datum *dataTypesSettingsCgm.LevelAlert) {
						datum.Units = pointer.FromString("mmol/L")
						datum.Level = pointer.FromFloat64(test.RandomFloat64())
					},
				),
				Entry("multiple errors",
					func(datum *dataTypesSettingsCgm.LevelAlert) {
						datum.Enabled = nil
						datum.Snooze.Duration = nil
						datum.Units = nil
						datum.Level = pointer.FromFloat64(test.RandomFloat64())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/snooze/duration"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
			)
		})
	})

	Context("ParseHighAlert", func() {
		// TODO
	})

	Context("NewHighAlert", func() {
		It("returns successfully with default values", func() {
			datum := dataTypesSettingsCgm.NewHighAlert()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Enabled).To(BeNil())
			Expect(datum.Snooze).To(BeNil())
			Expect(datum.Level).To(BeNil())
			Expect(datum.Units).To(BeNil())
		})
	})

	Context("HighAlert", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesSettingsCgm.HighAlert)) {
				datum := dataTypesSettingsCgmTest.RandomHighAlert()
				mutator(datum)
				test.ExpectSerializedBSON(datum, dataTypesSettingsCgmTest.NewObjectFromHighAlert(datum, test.ObjectFormatBSON))
				test.ExpectSerializedJSON(datum, dataTypesSettingsCgmTest.NewObjectFromHighAlert(datum, test.ObjectFormatJSON))
				dataTest.ExpectSerializedObject(datum, dataTypesSettingsCgmTest.NewObjectFromHighAlert(datum, test.ObjectFormatJSON),
					func(parser data.ObjectParser) interface{} { return dataTypesSettingsCgm.ParseHighAlert(parser) })
			},
			Entry("succeeds",
				func(datum *dataTypesSettingsCgm.HighAlert) {},
			),
			Entry("empty",
				func(datum *dataTypesSettingsCgm.HighAlert) { *datum = dataTypesSettingsCgm.HighAlert{} },
			),
		)

		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesSettingsCgm.HighAlert), expectedErrors ...error) {
					datum := dataTypesSettingsCgmTest.RandomHighAlert()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesSettingsCgm.HighAlert) {},
				),
				Entry("enabled missing",
					func(datum *dataTypesSettingsCgm.HighAlert) { datum.Enabled = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
				),
				Entry("enabled false",
					func(datum *dataTypesSettingsCgm.HighAlert) { datum.Enabled = pointer.FromBool(false) },
				),
				Entry("enabled true",
					func(datum *dataTypesSettingsCgm.HighAlert) { datum.Enabled = pointer.FromBool(true) },
				),
				Entry("snooze missing",
					func(datum *dataTypesSettingsCgm.HighAlert) { datum.Snooze = nil },
				),
				Entry("snooze invalid",
					func(datum *dataTypesSettingsCgm.HighAlert) { datum.Snooze.Duration = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/snooze/duration"),
				),
				Entry("snooze valid",
					func(datum *dataTypesSettingsCgm.HighAlert) { datum.Snooze = dataTypesSettingsCgmTest.RandomSnooze() },
				),
				Entry("units missing; level missing",
					func(datum *dataTypesSettingsCgm.HighAlert) {
						datum.Units = nil
						datum.Level = nil
					},
				),
				Entry("units missing; level exists",
					func(datum *dataTypesSettingsCgm.HighAlert) {
						datum.Units = nil
						datum.Level = pointer.FromFloat64(test.RandomFloat64())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid; level missing",
					func(datum *dataTypesSettingsCgm.HighAlert) {
						datum.Units = pointer.FromString("invalid")
						datum.Level = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/units"),
				),
				Entry("units invalid; level exists",
					func(datum *dataTypesSettingsCgm.HighAlert) {
						datum.Units = pointer.FromString("invalid")
						datum.Level = pointer.FromFloat64(test.RandomFloat64())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mg/dL", "mmol/L"}), "/units"),
				),
				Entry("units mg/dL; level missing",
					func(datum *dataTypesSettingsCgm.HighAlert) {
						datum.Units = pointer.FromString("mg/dL")
						datum.Level = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/units"),
				),
				Entry("units mg/dL; level out of range (lower)",
					func(datum *dataTypesSettingsCgm.HighAlert) {
						datum.Units = pointer.FromString("mg/dL")
						datum.Level = pointer.FromFloat64(99.9)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(99.9, 100.0, 400.0), "/level"),
				),
				Entry("units mg/dL; level in range (lower)",
					func(datum *dataTypesSettingsCgm.HighAlert) {
						datum.Units = pointer.FromString("mg/dL")
						datum.Level = pointer.FromFloat64(100.0)
					},
				),
				Entry("units mg/dL; level in range (upper)",
					func(datum *dataTypesSettingsCgm.HighAlert) {
						datum.Units = pointer.FromString("mg/dL")
						datum.Level = pointer.FromFloat64(400.0)
					},
				),
				Entry("units mg/dL; level out of range (upper)",
					func(datum *dataTypesSettingsCgm.HighAlert) {
						datum.Units = pointer.FromString("mg/dL")
						datum.Level = pointer.FromFloat64(400.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(400.1, 100.0, 400.0), "/level"),
				),
				Entry("units mmol/L; level missing",
					func(datum *dataTypesSettingsCgm.HighAlert) {
						datum.Units = pointer.FromString("mmol/L")
						datum.Level = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/units"),
				),
				Entry("units mmol/L; level out of range (lower)",
					func(datum *dataTypesSettingsCgm.HighAlert) {
						datum.Units = pointer.FromString("mmol/L")
						datum.Level = pointer.FromFloat64(5.55074)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(5.55074, 5.55075, 22.20299), "/level"),
				),
				Entry("units mmol/L; level in range (lower)",
					func(datum *dataTypesSettingsCgm.HighAlert) {
						datum.Units = pointer.FromString("mmol/L")
						datum.Level = pointer.FromFloat64(5.55075)
					},
				),
				Entry("units mmol/L; level in range (upper)",
					func(datum *dataTypesSettingsCgm.HighAlert) {
						datum.Units = pointer.FromString("mmol/L")
						datum.Level = pointer.FromFloat64(22.20299)
					},
				),
				Entry("units mmol/L; level out of range (upper)",
					func(datum *dataTypesSettingsCgm.HighAlert) {
						datum.Units = pointer.FromString("mmol/L")
						datum.Level = pointer.FromFloat64(22.20300)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(22.20300, 5.55075, 22.20299), "/level"),
				),
				Entry("multiple errors",
					func(datum *dataTypesSettingsCgm.HighAlert) {
						datum.Enabled = nil
						datum.Snooze.Duration = nil
						datum.Units = nil
						datum.Level = pointer.FromFloat64(test.RandomFloat64())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/snooze/duration"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
			)
		})
	})

	Context("HighAlertLevelRangeForUnits", func() {
		It("returns expected range for units missing", func() {
			minimum, maximum := dataTypesSettingsCgm.HighAlertLevelRangeForUnits(nil)
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units invalid", func() {
			minimum, maximum := dataTypesSettingsCgm.HighAlertLevelRangeForUnits(pointer.FromString("invalid"))
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units mg/dL", func() {
			minimum, maximum := dataTypesSettingsCgm.HighAlertLevelRangeForUnits(pointer.FromString("mg/dL"))
			Expect(minimum).To(Equal(100.0))
			Expect(maximum).To(Equal(400.0))
		})

		It("returns expected range for units mmol/L", func() {
			minimum, maximum := dataTypesSettingsCgm.HighAlertLevelRangeForUnits(pointer.FromString("mmol/L"))
			Expect(minimum).To(Equal(5.55075))
			Expect(maximum).To(Equal(22.20299))
		})
	})

	Context("ParseLowAlert", func() {
		// TODO
	})

	Context("NewLowAlert", func() {
		It("returns successfully with default values", func() {
			datum := dataTypesSettingsCgm.NewLowAlert()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Enabled).To(BeNil())
			Expect(datum.Snooze).To(BeNil())
			Expect(datum.Level).To(BeNil())
			Expect(datum.Units).To(BeNil())
		})
	})

	Context("LowAlert", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesSettingsCgm.LowAlert)) {
				datum := dataTypesSettingsCgmTest.RandomLowAlert()
				mutator(datum)
				test.ExpectSerializedBSON(datum, dataTypesSettingsCgmTest.NewObjectFromLowAlert(datum, test.ObjectFormatBSON))
				test.ExpectSerializedJSON(datum, dataTypesSettingsCgmTest.NewObjectFromLowAlert(datum, test.ObjectFormatJSON))
				dataTest.ExpectSerializedObject(datum, dataTypesSettingsCgmTest.NewObjectFromLowAlert(datum, test.ObjectFormatJSON),
					func(parser data.ObjectParser) interface{} { return dataTypesSettingsCgm.ParseLowAlert(parser) })
			},
			Entry("succeeds",
				func(datum *dataTypesSettingsCgm.LowAlert) {},
			),
			Entry("empty",
				func(datum *dataTypesSettingsCgm.LowAlert) { *datum = dataTypesSettingsCgm.LowAlert{} },
			),
		)

		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesSettingsCgm.LowAlert), expectedErrors ...error) {
					datum := dataTypesSettingsCgmTest.RandomLowAlert()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesSettingsCgm.LowAlert) {},
				),
				Entry("enabled missing",
					func(datum *dataTypesSettingsCgm.LowAlert) { datum.Enabled = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
				),
				Entry("enabled false",
					func(datum *dataTypesSettingsCgm.LowAlert) { datum.Enabled = pointer.FromBool(false) },
				),
				Entry("enabled true",
					func(datum *dataTypesSettingsCgm.LowAlert) { datum.Enabled = pointer.FromBool(true) },
				),
				Entry("snooze missing",
					func(datum *dataTypesSettingsCgm.LowAlert) { datum.Snooze = nil },
				),
				Entry("snooze invalid",
					func(datum *dataTypesSettingsCgm.LowAlert) { datum.Snooze.Duration = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/snooze/duration"),
				),
				Entry("snooze valid",
					func(datum *dataTypesSettingsCgm.LowAlert) { datum.Snooze = dataTypesSettingsCgmTest.RandomSnooze() },
				),
				Entry("units missing; level missing",
					func(datum *dataTypesSettingsCgm.LowAlert) {
						datum.Units = nil
						datum.Level = nil
					},
				),
				Entry("units missing; level exists",
					func(datum *dataTypesSettingsCgm.LowAlert) {
						datum.Units = nil
						datum.Level = pointer.FromFloat64(test.RandomFloat64())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid; level missing",
					func(datum *dataTypesSettingsCgm.LowAlert) {
						datum.Units = pointer.FromString("invalid")
						datum.Level = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/units"),
				),
				Entry("units invalid; level exists",
					func(datum *dataTypesSettingsCgm.LowAlert) {
						datum.Units = pointer.FromString("invalid")
						datum.Level = pointer.FromFloat64(test.RandomFloat64())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mg/dL", "mmol/L"}), "/units"),
				),
				Entry("units mg/dL; level missing",
					func(datum *dataTypesSettingsCgm.LowAlert) {
						datum.Units = pointer.FromString("mg/dL")
						datum.Level = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/units"),
				),
				Entry("units mg/dL; level out of range (lower)",
					func(datum *dataTypesSettingsCgm.LowAlert) {
						datum.Units = pointer.FromString("mg/dL")
						datum.Level = pointer.FromFloat64(49.9)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(49.9, 50.0, 150.0), "/level"),
				),
				Entry("units mg/dL; level in range (lower)",
					func(datum *dataTypesSettingsCgm.LowAlert) {
						datum.Units = pointer.FromString("mg/dL")
						datum.Level = pointer.FromFloat64(50.0)
					},
				),
				Entry("units mg/dL; level in range (upper)",
					func(datum *dataTypesSettingsCgm.LowAlert) {
						datum.Units = pointer.FromString("mg/dL")
						datum.Level = pointer.FromFloat64(150.0)
					},
				),
				Entry("units mg/dL; level out of range (upper)",
					func(datum *dataTypesSettingsCgm.LowAlert) {
						datum.Units = pointer.FromString("mg/dL")
						datum.Level = pointer.FromFloat64(150.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(150.1, 50.0, 150.0), "/level"),
				),
				Entry("units mmol/L; level missing",
					func(datum *dataTypesSettingsCgm.LowAlert) {
						datum.Units = pointer.FromString("mmol/L")
						datum.Level = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/units"),
				),
				Entry("units mmol/L; level out of range (lower)",
					func(datum *dataTypesSettingsCgm.LowAlert) {
						datum.Units = pointer.FromString("mmol/L")
						datum.Level = pointer.FromFloat64(2.77536)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(2.77536, 2.77537, 8.32612), "/level"),
				),
				Entry("units mmol/L; level in range (lower)",
					func(datum *dataTypesSettingsCgm.LowAlert) {
						datum.Units = pointer.FromString("mmol/L")
						datum.Level = pointer.FromFloat64(2.77537)
					},
				),
				Entry("units mmol/L; level in range (upper)",
					func(datum *dataTypesSettingsCgm.LowAlert) {
						datum.Units = pointer.FromString("mmol/L")
						datum.Level = pointer.FromFloat64(8.32612)
					},
				),
				Entry("units mmol/L; level out of range (upper)",
					func(datum *dataTypesSettingsCgm.LowAlert) {
						datum.Units = pointer.FromString("mmol/L")
						datum.Level = pointer.FromFloat64(8.32613)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(8.32613, 2.77537, 8.32612), "/level"),
				),
				Entry("multiple errors",
					func(datum *dataTypesSettingsCgm.LowAlert) {
						datum.Enabled = nil
						datum.Snooze.Duration = nil
						datum.Units = nil
						datum.Level = pointer.FromFloat64(test.RandomFloat64())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/snooze/duration"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
			)
		})
	})

	Context("LowAlertLevelRangeForUnits", func() {
		It("returns expected range for units missing", func() {
			minimum, maximum := dataTypesSettingsCgm.LowAlertLevelRangeForUnits(nil)
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units invalid", func() {
			minimum, maximum := dataTypesSettingsCgm.LowAlertLevelRangeForUnits(pointer.FromString("invalid"))
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units mg/dL", func() {
			minimum, maximum := dataTypesSettingsCgm.LowAlertLevelRangeForUnits(pointer.FromString("mg/dL"))
			Expect(minimum).To(Equal(50.0))
			Expect(maximum).To(Equal(150.0))
		})

		It("returns expected range for units mmol/L", func() {
			minimum, maximum := dataTypesSettingsCgm.LowAlertLevelRangeForUnits(pointer.FromString("mmol/L"))
			Expect(minimum).To(Equal(2.77537))
			Expect(maximum).To(Equal(8.32612))
		})
	})

	Context("ParseUrgentLowAlert", func() {
		// TODO
	})

	Context("NewUrgentLowAlert", func() {
		It("returns successfully with default values", func() {
			datum := dataTypesSettingsCgm.NewUrgentLowAlert()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Enabled).To(BeNil())
			Expect(datum.Snooze).To(BeNil())
			Expect(datum.Level).To(BeNil())
			Expect(datum.Units).To(BeNil())
		})
	})

	Context("UrgentLowAlert", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesSettingsCgm.UrgentLowAlert)) {
				datum := dataTypesSettingsCgmTest.RandomUrgentLowAlert()
				mutator(datum)
				test.ExpectSerializedBSON(datum, dataTypesSettingsCgmTest.NewObjectFromUrgentLowAlert(datum, test.ObjectFormatBSON))
				test.ExpectSerializedJSON(datum, dataTypesSettingsCgmTest.NewObjectFromUrgentLowAlert(datum, test.ObjectFormatJSON))
				dataTest.ExpectSerializedObject(datum, dataTypesSettingsCgmTest.NewObjectFromUrgentLowAlert(datum, test.ObjectFormatJSON),
					func(parser data.ObjectParser) interface{} { return dataTypesSettingsCgm.ParseUrgentLowAlert(parser) })
			},
			Entry("succeeds",
				func(datum *dataTypesSettingsCgm.UrgentLowAlert) {},
			),
			Entry("empty",
				func(datum *dataTypesSettingsCgm.UrgentLowAlert) { *datum = dataTypesSettingsCgm.UrgentLowAlert{} },
			),
		)

		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesSettingsCgm.UrgentLowAlert), expectedErrors ...error) {
					datum := dataTypesSettingsCgmTest.RandomUrgentLowAlert()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesSettingsCgm.UrgentLowAlert) {},
				),
				Entry("enabled missing",
					func(datum *dataTypesSettingsCgm.UrgentLowAlert) { datum.Enabled = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
				),
				Entry("enabled false",
					func(datum *dataTypesSettingsCgm.UrgentLowAlert) { datum.Enabled = pointer.FromBool(false) },
				),
				Entry("enabled true",
					func(datum *dataTypesSettingsCgm.UrgentLowAlert) { datum.Enabled = pointer.FromBool(true) },
				),
				Entry("snooze missing",
					func(datum *dataTypesSettingsCgm.UrgentLowAlert) { datum.Snooze = nil },
				),
				Entry("snooze invalid",
					func(datum *dataTypesSettingsCgm.UrgentLowAlert) { datum.Snooze.Duration = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/snooze/duration"),
				),
				Entry("snooze valid",
					func(datum *dataTypesSettingsCgm.UrgentLowAlert) {
						datum.Snooze = dataTypesSettingsCgmTest.RandomSnooze()
					},
				),
				Entry("units missing; level missing",
					func(datum *dataTypesSettingsCgm.UrgentLowAlert) {
						datum.Units = nil
						datum.Level = nil
					},
				),
				Entry("units missing; level exists",
					func(datum *dataTypesSettingsCgm.UrgentLowAlert) {
						datum.Units = nil
						datum.Level = pointer.FromFloat64(test.RandomFloat64())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid; level missing",
					func(datum *dataTypesSettingsCgm.UrgentLowAlert) {
						datum.Units = pointer.FromString("invalid")
						datum.Level = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/units"),
				),
				Entry("units invalid; level exists",
					func(datum *dataTypesSettingsCgm.UrgentLowAlert) {
						datum.Units = pointer.FromString("invalid")
						datum.Level = pointer.FromFloat64(test.RandomFloat64())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mg/dL", "mmol/L"}), "/units"),
				),
				Entry("units mg/dL; level missing",
					func(datum *dataTypesSettingsCgm.UrgentLowAlert) {
						datum.Units = pointer.FromString("mg/dL")
						datum.Level = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/units"),
				),
				Entry("units mg/dL; level out of range (lower)",
					func(datum *dataTypesSettingsCgm.UrgentLowAlert) {
						datum.Units = pointer.FromString("mg/dL")
						datum.Level = pointer.FromFloat64(39.9)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(39.9, 40.0, 80.0), "/level"),
				),
				Entry("units mg/dL; level in range (lower)",
					func(datum *dataTypesSettingsCgm.UrgentLowAlert) {
						datum.Units = pointer.FromString("mg/dL")
						datum.Level = pointer.FromFloat64(40.0)
					},
				),
				Entry("units mg/dL; level in range (upper)",
					func(datum *dataTypesSettingsCgm.UrgentLowAlert) {
						datum.Units = pointer.FromString("mg/dL")
						datum.Level = pointer.FromFloat64(80.0)
					},
				),
				Entry("units mg/dL; level out of range (upper)",
					func(datum *dataTypesSettingsCgm.UrgentLowAlert) {
						datum.Units = pointer.FromString("mg/dL")
						datum.Level = pointer.FromFloat64(80.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(80.1, 40.0, 80.0), "/level"),
				),
				Entry("units mmol/L; level missing",
					func(datum *dataTypesSettingsCgm.UrgentLowAlert) {
						datum.Units = pointer.FromString("mmol/L")
						datum.Level = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/units"),
				),
				Entry("units mmol/L; level out of range (lower)",
					func(datum *dataTypesSettingsCgm.UrgentLowAlert) {
						datum.Units = pointer.FromString("mmol/L")
						datum.Level = pointer.FromFloat64(2.22029)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(2.22029, 2.22030, 4.44060), "/level"),
				),
				Entry("units mmol/L; level in range (lower)",
					func(datum *dataTypesSettingsCgm.UrgentLowAlert) {
						datum.Units = pointer.FromString("mmol/L")
						datum.Level = pointer.FromFloat64(2.22030)
					},
				),
				Entry("units mmol/L; level in range (upper)",
					func(datum *dataTypesSettingsCgm.UrgentLowAlert) {
						datum.Units = pointer.FromString("mmol/L")
						datum.Level = pointer.FromFloat64(4.44060)
					},
				),
				Entry("units mmol/L; level out of range (upper)",
					func(datum *dataTypesSettingsCgm.UrgentLowAlert) {
						datum.Units = pointer.FromString("mmol/L")
						datum.Level = pointer.FromFloat64(4.44061)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(4.44061, 2.22030, 4.44060), "/level"),
				),
				Entry("multiple errors",
					func(datum *dataTypesSettingsCgm.UrgentLowAlert) {
						datum.Enabled = nil
						datum.Snooze.Duration = nil
						datum.Units = nil
						datum.Level = pointer.FromFloat64(test.RandomFloat64())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/snooze/duration"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
			)
		})
	})

	Context("UrgentLowAlertLevelRangeForUnits", func() {
		It("returns expected range for units missing", func() {
			minimum, maximum := dataTypesSettingsCgm.UrgentLowAlertLevelRangeForUnits(nil)
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units invalid", func() {
			minimum, maximum := dataTypesSettingsCgm.UrgentLowAlertLevelRangeForUnits(pointer.FromString("invalid"))
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units mg/dL", func() {
			minimum, maximum := dataTypesSettingsCgm.UrgentLowAlertLevelRangeForUnits(pointer.FromString("mg/dL"))
			Expect(minimum).To(Equal(40.0))
			Expect(maximum).To(Equal(80.0))
		})

		It("returns expected range for units mmol/L", func() {
			minimum, maximum := dataTypesSettingsCgm.UrgentLowAlertLevelRangeForUnits(pointer.FromString("mmol/L"))
			Expect(minimum).To(Equal(2.22030))
			Expect(maximum).To(Equal(4.44060))
		})
	})
})
