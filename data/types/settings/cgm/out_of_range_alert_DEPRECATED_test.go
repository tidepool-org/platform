package cgm_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/settings/cgm"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewOutOfRangeAlertDEPRECATED() *cgm.OutOfRangeAlertDEPRECATED {
	datum := cgm.NewOutOfRangeAlertDEPRECATED()
	datum.Enabled = pointer.FromBool(test.RandomBool())
	datum.Threshold = pointer.FromInt(test.RandomIntFromArray(cgm.OutOfRangeAlertDEPRECATEDThresholds()))
	return datum
}

func CloneOutOfRangeAlertDEPRECATED(datum *cgm.OutOfRangeAlertDEPRECATED) *cgm.OutOfRangeAlertDEPRECATED {
	if datum == nil {
		return nil
	}
	clone := cgm.NewOutOfRangeAlertDEPRECATED()
	clone.Enabled = test.CloneBool(datum.Enabled)
	clone.Threshold = test.CloneInt(datum.Threshold)
	return clone
}

var _ = Describe("OutOfRangeAlertDEPRECATED", func() {
	It("OutOfRangeAlertDEPRECATEDThresholds returns expected", func() {
		Expect(cgm.OutOfRangeAlertDEPRECATEDThresholds()).To(Equal([]int{
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
			Expect(cgm.NewOutOfRangeAlertDEPRECATED()).To(Equal(&cgm.OutOfRangeAlertDEPRECATED{}))
		})
	})

	Context("OutOfRangeAlertDEPRECATED", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *cgm.OutOfRangeAlertDEPRECATED), expectedErrors ...error) {
					datum := NewOutOfRangeAlertDEPRECATED()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *cgm.OutOfRangeAlertDEPRECATED) {},
				),
				Entry("enabled missing",
					func(datum *cgm.OutOfRangeAlertDEPRECATED) { datum.Enabled = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
				),
				Entry("enabled true",
					func(datum *cgm.OutOfRangeAlertDEPRECATED) { datum.Enabled = pointer.FromBool(true) },
				),
				Entry("enabled false",
					func(datum *cgm.OutOfRangeAlertDEPRECATED) { datum.Enabled = pointer.FromBool(false) },
				),
				Entry("threshold missing",
					func(datum *cgm.OutOfRangeAlertDEPRECATED) { datum.Threshold = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/snooze"),
				),
				Entry("threshold invalid",
					func(datum *cgm.OutOfRangeAlertDEPRECATED) { datum.Threshold = pointer.FromInt(1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueIntNotOneOf(1, cgm.OutOfRangeAlertDEPRECATEDThresholds()), "/snooze"),
				),
				Entry("threshold valid",
					func(datum *cgm.OutOfRangeAlertDEPRECATED) { datum.Threshold = pointer.FromInt(1200000) },
				),
				Entry("multiple errors",
					func(datum *cgm.OutOfRangeAlertDEPRECATED) {
						datum.Enabled = nil
						datum.Threshold = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/snooze"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *cgm.OutOfRangeAlertDEPRECATED), expectator func(datum *cgm.OutOfRangeAlertDEPRECATED, expectedDatum *cgm.OutOfRangeAlertDEPRECATED)) {
					for _, origin := range structure.Origins() {
						datum := NewOutOfRangeAlertDEPRECATED()
						mutator(datum)
						expectedDatum := CloneOutOfRangeAlertDEPRECATED(datum)
						normalizer := dataNormalizer.New()
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
					func(datum *cgm.OutOfRangeAlertDEPRECATED) {},
					nil,
				),
				Entry("does not modify the datum; enabled missing",
					func(datum *cgm.OutOfRangeAlertDEPRECATED) { datum.Enabled = nil },
					nil,
				),
				Entry("does not modify the datum; threshold missing",
					func(datum *cgm.OutOfRangeAlertDEPRECATED) { datum.Threshold = nil },
					nil,
				),
			)
		})
	})
})
