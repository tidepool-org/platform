package cgm

import (
	"math"

	"github.com/tidepool-org/platform/data"
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/structure"
)

const (
	HighLevelAlertLevelMgdLMaximum  float64 = 400
	HighLevelAlertLevelMgdLMinimum  float64 = 120
	HighLevelAlertLevelMmolLMaximum float64 = 22.20299
	HighLevelAlertLevelMmolLMinimum float64 = 6.66090
	LowLevelAlertLevelMgdLMaximum   float64 = 100
	LowLevelAlertLevelMgdLMinimum   float64 = 59
	LowLevelAlertLevelMmolLMaximum  float64 = 5.55075
	LowLevelAlertLevelMmolLMinimum  float64 = 3.27494
)

func LevelAlertSnoozes() []int {
	return []int{
		0, 900000, 1800000, 2700000, 3600000, 4500000, 5400000, 6300000,
		7200000, 8100000, 9000000, 9900000, 10800000, 11700000, 12600000,
		13500000, 14400000, 15300000, 16200000, 17100000, 18000000,
	}
}

type LevelAlert struct {
	Enabled *bool    `json:"enabled,omitempty" bson:"enabled,omitempty"`
	Level   *float64 `json:"level,omitempty" bson:"level,omitempty"`
	Snooze  *int     `json:"snooze,omitempty" bson:"snooze,omitempty"`
}

func (l *LevelAlert) Parse(parser data.ObjectParser) {
	l.Enabled = parser.ParseBoolean("enabled")
	l.Level = parser.ParseFloat("level")
	l.Snooze = parser.ParseInteger("snooze")
}

func (l *LevelAlert) Validate(validator structure.Validator, units *string) {
	validator.Bool("enabled", l.Enabled).Exists()
	validator.Float64("level", l.Level).Exists()
	validator.Int("snooze", l.Snooze).Exists().OneOf(LevelAlertSnoozes()...)
}

func (l *LevelAlert) Normalize(normalizer data.Normalizer, units *string) {
	if normalizer.Origin() == structure.OriginExternal {
		l.Level = dataBloodGlucose.NormalizeValueForUnits(l.Level, units)
	}
}

type HighLevelAlert struct {
	LevelAlert `bson:",inline"`
}

func ParseHighLevelAlert(parser data.ObjectParser) *HighLevelAlert {
	if parser.Object() == nil {
		return nil
	}
	highLevelAlert := NewHighLevelAlert()
	highLevelAlert.Parse(parser)
	parser.ProcessNotParsed()
	return highLevelAlert
}

func NewHighLevelAlert() *HighLevelAlert {
	return &HighLevelAlert{}
}

func (h *HighLevelAlert) Validate(validator structure.Validator, units *string) {
	h.LevelAlert.Validate(validator, units)

	validator.Float64("level", h.Level).InRange(h.LevelRangeForUnits(units))
}

func (h *HighLevelAlert) LevelRangeForUnits(units *string) (float64, float64) {
	if units != nil {
		switch *units {
		case dataBloodGlucose.MgdL, dataBloodGlucose.Mgdl:
			return HighLevelAlertLevelMgdLMinimum, HighLevelAlertLevelMgdLMaximum
		case dataBloodGlucose.MmolL, dataBloodGlucose.Mmoll:
			return HighLevelAlertLevelMmolLMinimum, HighLevelAlertLevelMmolLMaximum
		}
	}
	return -math.MaxFloat64, math.MaxFloat64
}

type LowLevelAlert struct {
	LevelAlert `bson:",inline"`
}

func ParseLowLevelAlert(parser data.ObjectParser) *LowLevelAlert {
	if parser.Object() == nil {
		return nil
	}
	lowLevelAlert := NewLowLevelAlert()
	lowLevelAlert.Parse(parser)
	parser.ProcessNotParsed()
	return lowLevelAlert
}

func NewLowLevelAlert() *LowLevelAlert {
	return &LowLevelAlert{}
}

func (l *LowLevelAlert) Validate(validator structure.Validator, units *string) {
	l.LevelAlert.Validate(validator, units)

	validator.Float64("level", l.Level).InRange(l.LevelRangeForUnits(units))
}

func (l *LowLevelAlert) LevelRangeForUnits(units *string) (float64, float64) {
	if units != nil {
		switch *units {
		case dataBloodGlucose.MgdL, dataBloodGlucose.Mgdl:
			return LowLevelAlertLevelMgdLMinimum, LowLevelAlertLevelMgdLMaximum
		case dataBloodGlucose.MmolL, dataBloodGlucose.Mmoll:
			return LowLevelAlertLevelMmolLMinimum, LowLevelAlertLevelMmolLMaximum
		}
	}
	return -math.MaxFloat64, math.MaxFloat64
}
