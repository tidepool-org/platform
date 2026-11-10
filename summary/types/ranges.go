package types

import (
	"errors"
	"fmt"
	"math"

	"github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/pointer"
)

// Thresholds are the thresholds between the glycemic ranges of a [RangeSet], in one unit.
//
// They're the clinical values at which the ranges meet. A reading exactly at VeryLow or Low
// is in the range above it, and one exactly at High or VeryHigh is in the range below
// it. VeryLow and VeryHigh are optional: without them, readings below Low are all low, and
// readings above High are all high. Readings aren't classified against thresholds directly,
// but against the [CutPoints] derived from them, which have the same four boundaries.
type Thresholds struct {
	VeryLow  *float64
	Low      float64
	High     float64
	VeryHigh *float64
}

// A thresholdName names one of the four thresholds.
type thresholdName string

const (
	thresholdVeryLow  thresholdName = "veryLow"
	thresholdLow      thresholdName = "low"
	thresholdHigh     thresholdName = "high"
	thresholdVeryHigh thresholdName = "veryHigh"
)

var thresholdNames = []thresholdName{
	thresholdVeryLow,
	thresholdLow,
	thresholdHigh,
	thresholdVeryHigh,
}

// threshold returns the named threshold, or nil when it isn't set, as the optional
// thresholds may not be. A name other than the four is treated as a threshold that isn't
// set.
func (t *Thresholds) threshold(name thresholdName) *float64 {
	switch name {
	case thresholdVeryLow:
		return t.VeryLow
	case thresholdLow:
		return &t.Low
	case thresholdHigh:
		return &t.High
	case thresholdVeryHigh:
		return t.VeryHigh
	default:
		return nil
	}
}

func (t Thresholds) clone() Thresholds {
	cloned := t
	cloned.VeryLow = pointer.CloneFloat64(t.VeryLow)
	cloned.VeryHigh = pointer.CloneFloat64(t.VeryHigh)
	return cloned
}

// validate reports an error if the units aren't [glucose.MgdL] or [glucose.MmolL], or if a
// threshold isn't a number, is outside the range of values of the units, isn't a multiple
// of the resolution of readings in the units, which the cut points depend on, or isn't
// above the threshold before it.
func (t Thresholds) validate(units string) error {
	if units != glucose.MgdL && units != glucose.MmolL {
		return fmt.Errorf("unsupported units %q: must be %q or %q", units, glucose.MgdL,
			glucose.MmolL)
	}
	minimum, maximum := glucose.ValueRangeForUnits(&units)
	previous := math.Inf(-1)
	for _, name := range thresholdNames {
		value := t.threshold(name)
		if value == nil {
			continue
		}
		if math.IsNaN(*value) {
			return fmt.Errorf("%s threshold isn't a number", name)
		}
		if *value < minimum || *value > maximum {
			return fmt.Errorf("%s threshold %v is outside the range of %s, %v to %v",
				name, *value, units, minimum, maximum)
		}
		if steps := *value / unitStep(units); math.Abs(steps-math.Round(steps)) > 1e-6 {
			return fmt.Errorf("%s threshold %v isn't a multiple of %v %s, the resolution "+
				"of readings", name, *value, unitStep(units), units)
		}
		if *value <= previous {
			return fmt.Errorf("thresholds aren't ascending: %s", t.describe(units))
		}
		previous = *value
	}
	return nil
}

func (t Thresholds) describe(units string) string {
	optional := func(value *float64) string {
		if value == nil {
			return "unset"
		}
		return fmt.Sprint(*value)
	}
	return fmt.Sprintf("%s: veryLow %s, low %v, high %v, veryHigh %s",
		units, optional(t.VeryLow), t.Low, t.High, optional(t.VeryHigh))
}

// validate reports an error if the range set doesn't define valid thresholds in both
// units.
func (rs RangeSet) validate() error {
	if len(rs.Thresholds) == 0 {
		return errors.New("range set has no thresholds")
	}
	for units, thresholds := range rs.Thresholds {
		if err := thresholds.validate(units); err != nil {
			return err
		}
	}
	if len(rs.Thresholds) != 2 {
		return errors.New("range set doesn't define thresholds in both units")
	}
	return nil
}

