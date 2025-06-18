package glucose

import (
	"math"

	"github.com/tidepool-org/platform/data"
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/types/blood"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
)

type Glucose struct {
	blood.Blood `bson:",inline"`
}

func New(typ string) Glucose {
	return Glucose{
		Blood: blood.New(typ),
	}
}

func (g *Glucose) Validate(validator structure.Validator) {
	g.Blood.Validate(validator)

	validator.String("units", g.Units).Exists().OneOf(dataBloodGlucose.Units()...)
	validator.Float64("value", g.Value).Exists().InRange(dataBloodGlucose.ValueRangeForUnits(g.Units))
}

func (g *Glucose) Normalize(normalizer data.Normalizer) {
	g.Blood.Normalize(normalizer)

	if normalizer.Origin() == structure.OriginExternal {
		units := g.Units
		g.Units = dataBloodGlucose.NormalizeUnits(units)
		g.Value = dataBloodGlucose.NormalizeValueForUnits(g.Value, units)
	}
}

// Classify the datum based on ADA thresholds.
//
// It pretends to handle values in mg/dL, but all values received by platform are normalized
// into mmol/L upon reception, so being able to classify values in other units is of limited
// use.
func (g *Glucose) Classify() (RangeClassification, error) {
	switch {
	case g.Units == nil:
		return RangeInvalid, errors.New("unhandled units: nil")
	case g.Value == nil:
		return RangeInvalid, errors.New("unhandled value: nil")
	case *g.Units == dataBloodGlucose.MgdL:
		return g.classify(thresholdsMgdL)
	case *g.Units == dataBloodGlucose.MmolL:
		return g.classify(thresholdsMmolL)
	default:
		return RangeInvalid, errors.Newf("unhandled units: %s", *g.Units)
	}
}

// RangeClassification of a blood glucose value, e.g. Low, Very Low, etc.
type RangeClassification int

const (
	RangeInvalid RangeClassification = iota
	RangeVeryLow
	RangeLow
	RangeTarget
	RangeHigh
	RangeVeryHigh
	RangeExtremelyHigh
)

type classificationThresholds struct {
	VeryLow  float64
	Low      float64
	Target   float64
	High     float64
	VeryHigh float64
	// Precision to round to for these thresholds/units.
	Precision int
}

func (g *Glucose) classify(thresholds classificationThresholds) (RangeClassification, error) {
	if g.Value == nil {
		return RangeInvalid, errors.New("unhandled value: nil")
	}
	// Rounded values are used when distinguishing between low, target, and high only. This
	// matches the algorithm used on the frontend to generate reports, and we want reports
	// generated from summaries to match those reports, so we need to do the same. It could
	// be argued that the ADA ranges are intended to be used with rounded values, but then
	// we wouldn't match the frontend, so any change in that policy should be coordinated
	// with the frontend.
	rounded := roundToEvenWithPrecision(*g.Value, thresholds.Precision)
	switch {
	case *g.Value < thresholds.VeryLow:
		return RangeVeryLow, nil
	case rounded < thresholds.Low:
		return RangeLow, nil
	case rounded <= thresholds.Target:
		return RangeTarget, nil
	case rounded <= thresholds.High:
		return RangeHigh, nil
	case *g.Value <= thresholds.VeryHigh:
		return RangeVeryHigh, nil
	default:
		return RangeExtremelyHigh, nil
	}
}

func roundToEvenWithPrecision(v float64, decimals int) float64 {
	if decimals < 1 {
		return math.RoundToEven(v)
	}
	coef := math.Pow(10, float64(decimals))
	return math.RoundToEven(v*coef) / coef
}

const (
	ThresholdMmolLVeryLow  float64 = 3    // Source: https://doi.org/10.2337/dc24-S006
	ThresholdMmolLLow      float64 = 3.9  // Source: https://doi.org/10.2337/dc24-S006
	ThresholdMmolLTarget   float64 = 10   // Source: https://doi.org/10.2337/dc24-S006
	ThresholdMmolLHigh     float64 = 13.9 // Source: https://doi.org/10.2337/dc24-S006
	ThresholdMmolLVeryHigh float64 = 19.4 // Source: BACK-2963
)

var thresholdsMmolL = classificationThresholds{
	VeryLow:   ThresholdMmolLVeryLow,
	Low:       ThresholdMmolLLow,
	Target:    ThresholdMmolLTarget,
	High:      ThresholdMmolLHigh,
	VeryHigh:  ThresholdMmolLVeryHigh,
	Precision: 1,
}

const (
	ThresholdMgdLVeryLow  float64 = 54  // Source: https://doi.org/10.2337/dc24-S006
	ThresholdMgdLLow      float64 = 70  // Source: https://doi.org/10.2337/dc24-S006
	ThresholdMgdLTarget   float64 = 180 // Source: https://doi.org/10.2337/dc24-S006
	ThresholdMgdLHigh     float64 = 250 // Source: https://doi.org/10.2337/dc24-S006
	ThresholdMgdLVeryHigh float64 = 350 // Source: BACK-2963
)

var thresholdsMgdL = classificationThresholds{
	VeryLow:   ThresholdMgdLVeryLow,
	Low:       ThresholdMgdLLow,
	Target:    ThresholdMgdLTarget,
	High:      ThresholdMgdLHigh,
	VeryHigh:  ThresholdMgdLVeryHigh,
	Precision: 0,
}
