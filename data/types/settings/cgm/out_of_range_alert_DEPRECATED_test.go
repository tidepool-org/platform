package cgm_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTypesSettingsCgm "github.com/tidepool-org/platform/data/types/settings/cgm"
	dataTypesSettingsCgmTest "github.com/tidepool-org/platform/data/types/settings/cgm/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("OutOfRangeAlertDEPRECATED", func() {
	It("OutOfRangeAlertDEPRECATEDThresholds returns expected", func() {
		Expect(dataTypesSettingsCgm.OutOfRangeAlertDEPRECATEDThresholds()).To(Equal([]int{
			1200000, 1500000, 1800000, 2100000, 2400000, 2700000, 3000000, 3300000,
			3600000, 3900000, 4200000, 4500000, 4800000, 5100000, 5400000, 5700000,
			6000000, 6300000, 6600000, 6900000, 7200000, 7500000, 7800000, 8100000,
			8400000, 8700000, 9000000, 9300000, 9600000, 9900000, 10200000,
			10500000, 10800000, 11100000, 11400000, 11700000, 12000000, 12300000,
			12600000, 12900000, 13200000, 13500000, 13800000, 14100000, 14400000}))
	})

	Context("ParseOutOfRangeAlertDEPRECATED", func() {
		// TODO
	})

	Context("NewOutOfRangeAlertDEPRECATED", func() {
		It("is successful", func() {
			Expect(dataTypesSettingsCgm.NewOutOfRangeAlertDEPRECATED()).To(Equal(&dataTypesSettingsCgm.OutOfRangeAlertDEPRECATED{}))
		})
	})

	Context("OutOfRangeAlertDEPRECATED", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesSettingsCgm.OutOfRangeAlertDEPRECATED), expectedErrors ...error) {
					datum := dataTypesSettingsCgmTest.RandomOutOfRangeAlertDEPRECATED()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesSettingsCgm.OutOfRangeAlertDEPRECATED) {},
				),
				Entry("enabled missing",
					func(datum *dataTypesSettingsCgm.OutOfRangeAlertDEPRECATED) { datum.Enabled = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
				),
				Entry("enabled false",
					func(datum *dataTypesSettingsCgm.OutOfRangeAlertDEPRECATED) { datum.Enabled = pointer.FromBool(false) },
				),
				Entry("enabled true",
					func(datum *dataTypesSettingsCgm.OutOfRangeAlertDEPRECATED) { datum.Enabled = pointer.FromBool(true) },
				),
				Entry("threshold missing",
					func(datum *dataTypesSettingsCgm.OutOfRangeAlertDEPRECATED) { datum.Threshold = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/snooze"),
				),
				Entry("threshold invalid",
					func(datum *dataTypesSettingsCgm.OutOfRangeAlertDEPRECATED) { datum.Threshold = pointer.FromInt(1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueIntNotOneOf(1, dataTypesSettingsCgm.OutOfRangeAlertDEPRECATEDThresholds()), "/snooze"),
				),
				Entry("threshold valid",
					func(datum *dataTypesSettingsCgm.OutOfRangeAlertDEPRECATED) {
						datum.Threshold = pointer.FromInt(1200000)
					},
				),
				Entry("multiple errors",
					func(datum *dataTypesSettingsCgm.OutOfRangeAlertDEPRECATED) {
						datum.Enabled = nil
						datum.Threshold = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/snooze"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *dataTypesSettingsCgm.OutOfRangeAlertDEPRECATED), expectator func(datum *dataTypesSettingsCgm.OutOfRangeAlertDEPRECATED, expectedDatum *dataTypesSettingsCgm.OutOfRangeAlertDEPRECATED)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesSettingsCgmTest.RandomOutOfRangeAlertDEPRECATED()
						mutator(datum)
						expectedDatum := dataTypesSettingsCgmTest.CloneOutOfRangeAlertDEPRECATED(datum)
						normalizer := dataNormalizer.New(logTest.NewLogger())
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						if expectator != nil {
							expectator(datum, expectedDatum)
						}
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *dataTypesSettingsCgm.OutOfRangeAlertDEPRECATED) {},
					nil,
				),
				Entry("does not modify the datum; enabled missing",
					func(datum *dataTypesSettingsCgm.OutOfRangeAlertDEPRECATED) { datum.Enabled = nil },
					nil,
				),
				Entry("does not modify the datum; threshold missing",
					func(datum *dataTypesSettingsCgm.OutOfRangeAlertDEPRECATED) { datum.Threshold = nil },
					nil,
				),
			)
		})
	})
})