// RangeSet defines a set of glycemic ranges, very low, low, in range, high, and very high,
// by the thresholds between them.
//
// Key identifies the set, as the clinic service's glycemicRanges.key does, such as
// "preset:adaStandard" or "custom:mg/dL:54,70,180,250". Thresholds holds the thresholds
// between the ranges in each unit, [glucose.MgdL] and [glucose.MmolL], so that readings
// recorded in either unit are classified by thresholds expressed in their own unit. The ADA
// presets, from [PresetRangeSet], define both units; a custom set, from
// [NewCustomRangeSet], is defined in one unit and converted to the other.
type RangeSet struct {
	Key        string
	Thresholds map[string]Thresholds
}

// NewCustomRangeSet returns a range set from thresholds defined in one unit, converting
// them to the other unit at the other unit's resolution, as the ADA presets are defined in
// both.
//
// A threshold in mg/dL converts to the nearest single decimal of mmol/L, and one in mmol/L
// to the smallest whole mg/dL that converts back to it; both reproduce the presets'
// definitions from each other.
//
// Readings in either unit are then classified by that unit's thresholds, so a reading of
// 10.0 mmol/L is in range for a high threshold of 180 mg/dL, as it is for the ADA standard,
// even though it converts to 180.16 mg/dL.
func NewCustomRangeSet(key, units string, thresholds Thresholds) (RangeSet, error) {
	units, ok := normalizeUnits(units)
	if !ok {
		return RangeSet{}, fmt.Errorf("unsupported units %q", units)
	}
	if err := thresholds.validate(units); err != nil {
		return RangeSet{}, err
	}
	other, convert := glucose.MmolL, mgdLToMmolL
	if units == glucose.MmolL {
		other, convert = glucose.MgdL, mmolLToMgdL
	}
	converted := convertThresholds(thresholds, convert)
	// Whole mg/dL values are closer together than single decimals of mmol/L, so mg/dL
	// thresholds less than 2 apart convert to the same mmol/L threshold.
	if err := converted.validate(other); err != nil {
		return RangeSet{}, fmt.Errorf("converted to %s, %w", other, err)
	}
	return RangeSet{
		Key:        key,
		Thresholds: map[string]Thresholds{units: thresholds.clone(), other: converted},
	}, nil
}

// normalizeUnits returns the units as [glucose.MgdL] or [glucose.MmolL], accepting the
// lowercase spellings the glucose package does, or false for any other units.
func normalizeUnits(units string) (string, bool) {
	switch units {
	case glucose.MgdL, glucose.Mgdl:
		return glucose.MgdL, true
	case glucose.MmolL, glucose.Mmoll:
		return glucose.MmolL, true
	}
	return units, false
}

func convertThresholds(t Thresholds, convert func(float64) float64) Thresholds {
	converted := Thresholds{Low: convert(t.Low), High: convert(t.High)}
	if t.VeryLow != nil {
		converted.VeryLow = pointer.FromFloat64(convert(*t.VeryLow))
	}
	if t.VeryHigh != nil {
		converted.VeryHigh = pointer.FromFloat64(convert(*t.VeryHigh))
	}
	return converted
}

// mgdLToMmolL converts a threshold in mg/dL to the nearest single decimal of mmol/L.
func mgdLToMmolL(mgdL float64) float64 {
	return math.Round(stored(mgdL, glucose.MgdL)*10) / 10
}

// mmolLToMgdL converts a threshold in mmol/L to the smallest whole mg/dL that converts back
// to it, which is how the ADA presets pair their thresholds: 7.8 mmol/L is 140 mg/dL,
// although 141 mg/dL also converts to 7.8. Every single decimal of mmol/L has such a value,
// as whole mg/dL values are closer together.
func mmolLToMgdL(mmolL float64) float64 {
	target := math.Round(mmolL*10) / 10
	mgdL := math.Max(0, math.Floor(mmolL*glucose.MmolLToMgdLConversionFactor)-1)
	for mgdL < glucose.MgdLMaximum && mgdLToMmolL(mgdL) != target {
		mgdL++
	}
	return mgdL
}

// The names of the ADA presets, as the clinic service's glycemicRangesPreset names them.
const (
	PresetADAStandard       = "adaStandard"
	PresetADAHighRisk       = "adaHighRisk"
	PresetADAPregnancyType1 = "adaPregnancyType1"
	PresetADAPregnancyType2 = "adaPregnancyType2"
)

const presetKeyPrefix = "preset:"

// StandardRangeSetKey is the key of the ADA standard preset, the range set the stored
// summaries are computed with.
const StandardRangeSetKey = presetKeyPrefix + PresetADAStandard

