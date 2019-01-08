package cgm_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataTypesSettingsCgm "github.com/tidepool-org/platform/data/types/settings/cgm"
	dataTypesSettingsCgmTest "github.com/tidepool-org/platform/data/types/settings/cgm/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Alert", func() {
	Context("ParseAlerts", func() {
		// TODO
	})

	Context("NewAlerts", func() {
		It("returns successfully with default values", func() {
			datum := dataTypesSettingsCgm.NewAlerts()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Enabled).To(BeNil())
			Expect(datum.UrgentLow).To(BeNil())
			Expect(datum.UrgentLowPredicted).To(BeNil())
			Expect(datum.Low).To(BeNil())
			Expect(datum.LowPredicted).To(BeNil())
			Expect(datum.High).To(BeNil())
			Expect(datum.HighPredicted).To(BeNil())
			Expect(datum.Fall).To(BeNil())
			Expect(datum.Rise).To(BeNil())
			Expect(datum.NoData).To(BeNil())
			Expect(datum.OutOfRange).To(BeNil())
		})
	})

	Context("Alerts", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesSettingsCgm.Alerts)) {
				datum := dataTypesSettingsCgmTest.RandomAlerts()
				mutator(datum)
				test.ExpectSerializedObjectBSON(datum, dataTypesSettingsCgmTest.NewObjectFromAlerts(datum, test.ObjectFormatBSON))
				test.ExpectSerializedObjectJSON(datum, dataTypesSettingsCgmTest.NewObjectFromAlerts(datum, test.ObjectFormatJSON))
			},
			Entry("succeeds",
				func(datum *dataTypesSettingsCgm.Alerts) {},
			),
			Entry("empty",
				func(datum *dataTypesSettingsCgm.Alerts) { *datum = dataTypesSettingsCgm.Alerts{} },
			),
		)

		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesSettingsCgm.Alerts), expectedErrors ...error) {
					datum := dataTypesSettingsCgmTest.RandomAlerts()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesSettingsCgm.Alerts) {},
				),
				Entry("enabled missing",
					func(datum *dataTypesSettingsCgm.Alerts) { datum.Enabled = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
				),
				Entry("enabled false",
					func(datum *dataTypesSettingsCgm.Alerts) { datum.Enabled = pointer.FromBool(false) },
				),
				Entry("enabled true",
					func(datum *dataTypesSettingsCgm.Alerts) { datum.Enabled = pointer.FromBool(true) },
				),
				Entry("urgent low missing",
					func(datum *dataTypesSettingsCgm.Alerts) { datum.UrgentLow = nil },
				),
				Entry("urgent low invalid",
					func(datum *dataTypesSettingsCgm.Alerts) { datum.UrgentLow.Enabled = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/urgentLow/enabled"),
				),
				Entry("urgent low valid",
					func(datum *dataTypesSettingsCgm.Alerts) {
						datum.UrgentLow = dataTypesSettingsCgmTest.RandomUrgentLowAlert()
					},
				),
				Entry("urgent low predicted missing",
					func(datum *dataTypesSettingsCgm.Alerts) { datum.UrgentLowPredicted = nil },
				),
				Entry("urgent low predicted invalid",
					func(datum *dataTypesSettingsCgm.Alerts) { datum.UrgentLowPredicted.Enabled = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/urgentLowPredicted/enabled"),
				),
				Entry("urgent low predicted valid",
					func(datum *dataTypesSettingsCgm.Alerts) {
						datum.UrgentLowPredicted = dataTypesSettingsCgmTest.RandomUrgentLowAlert()
					},
				),
				Entry("low missing",
					func(datum *dataTypesSettingsCgm.Alerts) { datum.Low = nil },
				),
				Entry("low invalid",
					func(datum *dataTypesSettingsCgm.Alerts) { datum.Low.Enabled = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/low/enabled"),
				),
				Entry("low valid",
					func(datum *dataTypesSettingsCgm.Alerts) {
						datum.Low = dataTypesSettingsCgmTest.RandomLowAlert()
					},
				),
				Entry("low predicted missing",
					func(datum *dataTypesSettingsCgm.Alerts) { datum.LowPredicted = nil },
				),
				Entry("low predicted invalid",
					func(datum *dataTypesSettingsCgm.Alerts) { datum.LowPredicted.Enabled = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/lowPredicted/enabled"),
				),
				Entry("low predicted valid",
					func(datum *dataTypesSettingsCgm.Alerts) {
						datum.LowPredicted = dataTypesSettingsCgmTest.RandomLowAlert()
					},
				),
				Entry("high missing",
					func(datum *dataTypesSettingsCgm.Alerts) { datum.High = nil },
				),
				Entry("high invalid",
					func(datum *dataTypesSettingsCgm.Alerts) { datum.High.Enabled = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/high/enabled"),
				),
				Entry("high valid",
					func(datum *dataTypesSettingsCgm.Alerts) {
						datum.High = dataTypesSettingsCgmTest.RandomHighAlert()
					},
				),
				Entry("high predicted missing",
					func(datum *dataTypesSettingsCgm.Alerts) { datum.HighPredicted = nil },
				),
				Entry("high predicted invalid",
					func(datum *dataTypesSettingsCgm.Alerts) { datum.HighPredicted.Enabled = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/highPredicted/enabled"),
				),
				Entry("high predicted valid",
					func(datum *dataTypesSettingsCgm.Alerts) {
						datum.HighPredicted = dataTypesSettingsCgmTest.RandomHighAlert()
					},
				),
				Entry("fall missing",
					func(datum *dataTypesSettingsCgm.Alerts) { datum.Fall = nil },
				),
				Entry("fall invalid",
					func(datum *dataTypesSettingsCgm.Alerts) { datum.Fall.Enabled = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/fall/enabled"),
				),
				Entry("fall valid",
					func(datum *dataTypesSettingsCgm.Alerts) {
						datum.Fall = dataTypesSettingsCgmTest.RandomFallAlert()
					},
				),
				Entry("rise missing",
					func(datum *dataTypesSettingsCgm.Alerts) { datum.Rise = nil },
				),
				Entry("rise invalid",
					func(datum *dataTypesSettingsCgm.Alerts) { datum.Rise.Enabled = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rise/enabled"),
				),
				Entry("rise valid",
					func(datum *dataTypesSettingsCgm.Alerts) {
						datum.Rise = dataTypesSettingsCgmTest.RandomRiseAlert()
					},
				),
				Entry("no data missing",
					func(datum *dataTypesSettingsCgm.Alerts) { datum.NoData = nil },
				),
				Entry("no data invalid",
					func(datum *dataTypesSettingsCgm.Alerts) { datum.NoData.Enabled = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/noData/enabled"),
				),
				Entry("no data valid",
					func(datum *dataTypesSettingsCgm.Alerts) {
						datum.NoData = dataTypesSettingsCgmTest.RandomNoDataAlert()
					},
				),
				Entry("out of range missing",
					func(datum *dataTypesSettingsCgm.Alerts) { datum.OutOfRange = nil },
				),
				Entry("out of range invalid",
					func(datum *dataTypesSettingsCgm.Alerts) { datum.OutOfRange.Enabled = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/outOfRange/enabled"),
				),
				Entry("out of range valid",
					func(datum *dataTypesSettingsCgm.Alerts) {
						datum.OutOfRange = dataTypesSettingsCgmTest.RandomOutOfRangeAlert()
					},
				),
				Entry("multiple errors",
					func(datum *dataTypesSettingsCgm.Alerts) {
						datum.Enabled = nil
						datum.UrgentLow.Enabled = nil
						datum.UrgentLowPredicted.Enabled = nil
						datum.Low.Enabled = nil
						datum.LowPredicted.Enabled = nil
						datum.High.Enabled = nil
						datum.HighPredicted.Enabled = nil
						datum.Fall.Enabled = nil
						datum.Rise.Enabled = nil
						datum.NoData.Enabled = nil
						datum.OutOfRange.Enabled = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/urgentLow/enabled"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/urgentLowPredicted/enabled"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/low/enabled"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/lowPredicted/enabled"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/high/enabled"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/highPredicted/enabled"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/fall/enabled"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rise/enabled"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/noData/enabled"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/outOfRange/enabled"),
				),
			)
		})
	})

	Context("Alert", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesSettingsCgm.Alert)) {
				datum := dataTypesSettingsCgmTest.RandomAlert()
				mutator(datum)
				test.ExpectSerializedObjectBSON(datum, dataTypesSettingsCgmTest.NewObjectFromAlert(datum, test.ObjectFormatBSON))
				test.ExpectSerializedObjectJSON(datum, dataTypesSettingsCgmTest.NewObjectFromAlert(datum, test.ObjectFormatJSON))
			},
			Entry("succeeds",
				func(datum *dataTypesSettingsCgm.Alert) {},
			),
			Entry("empty",
				func(datum *dataTypesSettingsCgm.Alert) { *datum = dataTypesSettingsCgm.Alert{} },
			),
		)

		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesSettingsCgm.Alert), expectedErrors ...error) {
					datum := dataTypesSettingsCgmTest.RandomAlert()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesSettingsCgm.Alert) {},
				),
				Entry("enabled missing",
					func(datum *dataTypesSettingsCgm.Alert) { datum.Enabled = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
				),
				Entry("enabled false",
					func(datum *dataTypesSettingsCgm.Alert) { datum.Enabled = pointer.FromBool(false) },
				),
				Entry("enabled true",
					func(datum *dataTypesSettingsCgm.Alert) { datum.Enabled = pointer.FromBool(true) },
				),
				Entry("snooze missing",
					func(datum *dataTypesSettingsCgm.Alert) { datum.Snooze = nil },
				),
				Entry("snooze invalid",
					func(datum *dataTypesSettingsCgm.Alert) { datum.Snooze.Duration = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/snooze/duration"),
				),
				Entry("snooze valid",
					func(datum *dataTypesSettingsCgm.Alert) {
						datum.Snooze = dataTypesSettingsCgmTest.RandomSnooze()
					},
				),
				Entry("multiple errors",
					func(datum *dataTypesSettingsCgm.Alert) {
						datum.Enabled = nil
						datum.Snooze.Duration = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/snooze/duration"),
				),
			)
		})
	})
})
