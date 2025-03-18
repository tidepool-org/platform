package pump

import (
	"github.com/tidepool-org/platform/data"
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	dataTypesDevice "github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	SubType = "pumpSettingsOverride" // TODO: Rename Type to "device/pumpSettingsOverride"; remove SubType

	BasalRateScaleFactorMaximum          = 10.0
	BasalRateScaleFactorMinimum          = 0.1
	CarbohydrateRatioScaleFactorMaximum  = 10.0
	CarbohydrateRatioScaleFactorMinimum  = 0.1
	DurationMaximum                      = 604800000 // 7 days in milliseconds
	DurationMinimum                      = 0
	InsulinSensitivityScaleFactorMaximum = 10.0
	InsulinSensitivityScaleFactorMinimum = 0.1
	MethodAutomatic                      = "automatic"
	MethodManual                         = "manual"
	OverridePresetLengthMaximum          = 100
	OverrideTypeCustom                   = "custom"
	OverrideTypePhysicalActivity         = "physicalActivity"
	OverrideTypePreprandial              = "preprandial"
	OverrideTypePreset                   = "preset"
	OverrideTypeSleep                    = "sleep"
)

func Methods() []string {
	return []string{
		MethodAutomatic,
		MethodManual,
	}
}

func OverrideTypes() []string {
	return []string{
		OverrideTypeCustom,
		OverrideTypePhysicalActivity,
		OverrideTypePreprandial,
		OverrideTypePreset,
		OverrideTypeSleep,
	}
}

type Pump struct {
	dataTypesDevice.Device `bson:",inline"`

	OverrideType                  *string                  `json:"overrideType,omitempty" bson:"overrideType,omitempty"`
	OverridePreset                *string                  `json:"overridePreset,omitempty" bson:"overridePreset,omitempty"`
	Method                        *string                  `json:"method,omitempty" bson:"method,omitempty"`
	Duration                      *int                     `json:"duration,omitempty" bson:"duration,omitempty"`
	DurationExpected              *int                     `json:"expectedDuration,omitempty" bson:"expectedDuration,omitempty"` // TODO: Rename durationExpected
	BloodGlucoseTarget            *dataBloodGlucose.Target `json:"bgTarget,omitempty" bson:"bgTarget,omitempty"`                 // TODO: Rename bloodGlucoseTarget
	BasalRateScaleFactor          *float64                 `json:"basalRateScaleFactor,omitempty" bson:"basalRateScaleFactor,omitempty"`
	CarbohydrateRatioScaleFactor  *float64                 `json:"carbRatioScaleFactor,omitempty" bson:"carbRatioScaleFactor,omitempty"` // TODO: Rename carbohydrateRatioScaleFactor
	InsulinSensitivityScaleFactor *float64                 `json:"insulinSensitivityScaleFactor,omitempty" bson:"insulinSensitivityScaleFactor,omitempty"`
	Units                         *Units                   `json:"units,omitempty" bson:"units,omitempty"`
}

func New() *Pump {
	return &Pump{
		Device: dataTypesDevice.New(SubType),
	}
}

func (p *Pump) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(p.Meta())
	}

	p.Device.Parse(parser)

	p.OverrideType = parser.String("overrideType")
	p.OverridePreset = parser.String("overridePreset")
	p.Method = parser.String("method")
	p.Duration = parser.Int("duration")
	p.DurationExpected = parser.Int("expectedDuration")
	p.BloodGlucoseTarget = dataBloodGlucose.ParseTarget(parser.WithReferenceObjectParser("bgTarget"))
	p.BasalRateScaleFactor = parser.Float64("basalRateScaleFactor")
	p.CarbohydrateRatioScaleFactor = parser.Float64("carbRatioScaleFactor")
	p.InsulinSensitivityScaleFactor = parser.Float64("insulinSensitivityScaleFactor")
	p.Units = ParseUnits(parser.WithReferenceObjectParser("units"))
}

func (p *Pump) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(p.Meta())
	}

	p.Device.Validate(validator)

	if p.SubType != "" {
		validator.String("subType", &p.SubType).EqualTo(SubType)
	}

	var unitsBloodGlucose *string
	if p.Units != nil {
		unitsBloodGlucose = p.Units.BloodGlucose
	}

	validator.String("overrideType", p.OverrideType).Exists().OneOf(OverrideTypes()...)
	if p.OverrideType != nil && *p.OverrideType == OverrideTypePreset {
		validator.String("overridePreset", p.OverridePreset).Exists().NotEmpty().LengthLessThanOrEqualTo(OverridePresetLengthMaximum)
	} else {
		validator.String("overridePreset", p.OverridePreset).NotExists()
	}
	validator.String("method", p.Method).OneOf(Methods()...)
	validator.Int("duration", p.Duration).InRange(DurationMinimum, DurationMaximum)
	expectedDurationValidator := validator.Int("expectedDuration", p.DurationExpected)
	if p.Duration != nil && *p.Duration >= DurationMinimum && *p.Duration <= DurationMaximum {
		expectedDurationValidator.InRange(*p.Duration, DurationMaximum)
	} else {
		expectedDurationValidator.InRange(DurationMinimum, DurationMaximum)
	}
	if p.BloodGlucoseTarget != nil {
		p.BloodGlucoseTarget.Validate(validator.WithReference("bgTarget"), unitsBloodGlucose)
	}
	validator.Float64("basalRateScaleFactor", p.BasalRateScaleFactor).InRange(BasalRateScaleFactorMinimum, BasalRateScaleFactorMaximum)
	validator.Float64("carbRatioScaleFactor", p.CarbohydrateRatioScaleFactor).InRange(CarbohydrateRatioScaleFactorMinimum, CarbohydrateRatioScaleFactorMaximum)
	validator.Float64("insulinSensitivityScaleFactor", p.InsulinSensitivityScaleFactor).InRange(InsulinSensitivityScaleFactorMinimum, InsulinSensitivityScaleFactorMaximum)
	if unitsValidator := validator.WithReference("units"); p.Units != nil {
		if p.BloodGlucoseTarget != nil {
			p.Units.Validate(unitsValidator)
		} else {
			unitsValidator.ReportError(structureValidator.ErrorValueExists())
		}
	} else if p.BloodGlucoseTarget != nil {
		unitsValidator.ReportError(structureValidator.ErrorValueNotExists())
	}
}

func (p *Pump) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(p.Meta())
	}

	p.Device.Normalize(normalizer)

	var unitsBloodGlucose *string
	if p.Units != nil {
		unitsBloodGlucose = p.Units.BloodGlucose
	}

	if p.BloodGlucoseTarget != nil {
		p.BloodGlucoseTarget.Normalize(normalizer.WithReference("bgTarget"), unitsBloodGlucose)
	}
	if p.Units != nil {
		p.Units.Normalize(normalizer.WithReference("units"))
	}
}
