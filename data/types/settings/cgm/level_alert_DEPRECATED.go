package cgm

import (
	"math"

	"github.com/tidepool-org/platform/data"
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/structure"
)

const (
	HighLevelAlertDEPRECATEDLevelMgdLMaximum  float64 = 400
	HighLevelAlertDEPRECATEDLevelMgdLMinimum  float64 = 100
	HighLevelAlertDEPRECATEDLevelMmolLMaximum float64 = 22.20299
	HighLevelAlertDEPRECATEDLevelMmolLMinimum float64 = 5.55075
	LowLevelAlertDEPRECATEDLevelMgdLMaximum   float64 = 150
	LowLevelAlertDEPRECATEDLevelMgdLMinimum   float64 = 59
	LowLevelAlertDEPRECATEDLevelMmolLMaximum  float64 = 8.32612
	LowLevelAlertDEPRECATEDLevelMmolLMinimum  float64 = 3.27494
)

func LevelAlertDEPRECATEDSnoozes() []int {
	return []int{
		0, 900000, 1200000, 1500000, 1800000, 2100000, 2400000, 2700000,
		3000000, 3300000, 3600000, 3900000, 4200000, 4500000, 4800000, 5100000,
		5400000, 5700000, 6000000, 6300000, 6600000, 6900000, 7200000, 7500000,
		7800000, 8100000, 8400000, 8700000, 9000000, 9300000, 9600000, 9900000,
		10200000, 10500000, 10800000, 11100000, 11400000, 11700000, 12000000,
		12300000, 12600000, 12900000, 13200000, 13500000, 13800000, 14100000,
		14400000, 15300000, 16200000, 17100000, 18000000,
	}
}

type LevelAlertDEPRECATED struct {
	Enabled *bool    `json:"enabled,omitempty" bson:"enabled,omitempty"`
	Level   *float64 `json:"level,omitempty" bson:"level,omitempty"`
	Snooze  *int     `json:"snooze,omitempty" bson:"snooze,omitempty"`
}

func (l *LevelAlertDEPRECATED) Parse(parser structure.ObjectParser) {
	l.Enabled = parser.Bool("enabled")
	l.Level = parser.Float64("level")
	l.Snooze = parser.Int("snooze")
}

func (l *LevelAlertDEPRECATED) Validate(validator structure.Validator, units *string) {
	validator.Bool("enabled", l.Enabled).Exists()
	validator.Float64("level", l.Level).Exists()
	validator.Int("snooze", l.Snooze).Exists().OneOf(LevelAlertDEPRECATEDSnoozes()...)
}

func (l *LevelAlertDEPRECATED) Normalize(normalizer data.Normalizer, units *string) {
	if normalizer.Origin() == structure.OriginExternal {
		l.Level = dataBloodGlucose.NormalizeValueForUnits(l.Level, units)
	}
}

type HighLevelAlertDEPRECATED struct {
	LevelAlertDEPRECATED `bson:",inline"`
}

func ParseHighLevelAlertDEPRECATED(parser structure.ObjectParser) *HighLevelAlertDEPRECATED {
	if !parser.Exists() {
		return nil
	}
	datum := NewHighLevelAlertDEPRECATED()
	parser.Parse(datum)
	return datum
}

func NewHighLevelAlertDEPRECATED() *HighLevelAlertDEPRECATED {
	return &HighLevelAlertDEPRECATED{}
}

func (h *HighLevelAlertDEPRECATED) Validate(validator structure.Validator, units *string) {
	h.LevelAlertDEPRECATED.Validate(validator, units)

	validator.Float64("level", h.Level).InRange(h.LevelRangeForUnits(units))
}

func (h *HighLevelAlertDEPRECATED) LevelRangeForUnits(units *string) (float64, float64) {
	if units != nil {
		switch *units {
		case dataBloodGlucose.MgdL, dataBloodGlucose.Mgdl:
			return HighLevelAlertDEPRECATEDLevelMgdLMinimum, HighLevelAlertDEPRECATEDLevelMgdLMaximum
		case dataBloodGlucose.MmolL, dataBloodGlucose.Mmoll:
			return HighLevelAlertDEPRECATEDLevelMmolLMinimum, HighLevelAlertDEPRECATEDLevelMmolLMaximum
		}
	}
	return -math.MaxFloat64, math.MaxFloat64
}

type LowLevelAlertDEPRECATED struct {
	LevelAlertDEPRECATED `bson:",inline"`
}

func ParseLowLevelAlertDEPRECATED(parser structure.ObjectParser) *LowLevelAlertDEPRECATED {
	if !parser.Exists() {
		return nil
	}
	datum := NewLowLevelAlertDEPRECATED()
	parser.Parse(datum)
	return datum
}

func NewLowLevelAlertDEPRECATED() *LowLevelAlertDEPRECATED {
	return &LowLevelAlertDEPRECATED{}
}

func (l *LowLevelAlertDEPRECATED) Validate(validator structure.Validator, units *string) {
	l.LevelAlertDEPRECATED.Validate(validator, units)

	validator.Float64("level", l.Level).InRange(l.LevelRangeForUnits(units))
}

func (l *LowLevelAlertDEPRECATED) LevelRangeForUnits(units *string) (float64, float64) {
	if units != nil {
		switch *units {
		case dataBloodGlucose.MgdL, dataBloodGlucose.Mgdl:
			return LowLevelAlertDEPRECATEDLevelMgdLMinimum, LowLevelAlertDEPRECATEDLevelMgdLMaximum
		case dataBloodGlucose.MmolL, dataBloodGlucose.Mmoll:
			return LowLevelAlertDEPRECATEDLevelMmolLMinimum, LowLevelAlertDEPRECATEDLevelMmolLMaximum
		}
	}
	return -math.MaxFloat64, math.MaxFloat64
}
