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

func NewOutOfRangeAlert() *cgm.OutOfRangeAlert {
	datum := cgm.NewOutOfRangeAlert()
	datum.Enabled = pointer.Bool(test.RandomBool())
	datum.Threshold = pointer.Int(test.RandomIntFromArray(cgm.OutOfRangeAlertThresholds()))
	return datum
}

func CloneOutOfRangeAlert(datum *cgm.OutOfRangeAlert) *cgm.OutOfRangeAlert {
	if datum == nil {
		return nil
	}
	clone := cgm.NewOutOfRangeAlert()
	clone.Enabled = test.CloneBool(datum.Enabled)
	clone.Threshold = test.CloneInt(datum.Threshold)
	return clone
}

var _ = Describe("OutOfRangeAlert", func() {
	It("OutOfRangeAlertThresholds returns expected", func() {
		Expect(cgm.OutOfRangeAlertThresholds()).To(Equal([]int{
			1200000, 1500000, 1800000, 2100000, 2400000, 2700000, 3000000, 3300000,
			3600000, 3900000, 4200000, 4500000, 4800000, 5100000, 5400000, 5700000,
			6000000, 6300000, 6600000, 6900000, 7200000, 7500000, 7800000, 8100000,
			8400000, 8700000, 9000000, 9300000, 9600000, 9900000, 10200000,
			10500000, 10800000, 11100000, 11400000, 11700000, 12000000, 12300000,
			12600000, 12900000, 13200000, 13500000, 13800000, 14100000, 14400000}))
	})

	Context("ParseOutOfRangeAlert", func() {
		// TODO
	})

	Context("NewOutOfRangeAlert", func() {
		It("is successful", func() {
			Expect(cgm.NewOutOfRangeAlert()).To(Equal(&cgm.OutOfRangeAlert{}))
		})
	})

	Context("OutOfRangeAlert", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *cgm.OutOfRangeAlert), expectedErrors ...error) {
					datum := NewOutOfRangeAlert()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *cgm.OutOfRangeAlert) {},
				),
				Entry("enabled missing",
					func(datum *cgm.OutOfRangeAlert) { datum.Enabled = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
				),
				Entry("enabled true",
					func(datum *cgm.OutOfRangeAlert) { datum.Enabled = pointer.Bool(true) },
				),
				Entry("enabled false",
					func(datum *cgm.OutOfRangeAlert) { datum.Enabled = pointer.Bool(false) },
				),
				Entry("threshold missing",
					func(datum *cgm.OutOfRangeAlert) { datum.Threshold = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/snooze"),
				),
				Entry("threshold invalid",
					func(datum *cgm.OutOfRangeAlert) { datum.Threshold = pointer.Int(1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueIntNotOneOf(1, cgm.OutOfRangeAlertThresholds()), "/snooze"),
				),
				Entry("threshold valid",
					func(datum *cgm.OutOfRangeAlert) { datum.Threshold = pointer.Int(1200000) },
				),
				Entry("multiple errors",
					func(datum *cgm.OutOfRangeAlert) {
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
				func(mutator func(datum *cgm.OutOfRangeAlert), expectator func(datum *cgm.OutOfRangeAlert, expectedDatum *cgm.OutOfRangeAlert)) {
					for _, origin := range structure.Origins() {
						datum := NewOutOfRangeAlert()
						mutator(datum)
						expectedDatum := CloneOutOfRangeAlert(datum)
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
					func(datum *cgm.OutOfRangeAlert) {},
					nil,
				),
				Entry("does not modify the datum; enabled missing",
					func(datum *cgm.OutOfRangeAlert) { datum.Enabled = nil },
					nil,
				),
				Entry("does not modify the datum; threshold missing",
					func(datum *cgm.OutOfRangeAlert) { datum.Threshold = nil },
					nil,
				),
			)
		})
	})
})
