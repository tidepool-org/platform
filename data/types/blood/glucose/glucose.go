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

func roundToEvenWithDecimalPlaces(v float64, decimals int) float64 {
	if decimals == 0 {
		return math.RoundToEven(v)
	}
	coef := math.Pow10(decimals)
	return math.RoundToEven(v*coef) / coef
}

type Classification string

const (
	ClassificationInvalid Classification = "invalid"

	VeryLow       = "very low"
	Low           = "low"
	InRange       = "in range"
	High          = "high"
	VeryHigh      = "very high"
	ExtremelyHigh = "extremely high"
)

type classificationThreshold struct {
	Name      Classification
	Value     float64
	Inclusive bool
}

type Classifier []classificationThreshold

func (c Classifier) Classify(g *Glucose) (Classification, error) {
	normalized, err := dataBloodGlucose.NormalizeValueForUnitsSafer(g.Value, g.Units)
	if err != nil {
		return ClassificationInvalid, errors.Wrap(err, "unable to classify")
	}
	// Rounded values are used for all classifications. To not do so risks introducing
	// inconsistency between frontend, backend, and/or other reports. See BACK-3800 for
	// details.
	rounded := roundToEvenWithDecimalPlaces(normalized, 1)
	for _, threshold := range c {
		if threshold.Includes(rounded) {
			return threshold.Name, nil
		}
	}
	// Ensure your highest threshold has a value like math.MaxFloat64 to avoid this.
	return ClassificationInvalid, errors.Newf("unable to classify value: %v", *g)
}

// Config helps summaries report the configured thresholds.
//
// These will get wrapped up into a Config returned with the summary report. A simple map
// provides flexibility until we better know how custom classification ranges are going to
// work out.
func (c Classifier) Config() map[Classification]float64 {
	config := map[Classification]float64{}
	for _, classification := range c {
		config[classification.Name] = classification.Value
	}
	return config
}

// TidepoolADAClassificationThresholdsMmolL for classifying glucose values.
//
// All values are normalized to MmolL before classification.
//
// In addition to the standard ADA ranges, the Tidepool-specifiic "extremely high" range is
// added.
//
// It is the author's responsibility to ensure the thresholds remain sorted from smallest to
// largest.
var TidepoolADAClassificationThresholdsMmolL = Classifier([]classificationThreshold{
	{Name: VeryLow, Value: 3, Inclusive: false},                    // Source: https://doi.org/10.2337/dc24-S006
	{Name: Low, Value: 3.9, Inclusive: false},                      // Source: https://doi.org/10.2337/dc24-S006
	{Name: InRange, Value: 10, Inclusive: true},                    // Source: https://doi.org/10.2337/dc24-S006
	{Name: High, Value: 13.9, Inclusive: true},                     // Source: https://doi.org/10.2337/dc24-S006
	{Name: VeryHigh, Value: 19.4, Inclusive: true},                 // Source: https://doi.org/10.2337/dc24-S006
	{Name: ExtremelyHigh, Value: math.MaxFloat64, Inclusive: true}, // Source: BACK-2963
})

func (c classificationThreshold) Includes(value float64) bool {
	if c.Inclusive && value <= c.Value {
		return true
	}
	return value < c.Value
}
