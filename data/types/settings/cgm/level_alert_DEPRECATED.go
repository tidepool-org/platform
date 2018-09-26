package cgm

import (
	"math"

	"github.com/tidepool-org/platform/data"
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/structure"
)

const (
	HighLevelAlertDEPRECATEDLevelMgdLMaximum  float64 = 400
	HighLevelAlertDEPRECATEDLevelMgdLMinimum  float64 = 120
	HighLevelAlertDEPRECATEDLevelMmolLMaximum float64 = 22.20299
	HighLevelAlertDEPRECATEDLevelMmolLMinimum float64 = 6.66090
	LowLevelAlertDEPRECATEDLevelMgdLMaximum   float64 = 100
	LowLevelAlertDEPRECATEDLevelMgdLMinimum   float64 = 59
	LowLevelAlertDEPRECATEDLevelMmolLMaximum  float64 = 5.55075
	LowLevelAlertDEPRECATEDLevelMmolLMinimum  float64 = 3.27494
)

func LevelAlertDEPRECATEDSnoozes() []int {
	return []int{
		0, 900000, 1800000, 2700000, 3600000, 4500000, 5400000, 6300000,
		7200000, 8100000, 9000000, 9900000, 10800000, 11700000, 12600000,
		13500000, 14400000, 15300000, 16200000, 17100000, 18000000,
	}
}

type LevelAlertDEPRECATED struct {
	Enabled *bool    `json:"enabled,omitempty" bson:"enabled,omitempty"`
	Level   *float64 `json:"level,omitempty" bson:"level,omitempty"`
	Snooze  *int     `json:"snooze,omitempty" bson:"snooze,omitempty"`
}

func (l *LevelAlertDEPRECATED) Parse(parser data.ObjectParser) {
	l.Enabled = parser.ParseBoolean("enabled")
	l.Level = parser.ParseFloat("level")
	l.Snooze = parser.ParseInteger("snooze")
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

func ParseHighLevelAlertDEPRECATED(parser data.ObjectParser) *HighLevelAlertDEPRECATED {
	if parser.Object() == nil {
		return nil
	}
	datum := NewHighLevelAlertDEPRECATED()
	datum.Parse(parser)
	parser.ProcessNotParsed()
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

func ParseLowLevelAlertDEPRECATED(parser data.ObjectParser) *LowLevelAlertDEPRECATED {
	if parser.Object() == nil {
		return nil
	}
	datum := NewLowLevelAlertDEPRECATED()
	datum.Parse(parser)
	parser.ProcessNotParsed()
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