// PresetRangeSet returns the range set of the named ADA preset, as the frontend (viz's
// glycemicRanges.js) defines it in each unit, or false for any other name. The pregnancy
// type 2 preset shares the pregnancy type 1 thresholds. Each call returns a new value,
// which the caller may modify.
func PresetRangeSet(preset string) (RangeSet, bool) {
	var thresholds map[string]Thresholds
	switch preset {
	case PresetADAStandard:
		thresholds = map[string]Thresholds{
			glucose.MgdL: {VeryLow: pointer.FromFloat64(54), Low: 70, High: 180,
				VeryHigh: pointer.FromFloat64(250)},
			glucose.MmolL: {VeryLow: pointer.FromFloat64(3), Low: 3.9, High: 10,
				VeryHigh: pointer.FromFloat64(13.9)},
		}
	case PresetADAHighRisk:
		thresholds = map[string]Thresholds{
			glucose.MgdL:  {Low: 70, High: 180, VeryHigh: pointer.FromFloat64(250)},
			glucose.MmolL: {Low: 3.9, High: 10, VeryHigh: pointer.FromFloat64(13.9)},
		}
	case PresetADAPregnancyType1, PresetADAPregnancyType2:
		thresholds = map[string]Thresholds{
			glucose.MgdL:  {VeryLow: pointer.FromFloat64(54), Low: 63, High: 140},
			glucose.MmolL: {VeryLow: pointer.FromFloat64(3), Low: 3.5, High: 7.8},
		}
	default:
		return RangeSet{}, false
	}
	return RangeSet{Key: presetKeyPrefix + preset, Thresholds: thresholds}, true
}

// CutPoints are the values at which readings normalized to mmol/L are split into the ranges
// of a range set.
//
// They're derived from the range set's [Thresholds] by [RangeSet.CutPoints], and share the
// thresholds' four boundaries, but differ in kind: a cut point is a value no reading takes,
// placed between the stored values on either side of its boundary, so that the thresholds'
// tie rule becomes a uniform one: a reading below VeryLow is very low, one below Low is
// low, one below High is in range, one below VeryHigh is high, and any other is very
// high. A cut point for an optional threshold the range set leaves unset is infinite, so
// that no reading falls in its range.
type CutPoints struct {
	VeryLow  float64
	Low      float64
	High     float64
	VeryHigh float64
}

// StandardCutPoints returns the cut points of the ADA standard ranges, with which the
// stored summaries are calculated. They're the constants defined in BACK-4158 rather than
// derived, so that the stored summaries, and the thresholds they report, don't change.
func StandardCutPoints() CutPoints {
	return CutPoints{
		VeryLow:  veryLowBloodGlucose,
		Low:      lowBloodGlucose,
		High:     highBloodGlucose,
		VeryHigh: veryHighBloodGlucose,
	}
}

// A Classification is the glycemic range that a reading falls in.
type Classification int

// The classifications, from the lowest range to the highest.
const (
	ClassificationVeryLow Classification = iota + 1
	ClassificationLow
	ClassificationInRange
	ClassificationHigh
	ClassificationVeryHigh
)

// Classify classifies a reading normalized to mmol/L.
func (cp CutPoints) Classify(n float64) Classification {
	switch {
	case n < cp.VeryLow:
		return ClassificationVeryLow
	case n < cp.Low:
		return ClassificationLow
	case n < cp.High:
		return ClassificationInRange
	case n < cp.VeryHigh:
		return ClassificationHigh
	default:
		return ClassificationVeryHigh
	}
}

// CutPoints derives the cut points of the range set's boundaries, generalizing the method
// of BACK-4158. Readings are classified after normalization to mmol/L, and every stored
// reading is assumed to be a whole number of mg/dL or a single decimal of mmol/L. Each cut
// point is placed in the gap between the stored values of the inputs neighboring its
// boundary, in both units, so that readings in each unit classify as that unit's
// thresholds intend. That requires the range set to define both units, as those from
// [PresetRangeSet] and [NewCustomRangeSet] do, with thresholds that agree to within a step
// of each unit's resolution, as the presets' do and as converted thresholds always do;
// otherwise no such gap exists, which is reported as an error.
func (rs RangeSet) CutPoints() (CutPoints, error) {
	if err := rs.validate(); err != nil {
		return CutPoints{}, err
	}

	veryLow, err := rs.cutPoint(thresholdVeryLow)
	if err != nil {
		return CutPoints{}, err
	}
	low, err := rs.cutPoint(thresholdLow)
	if err != nil {
		return CutPoints{}, err
	}
	high, err := rs.cutPoint(thresholdHigh)
	if err != nil {
		return CutPoints{}, err
	}
	veryHigh, err := rs.cutPoint(thresholdVeryHigh)
	if err != nil {
		return CutPoints{}, err
	}

	cutPoints := CutPoints{VeryLow: veryLow, Low: low, High: high, VeryHigh: veryHigh}
	if !(veryLow < low && low < high && high < veryHigh) {
		return CutPoints{},
			fmt.Errorf("range set thresholds aren't ascending: %+v", cutPoints)
	}
	return cutPoints, nil
}

