package pump

import (
	"sort"

	"github.com/tidepool-org/platform/data"
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	AbbreviationLengthMaximum            = 100
	BasalRateScaleFactorMaximum          = 10.0
	BasalRateScaleFactorMinimum          = 0.1
	CarbohydrateRatioScaleFactorMaximum  = 10.0
	CarbohydrateRatioScaleFactorMinimum  = 0.1
	DurationMaximum                      = 604800 // 7 days in seconds
	DurationMinimum                      = 0
	InsulinSensitivityScaleFactorMaximum = 10.0
	InsulinSensitivityScaleFactorMinimum = 0.1
	OverridePresetLengthMaximum          = 100
	OverridePresetNameLengthMaximum      = 100
)

type OverridePreset struct {
	Abbreviation                  *string                  `json:"abbreviation,omitempty" bson:"abbreviation,omitempty"`
	Duration                      *int                     `json:"duration,omitempty" bson:"duration,omitempty"`
	BloodGlucoseTarget            *dataBloodGlucose.Target `json:"bgTarget,omitempty" bson:"bgTarget,omitempty"`
	BasalRateScaleFactor          *float64                 `json:"basalRateScaleFactor,omitempty" bson:"basalRateScaleFactor,omitempty"`
	CarbohydrateRatioScaleFactor  *float64                 `json:"carbRatioScaleFactor,omitempty" bson:"carbRatioScaleFactor,omitempty"`
	InsulinSensitivityScaleFactor *float64                 `json:"insulinSensitivityScaleFactor,omitempty" bson:"insulinSensitivityScaleFactor,omitempty"`
}

func ParseOverridePreset(parser structure.ObjectParser) *OverridePreset {
	if !parser.Exists() {
		return nil
	}
	datum := NewOverridePreset()
	parser.Parse(datum)
	return datum
}

func NewOverridePreset() *OverridePreset {
	return &OverridePreset{}
}

func (o *OverridePreset) Parse(parser structure.ObjectParser) {
	o.Abbreviation = parser.String("abbreviation")
	o.Duration = parser.Int("duration")
	o.BloodGlucoseTarget = dataBloodGlucose.ParseTarget(parser.WithReferenceObjectParser("bgTarget"))
	o.BasalRateScaleFactor = parser.Float64("basalRateScaleFactor")
	o.CarbohydrateRatioScaleFactor = parser.Float64("carbRatioScaleFactor")
	o.InsulinSensitivityScaleFactor = parser.Float64("insulinSensitivityScaleFactor")
}

func (o *OverridePreset) Validate(validator structure.Validator, unitsBloodGlucose *string) {
	validator.String("abbreviation", o.Abbreviation).NotEmpty().LengthLessThanOrEqualTo(AbbreviationLengthMaximum)
	validator.Int("duration", o.Duration).InRange(DurationMinimum, DurationMaximum)
	if o.BloodGlucoseTarget != nil {
		o.BloodGlucoseTarget.Validate(validator.WithReference("bgTarget"), unitsBloodGlucose)
	}
	validator.Float64("basalRateScaleFactor", o.BasalRateScaleFactor).InRange(BasalRateScaleFactorMinimum, BasalRateScaleFactorMaximum)
	validator.Float64("carbRatioScaleFactor", o.CarbohydrateRatioScaleFactor).InRange(CarbohydrateRatioScaleFactorMinimum, CarbohydrateRatioScaleFactorMaximum)
	validator.Float64("insulinSensitivityScaleFactor", o.InsulinSensitivityScaleFactor).InRange(InsulinSensitivityScaleFactorMinimum, InsulinSensitivityScaleFactorMaximum)
}

func (o *OverridePreset) Normalize(normalizer data.Normalizer, unitsBloodGlucose *string) {
	if o.BloodGlucoseTarget != nil {
		o.BloodGlucoseTarget.Normalize(normalizer.WithReference("bgTarget"), unitsBloodGlucose)
	}
}

type OverridePresetMap map[string]*OverridePreset

func ParseOverridePresetMap(parser structure.ObjectParser) *OverridePresetMap {
	if !parser.Exists() {
		return nil
	}
	datum := NewOverridePresetMap()
	parser.Parse(datum)
	return datum
}

func NewOverridePresetMap() *OverridePresetMap {
	return &OverridePresetMap{}
}

func (o *OverridePresetMap) Parse(parser structure.ObjectParser) {
	for _, reference := range parser.References() {
		o.Set(reference, ParseOverridePreset(parser.WithReferenceObjectParser(reference)))
	}
}

func (o *OverridePresetMap) Validate(validator structure.Validator, unitsBloodGlucose *string) {
	for _, name := range o.sortedNames() {
		datumValidator := validator.WithReference(name)
		if datum := o.Get(name); datum != nil {
			datum.Validate(datumValidator, unitsBloodGlucose)
		} else {
			datumValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

func (o *OverridePresetMap) Normalize(normalizer data.Normalizer, unitsBloodGlucose *string) {
	for _, name := range o.sortedNames() {
		if datum := o.Get(name); datum != nil {
			datum.Normalize(normalizer.WithReference(name), unitsBloodGlucose)
		}
	}
}

func (o *OverridePresetMap) Get(name string) *OverridePreset {
	if datum, exists := (*o)[name]; exists {
		return datum
	}
	return nil
}

func (o *OverridePresetMap) Set(name string, datum *OverridePreset) {
	(*o)[name] = datum
}

func (o *OverridePresetMap) sortedNames() []string {
	names := []string{}
	for name := range *o {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
