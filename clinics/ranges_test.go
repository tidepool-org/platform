package clinics_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	clinic "github.com/tidepool-org/clinic/client"

	"github.com/tidepool-org/platform/clinics"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/summary/types"
)

var _ = Describe("RangeSetFor", func() {
	It("returns the standard preset for a record without glycemic ranges", func() {
		rangeSet, err := clinics.RangeSetFor(nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(rangeSet.Key).To(Equal(types.StandardRangeSetKey))
	})

	It("returns the standard preset for a record with empty glycemic ranges", func() {
		rangeSet, err := clinics.RangeSetFor(&clinic.GlycemicRangesV1{})
		Expect(err).ToNot(HaveOccurred())
		Expect(rangeSet.Key).To(Equal(types.StandardRangeSetKey))
	})

	DescribeTable("returns the range set of a preset",
		func(preset clinic.GlycemicRangesPresetV1) {
			rangeSet, err := clinics.RangeSetFor(&clinic.GlycemicRangesV1{Type: clinic.Preset, Preset: preset})
			Expect(err).ToNot(HaveOccurred())
			expected, ok := types.PresetRangeSet(string(preset))
			Expect(ok).To(BeTrue())
			Expect(rangeSet).To(Equal(expected))
		},
		Entry("ADA standard", clinic.ADAStandard),
		Entry("ADA high risk", clinic.ADAHighRisk),
		Entry("ADA pregnancy type 1", clinic.ADAPregnancyType1),
		Entry("ADA pregnancy type 2", clinic.ADAPregnancyType2),
	)

	It("returns an error for an unknown preset", func() {
		_, err := clinics.RangeSetFor(&clinic.GlycemicRangesV1{Type: clinic.Preset, Preset: "adaUnknown"})
		Expect(err).To(MatchError(ContainSubstring(`preset "adaUnknown" is unknown`)))
	})

	// The key is the one the clinic service derived, so that the summary matches the
	// records using the range set
	It("returns a custom range set with the key the clinic service derived", func() {
		ranges := &clinic.GlycemicRangesV1{
			Type: clinic.Custom,
			Key:  "custom:mmol/L:,3.9,10,",
			Custom: clinic.GlycemicRangesCustomV1{
				Name:       "Mine",
				Units:      clinic.GlycemicRangesCustomV1UnitsMmolL,
				Thresholds: clinic.GlycemicRangesThresholdsV1{Low: 3.9, High: 10},
			},
		}
		rangeSet, err := clinics.RangeSetFor(ranges)
		Expect(err).ToNot(HaveOccurred())
		Expect(rangeSet.Key).To(Equal("custom:mmol/L:,3.9,10,"))
		Expect(rangeSet.Thresholds).To(HaveKeyWithValue("mmol/L", types.Thresholds{Low: 3.9, High: 10}))
		Expect(rangeSet.Thresholds).To(HaveKey("mg/dL"))
	})

	It("returns an error for a custom range set without a key", func() {
		ranges := &clinic.GlycemicRangesV1{Type: clinic.Custom, Custom: clinic.GlycemicRangesCustomV1{
			Units: clinic.GlycemicRangesCustomV1UnitsMgdL, Thresholds: clinic.GlycemicRangesThresholdsV1{Low: 70, High: 180},
		}}
		_, err := clinics.RangeSetFor(ranges)
		Expect(err).To(MatchError(ContainSubstring("no key")))
	})

	It("returns an error for a custom range set platform cannot classify readings with", func() {
		ranges := &clinic.GlycemicRangesV1{Type: clinic.Custom, Key: "custom:mg/dL:104,105,180,", Custom: clinic.GlycemicRangesCustomV1{
			Units:      clinic.GlycemicRangesCustomV1UnitsMgdL,
			Thresholds: clinic.GlycemicRangesThresholdsV1{VeryLow: pointer.FromFloat64(104), Low: 105, High: 180},
		}}
		_, err := clinics.RangeSetFor(ranges)
		Expect(err).To(HaveOccurred())
	})

	It("returns an error for an unknown type", func() {
		_, err := clinics.RangeSetFor(&clinic.GlycemicRangesV1{Type: "other"})
		Expect(err).To(MatchError(ContainSubstring(`type "other" is unknown`)))
	})
})