// isLowerBoundary reports whether the named threshold defines a lower boundary, below
// which readings are in a lower range: Very Low / Low or Low / In Range. A reading at the
// threshold is in the range above a lower boundary, and in the range below an upper one.
func isLowerBoundary(name thresholdName) bool {
	return name == thresholdVeryLow || name == thresholdLow
}

// A gap is the interval of stored values, in one unit, between the inputs neighboring a
// boundary: below is the greatest stored value that must fall below the boundary's cut
// point, and above is the least that must fall at or above it, so the cut point lies in
// (below, above].
type gap struct {
	below float64
	above float64
}

// gaps returns the gap, in each of the range set's units, of the boundary that the named
// threshold defines. In each unit, the threshold and the input one step from it, at the
// unit's resolution, are the neighboring inputs. For a lower boundary, a reading at the
// threshold is in the range above, so the threshold is the gap's above, and the input a
// step below it is its below. For an upper boundary, a reading at the threshold is in
// the range below, so the threshold is the gap's below, and the input a step above it is
// its above.
//
// Given the ADA standard range set and thresholdLow, whose thresholds are 70 mg/dL and
// 3.9 mmol/L, the gaps are (3.83002, 3.88552] for mg/dL, the stored values of 69 and 70
// mg/dL, and (3.8, 3.9] for mmol/L. Given thresholdHigh, whose thresholds are 180 mg/dL
// and 10.0 mmol/L, they're (9.99135, 10.04685] for mg/dL, the stored values of 180 and
// 181 mg/dL, and (10.0, 10.1] for mmol/L. There are no gaps when the range set doesn't
// set the threshold.
func (rs RangeSet) gaps(name thresholdName) (map[string]gap, error) {
	gaps := map[string]gap{}
	for units, t := range rs.Thresholds {
		value := t.threshold(name)
		if value == nil {
			continue
		}
		threshold := stored(*value, units)
		if isLowerBoundary(name) {
			gaps[units] = gap{
				below: stored(*value-unitStep(units), units),
				above: threshold,
			}
		} else {
			gaps[units] = gap{
				below: threshold,
				above: stored(*value+unitStep(units), units),
			}
		}
	}
	if len(gaps) != 0 && len(gaps) != len(rs.Thresholds) {
		return nil, fmt.Errorf("range set sets %s in only some of its units", name)
	}
	return gaps, nil
}

// cutPoint returns the cut point of the boundary that the named threshold defines: the
// middle of the intersection of its gaps in each unit, so that in every unit a reading
// below it is in the range below the boundary, and one at or above it is in the range
// above. For a threshold the range set doesn't set, it's -Inf for a lower boundary or
// +Inf for an upper one, so that no reading is in the range beyond it.
func (rs RangeSet) cutPoint(name thresholdName) (float64, error) {
	gaps, err := rs.gaps(name)
	if err != nil {
		return 0, err
	}
	if len(gaps) == 0 {
		if isLowerBoundary(name) {
			return math.Inf(-1), nil
		}
		return math.Inf(1), nil
	}
	below, above := math.Inf(-1), math.Inf(1)
	for _, g := range gaps {
		below = max(below, g.below)
		above = min(above, g.above)
	}
	if below >= above {
		return 0, fmt.Errorf("range set's %s thresholds leave no gap for a cut point", name)
	}
	return (below + above) / 2, nil
}

func unitStep(units string) float64 {
	if units == glucose.MgdL {
		return 1
	}
	return 0.1
}

// stored returns the value as it's stored after normalization to mmol/L.
func stored(value float64, units string) float64 {
	return *glucose.NormalizeValueForUnits(&value, &units)
}
