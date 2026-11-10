package clinics

import (
	clinic "github.com/tidepool-org/clinic/client"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/summary/types"
)

// RangeSetFor returns the range set a patient record's glycemic ranges describe, as the
// clinic service reports them. A record without glycemic ranges uses the ADA standard
// preset. A custom range set keeps the key the clinic service derived, so that a summary
// computed with it matches the records using it. It reports an error for a preset
// platform doesn't know, and for a custom range set whose thresholds can't classify
// readings.
func RangeSetFor(ranges *clinic.GlycemicRangesV1) (types.RangeSet, error) {
	if ranges == nil {
		ranges = &clinic.GlycemicRangesV1{Type: clinic.Preset, Preset: clinic.ADAStandard}
	}
	switch ranges.Type {
	case clinic.Custom:
		if ranges.Key == "" {
			return types.RangeSet{}, errors.New("custom glycemic ranges have no key")
		}
		thresholds := ranges.Custom.Thresholds
		return types.NewCustomRangeSet(ranges.Key, string(ranges.Custom.Units), types.Thresholds{
			VeryLow:  thresholds.VeryLow,
			Low:      thresholds.Low,
			High:     thresholds.High,
			VeryHigh: thresholds.VeryHigh,
		})
	case clinic.Preset, "":
		preset := string(ranges.Preset)
		if preset == "" {
			preset = types.PresetADAStandard
		}
		rangeSet, ok := types.PresetRangeSet(preset)
		if !ok {
			return types.RangeSet{}, errors.Newf("glycemic ranges preset %q is unknown", preset)
		}
		return rangeSet, nil
	default:
		return types.RangeSet{}, errors.Newf("glycemic ranges type %q is unknown", ranges.Type)
	}
}
